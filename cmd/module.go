package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coreeng/tpm/pkg/builder"
	"github.com/coreeng/tpm/pkg/codegen"
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
	cmd.AddCommand(newModuleGenerateCmd())
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
				fmt.Fprintln(cmd.OutOrStdout(), mod.Name)
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
			fmt.Fprintf(out, "%s: ERROR: %v\n", path, err)
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
			fmt.Fprintf(out, "%s: validation failed\n", resolved.Name)
			for _, issue := range result.Issues {
				if issue.Level == validator.ErrorLevel {
					fmt.Fprintf(out, "  ERROR: %s [%s]: %s\n", issue.File, issue.Field, issue.Message)
				}
			}
			continue
		}
		fmt.Fprintf(out, "%s: ok\n", resolved.Name)
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
			fmt.Fprintln(out, "duplicate codes found:")
			for _, dup := range globalResult.Duplicates {
				fmt.Fprintf(out, "  %s\n", dup.Code)
				for _, loc := range dup.Locations {
					fmt.Fprintf(out, "    %s %s: %s\n", loc.ModuleName, loc.EntityType, loc.FilePath)
				}
			}
		}
	}

	if hadErrors {
		return fmt.Errorf("module validation failed")
	}
	fmt.Fprintf(out, "validated %d module(s)\n", len(modules))
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
		fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s\n", result.ModuleName, result.OutputFile)
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
				fmt.Fprintf(cmd.OutOrStdout(), "%s: generated %d code(s) in %d file(s)\n", path, result.CodesAdded, len(result.FilesModified))
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
					fmt.Fprintf(cmd.OutOrStdout(), "%s: created %d markdown file(s)\n", path, result.FilesCreated)
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
