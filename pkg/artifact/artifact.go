package artifact

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/coreeng/tpm/pkg/schema"
	"github.com/coreeng/tpm/pkg/validator"
)

const ModuleArtifactFile = "module.yaml"
const builtModuleSchemaFile = "module.schema.json"

// ResolveModuleArtifactPath accepts either a compiled module.yaml path or a
// directory containing module.yaml.
func ResolveModuleArtifactPath(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to inspect module artifact path %s: %w", path, err)
	}
	if info.IsDir() {
		path = filepath.Join(path, ModuleArtifactFile)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("failed to find %s in artifact directory: %w", ModuleArtifactFile, err)
		}
	}
	return path, nil
}

// ValidateModuleArtifact validates a compiled module artifact against the built
// module schema embedded in tpm. schemaDir is only for tpm development tests;
// callers should pass an empty string in normal use.
func ValidateModuleArtifact(path string, schemaDir string) (*validator.ValidationResult, error) {
	artifactPath, err := ResolveModuleArtifactPath(path)
	if err != nil {
		return nil, err
	}

	schemaValidator, err := schema.NewValidatorForKind(schema.BuiltSchemas, schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize built schema validator: %w", err)
	}

	schemaErrors, err := schemaValidator.ValidateYAMLFile(artifactPath, builtModuleSchemaFile)
	if err != nil {
		return nil, fmt.Errorf("built module schema validation error: %w", err)
	}

	result := &validator.ValidationResult{}
	for _, schemaErr := range schemaErrors {
		result.AddError(artifactPath, schemaErr.Field, schemaErr.Message)
	}
	return result, nil
}
