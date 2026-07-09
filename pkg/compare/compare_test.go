package compare

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareDetectsRemovedCode(t *testing.T) {
	oldPath := copyFixtureModule(t)
	newPath := copyFixtureModule(t)
	if err := os.RemoveAll(filepath.Join(newPath, "module", "01-cluster-fundamentals", "01-what-is-kubernetes")); err != nil {
		t.Fatal(err)
	}

	report, err := Compare(oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if !report.HasBreakingChanges() {
		t.Fatal("expected removed section to be breaking")
	}
	if len(report.Removed) != 1 || report.Removed[0].Code != "what-is-kubernetes" {
		t.Fatalf("removed = %#v, want what-is-kubernetes", report.Removed)
	}
}

func TestCompareTreatsSameParentReorderAsNonBreaking(t *testing.T) {
	oldPath := copyFixtureModule(t)
	newPath := copyFixtureModule(t)
	oldSection := filepath.Join(newPath, "module", "01-cluster-fundamentals", "01-what-is-kubernetes")
	newSection := filepath.Join(newPath, "module", "01-cluster-fundamentals", "02-what-is-kubernetes")
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
	newChapter := filepath.Join(newPath, "module", "04-next-chapter")
	if err := os.MkdirAll(newChapter, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newChapter, "chapter.yml"), []byte("code: next-chapter\ntitle: Next Chapter\nisDraft: false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newChapter, "description.md"), []byte("Next chapter\n"), 0600); err != nil {
		t.Fatal(err)
	}
	oldSection := filepath.Join(newPath, "module", "01-cluster-fundamentals", "01-what-is-kubernetes")
	newSection := filepath.Join(newChapter, "01-what-is-kubernetes")
	if err := os.Rename(oldSection, newSection); err != nil {
		t.Fatal(err)
	}

	report, err := Compare(oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Moved) != 1 || report.Moved[0].Code != "what-is-kubernetes" {
		t.Fatalf("moved = %#v, want section move", report.Moved)
	}
}

func TestCompareReportsDuplicateCodesClearly(t *testing.T) {
	oldPath := copyFixtureModule(t)
	newPath := copyFixtureModule(t)
	duplicateSection := filepath.Join(newPath, "module", "01-cluster-fundamentals", "01-what-is-kubernetes", "section.yaml")
	if err := os.WriteFile(duplicateSection, []byte("code: cluster-fundamentals\ntitle: Duplicate Code\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Compare(oldPath, newPath)
	if err == nil {
		t.Fatal("Compare returned nil error for duplicate codes")
	}
	for _, want := range []string{"duplicate code", "cluster-fundamentals", "chapter", "section"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not contain %q", err.Error(), want)
		}
	}
}

func TestCompareSupportsPathAtGitRef(t *testing.T) {
	fixture := filepath.Join("..", "builder", "testdata", "simple-module")
	report, err := Compare(fixture+"@HEAD", fixture+"@HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if report.HasBreakingChanges() {
		t.Fatalf("HEAD fixture compared to itself should not be breaking: %#v", report)
	}
}

func TestUntarRejectsArchiveEntriesOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "..", "outside.txt")
	err := untar(tarBuffer(t, "../outside.txt", "escaped"), root)
	if err == nil {
		t.Fatal("untar returned nil error for escaping archive entry")
	}
	if !strings.Contains(err.Error(), "escapes extraction root") {
		t.Fatalf("untar error = %v, want escape error", err)
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("outside file exists or stat failed unexpectedly: %v", err)
	}
}

func TestUntarRejectsBackslashArchiveEntries(t *testing.T) {
	root := t.TempDir()
	err := untar(tarBuffer(t, `..\outside.txt`, "escaped"), root)
	if err == nil {
		t.Fatal("untar returned nil error for backslash archive entry")
	}
	if !strings.Contains(err.Error(), "backslash") {
		t.Fatalf("untar error = %v, want backslash error", err)
	}
}

func TestUntarExtractsSafeArchiveEntries(t *testing.T) {
	root := t.TempDir()
	if err := untar(tarBuffer(t, "module/module.yaml", "code: demo\n"), root); err != nil {
		t.Fatalf("untar returned error for safe entry: %v", err)
	}
	// #nosec G304 -- test reads the known temp-file path it just extracted.
	data, err := os.ReadFile(filepath.Join(root, "module", "module.yaml"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(data) != "code: demo\n" {
		t.Fatalf("extracted file = %q", data)
	}
}

func copyFixtureModule(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "builder", "testdata", "simple-module")
	dst := filepath.Join(t.TempDir(), "simple-module")
	copyDir(t, src, dst)
	return dst
}

func tarBuffer(t *testing.T, name, contents string) *bytes.Reader {
	t.Helper()
	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	if err := writer.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0600,
		Size: int64(len(contents)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buffer.Bytes())
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0700); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyDir(t, srcPath, dstPath)
			continue
		}
		// #nosec G304 -- test fixture copy reads files from a controlled repository testdata path.
		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatal(err)
		}
		// #nosec G703 -- dstPath is constructed under the test temp directory by copyDir.
		if err := os.WriteFile(dstPath, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
