package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coreeng/tpm/pkg/module"
	"gopkg.in/yaml.v3"
)

func TestBuild_SimpleModule(t *testing.T) {
	// Create a temporary output directory
	outDir := t.TempDir()

	result, err := Build("testdata/simple-module", outDir, "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify output file exists
	outFile := filepath.Join(outDir, "simple-module", "module.yaml")
	if result.OutputFile != outFile {
		t.Fatalf("OutputFile = %s, want %s", result.OutputFile, outFile)
	}
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Fatalf("Output file not created: %s", outFile)
	}

	// Load and verify the output
	var mod module.BuiltModule
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if err := yaml.Unmarshal(data, &mod); err != nil {
		t.Fatalf("Failed to unmarshal output YAML: %v", err)
	}

	// Verify basic structure
	if mod.Code != "test-module-123" {
		t.Errorf("Expected module code 'test-module-123', got '%s'", mod.Code)
	}

	if mod.Title != "Test Module" {
		t.Errorf("Expected title 'Test Module', got '%s'", mod.Title)
	}

	if len(mod.Chapters) != 1 {
		t.Fatalf("Expected 1 chapter, got %d", len(mod.Chapters))
	}

	// Verify chapter
	ch := mod.Chapters[0]
	if ch.Code != "test-chapter-456" {
		t.Errorf("Expected chapter code 'test-chapter-456', got '%s'", ch.Code)
	}

	if ch.Index != 1 {
		t.Errorf("Expected chapter index 1, got %d", ch.Index)
	}

	// Verify section
	if len(ch.Sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(ch.Sections))
	}

	sec := ch.Sections[0]
	if sec.Code != "test-section-789" {
		t.Errorf("Expected section code 'test-section-789', got '%s'", sec.Code)
	}

	if sec.Index != 1 {
		t.Errorf("Expected section index 1, got %d", sec.Index)
	}

	// Verify isDraft is preserved
	if !ch.IsDraft {
		t.Errorf("Expected chapter isDraft to be true, got false")
	}

	t.Logf("✓ Build test passed successfully")
}

func TestAssignIndices(t *testing.T) {
	mod := &module.Module{
		Chapters: []module.Chapter{
			{
				Sections: []module.Section{
					{Title: "Section 1"},
					{Title: "Section 2"},
				},
			},
			{
				Sections: []module.Section{
					{Title: "Section 3"},
				},
			},
		},
	}

	assignIndices(mod)

	// Verify chapter indices
	if mod.Chapters[0].Index != 1 {
		t.Errorf("Expected chapter 0 index to be 1, got %d", mod.Chapters[0].Index)
	}
	if mod.Chapters[1].Index != 2 {
		t.Errorf("Expected chapter 1 index to be 2, got %d", mod.Chapters[1].Index)
	}

	// Verify section indices (each chapter starts at 1)
	if mod.Chapters[0].Sections[0].Index != 1 {
		t.Errorf("Expected section index 1, got %d", mod.Chapters[0].Sections[0].Index)
	}
	if mod.Chapters[0].Sections[1].Index != 2 {
		t.Errorf("Expected section index 2, got %d", mod.Chapters[0].Sections[1].Index)
	}
	if mod.Chapters[1].Sections[0].Index != 1 {
		t.Errorf("Expected section index 1 (new chapter), got %d", mod.Chapters[1].Sections[0].Index)
	}

	t.Logf("✓ Index assignment test passed")
}

func TestBuild_IsDraftPreserved(t *testing.T) {
	// Create a temporary output directory
	outDir := t.TempDir()

	_, err := Build("testdata/simple-module", outDir, "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Load and verify the output
	var mod module.BuiltModule
	data, err := os.ReadFile(filepath.Join(outDir, "simple-module", "module.yaml"))
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if err := yaml.Unmarshal(data, &mod); err != nil {
		t.Fatalf("Failed to unmarshal output YAML: %v", err)
	}

	// Verify isDraft is preserved when set to true
	if len(mod.Chapters) == 0 {
		t.Fatalf("Expected at least 1 chapter, got 0")
	}

	ch := mod.Chapters[0]
	if !ch.IsDraft {
		t.Errorf("Expected chapter isDraft to be true, got false")
	}

	t.Logf("✓ isDraft preservation test passed")
}

func TestBuild_IsDraftZeroValue(t *testing.T) {
	// Test that isDraft defaults to false (zero value) when not set
	// This tests the struct behavior directly without needing a full module build
	ch := module.Chapter{
		Code:  "test-chapter",
		Title: "Test Chapter",
		// IsDraft is not set, so it should be false (zero value)
	}

	if ch.IsDraft {
		t.Errorf("Expected chapter isDraft to be false (zero value) when not set, got true")
	}

	// Verify that isDraft is always present in YAML output (required field)
	data, err := yaml.Marshal(ch)
	if err != nil {
		t.Fatalf("Failed to marshal chapter: %v", err)
	}

	yamlStr := string(data)
	// isDraft is now required, so it should always appear in YAML output, even when false
	if !strings.Contains(yamlStr, "isDraft") {
		t.Errorf("Expected isDraft field to be present in YAML output (required field), but it was omitted")
	}
	if !strings.Contains(yamlStr, "isDraft: false") {
		t.Errorf("Expected isDraft to be false in YAML output, got: %s", yamlStr)
	}

	t.Logf("✓ isDraft zero value test passed - isDraft is always present (required field)")
}
