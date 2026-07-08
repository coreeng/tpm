package cmd

import (
	"fmt"
	"os"

	"github.com/coreeng/tpm/pkg/artifact"
	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/validator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newArtifactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "artifact",
		Short:        "Work with compiled module artifacts",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newArtifactValidateCmd())
	cmd.AddCommand(newArtifactInspectCmd())
	return cmd
}

type artifactValidateOptions struct {
	schemaDir string
}

func newArtifactValidateCmd() *cobra.Command {
	opts := &artifactValidateOptions{}
	cmd := &cobra.Command{
		Use:   "validate <module.yaml-or-dir>...",
		Short: "Validate compiled module artifacts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArtifactValidate(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.schemaDir, "schema-dir", "", "Path to built JSON schema directory")
	return cmd
}

func runArtifactValidate(cmd *cobra.Command, paths []string, opts *artifactValidateOptions) error {
	hadErrors := false
	for _, path := range paths {
		result, err := artifact.ValidateModuleArtifact(path, opts.schemaDir)
		if err != nil {
			if _, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "%s: ERROR: %v\n", path, err); writeErr != nil {
				return writeErr
			}
			hadErrors = true
			continue
		}
		if result.HasErrors() {
			hadErrors = true
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: validation failed\n", path); err != nil {
				return err
			}
			for _, issue := range result.Issues {
				if issue.Level == validator.ErrorLevel {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  ERROR: %s [%s]: %s\n", issue.File, issue.Field, issue.Message); err != nil {
						return err
					}
				}
			}
			continue
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: ok\n", path); err != nil {
			return err
		}
	}
	if hadErrors {
		return fmt.Errorf("artifact validation failed")
	}
	return nil
}

func newArtifactInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <module.yaml-or-dir>...",
		Short: "Print a summary of compiled module artifacts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, path := range args {
				mod, artifactPath, err := loadBuiltModule(path)
				if err != nil {
					return err
				}
				sections, labs, quizzes := countBuiltContent(mod)
				for _, line := range []string{
					fmt.Sprintf("%s\n", artifactPath),
					fmt.Sprintf("  code: %s\n", mod.Code),
					fmt.Sprintf("  title: %s\n", mod.Title),
					fmt.Sprintf("  chapters: %d\n", len(mod.Chapters)),
					fmt.Sprintf("  sections: %d\n", sections),
					fmt.Sprintf("  labs: %d\n", labs),
					fmt.Sprintf("  quizzes: %d\n", quizzes),
				} {
					if _, err := fmt.Fprint(cmd.OutOrStdout(), line); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}

func loadBuiltModule(path string) (*module.BuiltModule, string, error) {
	artifactPath, err := artifact.ResolveModuleArtifactPath(path)
	if err != nil {
		return nil, "", err
	}
	// #nosec G304 -- artifactPath is resolved from a local CLI argument by ResolveModuleArtifactPath.
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return nil, "", fmt.Errorf("read artifact %s: %w", artifactPath, err)
	}
	var mod module.BuiltModule
	if err := yaml.Unmarshal(data, &mod); err != nil {
		return nil, "", fmt.Errorf("parse artifact %s: %w", artifactPath, err)
	}
	return &mod, artifactPath, nil
}

func countBuiltContent(mod *module.BuiltModule) (sections, labs, quizzes int) {
	for _, chapter := range mod.Chapters {
		sections += len(chapter.Sections)
		labs += len(chapter.Assessments)
		quizzes += len(chapter.MultipleChoiceAssessments)
	}
	return sections, labs, quizzes
}
