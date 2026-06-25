package cmd

import (
	"fmt"
	"os"

	"github.com/coreeng/tpm/pkg/artifact"
	"github.com/coreeng/tpm/pkg/validator"
	"github.com/spf13/cobra"
)

type validateArtifactOptions struct {
	schemaDir string
}

var validateArtifactCmd = newValidateArtifactCmd()

func init() {
	rootCmd.AddCommand(validateArtifactCmd)
}

func newValidateArtifactCmd() *cobra.Command {
	opts := &validateArtifactOptions{}
	cmd := &cobra.Command{
		Use:   "validate-artifact <module.yaml-or-dir>",
		Short: "Validate a compiled module artifact",
		Long: `Validate a compiled module artifact against the built module schema embedded in tpm.

The argument can be either a module.yaml file or a directory containing
module.yaml, such as an extracted module bundle.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runValidateArtifact(args[0], opts)
		},
	}
	cmd.Flags().StringVar(&opts.schemaDir, "schema-dir", "", "Path to built JSON schema directory (defaults to schemas embedded in the tpm binary)")
	return cmd
}

func runValidateArtifact(path string, opts *validateArtifactOptions) {
	result, err := artifact.ValidateModuleArtifact(path, opts.schemaDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if result.HasErrors() {
		fmt.Println("✗ Module artifact validation failed")
		for _, issue := range result.Issues {
			if issue.Level == validator.ErrorLevel {
				fmt.Printf("  ERROR: %s: %s\n", issue.Field, issue.Message)
			}
		}
		os.Exit(1)
	}
	fmt.Println("✓ Module artifact validated successfully")
}
