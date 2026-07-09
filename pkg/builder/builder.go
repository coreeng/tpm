package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreeng/tpm/pkg/artifact"
	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/validator"
)

type BuildResult struct {
	ModuleName string
	OutputDir  string
	OutputFile string
}

// Build compiles a module from source into a single unified module.yaml artifact
// It validates that all codes exist, merges descriptions, assigns indices, and writes output
func Build(modulePath, outRoot, registryOverride, versionOverride string) (*BuildResult, error) {
	mod, resolved, built, err := Compile(modulePath, registryOverride, versionOverride)
	if err != nil {
		return nil, err
	}

	outDir := filepath.Join(outRoot, resolved.Name)
	fmt.Printf("\nWriting output to %s...\n", outDir)
	if err := writeModule(built, outDir); err != nil {
		return nil, fmt.Errorf("failed to write module: %w", err)
	}

	outputFile := filepath.Join(outDir, "module.yaml")
	fmt.Printf("\nBuilt module: %s -> %s\n", mod.Title, outputFile)
	return &BuildResult{ModuleName: resolved.Name, OutputDir: outDir, OutputFile: outputFile}, nil
}

func Compile(modulePath, registryOverride, versionOverride string) (*module.Module, module.ResolvedPath, *module.BuiltModule, error) {
	fmt.Printf("Loading module '%s'...\n", modulePath)
	mod, resolved, err := module.LoadPath(modulePath)
	if err != nil {
		return nil, module.ResolvedPath{}, nil, fmt.Errorf("failed to load module: %w", err)
	}
	fmt.Printf("  Loaded module: %s\n", mod.Title)

	fmt.Println("\nValidating module...")
	validationResult := validator.ValidateModule(mod, resolved.Name, "")

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
		return nil, module.ResolvedPath{}, nil, fmt.Errorf("validation failed with %d error(s)\n\nHint: Fix validation errors before building", validationResult.ErrorCount())
	}
	fmt.Println("  ✓ Module validation passed")

	fmt.Println("\nMerging description.md files...")
	warnings, err := MergeDescriptions(mod, resolved.SourcePath)
	if err != nil {
		return nil, module.ResolvedPath{}, nil, fmt.Errorf("failed to merge descriptions: %w", err)
	}
	fmt.Println("  Descriptions merged")

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
			return nil, module.ResolvedPath{}, nil, fmt.Errorf("failed to apply assessment overrides: %w", err)
		}
		if registryOverride != "" {
			fmt.Printf("  Registry override applied: %s\n", registryOverride)
		}
		if versionOverride != "" {
			fmt.Printf("  Version override applied: %s\n", versionOverride)
		}
	}

	fmt.Println("\nAssigning indices...")
	assignIndices(mod)
	fmt.Println("  Indices assigned")
	built := toBuiltModule(mod)

	fmt.Println("\nValidating built module...")
	builtValidationResult := validateBuiltModule(built)
	if builtValidationResult.HasErrors() {
		fmt.Println("\n❌ Built module validation errors:")
		for _, issue := range builtValidationResult.Issues {
			if issue.Level == validator.ErrorLevel {
				fmt.Printf("  ✗ %s: %s\n", issue.Field, issue.Message)
			}
		}
		return nil, module.ResolvedPath{}, nil, fmt.Errorf("built module validation failed with %d error(s)", builtValidationResult.ErrorCount())
	}
	fmt.Println("  Built module validation passed")

	return mod, resolved, built, nil
}

// validateBuiltModule validates the built module against the built schema.
func validateBuiltModule(mod *module.BuiltModule) *validator.ValidationResult {
	result := &validator.ValidationResult{}
	outDir, err := os.MkdirTemp("", "tpm-built-module-")
	if err != nil {
		result.AddError("", "schema", fmt.Sprintf("failed to create temp output directory: %v", err))
		return result
	}
	defer func() {
		_ = os.RemoveAll(outDir)
	}()
	if err := writeModule(mod, outDir); err != nil {
		result.AddError("", "schema", fmt.Sprintf("failed to write temp built module: %v", err))
		return result
	}
	artifactResult, err := artifact.ValidateModuleArtifact(outDir, "")
	if err != nil {
		result.AddError("", "schema", err.Error())
		return result
	}
	return artifactResult
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
