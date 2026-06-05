package moduleinit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldModuleSkeleton(t *testing.T) {
	dir := t.TempDir()

	if err := ScaffoldModuleSkeleton(dir, ModuleScaffoldOptions{Name: "config-map-module"}); err != nil {
		t.Fatalf("ScaffoldModuleSkeleton returned error: %v", err)
	}

	for _, name := range []string{
		"config-map-module/module/module.yaml",
		"config-map-module/module/description.md",
		"config-map-module/module/01-getting-started/chapter.yml",
		"config-map-module/module/01-getting-started/description.md",
	} {
		assertFileExists(t, filepath.Join(dir, name))
	}

	moduleYAML := readFile(t, filepath.Join(dir, "config-map-module/module/module.yaml"))
	for _, want := range []string{
		"title: Config Map Module",
		"shortDescription: Local lab module",
		"bannerImage: https://storage.googleapis.com/cdn-training-platform/local-lab-banner.png",
		"level: BEGINNER",
		"tags:",
		"  - local-lab",
	} {
		if !strings.Contains(moduleYAML, want) {
			t.Fatalf("module.yaml does not contain %q", want)
		}
	}
	if strings.Contains(moduleYAML, "description:") {
		t.Fatal("module.yaml contains unsupported source module description property")
	}

	moduleDescription := readFile(t, filepath.Join(dir, "config-map-module/module/description.md"))
	if !strings.Contains(moduleDescription, "# Config Map Module") {
		t.Fatal("module/description.md does not contain module title")
	}

	chapterYAML := readFile(t, filepath.Join(dir, "config-map-module/module/01-getting-started/chapter.yml"))
	for _, want := range []string{
		"title: Getting Started",
		"shortDescription: Start building local labs.",
		"isDraft: false",
	} {
		if !strings.Contains(chapterYAML, want) {
			t.Fatalf("chapter.yml does not contain %q", want)
		}
	}
	for _, reject := range []string{
		"index:",
		"description:",
	} {
		if strings.Contains(chapterYAML, reject) {
			t.Fatalf("chapter.yml contains unsupported source chapter property %q", reject)
		}
	}

	for _, name := range []string{
		"config-map-module/module/module.yaml",
		"config-map-module/module/description.md",
		"config-map-module/module/01-getting-started/chapter.yml",
		"config-map-module/module/01-getting-started/description.md",
	} {
		assertASCII(t, name, readFile(t, filepath.Join(dir, name)))
	}

	nonEmpty := filepath.Join(t.TempDir(), "existing-module")
	writeFile(t, filepath.Join(nonEmpty, "keep.txt"), "do not overwrite")
	if err := ScaffoldModuleSkeleton(filepath.Dir(nonEmpty), ModuleScaffoldOptions{Name: filepath.Base(nonEmpty)}); err == nil {
		t.Fatal("ScaffoldModuleSkeleton returned nil error for non-empty target")
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(contents)
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

func assertASCII(t *testing.T, name, contents string) {
	t.Helper()
	for i, r := range contents {
		if r > 127 {
			t.Fatalf("%s contains non-ASCII rune %q at byte %d", name, r, i)
		}
	}
}
