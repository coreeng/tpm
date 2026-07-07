package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestRootCommandRegistersModernGroupsOnly(t *testing.T) {
	for _, name := range []string{"module", "artifact", "lab"} {
		if _, _, err := rootCmd.Find([]string{name}); err != nil {
			t.Fatalf("root command missing %q: %v", name, err)
		}
	}
	for _, name := range []string{
		"list",
		"validate",
		"build",
		"validate-changes",
		"validate-artifact",
		"generate-codes",
		"generate-markdown",
		"init",
	} {
		if _, _, err := rootCmd.Find([]string{name}); err == nil {
			t.Fatalf("root command still exposes legacy %q", name)
		}
	}
}

func TestModuleInitListValidateBuildAndArtifactValidate(t *testing.T) {
	tempDir := t.TempDir()
	modulePath := filepath.Join(tempDir, "demo-module")
	output, err := executeRootCommand("module", "init", modulePath)
	if err != nil {
		t.Fatalf("module init returned error: %v\n%s", err, output)
	}
	assertCmdFileExists(t, filepath.Join(modulePath, "module", "module.yaml"))
	if !strings.Contains(output, "Created module") {
		t.Fatalf("module init output does not describe created module:\n%s", output)
	}

	output, err = executeRootCommand("module", "list", tempDir)
	if err != nil {
		t.Fatalf("module list returned error: %v\n%s", err, output)
	}
	if strings.TrimSpace(output) != "demo-module" {
		t.Fatalf("module list output = %q, want demo-module", output)
	}

	fixture := filepath.Join("..", "pkg", "builder", "testdata", "simple-module")
	output, err = executeRootCommand("module", "validate", fixture)
	if err != nil {
		t.Fatalf("module validate returned error: %v\n%s", err, output)
	}
	if !strings.Contains(output, "simple-module: ok") {
		t.Fatalf("module validate output does not report ok:\n%s", output)
	}

	outRoot := filepath.Join(tempDir, "artifacts")
	output, err = executeRootCommand("module", "build", fixture, "--out-root", outRoot)
	if err != nil {
		t.Fatalf("module build returned error: %v\n%s", err, output)
	}
	artifactPath := filepath.Join(outRoot, "simple-module", "module.yaml")
	assertCmdFileExists(t, artifactPath)
	if !strings.Contains(output, "simple-module -> "+artifactPath) {
		t.Fatalf("module build output does not report artifact path:\n%s", output)
	}

	output, err = executeRootCommand("artifact", "validate", filepath.Dir(artifactPath))
	if err != nil {
		t.Fatalf("artifact validate returned error: %v\n%s", err, output)
	}
	if !strings.Contains(output, filepath.Dir(artifactPath)+": ok") {
		t.Fatalf("artifact validate output does not report ok:\n%s", output)
	}
}

func TestLabInitOutlineAndRuntimeSurface(t *testing.T) {
	labPath := filepath.Join(t.TempDir(), "pod-image-lab")
	output, err := executeRootCommand("lab", "init", labPath)
	if err != nil {
		t.Fatalf("lab init returned error: %v\n%s", err, output)
	}
	assertCmdFileExists(t, filepath.Join(labPath, "lab.yaml"))
	assertCmdFileExists(t, filepath.Join(labPath, "validator", "main.go"))

	output, err = executeRootCommand("lab", "outline", labPath, "--codes", "--paths")
	if err != nil {
		t.Fatalf("lab outline returned error: %v\n%s", err, output)
	}
	for _, want := range []string{"Pod Image Lab", "Deploy a Pod from a built image", "metadata:", "runtime:", "code:"} {
		if !strings.Contains(output, want) {
			t.Fatalf("lab outline output does not contain %q:\n%s", want, output)
		}
	}

	output, err = executeRootCommand("lab", "start", "--help")
	if err != nil {
		t.Fatalf("lab start --help returned error: %v\n%s", err, output)
	}
	for _, want := range []string{"--chart-dir", "--chart-uri", "--state-dir", "--validator-registry"} {
		if !strings.Contains(output, want) {
			t.Fatalf("lab start help does not contain %q:\n%s", want, output)
		}
	}

	output, err = executeRootCommand("lab", "run")
	if err == nil {
		t.Fatalf("lab run unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("lab run error does not report unknown command: %v\n%s", err, output)
	}

	output, err = executeRootCommand("lab", "cleanup")
	if err == nil {
		t.Fatalf("lab cleanup unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("lab cleanup error does not report unknown command: %v\n%s", err, output)
	}
}

func TestLabStartValidatesChartFlagsBeforeLoadingLab(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "conflicting chart sources",
			args: []string{"lab", "start", "does-not-exist", "--chart-dir", "/tmp/local-chart", "--chart-uri", "oci://example.com/chart"},
			want: "set either chart-dir or chart-uri",
		},
		{
			name: "blank chart uri",
			args: []string{"lab", "start", "does-not-exist", "--chart-uri", " "},
			want: "chart-uri must not be blank",
		},
		{
			name: "blank chart version",
			args: []string{"lab", "start", "does-not-exist", "--chart-uri", "oci://example.com/chart", "--chart-version", "\t"},
			want: "chart-version must not be blank",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executeRootCommand(tt.args...)
			if err == nil {
				t.Fatalf("command returned nil error, output:\n%s", output)
			}
			if !strings.Contains(err.Error(), tt.want) && !strings.Contains(output, tt.want) {
				t.Fatalf("error/output does not contain %q\nerror: %v\noutput:\n%s", tt.want, err, output)
			}
		})
	}
}

func executeRootCommand(args ...string) (string, error) {
	var output bytes.Buffer
	resetCommandFlags(rootCmd)
	rootCmd.SetArgs(args)
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&output)
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.SilenceUsage = false
	rootCmd.SilenceErrors = false
	resetCommandFlags(rootCmd)
	return output.String(), err
}

func resetCommandFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
		_ = flag.Value.Set(flag.DefValue)
	})
	for _, child := range cmd.Commands() {
		resetCommandFlags(child)
	}
}

func assertCmdFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}
