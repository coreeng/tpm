package cmd

import (
	"fmt"
	"os"

	"github.com/coreeng/tpm/pkg/codegen"
	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
	"github.com/spf13/cobra"
)

var generateCodesCmd = &cobra.Command{
	Use:   "generate-codes",
	Short: "Generate missing UUID codes for all training platform module entities",
	Long: `Generate UUID codes for any entities that are missing a code field.

This command:
  - Scans all module entities (modules, chapters, sections, assessments, challenges, goals, MCQs, questions)
  - Generates a UUID code for any entity that doesn't have one
  - ONLY adds codes where missing - NEVER replaces existing codes
  - Preserves YAML formatting and comments

Entities that get UUID codes:
  - Modules (module.yaml)
  - Chapters (chapter.yml)
  - Sections (section.yaml)
  - Interactive Assessments (assessment.yaml)
  - Challenges (challenge.yaml)
  - Goals (within challenge.yaml)
  - Multiple Choice Assessments (within chapter.yml)
  - Questions (within chapter.yml)

Note: Challenge and goal codes can be semantic identifiers (e.g., 'DeployToStaging'),
but this command will generate UUIDs for any that are missing.`,
	Args: cobra.NoArgs,
	Run:  runGenerateCodes,
}

func runGenerateCodes(cmd *cobra.Command, args []string) {
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

	// Generate codes for all modules
	totalCodesAdded := 0
	totalFilesModified := 0
	hasErrors := false

	for _, moduleName := range moduleNames {
		fmt.Printf("Processing module: %s\n", moduleName)

		result, err := codegen.GenerateMissingCodes(rootDir, moduleName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed to process module: %v\n", err)
			hasErrors = true
			continue
		}

		if len(result.Errors) > 0 {
			fmt.Printf("  ⚠️  Encountered %d error(s):\n", len(result.Errors))
			for _, e := range result.Errors {
				fmt.Printf("    - %v\n", e)
			}
			hasErrors = true
		}

		if result.CodesAdded > 0 {
			fmt.Printf("  ✓ Generated %d code(s) across %d file(s)\n", result.CodesAdded, len(result.FilesModified))
			totalCodesAdded += result.CodesAdded
			totalFilesModified += len(result.FilesModified)
		} else {
			fmt.Printf("  ✓ No missing codes found\n")
		}
	}

	// Print summary
	fmt.Println()
	if totalCodesAdded > 0 {
		fmt.Printf("✓ Generated %d code(s) across %d file(s)\n", totalCodesAdded, totalFilesModified)
	} else {
		fmt.Println("✓ All modules have codes")
	}

	if hasErrors {
		fmt.Println("\n⚠️  Some errors occurred during code generation")
		os.Exit(1)
	}
}
