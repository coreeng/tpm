package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
	"github.com/coreeng/tpm/pkg/schema"
	"github.com/coreeng/tpm/pkg/validator"
)

// Build compiles a module from source into a single unified module.yaml artifact
// It validates that all codes exist, merges descriptions, assigns indices, and writes output
func Build(moduleName, outdir, registryOverride, versionOverride string) error {
	// Get repository root using shared utility
	rootDir, err := pathutil.GetRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to determine repository root: %w", err)
	}

	// 1. Load module using existing loader
	fmt.Printf("Loading module '%s'...\n", moduleName)
	mod, err := module.LoadModule(rootDir, moduleName)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}
	fmt.Printf("  ✓ Loaded module: %s\n", mod.Title)

	// 2. Run full validation FIRST (with schema validation enabled if schemas exist)
	fmt.Println("\nValidating module...")
	schemaDir := filepath.Join(rootDir, "schemas", "source")
	// Check if schema directory exists (may not exist in test environments)
	if _, err := os.Stat(schemaDir); os.IsNotExist(err) {
		schemaDir = "" // Skip schema validation if schemas don't exist
	}
	validationResult := validator.ValidateModule(mod, moduleName, schemaDir)

	// Print ALL validation output
	if validationResult.HasErrors() || validationResult.WarningCount() > 0 {
		if validationResult.HasErrors() {
			fmt.Println("\n❌ Validation errors:")
		}
		if validationResult.WarningCount() > 0 && !validationResult.HasErrors() {
			fmt.Println("\n⚠️  Validation warnings:")
		}
		for _, issue := range validationResult.Issues {
			var levelSymbol string
			if issue.Level == validator.ErrorLevel {
				levelSymbol = "  ✗"
			} else {
				levelSymbol = "  ⚠"
			}
			fmt.Printf("%s %s [%s]: %s\n", levelSymbol, issue.File, issue.Field, issue.Message)
		}
	}

	// Fail if validation errors found
	if validationResult.HasErrors() {
		return fmt.Errorf("validation failed with %d error(s)\n\nHint: Fix validation errors before building", validationResult.ErrorCount())
	}
	fmt.Println("  ✓ Module validation passed")

	// 3. Merge descriptions from description.md files
	fmt.Println("\nMerging description.md files...")
	modulePath := module.GetModulePath(rootDir, moduleName)
	warnings, err := MergeDescriptions(mod, modulePath)
	if err != nil {
		return fmt.Errorf("failed to merge descriptions: %w", err)
	}
	fmt.Println("  ✓ Descriptions merged")

	// Display warnings for double-declared descriptions
	if len(warnings) > 0 {
		fmt.Println("\n⚠️  Warnings:")
		for _, warning := range warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	// 3.5. Apply assessment overrides if specified
	if registryOverride != "" || versionOverride != "" {
		fmt.Println("\nApplying assessment overrides...")
		if err := applyAssessmentOverrides(mod, registryOverride, versionOverride); err != nil {
			return fmt.Errorf("failed to apply assessment overrides: %w", err)
		}
		if registryOverride != "" {
			fmt.Printf("  ✓ Registry override applied: %s\n", registryOverride)
		}
		if versionOverride != "" {
			fmt.Printf("  ✓ Version override applied: %s\n", versionOverride)
		}
	}

	// 4. Assign sequential indices
	fmt.Println("\nAssigning indices...")
	assignIndices(mod)
	fmt.Println("  ✓ Indices assigned")

	// 5. Validate built module against built schema
	fmt.Println("\nValidating built module...")
	builtSchemaDir := filepath.Join(rootDir, "schemas", "built")
	if _, err := os.Stat(builtSchemaDir); err == nil {
		// Built schema exists, validate against it
		builtValidationResult := validateBuiltModule(mod, builtSchemaDir)
		if builtValidationResult.HasErrors() {
			fmt.Println("\n❌ Built module validation errors:")
			for _, issue := range builtValidationResult.Issues {
				if issue.Level == validator.ErrorLevel {
					fmt.Printf("  ✗ %s: %s\n", issue.Field, issue.Message)
				}
			}
			return fmt.Errorf("built module validation failed with %d error(s)", builtValidationResult.ErrorCount())
		}
		fmt.Println("  ✓ Built module validation passed")
	}

	// 6. Write output
	fmt.Printf("\nWriting output to %s...\n", outdir)
	if err := writeModule(mod, outdir); err != nil {
		return fmt.Errorf("failed to write module: %w", err)
	}

	fmt.Printf("\n✅ Built module: %s → %s/module.yaml\n", mod.Title, outdir)
	return nil
}

// validateBuiltModule validates the built module against the built schema
func validateBuiltModule(mod *module.Module, schemaDir string) *validator.ValidationResult {
	result := &validator.ValidationResult{}

	// Create schema validator
	schemaValidator, err := schema.NewValidator(schemaDir)
	if err != nil {
		result.AddError("", "schema", fmt.Sprintf("failed to initialize schema validator: %v", err))
		return result
	}

	// Validate against built module schema
	errors, err := schemaValidator.ValidateStruct(mod, "module.schema.json")
	if err != nil {
		result.AddError("", "schema", fmt.Sprintf("schema validation error: %v", err))
		return result
	}

	// Add any schema errors to result
	for _, schemaErr := range errors {
		result.AddError("built module", schemaErr.Field, schemaErr.Message)
	}

	return result
}

// applyAssessmentOverrides applies registry and version overrides to all assessments in the module
func applyAssessmentOverrides(mod *module.Module, registryOverride, versionOverride string) error {
	for i := range mod.Chapters {
		chapter := &mod.Chapters[i]

		for j := range chapter.Assessments {
			assessment := &chapter.Assessments[j]

			// Override registry in image URIs
			if registryOverride != "" {
				if assessment.StarterImageURI != "" {
					assessment.StarterImageURI = replaceRegistry(assessment.StarterImageURI, registryOverride)
				}
				if assessment.ValidatorImageURI != "" {
					assessment.ValidatorImageURI = replaceRegistry(assessment.ValidatorImageURI, registryOverride)
				}
			}

			// Override image version
			if versionOverride != "" {
				assessment.ImageVersion = versionOverride
			}
		}
	}
	return nil
}

// replaceRegistry replaces the registry portion of an image URI with a new registry path
// Handles both full URIs (registry/path/image) and bare image names (image)
func replaceRegistry(originalURI, newRegistry string) string {
	// Extract image name (everything after last '/')
	lastSlash := strings.LastIndex(originalURI, "/")

	var imageName string
	if lastSlash == -1 {
		// No slash found - entire string is the image name
		imageName = originalURI
	} else {
		// Extract everything after the last slash
		imageName = originalURI[lastSlash+1:]
	}

	// Construct new URI with override registry
	return fmt.Sprintf("%s/%s", newRegistry, imageName)
}
