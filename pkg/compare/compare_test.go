package compare

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompareDetectsRemovedCode(t *testing.T) {
	oldPath := copyFixtureModule(t)
	newPath := copyFixtureModule(t)
	if err := os.RemoveAll(filepath.Join(newPath, "module", "01-chapter", "01-section")); err != nil {
		t.Fatal(err)
	}

	report, err := Compare(oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if !report.HasBreakingChanges() {
		t.Fatal("expected removed section to be breaking")
	}
	if len(report.Removed) != 1 || report.Removed[0].Code != "test-section-789" {
		t.Fatalf("removed = %#v, want test-section-789", report.Removed)
	}
}

func TestCompareTreatsSameParentReorderAsNonBreaking(t *testing.T) {
	oldPath := copyFixtureModule(t)
	newPath := copyFixtureModule(t)
	oldSection := filepath.Join(newPath, "module", "01-chapter", "01-section")
	newSection := filepath.Join(newPath, "module", "01-chapter", "02-section")
	if err := os.Rename(oldSection, newSection); err != nil {
		t.Fatal(err)
	}

	report, err := Compare(oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if report.HasBreakingChanges() {
		t.Fatalf("same-parent reorder should not be breaking: %#v", report)
	}
}

func TestCompareDetectsCrossParentMove(t *testing.T) {
	oldPath := copyFixtureModule(t)
	newPath := copyFixtureModule(t)
	newChapter := filepath.Join(newPath, "module", "02-next-chapter")
	if err := os.MkdirAll(newChapter, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newChapter, "chapter.yml"), []byte("code: next-chapter\ntitle: Next Chapter\nisDraft: false\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newChapter, "description.md"), []byte("Next chapter\n"), 0644); err != nil {
		t.Fatal(err)
	}
	oldSection := filepath.Join(newPath, "module", "01-chapter", "01-section")
	newSection := filepath.Join(newChapter, "01-section")
	if err := os.Rename(oldSection, newSection); err != nil {
		t.Fatal(err)
	}

	report, err := Compare(oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Moved) != 1 || report.Moved[0].Code != "test-section-789" {
		t.Fatalf("moved = %#v, want section move", report.Moved)
	}
}

func TestCompareSupportsPathAtGitRef(t *testing.T) {
	fixture := filepath.Join("..", "builder", "testdata", "simple-module")
	report, err := Compare(fixture+"@HEAD", fixture)
	if err != nil {
		t.Fatal(err)
	}
	if report.HasBreakingChanges() {
		t.Fatalf("HEAD fixture compared to working tree should not be breaking: %#v", report)
	}
}

func copyFixtureModule(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "builder", "testdata", "simple-module")
	dst := filepath.Join(t.TempDir(), "simple-module")
	copyDir(t, src, dst)
	return dst
}

func copyDir(t *testing.T, src, dst string) {
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
			copyDir(t, srcPath, dstPath)
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
