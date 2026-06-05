package lab

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectStandaloneLab(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "lab.yaml"), `title: Standalone Lab
code: standalone-lab
timeLimit: 30m
challenges:
  - code: FirstChallenge
    title: First Challenge
    goals:
      - code: FirstGoal
        title: First Goal
      - code: SecondGoal
        title: Second Goal
`)
	makeRuntimeDirs(t, root, true)

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Format != FormatStandalone {
		t.Fatalf("Format = %q, want %q", loaded.Format, FormatStandalone)
	}
	if loaded.RootPath != root {
		t.Errorf("RootPath = %q, want %q", loaded.RootPath, root)
	}
	if loaded.RuntimePath != root {
		t.Errorf("RuntimePath = %q, want %q", loaded.RuntimePath, root)
	}
	if loaded.MetadataPath != filepath.Join(root, "lab.yaml") {
		t.Errorf("MetadataPath = %q, want lab.yaml path", loaded.MetadataPath)
	}
	if loaded.Title != "Standalone Lab" {
		t.Errorf("Title = %q, want Standalone Lab", loaded.Title)
	}
	if loaded.Code != "standalone-lab" {
		t.Errorf("Code = %q, want standalone-lab", loaded.Code)
	}
	if loaded.TimeLimit != "30m" {
		t.Errorf("TimeLimit = %q, want 30m", loaded.TimeLimit)
	}
	assertRuntimePaths(t, loaded, root)
	assertChallenges(t, loaded.Challenges)
}

func TestDetectModuleBackedLabFromRuntimePath(t *testing.T) {
	repo := t.TempDir()
	metadataRoot := filepath.Join(repo, "path-to-production", "module", "01-chapter", "assessments", "01-lab")
	writeFile(t, filepath.Join(metadataRoot, "assessment.yaml"), `title: Module Backed Lab
code: module-backed-lab
timeLimit: 45m
`)
	writeFile(t, filepath.Join(metadataRoot, "01-challenge", "challenge.yml"), `title: First Challenge
code: ZFirstChallenge
goals:
  - code: FirstGoal
    title: First Goal
  - code: SecondGoal
    title: Second Goal
`)
	writeFile(t, filepath.Join(metadataRoot, "02-challenge", "challenge.yaml"), `title: Second Challenge
code: ASecondChallenge
goals:
  - code: ThirdGoal
    title: Third Goal
`)

	runtimeRoot := filepath.Join(repo, "path-to-production", "assessments", "01-chapter", "01-lab")
	makeRuntimeDirs(t, runtimeRoot, true)

	loaded, err := Load(runtimeRoot)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Format != FormatModuleBacked {
		t.Fatalf("Format = %q, want %q", loaded.Format, FormatModuleBacked)
	}
	if loaded.RootPath != metadataRoot {
		t.Errorf("RootPath = %q, want %q", loaded.RootPath, metadataRoot)
	}
	if loaded.RuntimePath != runtimeRoot {
		t.Errorf("RuntimePath = %q, want %q", loaded.RuntimePath, runtimeRoot)
	}
	if loaded.MetadataPath != filepath.Join(metadataRoot, "assessment.yaml") {
		t.Errorf("MetadataPath = %q, want assessment.yaml path", loaded.MetadataPath)
	}
	if loaded.Title != "Module Backed Lab" {
		t.Errorf("Title = %q, want Module Backed Lab", loaded.Title)
	}
	if loaded.Code != "module-backed-lab" {
		t.Errorf("Code = %q, want module-backed-lab", loaded.Code)
	}
	if loaded.TimeLimit != "45m" {
		t.Errorf("TimeLimit = %q, want 45m", loaded.TimeLimit)
	}
	assertRuntimePaths(t, loaded, runtimeRoot)
	assertModuleBackedChallenges(t, loaded.Challenges)
}

func TestLoadLabRejectsMissingValidator(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "lab.yaml"), `title: Standalone Lab
code: standalone-lab
`)
	makeRuntimeDirs(t, root, false)

	_, err := Load(root)
	if err == nil {
		t.Fatal("Load returned nil error, want missing validator error")
	}
	if !strings.Contains(err.Error(), "validator") {
		t.Fatalf("error %q does not mention validator", err.Error())
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error %q does not wrap os.ErrNotExist", err.Error())
	}
}

func makeRuntimeDirs(t *testing.T, root string, includeValidator bool) {
	t.Helper()
	for _, name := range []string{"starter-content", "solution"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0755); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	if includeValidator {
		if err := os.MkdirAll(filepath.Join(root, "validator"), 0755); err != nil {
			t.Fatalf("create validator: %v", err)
		}
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create parent for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertRuntimePaths(t *testing.T, loaded *Lab, root string) {
	t.Helper()
	if loaded.StarterPath != filepath.Join(root, "starter-content") {
		t.Errorf("StarterPath = %q, want starter-content path", loaded.StarterPath)
	}
	if loaded.SolutionPath != filepath.Join(root, "solution") {
		t.Errorf("SolutionPath = %q, want solution path", loaded.SolutionPath)
	}
	if loaded.ValidatorPath != filepath.Join(root, "validator") {
		t.Errorf("ValidatorPath = %q, want validator path", loaded.ValidatorPath)
	}
}

func assertChallenges(t *testing.T, challenges []Challenge) {
	t.Helper()
	if len(challenges) != 1 {
		t.Fatalf("len(Challenges) = %d, want 1", len(challenges))
	}
	challenge := challenges[0]
	if challenge.Code != "FirstChallenge" {
		t.Errorf("Challenge.Code = %q, want FirstChallenge", challenge.Code)
	}
	if challenge.Title != "First Challenge" {
		t.Errorf("Challenge.Title = %q, want First Challenge", challenge.Title)
	}
	if len(challenge.Goals) != 2 {
		t.Fatalf("len(Goals) = %d, want 2", len(challenge.Goals))
	}
	if challenge.Goals[0].Code != "FirstGoal" || challenge.Goals[0].Title != "First Goal" {
		t.Errorf("first goal = %#v, want FirstGoal/First Goal", challenge.Goals[0])
	}
	if challenge.Goals[1].Code != "SecondGoal" || challenge.Goals[1].Title != "Second Goal" {
		t.Errorf("second goal = %#v, want SecondGoal/Second Goal", challenge.Goals[1])
	}
}

func assertModuleBackedChallenges(t *testing.T, challenges []Challenge) {
	t.Helper()
	if len(challenges) != 2 {
		t.Fatalf("len(Challenges) = %d, want 2", len(challenges))
	}
	if challenges[0].Code != "ZFirstChallenge" || challenges[0].Title != "First Challenge" {
		t.Fatalf("first challenge = %#v, want 01-challenge yml metadata", challenges[0])
	}
	if challenges[1].Code != "ASecondChallenge" || challenges[1].Title != "Second Challenge" {
		t.Fatalf("second challenge = %#v, want 02-challenge yaml metadata", challenges[1])
	}
	if len(challenges[0].Goals) != 2 {
		t.Fatalf("len(first challenge Goals) = %d, want 2", len(challenges[0].Goals))
	}
	if challenges[0].Goals[0].Code != "FirstGoal" || challenges[0].Goals[0].Title != "First Goal" {
		t.Errorf("first goal = %#v, want FirstGoal/First Goal", challenges[0].Goals[0])
	}
	if challenges[0].Goals[1].Code != "SecondGoal" || challenges[0].Goals[1].Title != "Second Goal" {
		t.Errorf("second goal = %#v, want SecondGoal/Second Goal", challenges[0].Goals[1])
	}
}
