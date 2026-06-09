package lab

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestListGroupsNamespacesByRunID(t *testing.T) {
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse([]byte(`{
		"items": [
			{"metadata": {"name": "lab-def456-workspace", "labels": {"training-platform.coreeng.io/managed-by": "tpm", "training-platform.coreeng.io/lab-run-id": "def456", "training-platform.coreeng.io/lab-code": "other-lab", "training-platform.coreeng.io/lab-namespace-role": "workspace"}}},
			{"metadata": {"name": "lab-abc123-system", "labels": {"training-platform.coreeng.io/managed-by": "tpm", "training-platform.coreeng.io/lab-run-id": "abc123", "training-platform.coreeng.io/lab-code": "best-lap", "training-platform.coreeng.io/lab-namespace-role": "system"}}},
			{"metadata": {"name": "lab-abc123-workspace", "labels": {"training-platform.coreeng.io/managed-by": "tpm", "training-platform.coreeng.io/lab-run-id": "abc123", "training-platform.coreeng.io/lab-code": "best-lap", "training-platform.coreeng.io/lab-namespace-role": "workspace"}}}
		]
	}`), nil)

	output, err := List(context.Background(), Options{Runner: runner})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	wantLines := []string{
		"RUN ID",
		"LAB CODE",
		"SYSTEM NAMESPACE",
		"WORKSPACE NAMESPACE",
		"abc123",
		"best-lap",
		"lab-abc123-system",
		"lab-abc123-workspace",
		"def456",
		"other-lab",
		"-",
		"lab-def456-workspace",
	}
	assertContainsInOrder(t, output, wantLines)
}

func TestListIncludesCreatedTimeAndSortsNewestFirst(t *testing.T) {
	repoRoot := t.TempDir()
	stateDir := StateDir(repoRoot)
	olderCreatedAt := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	newerCreatedAt := time.Date(2026, 6, 2, 11, 0, 0, 0, time.UTC)
	for _, state := range []RunState{
		{RunID: "older", LabPath: repoRoot + "/labs/older", CreatedAt: olderCreatedAt},
		{RunID: "newer", LabPath: repoRoot + "/labs/newer", CreatedAt: newerCreatedAt},
	} {
		if err := SaveState(stateDir, state); err != nil {
			t.Fatalf("SaveState returned error: %v", err)
		}
	}
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse([]byte(`{
		"items": [
			{"metadata": {"name": "lab-older-system", "labels": {"training-platform.coreeng.io/managed-by": "tpm", "training-platform.coreeng.io/lab-run-id": "older", "training-platform.coreeng.io/lab-code": "older-lab", "training-platform.coreeng.io/lab-namespace-role": "system"}}},
			{"metadata": {"name": "lab-newer-system", "labels": {"training-platform.coreeng.io/managed-by": "tpm", "training-platform.coreeng.io/lab-run-id": "newer", "training-platform.coreeng.io/lab-code": "newer-lab", "training-platform.coreeng.io/lab-namespace-role": "system"}}}
		]
	}`), nil)

	output, err := List(context.Background(), Options{RepoRoot: repoRoot, StateDir: stateDir, Runner: runner})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	assertContainsInOrder(t, output, []string{
		"CREATED",
		"newer",
		newerCreatedAt.Format(time.RFC3339),
		"older",
		olderCreatedAt.Format(time.RFC3339),
	})
}

func TestListAlignsColumnsForUUIDRunIDs(t *testing.T) {
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse([]byte(`{"items":[{"metadata":{"name":"lab-3c1a2e36-a04a-4420-acf0-1830b74a2e25-system","labels":{"training-platform.coreeng.io/managed-by":"tpm","training-platform.coreeng.io/lab-run-id":"3c1a2e36-a04a-4420-acf0-1830b74a2e25","training-platform.coreeng.io/lab-code":"best-lap","training-platform.coreeng.io/lab-namespace-role":"system"}}}]}`), nil)

	output, err := List(context.Background(), Options{Runner: runner})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("output lines = %#v, want header and one row", lines)
	}
	headerLabCode := strings.Index(lines[0], "LAB CODE")
	rowLabCode := strings.Index(lines[1], "best-lap")
	if headerLabCode == -1 || rowLabCode == -1 {
		t.Fatalf("output = %q, missing LAB CODE or best-lap", output)
	}
	if rowLabCode != headerLabCode {
		t.Fatalf("LAB CODE column starts at %d, row starts at %d in output:\n%s", headerLabCode, rowLabCode, output)
	}
}

func TestListRendersMissingWorkspaceNamespace(t *testing.T) {
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse([]byte(`{"items":[{"metadata":{"name":"lab-abc123-system","labels":{"training-platform.coreeng.io/managed-by":"tpm","training-platform.coreeng.io/lab-run-id":"abc123","training-platform.coreeng.io/lab-code":"best-lap","training-platform.coreeng.io/lab-namespace-role":"system"}}}]}`), nil)

	output, err := List(context.Background(), Options{Runner: runner})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	assertContainsInOrder(t, output, []string{"abc123", "best-lap", "lab-abc123-system", "-"})
}

func TestListReportsNoActiveLabs(t *testing.T) {
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse([]byte(`{"items":[]}`), nil)

	output, err := List(context.Background(), Options{Runner: runner})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if strings.TrimSpace(output) != "No active labs found" {
		t.Fatalf("output = %q, want no active labs message", output)
	}
}

func TestListRejectsNonKindContextByDefault(t *testing.T) {
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("prod-cluster\n"), nil)

	_, err := List(context.Background(), Options{Runner: runner})
	if err == nil {
		t.Fatal("List returned nil error for non-kind context")
	}
	if !strings.Contains(err.Error(), "kind-") {
		t.Fatalf("error %q does not mention kind- context requirement", err.Error())
	}
}
