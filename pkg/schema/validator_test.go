package schema

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	// Empty schemaDir uses the schemas embedded in the binary.
	schemaDir := ""

	validator, err := NewValidator(schemaDir)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	if validator == nil {
		t.Fatal("Expected non-nil validator")
	}

	if validator.schemaDir == "" {
		t.Fatal("Expected non-empty schema directory")
	}
}

func TestLoadSchema(t *testing.T) {
	schemaDir := ""

	validator, err := NewValidator(schemaDir)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test loading a schema that should exist
	schema, err := validator.LoadSchema("module.schema.json")
	if err != nil {
		t.Fatalf("Failed to load module schema: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}
}

func TestLoadNonExistentSchema(t *testing.T) {
	schemaDir := ""

	validator, err := NewValidator(schemaDir)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test loading a schema that doesn't exist
	_, err = validator.LoadSchema("nonexistent.schema.json")
	if err == nil {
		t.Fatal("Expected error when loading non-existent schema")
	}
}

func TestValidateStruct(t *testing.T) {
	schemaDir := ""

	validator, err := NewValidator(schemaDir)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test with a valid module structure
	validModule := map[string]interface{}{
		"code":             "550e8400-e29b-41d4-a716-446655440000",
		"title":            "Test Module",
		"shortDescription": "A test module",
		"level":            "BEGINNER",
		"bannerImage":      "https://example.com/banner.png",
		"bannerVideo":      "https://www.youtube.com/watch?v=test",
		"tags":             []string{"test"},
	}

	errors, err := validator.ValidateStruct(validModule, "module.schema.json")
	if err != nil {
		t.Fatalf("Failed to validate struct: %v", err)
	}

	if len(errors) > 0 {
		t.Fatalf("Expected no validation errors, got: %v", errors)
	}
}

func TestValidateStructWithErrors(t *testing.T) {
	schemaDir := ""

	validator, err := NewValidator(schemaDir)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test with an invalid module structure (missing required fields)
	invalidModule := map[string]interface{}{
		"code":  "550e8400-e29b-41d4-a716-446655440000",
		"title": "Test Module",
		// Missing required fields: shortDescription, level, bannerImage, bannerVideo, tags
	}

	errors, err := validator.ValidateStruct(invalidModule, "module.schema.json")
	if err != nil {
		t.Fatalf("Failed to validate struct: %v", err)
	}

	if len(errors) == 0 {
		t.Fatal("Expected validation errors for incomplete module")
	}
}
