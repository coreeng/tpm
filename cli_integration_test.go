package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIRealCommandsAgainstFixtures(t *testing.T) {
	tempDir := t.TempDir()
	bin := filepath.Join(tempDir, "tpm")
	// #nosec G204 -- integration test builds the local package with a fixed executable.
	if output, err := exec.Command("go", "build", "-buildvcs=false", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	modulesDir := filepath.Join(tempDir, "modules")
	modulePath := filepath.Join(modulesDir, "kubernetes-101")
	copyIntegrationDir(t, filepath.Join("examples", "modules", "kubernetes-101"), modulePath)

	output := runCLI(t, bin, "module", "list", modulesDir)
	if strings.TrimSpace(output) != "kubernetes-101" {
		t.Fatalf("module list output = %q, want kubernetes-101", output)
	}

	outRoot := filepath.Join(tempDir, "artifacts")
	output = runCLI(t, bin, "module", "build", modulePath, "--out-root", outRoot)
	artifactPath := filepath.Join(outRoot, "kubernetes-101", "module.yaml")
	assertIntegrationFileExists(t, artifactPath)
	if !strings.Contains(output, "kubernetes-101 -> "+artifactPath) {
		t.Fatalf("module build output does not include artifact path:\n%s", output)
	}

	output = runCLI(t, bin, "artifact", "validate", filepath.Dir(artifactPath))
	if !strings.Contains(output, filepath.Dir(artifactPath)+": ok") {
		t.Fatalf("artifact validate output does not report ok:\n%s", output)
	}

	output = runCLI(t, bin, "module", "compare", modulePath, modulePath)
	if !strings.Contains(output, "removed: 0") || !strings.Contains(output, "moved between parents: 0") {
		t.Fatalf("module compare output does not report no breaking changes:\n%s", output)
	}

	labPath := filepath.Join(tempDir, "pod-image-lab")
	output = runCLI(t, bin, "lab", "init", labPath)
	if !strings.Contains(output, "Created standalone lab") {
		t.Fatalf("lab init output does not describe created lab:\n%s", output)
	}
	output = runCLI(t, bin, "lab", "outline", labPath, "--codes")
	if !strings.Contains(output, "Deploy a Pod from a built image") || !strings.Contains(output, "code:") {
		t.Fatalf("lab outline output does not contain expected lab content:\n%s", output)
	}
}

func runCLI(t *testing.T, bin string, args ...string) string {
	t.Helper()
	// #nosec G204 -- integration test invokes the just-built TPM binary with test-controlled arguments.
	cmd := exec.Command(bin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", strings.Join(append([]string{bin}, args...), " "), err, output)
	}
	return string(output)
}

func copyIntegrationDir(t *testing.T, src, dst string) {
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
			copyIntegrationDir(t, srcPath, dstPath)
			continue
		}
		// #nosec G304 -- integration test copies controlled repository fixtures.
		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatal(err)
		}
		// #nosec G703 -- dstPath is constructed under this test's temp directory.
		if err := os.WriteFile(dstPath, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func assertIntegrationFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}
