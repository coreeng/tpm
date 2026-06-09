package cmd

import (
	"strings"
	"testing"
)

func TestLabHelpListsSubcommands(t *testing.T) {
	output, err := executeRootCommand("lab", "--help")
	if err != nil {
		t.Fatalf("lab --help returned error: %v\n%s", err, output)
	}

	for _, want := range []string{"run", "list", "status", "cleanup"} {
		if !strings.Contains(output, want) {
			t.Fatalf("lab help does not contain %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "apply-solution") {
		t.Fatalf("lab help contains removed apply-solution command:\n%s", output)
	}
	if strings.Contains(strings.ToLower(output), "assessment") {
		t.Fatalf("lab help contains user-facing assessment wording:\n%s", output)
	}
}

func TestLabRootHelpListsLabCommand(t *testing.T) {
	output, err := executeRootCommand("--help")
	if err != nil {
		t.Fatalf("--help returned error: %v\n%s", err, output)
	}

	if !strings.Contains(output, "lab") {
		t.Fatalf("root help does not list lab command:\n%s", output)
	}
}

func TestLabRunHelpShowsRuntimeFlags(t *testing.T) {
	output, err := executeRootCommand("lab", "run", "--help")
	if err != nil {
		t.Fatalf("lab run --help returned error: %v\n%s", err, output)
	}

	for _, want := range []string{
		"--id",
		"--chart-dir",
		"--chart-uri",
		"--chart-version",
		"--allow-non-kind",
		"--assume-image-accessible",
		"--check-interval",
		"--registry-domain",
		"--state-dir",
		"--validator-registry",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("lab run help does not contain %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "--local-registry-node-port") {
		t.Fatalf("lab run help contains removed local registry NodePort flag:\n%s", output)
	}
	assertStateDirHelpShowsConfigDefault(t, output)
}

func TestLabRunAcceptsChartDirWithoutChartVersion(t *testing.T) {
	output, err := executeRootCommand("lab", "run", "does-not-exist", "--chart-dir", "/tmp/local-chart")
	if err == nil {
		// The lab path is intentionally invalid, so success would be unexpected.
		t.Fatalf("lab run returned nil error, output:\n%s", output)
	}
	if strings.Contains(err.Error(), "chart-uri") || strings.Contains(output, "chart-uri") || strings.Contains(err.Error(), "chart-version") || strings.Contains(output, "chart-version") {
		t.Fatalf("chart-dir should not require chart-uri/chart-version\nerror: %v\noutput:\n%s", err, output)
	}
}

func TestLabRunRejectsConflictingChartSources(t *testing.T) {
	output, err := executeRootCommand("lab", "run", "does-not-exist", "--chart-dir", "/tmp/local-chart", "--chart-uri", "oci://example.com/chart", "--chart-version", "1.2.3")
	if err == nil {
		t.Fatalf("lab run returned nil error, output:\n%s", output)
	}
	if !strings.Contains(err.Error(), "set either chart-dir or chart-uri") && !strings.Contains(output, "set either chart-dir or chart-uri") {
		t.Fatalf("error/output does not report conflicting chart sources\nerror: %v\noutput:\n%s", err, output)
	}
}

func TestLabCleanupHelpShowsSupportedFlags(t *testing.T) {
	output, err := executeRootCommand("lab", "cleanup", "--help")
	if err != nil {
		t.Fatalf("lab cleanup --help returned error: %v\n%s", err, output)
	}

	for _, want := range []string{"--id", "--allow-non-kind", "--state-dir"} {
		if !strings.Contains(output, want) {
			t.Fatalf("lab cleanup help does not contain %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "--all ") || strings.Contains(output, "--all=") {
		t.Fatalf("lab cleanup exposes unsupported --all flag:\n%s", output)
	}
	assertStateDirHelpShowsConfigDefault(t, output)
}

func TestLabStatusHelpShowsConfigStateDirDefault(t *testing.T) {
	output, err := executeRootCommand("lab", "status", "--help")
	if err != nil {
		t.Fatalf("lab status --help returned error: %v\n%s", err, output)
	}
	assertStateDirHelpShowsConfigDefault(t, output)
}

func TestLabRunValidatesLabPathArgument(t *testing.T) {
	output, err := executeRootCommand("lab", "run")
	if err == nil {
		t.Fatalf("lab run returned nil error, output:\n%s", output)
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") && !strings.Contains(output, "accepts 1 arg") {
		t.Fatalf("error/output does not report missing lab path\nerror: %v\noutput:\n%s", err, output)
	}
}

func assertStateDirHelpShowsConfigDefault(t *testing.T, output string) {
	t.Helper()
	if !strings.Contains(output, "--state-dir") {
		t.Fatalf("lab help does not contain --state-dir:\n%s", output)
	}
	if !strings.Contains(output, "default ~/.config/tpm/labs") {
		t.Fatalf("lab help does not describe config-dir state default:\n%s", output)
	}
	if strings.Contains(output, "default .build/tpm/labs") {
		t.Fatalf("lab help still describes repo-local state default:\n%s", output)
	}
}

func TestLabRunRejectsBlankChartFlagsBeforeLoadingLab(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "blank chart uri",
			args: []string{"lab", "run", "does-not-exist", "--chart-uri", " ", "--chart-version", "1.2.3"},
			want: "chart-uri must not be blank",
		},
		{
			name: "blank chart version",
			args: []string{"lab", "run", "does-not-exist", "--chart-uri", "oci://example.com/chart", "--chart-version", "\t"},
			want: "chart-version must not be blank",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executeRootCommand(tt.args...)
			if err == nil {
				t.Fatalf("lab run returned nil error, output:\n%s", output)
			}
			if !strings.Contains(err.Error(), tt.want) && !strings.Contains(output, tt.want) {
				t.Fatalf("error/output does not contain %q\nerror: %v\noutput:\n%s", tt.want, err, output)
			}
		})
	}
}

func TestLabApplySolutionCommandIsRemoved(t *testing.T) {
	output, err := executeRootCommand("lab", "apply-solution")
	if err == nil {
		t.Fatalf("lab apply-solution returned nil error, output:\n%s", output)
	}
	if !strings.Contains(err.Error(), "unknown command") && !strings.Contains(output, "unknown command") {
		t.Fatalf("error/output does not report unknown command\nerror: %v\noutput:\n%s", err, output)
	}
}
