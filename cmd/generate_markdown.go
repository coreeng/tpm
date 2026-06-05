package cmd

import (
	"fmt"
	"os"

	"github.com/coreeng/tpm/pkg/markdowngen"
	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
	"github.com/spf13/cobra"
)

var generateMarkdownCmd = &cobra.Command{
	Use:   "generate-markdown",
	Short: "Generate missing markdown files for training platform modules",
	Long: `Generate missing markdown files (description.md, successMessage.md) for all module entities.

This command creates placeholder markdown files where they are missing:
  - If markdown file exists → OK (nothing to do)
  - If markdown file doesn't exist → CREATE placeholder markdown file

Required markdown files:
  - description.md: Required for modules, chapters, sections, assessments, and challenges
  - successMessage.md: Required for challenges only

Note: Schema validation ensures description/successMessage fields are NOT in YAML files,
so this command only creates missing placeholder files (no content migration from YAML).`,
	Args: cobra.NoArgs,
	Run:  runGenerateMarkdown,
}

func runGenerateMarkdown(cmd *cobra.Command, args []string) {
	// Get repository root using shared utility
	rootDir, err := pathutil.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Find all modules
	fmt.Println("Scanning repository for modules...")
	moduleNames, err := module.FindModules(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding modules: %v\n", err)
		os.Exit(1)
	}

	if len(moduleNames) == 0 {
		fmt.Println("No modules found in repository")
		return
	}

	fmt.Printf("Found %d module(s): %v\n\n", len(moduleNames), moduleNames)

	// Generate markdown for all modules
	totalCreated := 0
	hasErrors := false

	for _, moduleName := range moduleNames {
		fmt.Printf("Processing module: %s\n", moduleName)

		result, err := markdowngen.GenerateMissingMarkdown(rootDir, moduleName)
		if err != nil || len(result.Errors) > 0 {
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %v\n", err)
			}
			if len(result.Errors) > 0 {
				fmt.Printf("  ⚠️  Encountered %d error(s):\n", len(result.Errors))
				for _, e := range result.Errors {
					fmt.Printf("    - %v\n", e)
				}
			}
			hasErrors = true
		}

		if result.FilesCreated > 0 {
			fmt.Printf("  ✓ Created %d placeholder file(s)\n", result.FilesCreated)
			totalCreated += result.FilesCreated
		} else {
			fmt.Printf("  ✓ All markdown files exist\n")
		}
	}

	// Print summary
	fmt.Println()
	if totalCreated > 0 {
		fmt.Printf("✓ Created %d placeholder markdown file(s)\n", totalCreated)
	} else {
		fmt.Println("✓ All markdown files exist")
	}

	if hasErrors {
		fmt.Println("\n⚠️  Some errors occurred during markdown generation")
		os.Exit(1)
	}
}
