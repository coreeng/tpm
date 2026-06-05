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

func TestInitLabStandaloneCreatesLabAndPrintsNextStep(t *testing.T) {
	target := filepath.Join(t.TempDir(), "pod-image-lab")
	output, err := executeRootCommand("init", "lab", target)
	if err != nil {
		t.Fatalf("init lab returned error: %v\n%s", err, output)
	}

	assertCmdFileExists(t, filepath.Join(target, "lab.yaml"))
	assertCmdFileExists(t, filepath.Join(target, "validator", "main.go"))
	if !strings.Contains(output, "Created standalone lab") {
		t.Fatalf("output does not describe created lab:\n%s", output)
	}
	if !strings.Contains(output, "tpm lab run "+target) {
		t.Fatalf("output does not include next lab run command:\n%s", output)
	}
	if strings.Contains(strings.ToLower(output), "assessment") {
		t.Fatalf("output contains user-facing assessment wording:\n%s", output)
	}
}

func TestInitLabModuleBackedCreatesLabAndPrintsNextStep(t *testing.T) {
	moduleDir := t.TempDir()
	output, err := executeRootCommand("init", "lab", moduleDir, "01-pod-images", "01-pod-image-lab", "--module-backed")
	if err != nil {
		t.Fatalf("init lab --module-backed returned error: %v\n%s", err, output)
	}

	assertCmdFileExists(t, filepath.Join(moduleDir, "module", "01-pod-images", "assessments", "01-pod-image-lab", "assessment.yaml"))
	assertCmdFileExists(t, filepath.Join(moduleDir, "assessments", "01-pod-images", "01-pod-image-lab", "validator", "main.go"))
	if !strings.Contains(output, "Created module-backed lab") {
		t.Fatalf("output does not describe created module-backed lab:\n%s", output)
	}
	if !strings.Contains(output, "tpm lab run "+filepath.Join(moduleDir, "assessments", "01-pod-images", "01-pod-image-lab")) {
		t.Fatalf("output does not include next lab run command:\n%s", output)
	}
}

func TestInitLabValidatesArguments(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "standalone rejects extra args",
			args: []string{"init", "lab", "one", "two"},
			want: "standalone lab requires exactly 1 argument",
		},
		{
			name: "module backed requires three args",
			args: []string{"init", "lab", "module", "chapter", "--module-backed"},
			want: "module-backed lab requires exactly 3 arguments",
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

func TestInitHelpDescribesGenericStandaloneLabs(t *testing.T) {
	output, err := executeRootCommand("init", "--help")
	if err != nil {
		t.Fatalf("init --help returned error: %v\n%s", err, output)
	}

	if !strings.Contains(output, "Use standalone labs when you want a local lab outside a module.") {
		t.Fatalf("init help does not describe standalone labs generically:\n%s", output)
	}
	if strings.Contains(output, "local ConfigMap lab") {
		t.Fatalf("init help describes standalone labs as ConfigMap-specific:\n%s", output)
	}
}

func TestInitLabHelpDescribesPodImageLab(t *testing.T) {
	output, err := executeRootCommand("init", "lab", "--help")
	if err != nil {
		t.Fatalf("init lab --help returned error: %v\n%s", err, output)
	}

	if !strings.Contains(output, "Create a Pod image lab skeleton") {
		t.Fatalf("init lab help does not describe Pod image lab skeleton:\n%s", output)
	}
	if !strings.Contains(output, "--artifact-registry") {
		t.Fatalf("init lab help does not include artifact registry flag:\n%s", output)
	}
	if !strings.Contains(output, "Usage:\n  tpm init lab <path> [flags]") {
		t.Fatalf("init lab help does not use clean standalone usage:\n%s", output)
	}
	if strings.Contains(output, "lab <path> | lab <module>") {
		t.Fatalf("init lab help still uses ambiguous alternative usage:\n%s", output)
	}
	if strings.Contains(output, "Create a ConfigMap lab skeleton") {
		t.Fatalf("init lab help still describes ConfigMap lab skeleton:\n%s", output)
	}
}

func TestInitModuleCreatesModuleSkeletonAndPrintsNextStep(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo-module")
	output, err := executeRootCommand("init", "module", target)
	if err != nil {
		t.Fatalf("init module returned error: %v\n%s", err, output)
	}

	assertCmdFileExists(t, filepath.Join(target, "module", "module.yaml"))
	assertCmdFileExists(t, filepath.Join(target, "module", "01-getting-started", "chapter.yml"))
	if !strings.Contains(output, "Created module") {
		t.Fatalf("output does not describe created module:\n%s", output)
	}
	if !strings.Contains(output, "tpm init lab . <chapter> <lab-name> --module-backed") {
		t.Fatalf("output does not include next module-backed lab command:\n%s", output)
	}
}

func TestInitModuleRelativePathPrintsCoherentNextStep(t *testing.T) {
	t.Chdir(t.TempDir())

	output, err := executeRootCommand("init", "module", "demo-module")
	if err != nil {
		t.Fatalf("init module returned error: %v\n%s", err, output)
	}

	assertCmdFileExists(t, filepath.Join("demo-module", "module", "module.yaml"))
	if !strings.Contains(output, "cd demo-module") {
		t.Fatalf("output does not include cd into relative module path:\n%s", output)
	}
	if !strings.Contains(output, "tpm init lab . <chapter> <lab-name> --module-backed") {
		t.Fatalf("output does not use current directory for module-backed lab after cd:\n%s", output)
	}
	if strings.Contains(output, "tpm init lab demo-module <chapter> <lab-name> --module-backed") {
		t.Fatalf("output uses nested relative module path after cd:\n%s", output)
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
