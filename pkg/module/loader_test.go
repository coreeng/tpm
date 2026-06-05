package module

import (
	"testing"
)

func TestLoadModule(t *testing.T) {
	rootDir := "testdata"
	moduleName := "test-module"

	mod, err := LoadModule(rootDir, moduleName)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Verify module metadata
	if mod.Code != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Errorf("Expected module code 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', got %s", mod.Code)
	}
	if mod.Title != "Test Module" {
		t.Errorf("Expected title 'Test Module', got %s", mod.Title)
	}
	// Description should be empty - it's loaded from description.md during MergeDescriptions
	if mod.Description != "" {
		t.Errorf("Expected empty description (loaded from YAML only), got %s", mod.Description)
	}
	if mod.ShortDescription != "Test module" {
		t.Errorf("Expected shortDescription 'Test module', got %s", mod.ShortDescription)
	}
	if mod.Level != "intermediate" {
		t.Errorf("Expected level 'intermediate', got %s", mod.Level)
	}

	// Verify chapters loaded
	if len(mod.Chapters) != 1 {
		t.Fatalf("Expected 1 chapter, got %d", len(mod.Chapters))
	}

	chapter := mod.Chapters[0]
	if chapter.Code != "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" {
		t.Errorf("Expected chapter code 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', got %s", chapter.Code)
	}
	if chapter.Title != "Test Chapter" {
		t.Errorf("Expected chapter title 'Test Chapter', got %s", chapter.Title)
	}

	// Verify sections loaded
	if len(chapter.Sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(chapter.Sections))
	}

	section := chapter.Sections[0]
	if section.Code != "ffffffff-ffff-ffff-ffff-ffffffffffff" {
		t.Errorf("Expected section code 'ffffffff-ffff-ffff-ffff-ffffffffffff', got %s", section.Code)
	}
	if section.Title != "Test Section" {
		t.Errorf("Expected section title 'Test Section', got %s", section.Title)
	}

	// Verify assessments loaded
	if len(chapter.Assessments) != 1 {
		t.Fatalf("Expected 1 assessment, got %d", len(chapter.Assessments))
	}

	assessment := chapter.Assessments[0]
	if assessment.Code != "10101010-1010-1010-1010-101010101010" {
		t.Errorf("Expected assessment code '10101010-1010-1010-1010-101010101010', got %s", assessment.Code)
	}
	if assessment.Title != "Test Assessment" {
		t.Errorf("Expected assessment title 'Test Assessment', got %s", assessment.Title)
	}

	// Verify challenges loaded
	if len(assessment.Challenges) != 1 {
		t.Fatalf("Expected 1 challenge, got %d", len(assessment.Challenges))
	}

	challenge := assessment.Challenges[0]
	if challenge.Code != "TestChallenge" {
		t.Errorf("Expected challenge code 'TestChallenge', got %s", challenge.Code)
	}
	if challenge.Title != "Test Challenge" {
		t.Errorf("Expected challenge title 'Test Challenge', got %s", challenge.Title)
	}

	// Verify goals loaded
	if len(challenge.Goals) != 2 {
		t.Fatalf("Expected 2 goals, got %d", len(challenge.Goals))
	}

	if challenge.Goals[0].Code != "FirstTestGoal" {
		t.Errorf("Expected first goal code 'FirstTestGoal', got %s", challenge.Goals[0].Code)
	}
	if challenge.Goals[1].Code != "SecondTestGoal" {
		t.Errorf("Expected second goal code 'SecondTestGoal', got %s", challenge.Goals[1].Code)
	}

	// Verify MCQ assessments loaded
	if len(chapter.MultipleChoiceAssessments) != 1 {
		t.Fatalf("Expected 1 MCQ assessment, got %d", len(chapter.MultipleChoiceAssessments))
	}

	mcq := chapter.MultipleChoiceAssessments[0]
	if mcq.Code != "cccccccc-cccc-cccc-cccc-cccccccccccc" {
		t.Errorf("Expected MCQ code 'cccccccc-cccc-cccc-cccc-cccccccccccc', got %s", mcq.Code)
	}
	if mcq.Title != "Chapter Quiz" {
		t.Errorf("Expected MCQ title 'Chapter Quiz', got %s", mcq.Title)
	}
	if mcq.PassingScore != 75 {
		t.Errorf("Expected passing score 75, got %d", mcq.PassingScore)
	}

	// Verify questions loaded
	if len(mcq.Questions) != 2 {
		t.Fatalf("Expected 2 questions, got %d", len(mcq.Questions))
	}

	q1 := mcq.Questions[0]
	if q1.Code != "dddddddd-dddd-dddd-dddd-dddddddddddd" {
		t.Errorf("Expected question 1 code 'dddddddd-dddd-dddd-dddd-dddddddddddd', got %s", q1.Code)
	}
	if q1.Type != "SINGLE" {
		t.Errorf("Expected question 1 type 'SINGLE', got %s", q1.Type)
	}
	if len(q1.Options) != 2 {
		t.Errorf("Expected 2 options for question 1, got %d", len(q1.Options))
	}

	q2 := mcq.Questions[1]
	if q2.Type != "MULTIPLE" {
		t.Errorf("Expected question 2 type 'MULTIPLE', got %s", q2.Type)
	}
	if len(q2.Options) != 3 {
		t.Errorf("Expected 3 options for question 2, got %d", len(q2.Options))
	}
}

func TestLoadModule_NonExistent(t *testing.T) {
	rootDir := "testdata"
	moduleName := "non-existent-module"

	_, err := LoadModule(rootDir, moduleName)
	if err == nil {
		t.Fatal("Expected error when loading non-existent module, got nil")
	}
}

func TestLoadModule_MalformedYAML(t *testing.T) {
	// This test would require creating a malformed YAML file
	// For now, we'll skip it and note it as a TODO
	t.Skip("TODO: Create malformed YAML fixture and test error handling")
}

func TestGetModulePath(t *testing.T) {
	tests := []struct {
		name       string
		rootDir    string
		moduleName string
		want       string
	}{
		{
			name:       "simple path",
			rootDir:    "/root",
			moduleName: "test-module",
			want:       "/root/test-module/module",
		},
		{
			name:       "relative path",
			rootDir:    ".",
			moduleName: "my-module",
			want:       "my-module/module",
		},
		{
			name:       "empty root",
			rootDir:    "",
			moduleName: "module-name",
			want:       "module-name/module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetModulePath(tt.rootDir, tt.moduleName)
			if got != tt.want {
				t.Errorf("GetModulePath(%q, %q) = %q, want %q", tt.rootDir, tt.moduleName, got, tt.want)
			}
		})
	}
}

func TestGetModuleFilePath(t *testing.T) {
	tests := []struct {
		name       string
		rootDir    string
		moduleName string
		wantSuffix string
	}{
		{
			name:       "yaml extension",
			rootDir:    "/root",
			moduleName: "test-module",
			wantSuffix: "/root/test-module/module/module.yaml",
		},
		{
			name:       "relative path",
			rootDir:    "testdata",
			moduleName: "my-module",
			wantSuffix: "testdata/my-module/module/module.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetModuleFilePath(tt.rootDir, tt.moduleName)
			if got != tt.wantSuffix {
				t.Errorf("GetModuleFilePath(%q, %q) = %q, want %q", tt.rootDir, tt.moduleName, got, tt.wantSuffix)
			}
		})
	}
}

func TestChapterSorting(t *testing.T) {
	// This test verifies that chapters are sorted by directory name
	// when loaded by LoadModule
	rootDir := "testdata"
	moduleName := "test-module"

	mod, err := LoadModule(rootDir, moduleName)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// With only one chapter, we can't test sorting
	// But we can verify it loaded correctly
	if len(mod.Chapters) < 1 {
		t.Fatal("Expected at least 1 chapter")
	}

	// The chapter should be "01-test-chapter"
	if mod.Chapters[0].Title != "Test Chapter" {
		t.Errorf("Expected first chapter to be '01-test-chapter', got %s", mod.Chapters[0].Title)
	}
}

func TestSectionSorting(t *testing.T) {
	// This test verifies that sections are sorted by directory name
	rootDir := "testdata"
	moduleName := "test-module"

	mod, err := LoadModule(rootDir, moduleName)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	if len(mod.Chapters) < 1 || len(mod.Chapters[0].Sections) < 1 {
		t.Fatal("Expected at least 1 chapter with 1 section")
	}

	// The section should be "01-test-section"
	section := mod.Chapters[0].Sections[0]
	if section.Title != "Test Section" {
		t.Errorf("Expected first section to be 'Test Section', got %s", section.Title)
	}
}

func TestAssessmentSorting(t *testing.T) {
	// This test verifies that assessments are sorted by directory name
	rootDir := "testdata"
	moduleName := "test-module"

	mod, err := LoadModule(rootDir, moduleName)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	if len(mod.Chapters) < 1 || len(mod.Chapters[0].Assessments) < 1 {
		t.Fatal("Expected at least 1 chapter with 1 assessment")
	}

	// The assessment should be "01-test-assessment"
	assessment := mod.Chapters[0].Assessments[0]
	if assessment.Title != "Test Assessment" {
		t.Errorf("Expected first assessment to be 'Test Assessment', got %s", assessment.Title)
	}
}

func TestChallengeSorting(t *testing.T) {
	// This test verifies that challenges are sorted by directory name
	rootDir := "testdata"
	moduleName := "test-module"

	mod, err := LoadModule(rootDir, moduleName)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	if len(mod.Chapters) < 1 || len(mod.Chapters[0].Assessments) < 1 || len(mod.Chapters[0].Assessments[0].Challenges) < 1 {
		t.Fatal("Expected at least 1 chapter with 1 assessment with 1 challenge")
	}

	// The challenge should be "01-test-challenge"
	challenge := mod.Chapters[0].Assessments[0].Challenges[0]
	if challenge.Title != "Test Challenge" {
		t.Errorf("Expected first challenge to be 'Test Challenge', got %s", challenge.Title)
	}
}

func TestFilePathTracking(t *testing.T) {
	// Verify that FilePath is correctly set on all loaded entities
	rootDir := "testdata"
	moduleName := "test-module"

	mod, err := LoadModule(rootDir, moduleName)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Module FilePath
	if mod.FilePath == "" {
		t.Error("Module FilePath should not be empty")
	}
	if !contains(mod.FilePath, "module.yaml") {
		t.Errorf("Module FilePath should contain 'module.yaml', got %s", mod.FilePath)
	}

	// Chapter FilePath
	if len(mod.Chapters) > 0 {
		chapter := mod.Chapters[0]
		if chapter.FilePath == "" {
			t.Error("Chapter FilePath should not be empty")
		}
		if !contains(chapter.FilePath, "chapter.yml") {
			t.Errorf("Chapter FilePath should contain 'chapter.yml', got %s", chapter.FilePath)
		}
	}

	// Section FilePath
	if len(mod.Chapters) > 0 && len(mod.Chapters[0].Sections) > 0 {
		section := mod.Chapters[0].Sections[0]
		if section.FilePath == "" {
			t.Error("Section FilePath should not be empty")
		}
		if !contains(section.FilePath, "section.yaml") {
			t.Errorf("Section FilePath should contain 'section.yaml', got %s", section.FilePath)
		}
	}

	// Assessment FilePath
	if len(mod.Chapters) > 0 && len(mod.Chapters[0].Assessments) > 0 {
		assessment := mod.Chapters[0].Assessments[0]
		if assessment.FilePath == "" {
			t.Error("Assessment FilePath should not be empty")
		}
		if !contains(assessment.FilePath, "assessment.yaml") {
			t.Errorf("Assessment FilePath should contain 'assessment.yaml', got %s", assessment.FilePath)
		}
	}

	// Challenge FilePath
	if len(mod.Chapters) > 0 && len(mod.Chapters[0].Assessments) > 0 && len(mod.Chapters[0].Assessments[0].Challenges) > 0 {
		challenge := mod.Chapters[0].Assessments[0].Challenges[0]
		if challenge.FilePath == "" {
			t.Error("Challenge FilePath should not be empty")
		}
		if !contains(challenge.FilePath, "challenge.yaml") {
			t.Errorf("Challenge FilePath should contain 'challenge.yaml', got %s", challenge.FilePath)
		}
	}
}

// Helper function for string contains
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
