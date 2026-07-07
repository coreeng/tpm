package authoring

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditRejectsMarkdownBackedDescription(t *testing.T) {
	modulePath := copyAuthoringFixture(t)
	err := Edit(modulePath, "section", Options{
		Chapter: 1,
		Section: 1,
		Sets:    []string{"description=not allowed"},
	})
	if err == nil || !strings.Contains(err.Error(), "not editable YAML") {
		t.Fatalf("Edit description error = %v, want not editable YAML", err)
	}
}

func TestEditCodeRequiresBreakingPolicy(t *testing.T) {
	modulePath := copyAuthoringFixture(t)
	err := Edit(modulePath, "section", Options{
		Chapter:        1,
		Section:        1,
		BreakingPolicy: BreakingPolicyError,
		Sets:           []string{"code=changed-section"},
	})
	if err == nil || !strings.Contains(err.Error(), "breaking change") {
		t.Fatalf("Edit code error = %v, want breaking change", err)
	}

	err = Edit(modulePath, "section", Options{
		Chapter:        1,
		Section:        1,
		BreakingPolicy: BreakingPolicyWarn,
		Sets:           []string{"code=changed-section"},
	})
	if err != nil {
		t.Fatalf("Edit code with warn policy returned error: %v", err)
	}
}

func TestAddMoveRemoveSectionByExplicitIndex(t *testing.T) {
	modulePath := copyAuthoringFixture(t)
	err := Add(modulePath, "section", Options{
		Chapter: 1,
		At:      1,
		Sets:    []string{"code=new-section", "title=New Section"},
	})
	if err != nil {
		t.Fatalf("Add section returned error: %v", err)
	}
	assertAuthoringFileExists(t, filepath.Join(modulePath, "module", "01-chapter", "01-new-section", "section.yaml"))
	assertAuthoringFileExists(t, filepath.Join(modulePath, "module", "01-chapter", "02-section", "section.yaml"))

	err = Move(modulePath, "section", Options{Chapter: 1, From: 2, To: 1})
	if err != nil {
		t.Fatalf("Move section returned error: %v", err)
	}
	assertAuthoringFileExists(t, filepath.Join(modulePath, "module", "01-chapter", "01-section", "section.yaml"))
	assertAuthoringFileExists(t, filepath.Join(modulePath, "module", "01-chapter", "02-new-section", "section.yaml"))

	err = Remove(modulePath, "section", Options{Chapter: 1, From: 2, Yes: true, BreakingPolicy: BreakingPolicyError})
	if err == nil || !strings.Contains(err.Error(), "breaking change") {
		t.Fatalf("Remove section without policy error = %v, want breaking change", err)
	}
	err = Remove(modulePath, "section", Options{Chapter: 1, From: 2, Yes: true, BreakingPolicy: BreakingPolicyWarn})
	if err != nil {
		t.Fatalf("Remove section with warn policy returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(modulePath, "module", "01-chapter", "02-new-section")); !os.IsNotExist(err) {
		t.Fatalf("removed section still exists or stat failed unexpectedly: %v", err)
	}
}

func copyAuthoringFixture(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "builder", "testdata", "simple-module")
	dst := filepath.Join(t.TempDir(), "simple-module")
	copyAuthoringDir(t, src, dst)
	return dst
}

func copyAuthoringDir(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyAuthoringDir(t, srcPath, dstPath)
			continue
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func assertAuthoringFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}
