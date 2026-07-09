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
	// #nosec G304 -- test reads the output path it just built.
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if err := yaml.Unmarshal(data, &mod); err != nil {
		t.Fatalf("Failed to unmarshal output YAML: %v", err)
	}

	if mod.Code != "kubernetes-101" {
		t.Errorf("Expected module code 'kubernetes-101', got '%s'", mod.Code)
	}

	if mod.Title != "Kubernetes 101" {
		t.Errorf("Expected title 'Kubernetes 101', got '%s'", mod.Title)
	}

	if len(mod.Chapters) != 3 {
		t.Fatalf("Expected 3 chapters, got %d", len(mod.Chapters))
	}

	// Verify chapter
	ch := mod.Chapters[0]
	if ch.Code != "cluster-fundamentals" {
		t.Errorf("Expected chapter code 'cluster-fundamentals', got '%s'", ch.Code)
	}

	if ch.Index != 1 {
		t.Errorf("Expected chapter index 1, got %d", ch.Index)
	}

	// Verify the fixture intentionally exercises uneven chapter sizes.
	expectedSectionCounts := []int{2, 3, 1}
	for i, want := range expectedSectionCounts {
		if got := len(mod.Chapters[i].Sections); got != want {
			t.Fatalf("Expected chapter %d to have %d section(s), got %d", i+1, want, got)
		}
	}

	sec := ch.Sections[0]
	if sec.Code != "what-is-kubernetes" {
		t.Errorf("Expected section code 'what-is-kubernetes', got '%s'", sec.Code)
	}

	if sec.Index != 1 {
		t.Errorf("Expected section index 1, got %d", sec.Index)
	}

	if ch.IsDraft {
		t.Errorf("Expected chapter isDraft to be false, got true")
	}

	quiz := mod.Chapters[2].MultipleChoiceAssessments[0]
	if len(quiz.Questions) != 3 {
		t.Fatalf("Expected operations quiz to have 3 questions, got %d", len(quiz.Questions))
	}
	if quiz.Questions[0].Type != "SINGLE" {
		t.Errorf("Expected first quiz question to be SINGLE, got %s", quiz.Questions[0].Type)
	}
	if quiz.Questions[2].Type != "MULTIPLE" {
		t.Errorf("Expected third quiz question to be MULTIPLE, got %s", quiz.Questions[2].Type)
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

func TestBuild_IsDraftFalsePreserved(t *testing.T) {
	// Create a temporary output directory
	outDir := t.TempDir()

	_, err := Build("testdata/simple-module", outDir, "", "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Load and verify the output
	var mod module.BuiltModule
	// #nosec G304 -- test reads the output path it just built.
	data, err := os.ReadFile(filepath.Join(outDir, "simple-module", "module.yaml"))
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if err := yaml.Unmarshal(data, &mod); err != nil {
		t.Fatalf("Failed to unmarshal output YAML: %v", err)
	}

	// Verify isDraft is preserved when set to false
	if len(mod.Chapters) == 0 {
		t.Fatalf("Expected at least 1 chapter, got 0")
	}

	ch := mod.Chapters[0]
	if ch.IsDraft {
		t.Errorf("Expected chapter isDraft to be false, got true")
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
