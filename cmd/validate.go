package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/coreeng/tpm/pkg/builder"
	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
	"github.com/coreeng/tpm/pkg/validator"
	"github.com/spf13/cobra"
)

var (
	schemaDir string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate all training platform modules",
	Long: `Validate all training platform modules in the repository.

Performs the following validations for each module:
  - JSON Schema validation (structure, required fields, types, patterns)
  - Required markdown files exist (description.md, successMessage.md)

Global validations across all modules:
  - Checks that all codes are unique across all modules
  - Reports any duplicate codes with their locations

The JSON schema validation automatically enforces:
  - All entities have required 'code' fields
  - Description/successMessage fields are NOT in YAML (must be in .md files)
  - Code format validation (UUIDs for structural entities, patterns for others)
  - Field types are correct (strings, arrays, enums, etc.)
  - All required properties are present
  - No additional properties beyond schema definition

Returns exit code 0 if validation passes, 1 if errors are found.
Warnings do not cause validation to fail.`,
	Args: cobra.NoArgs,
	Run:  runValidate,
}

func init() {
	validateCmd.Flags().StringVar(&schemaDir, "schema-dir", "", "Path to JSON schema directory (defaults to schemas embedded in the binary)")
}

func runValidate(cmd *cobra.Command, args []string) {
	// Get repository root using shared utility
	rootDir, err := pathutil.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate all modules
	runValidateAllModules(rootDir)
}

func runValidateAllModules(rootDir string) {
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

	// Load and validate all modules
	fmt.Println("Validating modules...")
	modules := make([]*module.Module, 0, len(moduleNames))
	loadErrors := make([]error, 0)
	hasValidationErrors := false

	for _, moduleName := range moduleNames {
		mod, err := module.LoadModule(rootDir, moduleName)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load %s: %w", moduleName, err))
			continue
		}

		// Validate individual module BEFORE merging descriptions
		// This ensures we check the raw YAML state before markdown files overwrite the fields
		result := validator.ValidateModule(mod, moduleName, schemaDir)

		// Merge descriptions to check for double declarations
		modulePath := module.GetModulePath(rootDir, moduleName)
		warnings, err := builder.MergeDescriptions(mod, modulePath)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to merge descriptions for %s: %w", moduleName, err))
			continue
		}
		if result.HasErrors() {
			hasValidationErrors = true
			fmt.Printf("\n[%s] Validation errors:\n", moduleName)
			for _, issue := range result.Issues {
				if issue.Level == validator.ErrorLevel {
					fmt.Printf("  ERROR: %s: %s\n", issue.File, issue.Message)
				}
			}
		} else {
			fmt.Printf("  ✓ %s\n", moduleName)
		}

		// Display warnings for double-declared descriptions
		if len(warnings) > 0 {
			fmt.Printf("\n[%s] ⚠️  Warnings:\n", moduleName)
			for _, warning := range warnings {
				fmt.Printf("  - %s\n", warning)
			}
		}

		modules = append(modules, mod)
	}

	if len(loadErrors) > 0 {
		fmt.Println("\n⚠ Some modules failed to load:")
		for _, err := range loadErrors {
			fmt.Printf("  - %v\n", err)
		}
		if len(modules) == 0 {
			fmt.Println("\nNo modules loaded successfully. Cannot proceed with validation.")
			os.Exit(1)
		}
		fmt.Println()
	}

	// Check for duplicate codes across all modules
	fmt.Println("\nChecking code uniqueness across all modules...")
	globalResult := validator.ValidateAllModules(modules, moduleNames)

	fmt.Printf("Total unique codes: %d\n", globalResult.TotalCodes)

	if globalResult.HasDuplicates() {
		hasValidationErrors = true
		fmt.Printf("\n✗ Found %d duplicate code(s):\n\n", len(globalResult.Duplicates))

		// Sort duplicates by code for consistent output
		sort.Slice(globalResult.Duplicates, func(i, j int) bool {
			return globalResult.Duplicates[i].Code < globalResult.Duplicates[j].Code
		})

		for _, dup := range globalResult.Duplicates {
			fmt.Printf("Code '%s' appears in %d location(s):\n", dup.Code, len(dup.Locations))
			for _, loc := range dup.Locations {
				fmt.Printf("  - [%s] %s: %s\n", loc.ModuleName, loc.EntityType, loc.FilePath)
			}
			fmt.Println()
		}
	}

	// Print final summary
	if hasValidationErrors {
		fmt.Println("✗ Validation failed")
		os.Exit(1)
	} else {
		fmt.Println("\n✓ All modules validated successfully!")
	}
}
