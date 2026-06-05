package cmd

import (
	"fmt"
	"os"

	"github.com/coreeng/tpm/pkg/builder"
	"github.com/spf13/cobra"
)

var (
	outDir                     string
	assessmentRegistryOverride string
	assessmentVersionOverride  string
)

var buildCmd = &cobra.Command{
	Use:   "build <module-name>",
	Short: "Build a training module into a unified module.yaml",
	Long: `Compiles module source files into a single module.yaml artifact.

This command performs full validation first (same as 'tpm validate'):
- JSON schema validation (structure, required fields, types, patterns)
- Required markdown files exist
- No duplicate codes

Only proceeds with build if validation passes. Build steps:
- Merges description.md files into YAML
- Assigns sequential indices
- Outputs unified module.yaml

If validation fails, fix the errors before building. Use 'tpm generate-codes'
and 'tpm generate-markdown' to auto-generate missing codes and markdown files.

Examples:
  tpm build k8s-for-app-devs --outdir .build/module
  tpm build path-to-production -o /tmp/output`,
	Args: cobra.ExactArgs(1),
	Run:  runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&outDir, "outdir", "o", "", "Output directory (required)")
	buildCmd.Flags().StringVar(&assessmentRegistryOverride, "assessment-registry-override", "", "Override registry path for all assessments")
	buildCmd.Flags().StringVar(&assessmentVersionOverride, "assessment-version-override", "", "Override imageVersion for all assessments")
	if err := buildCmd.MarkFlagRequired("outdir"); err != nil {
		panic(fmt.Sprintf("failed to mark outdir flag as required: %v", err))
	}
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) {
	moduleName := args[0]

	if err := builder.Build(moduleName, outDir, assessmentRegistryOverride, assessmentVersionOverride); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
