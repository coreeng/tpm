package validator

import (
	"testing"

	"github.com/coreeng/tpm/pkg/module"
)

func TestValidateModule_ValidModule(t *testing.T) {
	rootDir := "testdata"
	moduleName := "valid-module"

	mod, err := module.LoadModule(rootDir, moduleName)
	if err != nil {
		t.Fatalf("Failed to load valid module: %v", err)
	}

	result := ValidateModule(mod, moduleName, "")

	if result.HasErrors() {
		t.Errorf("Expected no errors for valid module, but got:")
		for _, issue := range result.Issues {
			if issue.Level == ErrorLevel {
				t.Errorf("  ERROR: %s: %s - %s", issue.File, issue.Field, issue.Message)
			}
		}
	}
}

func TestValidateModule_MissingModuleCode(t *testing.T) {
	rootDir := "testdata"
	moduleName := "invalid-module-missing-code"

	mod, err := module.LoadModule(rootDir, moduleName)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	result := ValidateModule(mod, moduleName, "")

	if !result.HasErrors() {
		t.Fatal("Expected errors for module with missing code, but got none")
	}

	// Check that we have an error about missing code
	found := false
	for _, issue := range result.Issues {
		if issue.Level == ErrorLevel && contains(issue.Message, "missing properties: 'code'") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected error about missing 'code' property, but didn't find it")
		t.Logf("Got issues:")
		for _, issue := range result.Issues {
			t.Logf("  %v: %s: %s - %s", issue.Level, issue.File, issue.Field, issue.Message)
		}
	}
}

func TestValidationResult_HasErrors(t *testing.T) {
	tests := []struct {
		name   string
		result *ValidationResult
		want   bool
	}{
		{
			name:   "no issues",
			result: &ValidationResult{Issues: []ValidationIssue{}},
			want:   false,
		},
		{
			name: "only warnings",
			result: &ValidationResult{
				Issues: []ValidationIssue{
					{Level: WarningLevel, File: "test.yaml", Field: "optional", Message: "warning"},
				},
			},
			want: false,
		},
		{
			name: "has errors",
			result: &ValidationResult{
				Issues: []ValidationIssue{
					{Level: ErrorLevel, File: "test.yaml", Field: "required", Message: "error"},
				},
			},
			want: true,
		},
		{
			name: "mixed errors and warnings",
			result: &ValidationResult{
				Issues: []ValidationIssue{
					{Level: WarningLevel, File: "test.yaml", Field: "optional", Message: "warning"},
					{Level: ErrorLevel, File: "test.yaml", Field: "required", Message: "error"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.HasErrors()
			if got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{}

	result.AddError("test.yaml", "field1", "error message")

	if len(result.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if issue.Level != ErrorLevel {
		t.Errorf("Expected ErrorLevel, got %v", issue.Level)
	}
	if issue.File != "test.yaml" {
		t.Errorf("Expected file 'test.yaml', got %s", issue.File)
	}
	if issue.Field != "field1" {
		t.Errorf("Expected field 'field1', got %s", issue.Field)
	}
	if issue.Message != "error message" {
		t.Errorf("Expected message 'error message', got %s", issue.Message)
	}
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := &ValidationResult{}

	result.AddWarning("test.yaml", "field1", "warning message")

	if len(result.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if issue.Level != WarningLevel {
		t.Errorf("Expected WarningLevel, got %v", issue.Level)
	}
	if issue.File != "test.yaml" {
		t.Errorf("Expected file 'test.yaml', got %s", issue.File)
	}
	if issue.Field != "field1" {
		t.Errorf("Expected field 'field1', got %s", issue.Field)
	}
	if issue.Message != "warning message" {
		t.Errorf("Expected message 'warning message', got %s", issue.Message)
	}
}

func TestValidationLevel_Values(t *testing.T) {
	// Just test that the constants are defined and different
	if ErrorLevel == WarningLevel {
		t.Error("ErrorLevel and WarningLevel should be different values")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
