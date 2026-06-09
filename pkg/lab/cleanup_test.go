package lab

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestCleanupDeletesOnlyRecordedNamespaces(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := filepath.Join(repoRoot, "state")
	labPath := filepath.Join(repoRoot, "labs", "create-config-map")
	createLabDirs(t, labPath)

	state := RunState{
		LabPath:            labPath,
		RunID:              "abc123",
		SystemNamespace:    "lab-abc123-system",
		WorkspaceNamespace: "lab-abc123-workspace",
		ValidatorImageTag:  "validator:abc123",
		HelmReleaseName:    "lab-abc123",
		ChartURI:           "oci://example.com/chart",
		ChartVersion:       "1.0.0",
		CreatedAt:          time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC),
	}
	if err := SaveState(stateDir, state); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("kind-local\n"), nil)

	if err := Cleanup(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: "abc123", Runner: runner}); err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}

	want := []Command{
		{Name: "kubectl", Args: []string{"config", "current-context"}},
		{Name: "helm", Args: []string{"uninstall", "lab-abc123", "-n", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"delete", "namespace", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"delete", "namespace", "lab-abc123-workspace"}},
	}
	if diff := cmp.Diff(want, runner.Commands); diff != "" {
		t.Fatalf("commands mismatch (-want +got):\n%s", diff)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "abc123.yaml")); !os.IsNotExist(err) {
		t.Fatalf("state file still exists after cleanup, stat err: %v", err)
	}
}

func TestCleanupRejectsNonKindContextByDefault(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := filepath.Join(repoRoot, "state")
	labPath := filepath.Join(repoRoot, "labs", "create-config-map")
	createLabDirs(t, labPath)
	state := cleanupTestState(labPath)
	if err := SaveState(stateDir, state); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("docker-desktop\n"), nil)

	err := Cleanup(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: "abc123", Runner: runner})
	if err == nil {
		t.Fatal("Cleanup returned nil, want non-kind context error")
	}
	if !strings.Contains(err.Error(), "does not start with kind-") {
		t.Fatalf("Cleanup error = %v, want kind context message", err)
	}

	want := []Command{
		{Name: "kubectl", Args: []string{"config", "current-context"}},
	}
	if diff := cmp.Diff(want, runner.Commands); diff != "" {
		t.Fatalf("commands mismatch (-want +got):\n%s", diff)
	}
}

func TestCleanupAllowsNonKindContextWhenExplicitlyAllowed(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := filepath.Join(repoRoot, "state")
	labPath := filepath.Join(repoRoot, "labs", "create-config-map")
	createLabDirs(t, labPath)
	state := cleanupTestState(labPath)
	if err := SaveState(stateDir, state); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("docker-desktop\n"), nil)

	if err := Cleanup(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: "abc123", Runner: runner, AllowNonKind: true}); err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}

	want := []Command{
		{Name: "kubectl", Args: []string{"config", "current-context"}},
		{Name: "helm", Args: []string{"uninstall", "lab-abc123", "-n", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"delete", "namespace", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"delete", "namespace", "lab-abc123-workspace"}},
	}
	if diff := cmp.Diff(want, runner.Commands); diff != "" {
		t.Fatalf("commands mismatch (-want +got):\n%s", diff)
	}
}

func TestCleanupAttemptsAllDeletesAndKeepsStateOnFailure(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := filepath.Join(repoRoot, "state")
	labPath := filepath.Join(repoRoot, "labs", "create-config-map")
	createLabDirs(t, labPath)
	state := cleanupTestState(labPath)
	if err := SaveState(stateDir, state); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse(nil, errors.New("helm failed"))
	runner.QueueResponse(nil, errors.New("system delete failed"))
	runner.QueueResponse(nil, nil)

	err := Cleanup(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: "abc123", Runner: runner})
	if err == nil {
		t.Fatal("Cleanup returned nil, want aggregate cleanup error")
	}
	for _, wantMessage := range []string{"helm failed", "system delete failed"} {
		if !strings.Contains(err.Error(), wantMessage) {
			t.Fatalf("Cleanup error = %v, want message %q", err, wantMessage)
		}
	}

	want := []Command{
		{Name: "kubectl", Args: []string{"config", "current-context"}},
		{Name: "helm", Args: []string{"uninstall", "lab-abc123", "-n", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"delete", "namespace", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"delete", "namespace", "lab-abc123-workspace"}},
	}
	if diff := cmp.Diff(want, runner.Commands); diff != "" {
		t.Fatalf("commands mismatch (-want +got):\n%s", diff)
	}
	assertFileExists(t, filepath.Join(stateDir, "abc123.yaml"))
}

func TestCleanupRemovesStateWhenResourcesAlreadyDeleted(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := filepath.Join(repoRoot, "state")
	labPath := filepath.Join(repoRoot, "labs", "create-config-map")
	createLabDirs(t, labPath)
	state := cleanupTestState(labPath)
	if err := SaveState(stateDir, state); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse(nil, errors.New("release: not found"))
	runner.QueueResponse(nil, errors.New("Error from server (NotFound): namespaces \"lab-abc123-system\" not found"))
	runner.QueueResponse(nil, errors.New("Error from server (NotFound): namespaces \"lab-abc123-workspace\" not found"))

	if err := Cleanup(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: "abc123", Runner: runner}); err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "abc123.yaml")); !os.IsNotExist(err) {
		t.Fatalf("state file still exists after cleanup, stat err: %v", err)
	}
}

func cleanupTestState(labPath string) RunState {
	return RunState{
		LabPath:            labPath,
		RunID:              "abc123",
		SystemNamespace:    "lab-abc123-system",
		WorkspaceNamespace: "lab-abc123-workspace",
		ValidatorImageTag:  "validator:abc123",
		HelmReleaseName:    "lab-abc123",
		ChartURI:           "oci://example.com/chart",
		ChartVersion:       "1.0.0",
		CreatedAt:          time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC),
	}
}

func createLabDirs(t *testing.T, labPath string) {
	t.Helper()
	for _, name := range []string{"starter-content", "solution", "validator"} {
		if err := os.MkdirAll(filepath.Join(labPath, name), 0755); err != nil {
			t.Fatalf("create %s dir: %v", name, err)
		}
	}
	metadata := []byte("title: Create ConfigMap\ncode: create-config-map\ntimeLimit: 30m\n")
	if err := os.WriteFile(filepath.Join(labPath, "lab.yaml"), metadata, 0644); err != nil {
		t.Fatalf("write lab metadata: %v", err)
	}
}
