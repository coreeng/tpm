package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coreeng/tpm/pkg/authoring"
	"github.com/coreeng/tpm/pkg/builder"
	"github.com/coreeng/tpm/pkg/codegen"
	"github.com/coreeng/tpm/pkg/compare"
	moduleinit "github.com/coreeng/tpm/pkg/init"
	"github.com/coreeng/tpm/pkg/markdowngen"
	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/validator"
	"github.com/spf13/cobra"
)

func newModuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "module",
		Short:        "Work with module source directories",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newModuleListCmd())
	cmd.AddCommand(newModuleInitCmd())
	cmd.AddCommand(newModuleValidateCmd())
	cmd.AddCommand(newModuleBuildCmd())
	cmd.AddCommand(newModulePreviewCmd())
	cmd.AddCommand(newModuleGenerateCmd())
	cmd.AddCommand(newModuleCompareCmd())
	cmd.AddCommand(newModuleAddCmd())
	cmd.AddCommand(newModuleEditCmd())
	cmd.AddCommand(newModuleMoveCmd())
	cmd.AddCommand(newModuleRemoveCmd())
	return cmd
}

func newModuleInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <module-path>",
		Short: "Create a module skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := filepath.Clean(args[0])
			name := filepath.Base(target)
			dir := filepath.Dir(target)
			if err := moduleinit.ScaffoldModuleSkeleton(dir, moduleinit.ModuleScaffoldOptions{Name: name}); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Created module %s\n", target)
			return err
		},
	}
}

func newModuleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <dir>",
		Short: "List modules directly under a directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modules, err := module.FindPaths(args[0])
			if err != nil {
				return err
			}
			for _, mod := range modules {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), mod.Name); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

type moduleValidateOptions struct {
	schemaDir string
}

func newModuleValidateCmd() *cobra.Command {
	opts := &moduleValidateOptions{}
	cmd := &cobra.Command{
		Use:   "validate <module-path>...",
		Short: "Validate one or more module source directories",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModuleValidate(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.schemaDir, "schema-dir", "", "Path to source JSON schema directory")
	return cmd
}

func runModuleValidate(cmd *cobra.Command, paths []string, opts *moduleValidateOptions) error {
	out := cmd.OutOrStdout()
	modules := make([]*module.Module, 0, len(paths))
	names := make([]string, 0, len(paths))
	hadErrors := false

	for _, path := range paths {
		mod, resolved, err := module.LoadPath(path)
		if err != nil {
			if _, writeErr := fmt.Fprintf(out, "%s: ERROR: %v\n", path, err); writeErr != nil {
				return writeErr
			}
			hadErrors = true
			continue
		}
		result := validator.ValidateModule(mod, resolved.Name, opts.schemaDir)
		if !result.HasErrors() {
			if _, err := builder.MergeDescriptions(mod, resolved.SourcePath); err != nil {
				result.AddError(resolved.ModuleFilePath, "markdown", err.Error())
			}
		}
		if result.HasErrors() {
			hadErrors = true
			if _, err := fmt.Fprintf(out, "%s: validation failed\n", resolved.Name); err != nil {
				return err
			}
			for _, issue := range result.Issues {
				if issue.Level == validator.ErrorLevel {
					if _, err := fmt.Fprintf(out, "  ERROR: %s [%s]: %s\n", issue.File, issue.Field, issue.Message); err != nil {
						return err
					}
				}
			}
			continue
		}
		if _, err := fmt.Fprintf(out, "%s: ok\n", resolved.Name); err != nil {
			return err
		}
		modules = append(modules, mod)
		names = append(names, resolved.Name)
	}

	if len(modules) > 0 {
		globalResult := validator.ValidateAllModules(modules, names)
		if globalResult.HasDuplicates() {
			hadErrors = true
			sort.Slice(globalResult.Duplicates, func(i, j int) bool {
				return globalResult.Duplicates[i].Code < globalResult.Duplicates[j].Code
			})
			if _, err := fmt.Fprintln(out, "duplicate codes found:"); err != nil {
				return err
			}
			for _, dup := range globalResult.Duplicates {
				if _, err := fmt.Fprintf(out, "  %s\n", dup.Code); err != nil {
					return err
				}
				for _, loc := range dup.Locations {
					if _, err := fmt.Fprintf(out, "    %s %s: %s\n", loc.ModuleName, loc.EntityType, loc.FilePath); err != nil {
						return err
					}
				}
			}
		}
	}

	if hadErrors {
		return fmt.Errorf("module validation failed")
	}
	if _, err := fmt.Fprintf(out, "validated %d module(s)\n", len(modules)); err != nil {
		return err
	}
	return nil
}

type moduleBuildOptions struct {
	outRoot                    string
	assessmentRegistryOverride string
	assessmentVersionOverride  string
}

func newModuleBuildCmd() *cobra.Command {
	opts := &moduleBuildOptions{}
	cmd := &cobra.Command{
		Use:   "build <module-path>...",
		Short: "Build module source directories into module.yaml artifacts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModuleBuild(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.outRoot, "out-root", "", "Output root directory")
	cmd.Flags().StringVar(&opts.assessmentRegistryOverride, "assessment-registry-override", "", "Override registry path for all labs")
	cmd.Flags().StringVar(&opts.assessmentVersionOverride, "assessment-version-override", "", "Override imageVersion for all labs")
	if err := cmd.MarkFlagRequired("out-root"); err != nil {
		panic(fmt.Sprintf("failed to mark out-root flag as required: %v", err))
	}
	return cmd
}

func runModuleBuild(cmd *cobra.Command, paths []string, opts *moduleBuildOptions) error {
	seen := map[string]string{}
	for _, path := range paths {
		resolved, err := module.ResolvePath(path)
		if err != nil {
			return err
		}
		if existing, ok := seen[resolved.Name]; ok {
			return fmt.Errorf("module paths %s and %s both build to %s/%s", existing, path, opts.outRoot, resolved.Name)
		}
		seen[resolved.Name] = path
	}

	for _, path := range paths {
		result, err := builder.Build(path, opts.outRoot, opts.assessmentRegistryOverride, opts.assessmentVersionOverride)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s\n", result.ModuleName, result.OutputFile); err != nil {
			return err
		}
	}
	return nil
}

func newModuleGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "generate",
		Short:        "Generate module source helpers",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newModuleGenerateCodesCmd())
	cmd.AddCommand(newModuleGenerateMarkdownCmd())
	return cmd
}

func newModuleGenerateCodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "codes <module-path>...",
		Short: "Generate missing codes for module source entities",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var failures []string
			for _, path := range args {
				result, err := codegen.GenerateMissingCodesPath(path)
				if err != nil {
					failures = append(failures, fmt.Sprintf("%s: %v", path, err))
					continue
				}
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: generated %d code(s) in %d file(s)\n", path, result.CodesAdded, len(result.FilesModified)); err != nil {
					return err
				}
				for _, err := range result.Errors {
					failures = append(failures, fmt.Sprintf("%s: %v", path, err))
				}
			}
			if len(failures) > 0 {
				return errors.New(strings.Join(failures, "\n"))
			}
			return nil
		},
	}
}

func newModuleGenerateMarkdownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "markdown <module-path>...",
		Short: "Generate missing module markdown files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var failures []string
			for _, path := range args {
				result, err := markdowngen.GenerateMissingMarkdownPath(path)
				if err != nil {
					failures = append(failures, fmt.Sprintf("%s: %v", path, err))
				}
				if result != nil {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: created %d markdown file(s)\n", path, result.FilesCreated); err != nil {
						return err
					}
					for _, err := range result.Errors {
						failures = append(failures, fmt.Sprintf("%s: %v", path, err))
					}
				}
			}
			if len(failures) > 0 {
				return errors.New(strings.Join(failures, "\n"))
			}
			return nil
		},
	}
}

type moduleCompareOptions struct {
	breakingPolicy string
	allowBreaking  bool
}

func newModuleCompareCmd() *cobra.Command {
	opts := &moduleCompareOptions{breakingPolicy: string(compare.BreakingPolicyError)}
	cmd := &cobra.Command{
		Use:   "compare <old-location> <new-location>",
		Short: "Compare module code compatibility between paths or git refs",
		Long: `Compare module code compatibility between two locations.

Locations can be local paths or path@ref git locations. Local paths can point to
module source directories, built artifact directories, or built module.yaml files.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.allowBreaking {
				opts.breakingPolicy = string(compare.BreakingPolicyWarn)
			}
			policy, err := compare.ValidateBreakingPolicy(opts.breakingPolicy)
			if err != nil {
				return err
			}
			report, err := compare.Compare(args[0], args[1])
			if err != nil {
				return err
			}
			if err := writeCompareReport(cmd, report, policy); err != nil {
				return err
			}
			if report.HasBreakingChanges() && policy == compare.BreakingPolicyError {
				return fmt.Errorf("breaking module changes detected")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.breakingPolicy, "breaking-policy", string(compare.BreakingPolicyError), "Breaking change policy: error, warn, or ignore")
	cmd.Flags().BoolVar(&opts.allowBreaking, "allow-breaking", false, "Alias for --breaking-policy=warn")
	return cmd
}

func writeCompareReport(cmd *cobra.Command, report *compare.Report, policy compare.BreakingPolicy) error {
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "old: %s (%d code(s))\n", report.OldLocation, report.OldCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "new: %s (%d code(s))\n", report.NewLocation, report.NewCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "added: %d\n", len(report.Added)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "removed: %d\n", len(report.Removed)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "moved between parents: %d\n", len(report.Moved)); err != nil {
		return err
	}
	if report.HasBreakingChanges() {
		switch policy {
		case compare.BreakingPolicyWarn:
			if _, err := fmt.Fprintln(out, "WARNING: breaking module changes detected"); err != nil {
				return err
			}
		case compare.BreakingPolicyIgnore:
			if _, err := fmt.Fprintln(out, "breaking module changes detected; policy is ignore"); err != nil {
				return err
			}
		default:
			if _, err := fmt.Fprintln(out, "ERROR: breaking module changes detected"); err != nil {
				return err
			}
		}
	}
	for _, info := range report.Removed {
		if _, err := fmt.Fprintf(out, "  removed %s (%s): %s\n", info.Code, info.EntityType, info.FilePath); err != nil {
			return err
		}
	}
	for _, moved := range report.Moved {
		if _, err := fmt.Fprintf(out, "  moved %s (%s): %s/%s -> %s/%s\n", moved.Code, moved.EntityType, moved.Old.ParentType, moved.Old.ParentCode, moved.New.ParentType, moved.New.ParentCode); err != nil {
			return err
		}
	}
	for _, info := range report.Added {
		if _, err := fmt.Fprintf(out, "  added %s (%s): %s\n", info.Code, info.EntityType, info.FilePath); err != nil {
			return err
		}
	}
	return nil
}

type moduleAuthoringOptions struct {
	authoring.Options
	allowBreaking bool
}

func newModuleAddCmd() *cobra.Command {
	opts := &moduleAuthoringOptions{}
	cmd := &cobra.Command{
		Use:   "add <type> <module-path>",
		Short: "Add a YAML-backed module resource at an explicit index",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := authoring.Add(args[1], args[0], opts.Options); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "added %s\n", args[0])
			return err
		},
	}
	addAuthoringIndexFlags(cmd, opts)
	cmd.Flags().IntVar(&opts.At, "at", 0, "1-based insertion index")
	cmd.Flags().StringArrayVar(&opts.Sets, "set", nil, "YAML field assignment as field=value")
	if err := cmd.MarkFlagRequired("at"); err != nil {
		panic(fmt.Sprintf("failed to mark at flag as required: %v", err))
	}
	return cmd
}

func newModuleEditCmd() *cobra.Command {
	opts := &moduleAuthoringOptions{Options: authoring.Options{BreakingPolicy: authoring.BreakingPolicyError}}
	cmd := &cobra.Command{
		Use:   "edit <type> <module-path>",
		Short: "Edit YAML fields on a module resource",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.allowBreaking {
				opts.BreakingPolicy = authoring.BreakingPolicyWarn
			}
			if err := validateAuthoringPolicy(opts.BreakingPolicy); err != nil {
				return err
			}
			if err := authoring.Edit(args[1], args[0], opts.Options); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "edited %s\n", args[0])
			return err
		},
	}
	addAuthoringIndexFlags(cmd, opts)
	addAuthoringBreakingFlags(cmd, opts)
	cmd.Flags().StringArrayVar(&opts.Sets, "set", nil, "YAML field assignment as field=value")
	return cmd
}

func newModuleMoveCmd() *cobra.Command {
	opts := &moduleAuthoringOptions{}
	cmd := &cobra.Command{
		Use:   "move <type> <module-path>",
		Short: "Move a module resource within its current parent",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := authoring.Move(args[1], args[0], opts.Options); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "moved %s\n", args[0])
			return err
		},
	}
	addAuthoringIndexFlags(cmd, opts)
	cmd.Flags().IntVar(&opts.From, "from", 0, "1-based source index")
	cmd.Flags().IntVar(&opts.To, "to", 0, "1-based destination index")
	if err := cmd.MarkFlagRequired("from"); err != nil {
		panic(fmt.Sprintf("failed to mark from flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("to"); err != nil {
		panic(fmt.Sprintf("failed to mark to flag as required: %v", err))
	}
	return cmd
}

func newModuleRemoveCmd() *cobra.Command {
	opts := &moduleAuthoringOptions{Options: authoring.Options{BreakingPolicy: authoring.BreakingPolicyError}}
	cmd := &cobra.Command{
		Use:   "remove <type> <module-path>",
		Short: "Remove a module resource by explicit index",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.allowBreaking {
				opts.BreakingPolicy = authoring.BreakingPolicyWarn
			}
			if err := validateAuthoringPolicy(opts.BreakingPolicy); err != nil {
				return err
			}
			if err := authoring.Remove(args[1], args[0], opts.Options); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", args[0])
			return err
		},
	}
	addAuthoringIndexFlags(cmd, opts)
	addAuthoringBreakingFlags(cmd, opts)
	cmd.Flags().IntVar(&opts.From, "from", 0, "1-based resource index to remove")
	cmd.Flags().BoolVar(&opts.Yes, "yes", false, "Confirm removal")
	if err := cmd.MarkFlagRequired("from"); err != nil {
		panic(fmt.Sprintf("failed to mark from flag as required: %v", err))
	}
	return cmd
}

func addAuthoringIndexFlags(cmd *cobra.Command, opts *moduleAuthoringOptions) {
	cmd.Flags().IntVar(&opts.Chapter, "chapter", 0, "1-based chapter index")
	cmd.Flags().IntVar(&opts.Section, "section", 0, "1-based section index")
	cmd.Flags().IntVar(&opts.Lab, "lab", 0, "1-based lab index")
	cmd.Flags().IntVar(&opts.Challenge, "challenge", 0, "1-based challenge index")
	cmd.Flags().IntVar(&opts.Goal, "goal", 0, "1-based goal index")
	cmd.Flags().IntVar(&opts.Quiz, "quiz", 0, "1-based quiz index")
	cmd.Flags().IntVar(&opts.Question, "question", 0, "1-based question index")
	cmd.Flags().IntVar(&opts.Option, "option", 0, "1-based option index")
}

func addAuthoringBreakingFlags(cmd *cobra.Command, opts *moduleAuthoringOptions) {
	cmd.Flags().Var((*authoringPolicyValue)(&opts.BreakingPolicy), "breaking-policy", "Breaking change policy: error, warn, or ignore")
	cmd.Flags().BoolVar(&opts.allowBreaking, "allow-breaking", false, "Alias for --breaking-policy=warn")
}

type authoringPolicyValue authoring.BreakingPolicy

func (v *authoringPolicyValue) String() string {
	if v == nil || *v == "" {
		return string(authoring.BreakingPolicyError)
	}
	return string(*v)
}

func (v *authoringPolicyValue) Set(value string) error {
	policy := authoring.BreakingPolicy(value)
	if err := validateAuthoringPolicy(policy); err != nil {
		return err
	}
	*v = authoringPolicyValue(policy)
	return nil
}

func (v *authoringPolicyValue) Type() string {
	return "policy"
}

func validateAuthoringPolicy(policy authoring.BreakingPolicy) error {
	switch policy {
	case authoring.BreakingPolicyError, authoring.BreakingPolicyWarn, authoring.BreakingPolicyIgnore:
		return nil
	default:
		return fmt.Errorf("breaking-policy must be one of: error, warn, ignore")
	}
}
