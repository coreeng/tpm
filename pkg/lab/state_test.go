package lab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestRunNamesUseLabPrefix(t *testing.T) {
	names := NewRunNames("abc123", "create-config-map")

	if names.SystemNamespace != "lab-abc123-system" {
		t.Errorf("SystemNamespace = %q, want lab-abc123-system", names.SystemNamespace)
	}
	if names.WorkspaceNamespace != "lab-abc123-workspace" {
		t.Errorf("WorkspaceNamespace = %q, want lab-abc123-workspace", names.WorkspaceNamespace)
	}
	if strings.Contains(names.SystemNamespace, "assessment") {
		t.Errorf("SystemNamespace %q contains assessment", names.SystemNamespace)
	}
	if strings.Contains(names.WorkspaceNamespace, "assessment") {
		t.Errorf("WorkspaceNamespace %q contains assessment", names.WorkspaceNamespace)
	}
}

func TestStateRoundTrip(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "state")
	createdAt := time.Date(2026, 6, 2, 10, 11, 12, 0, time.UTC)
	state := RunState{
		LabPath:            "labs/create-config-map",
		RunID:              "abc123",
		SystemNamespace:    "lab-abc123-system",
		WorkspaceNamespace: "lab-abc123-workspace",
		ValidatorImageTag:  "validator:abc123",
		HelmReleaseName:    "lab-abc123",
		ChartURI:           "oci://example.com/charts/training-platform-assessment",
		ChartVersion:       "1.2.3",
		CreatedAt:          createdAt,
	}

	if err := SaveState(stateDir, state); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}

	loaded, err := LoadState(filepath.Join(stateDir, "abc123.yaml"))
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}
	if diff := cmp.Diff(&state, loaded); diff != "" {
		t.Fatalf("loaded state mismatch (-want +got):\n%s", diff)
	}
}

func TestStateDirDefaultsToUserConfigDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	stateDir := StateDir(t.TempDir())

	want := filepath.Join(home, ".config", "tpm", "labs")
	if stateDir != want {
		t.Fatalf("StateDir() = %q, want %q", stateDir, want)
	}
	if strings.Contains(stateDir, string(os.PathSeparator)+".build"+string(os.PathSeparator)) {
		t.Fatalf("StateDir() = %q, should not use repo-local .build", stateDir)
	}
}

func TestLatestStateForLabPath(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "state")
	older := RunState{
		LabPath:            "labs/create-config-map",
		RunID:              "old",
		SystemNamespace:    "lab-old-system",
		WorkspaceNamespace: "lab-old-workspace",
		ValidatorImageTag:  "validator:old",
		HelmReleaseName:    "lab-old",
		ChartURI:           "oci://example.com/chart",
		ChartVersion:       "1.0.0",
		CreatedAt:          time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC),
	}
	newer := RunState{
		LabPath:            "labs/create-config-map",
		RunID:              "new",
		SystemNamespace:    "lab-new-system",
		WorkspaceNamespace: "lab-new-workspace",
		ValidatorImageTag:  "validator:new",
		HelmReleaseName:    "lab-new",
		ChartURI:           "oci://example.com/chart",
		ChartVersion:       "1.0.0",
		CreatedAt:          time.Date(2026, 6, 2, 11, 0, 0, 0, time.UTC),
	}
	otherLab := RunState{
		LabPath:            "labs/other",
		RunID:              "other",
		SystemNamespace:    "lab-other-system",
		WorkspaceNamespace: "lab-other-workspace",
		ValidatorImageTag:  "validator:other",
		HelmReleaseName:    "lab-other",
		ChartURI:           "oci://example.com/chart",
		ChartVersion:       "1.0.0",
		CreatedAt:          time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC),
	}

	for _, state := range []RunState{older, newer, otherLab} {
		if err := SaveState(stateDir, state); err != nil {
			t.Fatalf("SaveState(%s) returned error: %v", state.RunID, err)
		}
	}

	latest, err := FindLatestState(stateDir, "labs/create-config-map")
	if err != nil {
		t.Fatalf("FindLatestState returned error: %v", err)
	}
	if diff := cmp.Diff(&newer, latest); diff != "" {
		t.Fatalf("latest state mismatch (-want +got):\n%s", diff)
	}
}

func TestLatestStateForMissingStateDir(t *testing.T) {
	latest, err := FindLatestState(filepath.Join(t.TempDir(), "state"), "labs/create-config-map")
	if err != nil {
		t.Fatalf("FindLatestState returned error: %v", err)
	}
	if latest != nil {
		t.Fatalf("FindLatestState returned %#v, want nil", latest)
	}
}
