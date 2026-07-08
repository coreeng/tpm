package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coreeng/tpm/pkg/builder"
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

	output, err = executeRootCommand("lab", "preview", "--help")
	if err != nil {
		t.Fatalf("lab preview --help returned error: %v\n%s", err, output)
	}
	for _, want := range []string{"--addr", "--no-open-browser", "--watch", "--chart-dir", "--chart-uri", "--state-dir", "--validator-registry"} {
		if !strings.Contains(output, want) {
			t.Fatalf("lab preview help does not contain %q:\n%s", want, output)
		}
	}

	output, err = executeRootCommand("lab", "cleanup", "--help")
	if err != nil {
		t.Fatalf("lab cleanup --help returned error: %v\n%s", err, output)
	}
	if !strings.Contains(output, "--id") || !strings.Contains(output, "--state-dir") {
		t.Fatalf("lab cleanup help does not contain expected flags:\n%s", output)
	}

	output, err = executeRootCommand("lab", "run")
	if err == nil {
		t.Fatalf("lab run unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("lab run error does not report unknown command: %v\n%s", err, output)
	}

	output, err = executeRootCommand("lab", "start")
	if err == nil {
		t.Fatalf("lab start unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("lab start error does not report unknown command: %v\n%s", err, output)
	}

	output, err = executeRootCommand("lab", "stop")
	if err == nil {
		t.Fatalf("lab stop unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("lab stop error does not report unknown command: %v\n%s", err, output)
	}
}

func TestModulePreviewHelpAndTemplate(t *testing.T) {
	output, err := executeRootCommand("module", "preview", "--help")
	if err != nil {
		t.Fatalf("module preview --help returned error: %v\n%s", err, output)
	}
	for _, want := range []string{"--addr", "--no-open-browser", "--watch"} {
		if !strings.Contains(output, want) {
			t.Fatalf("module preview help does not contain %q:\n%s", want, output)
		}
	}

	fixture := filepath.Join("..", "pkg", "builder", "testdata", "simple-module")
	_, _, built, err := builder.Compile(fixture, "", "")
	if err != nil {
		t.Fatalf("Compile fixture returned error: %v", err)
	}
	var rendered bytes.Buffer
	if err := modulePreviewTemplate.Execute(&rendered, built); err != nil {
		t.Fatalf("module preview template returned error: %v", err)
	}
	for _, want := range []string{"Kubernetes 101", `<span class="toc-number">1.</span>`, `<span class="toc-number">2.</span>`, `<span class="toc-number">3.</span>`, "Cluster fundamentals", "Pods, Deployments, and Services", "Kubernetes operations check"} {
		if !strings.Contains(rendered.String(), want) {
			t.Fatalf("module preview render does not contain %q:\n%s", want, rendered.String())
		}
	}
}

func TestLabPreviewValidatesChartFlagsBeforeLoadingLab(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "conflicting chart sources",
			args: []string{"lab", "preview", "does-not-exist", "--chart-dir", "/tmp/local-chart", "--chart-uri", "oci://example.com/chart"},
			want: "set either chart-dir or chart-uri",
		},
		{
			name: "blank chart uri",
			args: []string{"lab", "preview", "does-not-exist", "--chart-uri", " "},
			want: "chart-uri must not be blank",
		},
		{
			name: "blank chart version",
			args: []string{"lab", "preview", "does-not-exist", "--chart-uri", "oci://example.com/chart", "--chart-version", "\t"},
			want: "chart-version must not be blank",
		},
		{
			name: "invalid preview address",
			args: []string{"lab", "preview", "does-not-exist", "--chart-uri", "oci://example.com/chart", "--addr", "127.0.0.1:99999"},
			want: "invalid port",
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

func TestModuleCompareBreakingPolicy(t *testing.T) {
	oldPath := copyCmdFixtureModule(t)
	newPath := copyCmdFixtureModule(t)
	if err := os.RemoveAll(filepath.Join(newPath, "module", "01-cluster-fundamentals", "01-what-is-kubernetes")); err != nil {
		t.Fatal(err)
	}

	output, err := executeRootCommand("module", "compare", oldPath, newPath)
	if err == nil {
		t.Fatalf("module compare unexpectedly succeeded:\n%s", output)
	}
	if !strings.Contains(output, "ERROR: breaking module changes detected") {
		t.Fatalf("compare output does not report breaking error:\n%s", output)
	}

	output, err = executeRootCommand("module", "compare", oldPath, newPath, "--breaking-policy", "warn")
	if err != nil {
		t.Fatalf("module compare --breaking-policy warn returned error: %v\n%s", err, output)
	}
	if !strings.Contains(output, "WARNING: breaking module changes detected") {
		t.Fatalf("compare output does not report warning:\n%s", output)
	}

	output, err = executeRootCommand("module", "compare", oldPath, newPath, "--allow-breaking")
	if err != nil {
		t.Fatalf("module compare --allow-breaking returned error: %v\n%s", err, output)
	}
	if !strings.Contains(output, "WARNING: breaking module changes detected") {
		t.Fatalf("allow-breaking output does not use warning policy:\n%s", output)
	}
}

func TestModuleAuthoringCommands(t *testing.T) {
	modulePath := copyCmdFixtureModule(t)
	output, err := executeRootCommand("module", "add", "section", modulePath, "--chapter", "1", "--at", "1", "--set", "code=new-section", "--set", "title=New Section")
	if err != nil {
		t.Fatalf("module add section returned error: %v\n%s", err, output)
	}
	assertCmdFileExists(t, filepath.Join(modulePath, "module", "01-cluster-fundamentals", "01-new-section", "section.yaml"))

	output, err = executeRootCommand("module", "edit", "section", modulePath, "--chapter", "1", "--section", "1", "--set", "title=Edited Section")
	if err != nil {
		t.Fatalf("module edit section returned error: %v\n%s", err, output)
	}
	// #nosec G304 -- test reads a fixture output path under its temp module.
	sectionYAML, err := os.ReadFile(filepath.Join(modulePath, "module", "01-cluster-fundamentals", "01-new-section", "section.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sectionYAML), "title: Edited Section") {
		t.Fatalf("section title was not edited:\n%s", sectionYAML)
	}

	output, err = executeRootCommand("module", "move", "section", modulePath, "--chapter", "1", "--from", "2", "--to", "1")
	if err != nil {
		t.Fatalf("module move section returned error: %v\n%s", err, output)
	}
	assertCmdFileExists(t, filepath.Join(modulePath, "module", "01-cluster-fundamentals", "01-what-is-kubernetes", "section.yaml"))

	output, err = executeRootCommand("module", "remove", "section", modulePath, "--chapter", "1", "--from", "2", "--yes", "--breaking-policy", "warn")
	if err != nil {
		t.Fatalf("module remove section returned error: %v\n%s", err, output)
	}
}

func TestModuleAuthoringInlineQuizAndLabResources(t *testing.T) {
	modulePath := copyCmdFixtureModule(t)
	commands := [][]string{
		{"module", "add", "quiz", modulePath, "--chapter", "1", "--at", "1", "--set", "code=quiz-one", "--set", "title=Quiz One", "--set", "description=Quiz description", "--set", "passingScore=80"},
		{"module", "add", "question", modulePath, "--chapter", "1", "--quiz", "1", "--at", "1", "--set", "code=question-one", "--set", "question=Pick one", "--set", "type=SINGLE"},
		{"module", "add", "option", modulePath, "--chapter", "1", "--quiz", "1", "--question", "1", "--at", "1", "--set", "text=First", "--set", "correct=true"},
		{"module", "add", "option", modulePath, "--chapter", "1", "--quiz", "1", "--question", "1", "--at", "2", "--set", "text=Second", "--set", "correct=false"},
		{"module", "move", "option", modulePath, "--chapter", "1", "--quiz", "1", "--question", "1", "--from", "2", "--to", "1"},
		{"module", "edit", "option", modulePath, "--chapter", "1", "--quiz", "1", "--question", "1", "--option", "1", "--set", "text=Edited Second"},
		{"module", "add", "lab", modulePath, "--chapter", "1", "--at", "1", "--set", "code=lab-one", "--set", "title=Lab One", "--set", "starterImageUri=oci://example/starter", "--set", "validatorImageUri=oci://example/validator", "--set", "imageVersion=0.0.1"},
		{"module", "add", "challenge", modulePath, "--chapter", "1", "--lab", "1", "--at", "1", "--set", "code=ChallengeOne", "--set", "title=Challenge One"},
		{"module", "add", "goal", modulePath, "--chapter", "1", "--lab", "1", "--challenge", "1", "--at", "1", "--set", "code=GoalOne", "--set", "title=Goal One", "--set", "description=Goal description"},
		{"module", "edit", "goal", modulePath, "--chapter", "1", "--lab", "1", "--challenge", "1", "--goal", "1", "--set", "description=Updated goal"},
	}
	for _, args := range commands {
		output, err := executeRootCommand(args...)
		if err != nil {
			t.Fatalf("%s returned error: %v\n%s", strings.Join(args, " "), err, output)
		}
	}

	// #nosec G304 -- test reads a fixture output path under its temp module.
	chapterYAML, err := os.ReadFile(filepath.Join(modulePath, "module", "01-cluster-fundamentals", "chapter.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"code: quiz-one", "text: Edited Second", "passingScore: 80"} {
		if !strings.Contains(string(chapterYAML), want) {
			t.Fatalf("chapter YAML does not contain %q:\n%s", want, chapterYAML)
		}
	}
	// #nosec G304 -- test reads a fixture output path under its temp module.
	challengeYAML, err := os.ReadFile(filepath.Join(modulePath, "module", "01-cluster-fundamentals", "assessments", "01-lab-one", "01-challenge-one", "challenge.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(challengeYAML), "description: Updated goal") {
		t.Fatalf("challenge YAML does not contain updated goal:\n%s", challengeYAML)
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
		if sliceValue, ok := flag.Value.(interface{ Replace([]string) error }); ok {
			_ = sliceValue.Replace([]string{})
			flag.Changed = false
			return
		}
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

func copyCmdFixtureModule(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "pkg", "builder", "testdata", "simple-module")
	dst := filepath.Join(t.TempDir(), "simple-module")
	copyCmdDir(t, src, dst)
	return dst
}

func copyCmdDir(t *testing.T, src, dst string) {
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
			copyCmdDir(t, srcPath, dstPath)
			continue
		}
		// #nosec G304 -- test copies controlled repository fixtures.
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
