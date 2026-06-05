package lab

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestStatusRendersLabTerms(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := StateDir(repoRoot)
	labPath := repoRoot + "/labs/pod-image-lab"
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
	runner.QueueResponse([]byte(`{
		"items": [{
			"metadata": {"name": "validator-abc123"},
			"status": {"conditions": [
				{"type": "ContainersReady", "status": "True"},
				{"type": "IA_Ready", "status": "True", "reason": "Ready", "message": "Validator is ready"},
				{"type": "IAC_DeployPodFromImage", "status": "False", "reason": "WaitingForGoals", "message": "Challenge is waiting for goals"},
				{"type": "IAG_DeployPodFromImage_PodUsesBuiltImage", "status": "True", "reason": "PodUsesBuiltImage", "message": "Pod uses the built image"},
				{"type": "IA_Completed", "status": "False", "reason": "Incomplete", "message": "Lab is still in progress"}
			]}
		}]
	}`), nil)

	status, err := Status(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: "abc123", Runner: runner})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	for _, want := range []string{
		"Lab abc123 status",
		"Pod validator-abc123",
		"Lab Progress",
		"TYPE",
		"NAME",
		"STATUS",
		"REASON",
		"MESSAGE",
		"lab",
		"Ready",
		"Validator is ready",
		"challenge",
		"DeployPodFromImage",
		"WaitingForGoals",
		"goal",
		"DeployPodFromImage/PodUsesBuiltImage",
		"PodUsesBuiltImage",
		"Validator Pod Conditions",
		"ContainersReady: True",
	} {
		if !strings.Contains(status, want) {
			t.Fatalf("status does not contain %q:\n%s", want, status)
		}
	}
	if strings.Contains(status, "lab Ready: True") || strings.Contains(status, "challenge DeployPodFromImage: False") {
		t.Fatalf("status still renders lab progress as plain condition lines:\n%s", status)
	}
	if strings.Contains(strings.ToLower(status), "assessment") {
		t.Fatalf("status contains assessment terminology: %q", status)
	}

	wantCommands := []Command{
		{Name: "kubectl", Args: []string{"get", "pods", "-n", "lab-abc123-system", "-l", "app.kubernetes.io/component=validator", "-o", "json"}},
	}
	if diff := cmp.Diff(wantCommands, runner.Commands); diff != "" {
		t.Fatalf("commands mismatch (-want +got):\n%s", diff)
	}
}

func TestStatusRendersFailedProvisioningWhenValidatorPodMissing(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := StateDir(repoRoot)
	state := testRunState(t, stateDir, repoRoot+"/labs/pod-image-lab")
	runner := NewFakeRunner()
	runner.QueueResponse([]byte(`{"items": []}`), nil)

	status, err := Status(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: state.RunID, Runner: runner})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	for _, want := range []string{
		"Lab abc123 status",
		"Provisioning: Failed",
		"No validator pod found in system namespace lab-abc123-system.",
		"The lab runtime chart likely failed to install or was uninstalled before the validator started.",
		"helm status lab-abc123 -n lab-abc123-system",
	} {
		if !strings.Contains(status, want) {
			t.Fatalf("status does not contain %q:\n%s", want, status)
		}
	}
}

func TestStatusRejectsMultipleValidatorPods(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := StateDir(repoRoot)
	state := testRunState(t, stateDir, repoRoot+"/labs/pod-image-lab")
	runner := NewFakeRunner()
	runner.QueueResponse([]byte(`{"items": [{"metadata": {"name": "validator-a"}}, {"metadata": {"name": "validator-b"}}]}`), nil)

	_, err := Status(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: state.RunID, Runner: runner})
	if err == nil || !strings.Contains(err.Error(), "expected one lab validator pod") {
		t.Fatalf("Status error = %v, want multiple validator pods error", err)
	}
}

func TestStatusSortsConditions(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := StateDir(repoRoot)
	state := testRunState(t, stateDir, repoRoot+"/labs/pod-image-lab")
	runner := NewFakeRunner()
	runner.QueueResponse([]byte(`{
		"items": [{
			"metadata": {"name": "validator-abc123"},
			"status": {"conditions": [
				{"type": "ZZ_Custom", "status": "False"},
				{"type": "IA_Completed", "status": "False"},
				{"type": "IAG_DeployPodFromImage_PodUsesBuiltImage", "status": "True"},
				{"type": "IA_Custom", "status": "False"},
				{"type": "IA_Ready", "status": "True"},
				{"type": "IAC_DeployPodFromImage", "status": "False"}
			]}
		}]
	}`), nil)

	status, err := Status(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: state.RunID, Runner: runner})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	readyIndex := assertStatusContainsToken(t, status, "Ready")
	challengeIndex := assertStatusContainsToken(t, status, "challenge")
	goalIndex := assertStatusContainsToken(t, status, "goal")
	completedIndex := assertStatusContainsToken(t, status, "Completed")
	customLabIndex := assertStatusContainsToken(t, status, "Custom")
	unknownIndex := assertStatusContainsToken(t, status, "ZZ_Custom")
	if !(readyIndex < challengeIndex && challengeIndex < goalIndex && goalIndex < completedIndex && completedIndex < customLabIndex && customLabIndex < unknownIndex) {
		t.Fatalf("conditions not sorted by type in status:\n%s", status)
	}
}

func TestStatusRendersConditionLabelEdges(t *testing.T) {
	for _, tt := range []struct {
		conditionType string
		want          string
	}{
		{conditionType: "IAG_PodUsesBuiltImage", want: "goal PodUsesBuiltImage"},
		{conditionType: "UnknownCondition", want: "UnknownCondition"},
	} {
		if got := renderConditionLabel(tt.conditionType); got != tt.want {
			t.Fatalf("renderConditionLabel(%q) = %q, want %q", tt.conditionType, got, tt.want)
		}
	}
}

func TestStatusIndicatesCompletedFromLabProgressTable(t *testing.T) {
	status := `Lab abc123 status
Pod validator-abc123
Lab Progress
+------+-----------+--------+--------+---------+
| TYPE | NAME      | STATUS | REASON | MESSAGE |
+------+-----------+--------+--------+---------+
| lab  | Completed | True   | Done   | done    |
+------+-----------+--------+--------+---------+`

	if !statusIndicatesCompleted(status) {
		t.Fatalf("statusIndicatesCompleted returned false for completed lab table:\n%s", status)
	}
	if statusIndicatesCompleted(strings.Replace(status, "True", "False", 1)) {
		t.Fatalf("statusIndicatesCompleted returned true for incomplete lab table")
	}
}

func TestStatusReturnsKubectlAndJSONErrors(t *testing.T) {
	ctx := context.Background()
	repoRoot := t.TempDir()
	stateDir := StateDir(repoRoot)
	state := testRunState(t, stateDir, repoRoot+"/labs/pod-image-lab")

	runner := NewFakeRunner()
	runner.QueueResponse(nil, errBoom{})
	_, err := Status(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: state.RunID, Runner: runner})
	if err == nil || !strings.Contains(err.Error(), "get lab validator pods") {
		t.Fatalf("Status error = %v, want kubectl error", err)
	}

	runner = NewFakeRunner()
	runner.QueueResponse([]byte(`not-json`), nil)
	_, err = Status(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: state.RunID, Runner: runner})
	if err == nil || !strings.Contains(err.Error(), "parse lab validator pods") {
		t.Fatalf("Status error = %v, want parse error", err)
	}
}

func testRunState(t *testing.T, stateDir, labPath string) RunState {
	t.Helper()
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
	return state
}

func assertStatusContainsToken(t *testing.T, status, token string) int {
	t.Helper()
	index := strings.Index(status, token)
	if index == -1 {
		t.Fatalf("status does not contain sort token %q:\n%s", token, status)
	}
	return index
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }
