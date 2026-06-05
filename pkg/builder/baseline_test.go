package builder

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/coreeng/tpm/pkg/module"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

// Flag to update baseline files
var updateBaseline = flag.Bool("update-baseline", false, "Update baseline files instead of comparing")

func TestBuildModule_Baseline(t *testing.T) {
	// Create a temporary output directory
	outDir := t.TempDir()

	// Build the module - use path relative to repo root
	err := Build("pkg/builder/testdata/simple-module", outDir, "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Read the built module
	outFile := filepath.Join(outDir, "module.yaml")
	actualBytes, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Baseline path (relative to current test directory)
	baselinePath := filepath.Join("testdata", "simple-module", "baseline-module.yaml")

	// If update flag is set, write the baseline and exit
	if *updateBaseline {
		if err := os.WriteFile(baselinePath, actualBytes, 0644); err != nil {
			t.Fatalf("Failed to write baseline file: %v", err)
		}
		t.Logf("✓ Baseline file updated: %s", baselinePath)
		t.Logf("  Baseline size: %d bytes", len(actualBytes))
		return
	}

	// Read the baseline file
	baselineBytes, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("Failed to read baseline file: %v\n"+
			"To generate baseline, run: go test -run TestBuildModule_Baseline -update-baseline", err)
	}

	// Parse both YAML files for structured comparison
	var actual, baseline map[string]interface{}
	if err := yaml.Unmarshal(actualBytes, &actual); err != nil {
		t.Fatalf("Failed to unmarshal actual YAML: %v", err)
	}
	if err := yaml.Unmarshal(baselineBytes, &baseline); err != nil {
		t.Fatalf("Failed to unmarshal baseline YAML: %v", err)
	}

	// Compare the structures
	if diff := cmp.Diff(baseline, actual); diff != "" {
		t.Errorf("Build output differs from baseline (-baseline +actual):\n%s\n"+
			"To update baseline, run: go test -run TestBuildModule_Baseline -update-baseline", diff)

		// Also write the actual output for debugging
		debugPath := filepath.Join("simple-module", "actual-output.yaml")
		if err := os.WriteFile(debugPath, actualBytes, 0644); err == nil {
			t.Logf("Actual output written to: %s", debugPath)
		}
	}
}

func TestBuildModule_BaselineStructure(t *testing.T) {
	// Create a temporary output directory
	outDir := t.TempDir()

	// Build the module - use path relative to repo root
	err := Build("pkg/builder/testdata/simple-module", outDir, "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Read the built module
	outFile := filepath.Join(outDir, "module.yaml")
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Parse the module
	var mod module.Module
	if err := yaml.Unmarshal(data, &mod); err != nil {
		t.Fatalf("Failed to unmarshal output YAML: %v", err)
	}

	// Verify basic structure
	if mod.Code != "test-module-123" {
		t.Errorf("Expected module code 'test-module-123', got '%s'", mod.Code)
	}

	if mod.Title != "Test Module" {
		t.Errorf("Expected module title 'Test Module', got '%s'", mod.Title)
	}

	// Verify description was merged from markdown
	if mod.Description == "" {
		t.Error("Expected description to be merged from description.md")
	}
	// Just verify it's not empty - exact content may vary
	if len(mod.Description) == 0 {
		t.Error("Description should not be empty")
	}

	// Verify chapters
	if len(mod.Chapters) != 1 {
		t.Fatalf("Expected 1 chapter, got %d", len(mod.Chapters))
	}

	chapter := mod.Chapters[0]
	if chapter.Code != "test-chapter-456" {
		t.Errorf("Expected chapter code 'test-chapter-456', got '%s'", chapter.Code)
	}

	// Verify chapter description was merged from markdown
	if chapter.Description == "" {
		t.Error("Expected chapter description to be merged from description.md")
	}

	// Verify sections
	if len(chapter.Sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(chapter.Sections))
	}

	section := chapter.Sections[0]
	if section.Code != "test-section-789" {
		t.Errorf("Expected section code 'test-section-789', got '%s'", section.Code)
	}

	// Verify section description was merged from markdown
	if section.Description == "" {
		t.Error("Expected section description to be merged from description.md")
	}

	// Verify indices are assigned
	if chapter.Index != 1 {
		t.Errorf("Expected chapter index 1, got %d", chapter.Index)
	}
	if section.Index != 1 {
		t.Errorf("Expected section index 1, got %d", section.Index)
	}
}
