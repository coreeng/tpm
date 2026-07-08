package schema

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"gopkg.in/yaml.v3"
)

// SchemaKind selects one of the schema sets owned and embedded by tpm.
type SchemaKind string

const (
	SourceSchemas SchemaKind = "source"
	BuiltSchemas  SchemaKind = "built"
)

// embeddedSchemas holds the canonical schemas baked into the tpm binary so the
// CLI is self-contained and works without schemas living alongside it on disk
// (e.g. when installed via Homebrew). Callers can still override them by
// passing an explicit schema directory.
//
//go:embed schemas/source/*.json schemas/built/*.json
var embeddedSchemas embed.FS

var (
	embeddedDirsMu sync.Mutex
	embeddedDirs   = map[SchemaKind]string{}
)

func schemaSubdir(kind SchemaKind) (string, error) {
	switch kind {
	case SourceSchemas, BuiltSchemas:
		return "schemas/" + string(kind), nil
	default:
		return "", fmt.Errorf("unknown schema kind %q", kind)
	}
}

// EmbeddedSchemaDir materializes the embedded source schemas into a temporary
// directory. It is kept for callers that predate explicit schema kinds.
func EmbeddedSchemaDir() (string, error) {
	return EmbeddedSchemaDirForKind(SourceSchemas)
}

// EmbeddedSchemaDirForKind materializes the embedded schemas into a temporary
// directory (once per schema kind per process) and returns its path. This lets
// the rest of the validator reuse its on-disk loading logic unchanged.
func EmbeddedSchemaDirForKind(kind SchemaKind) (string, error) {
	embeddedDirsMu.Lock()
	defer embeddedDirsMu.Unlock()

	if dir := embeddedDirs[kind]; dir != "" {
		return dir, nil
	}

	subdir, err := schemaSubdir(kind)
	if err != nil {
		return "", err
	}

	dir, err := os.MkdirTemp("", "tpm-"+string(kind)+"-schemas-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp schema dir: %w", err)
	}

	if err := copyEmbeddedSchemas(subdir, dir); err != nil {
		return "", err
	}

	embeddedDirs[kind] = dir
	return dir, nil
}

func copyEmbeddedSchemas(subdir, targetDir string) error {
	entries, err := fs.ReadDir(embeddedSchemas, subdir)
	if err != nil {
		return fmt.Errorf("failed to read embedded schemas: %w", err)
	}
	for _, entry := range entries {
		data, err := embeddedSchemas.ReadFile(subdir + "/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read embedded schema %s: %w", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(targetDir, entry.Name()), data, 0o600); err != nil {
			return fmt.Errorf("failed to write embedded schema %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// Validator handles JSON schema validation for module files
type Validator struct {
	schemaDir string
	compiler  *jsonschema.Compiler
	schemas   map[string]*jsonschema.Schema
}

// ValidationError represents a schema validation error
type ValidationError struct {
	Path    string // JSONPath to the error location
	Field   string // Field name
	Message string // Error message
}

// NewValidator creates a new source schema validator. When schemaDir is empty
// the source schemas embedded in the tpm binary are used, so the CLI works
// standalone.
func NewValidator(schemaDir string) (*Validator, error) {
	return NewValidatorForKind(SourceSchemas, schemaDir)
}

// NewValidatorForKind creates a validator for the requested schema kind. When
// schemaDir is empty, the matching schemas embedded in the tpm binary are used.
func NewValidatorForKind(kind SchemaKind, schemaDir string) (*Validator, error) {
	if schemaDir == "" {
		dir, err := EmbeddedSchemaDirForKind(kind)
		if err != nil {
			return nil, err
		}
		schemaDir = dir
	}

	absPath, err := filepath.Abs(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("invalid schema directory: %w", err)
	}

	// Check if schema directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("schema directory does not exist: %s", absPath)
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	if err := registerSchemaResources(compiler, kind, absPath); err != nil {
		return nil, err
	}

	return &Validator{
		schemaDir: absPath,
		compiler:  compiler,
		schemas:   make(map[string]*jsonschema.Schema),
	}, nil
}

func registerSchemaResources(compiler *jsonschema.Compiler, kind SchemaKind, schemaDir string) error {
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return fmt.Errorf("failed to read schema directory %s: %w", schemaDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		localPath := filepath.Join(schemaDir, entry.Name())
		// #nosec G304 -- localPath is constructed from a schema directory entry found by os.ReadDir.
		data, err := os.ReadFile(localPath)
		if err != nil {
			return fmt.Errorf("failed to read schema %s: %w", entry.Name(), err)
		}

		for _, url := range schemaResourceURLs(kind, entry.Name()) {
			reader := io.NopCloser(bytes.NewReader(data))
			if err := compiler.AddResource(url, reader); err != nil {
				return fmt.Errorf("failed to register schema %s as %s: %w", entry.Name(), url, err)
			}
		}
	}
	return nil
}

func schemaResourceURLs(kind SchemaKind, filename string) []string {
	kindPath := string(kind)
	return []string{
		fmt.Sprintf("https://raw.githubusercontent.com/coreeng/tpm/main/pkg/schema/schemas/%s/%s", kindPath, filename),
		fmt.Sprintf("https://raw.githubusercontent.com/coreeng/training-platform-modules/main/p2p_assessments/schemas/%s/%s", kindPath, filename),
	}
}

// LoadSchema loads a JSON schema from the schema directory
func (v *Validator) LoadSchema(schemaName string) (*jsonschema.Schema, error) {
	// Check cache first
	if schema, ok := v.schemas[schemaName]; ok {
		return schema, nil
	}

	schemaPath := filepath.Join(v.schemaDir, schemaName)

	// Check if file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("schema file not found: %s", schemaPath)
	}

	// Load schema using file:// URL
	schemaURL := "file://" + schemaPath
	schema, err := v.compiler.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema %s: %w", schemaName, err)
	}

	// Cache the schema
	v.schemas[schemaName] = schema
	return schema, nil
}

// ValidateYAMLFile validates a YAML file against a JSON schema
func (v *Validator) ValidateYAMLFile(filePath, schemaName string) ([]ValidationError, error) {
	// Load the schema
	schema, err := v.LoadSchema(schemaName)
	if err != nil {
		return nil, err
	}

	// Read YAML file
	// #nosec G304 -- validation intentionally reads the local YAML file selected by the CLI user.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse YAML to interface{}
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
	}

	// Validate
	if err := schema.Validate(yamlData); err != nil {
		return convertSchemaErrors(err), nil
	}

	return nil, nil
}

// ValidateData validates arbitrary data against a schema
func (v *Validator) ValidateData(data interface{}, schemaName string) ([]ValidationError, error) {
	schema, err := v.LoadSchema(schemaName)
	if err != nil {
		return nil, err
	}

	if err := schema.Validate(data); err != nil {
		return convertSchemaErrors(err), nil
	}

	return nil, nil
}

// ValidateStruct validates a Go struct by converting it to JSON first
func (v *Validator) ValidateStruct(data interface{}, schemaName string) ([]ValidationError, error) {
	// Convert struct to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// Parse JSON back to interface{}
	var jsonData interface{}
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	// Validate
	return v.ValidateData(jsonData, schemaName)
}

// convertSchemaErrors converts jsonschema validation errors to our ValidationError format
func convertSchemaErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErr, ok := err.(*jsonschema.ValidationError); ok {
		errors = append(errors, extractErrors(validationErr)...)
	} else {
		// If it's not a validation error, treat as a generic error
		errors = append(errors, ValidationError{
			Path:    "",
			Field:   "",
			Message: err.Error(),
		})
	}

	return errors
}

// extractErrors recursively extracts all validation errors
func extractErrors(err *jsonschema.ValidationError) []ValidationError {
	var errors []ValidationError

	// Add the current error
	errors = append(errors, ValidationError{
		Path:    err.InstanceLocation,
		Field:   getFieldName(err.InstanceLocation),
		Message: err.Message,
	})

	// Recursively add causes
	for _, cause := range err.Causes {
		errors = append(errors, extractErrors(cause)...)
	}

	return errors
}

// getFieldName extracts the field name from a JSON path
func getFieldName(path string) string {
	if path == "" {
		return ""
	}

	// Remove leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Get the last segment (field name)
	parts := splitPath(path)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return path
}

// splitPath splits a JSON path into segments
func splitPath(path string) []string {
	var parts []string
	current := ""

	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
