package lab

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestKindIntegration(t *testing.T) {
	if os.Getenv("TPM_KIND_INTEGRATION") != "true" {
		t.Skip("set TPM_KIND_INTEGRATION=true to run kind integration test")
	}

	chartURI := os.Getenv("TPM_LAB_CHART_URI")
	if chartURI == "" {
		t.Fatal("TPM_LAB_CHART_URI is required when TPM_KIND_INTEGRATION=true")
	}
	chartVersion := os.Getenv("TPM_LAB_CHART_VERSION")
	if chartVersion == "" {
		t.Fatal("TPM_LAB_CHART_VERSION is required when TPM_KIND_INTEGRATION=true")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	repoRoot := t.TempDir()
	stateDir := filepath.Join(repoRoot, ".state")
	runID := "kind-integration"
	defer cleanupKindIntegration(t, repoRoot, stateDir, runID)

	labPath := filepath.Join(repoRoot, "labs", "config-map-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "config-map-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}

	state, err := Run(ctx, Options{
		LabPath:       labPath,
		RepoRoot:      repoRoot,
		StateDir:      stateDir,
		ID:            runID,
		ChartURI:      chartURI,
		ChartVersion:  chartVersion,
		CheckInterval: time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	runner := ExecRunner{}
	if err := runner.Run(ctx, "kubectl", "apply", "-n", state.WorkspaceNamespace, "-f", filepath.Join(labPath, "solution")); err != nil {
		t.Fatalf("apply lab solution returned error: %v", err)
	}

	status, err := waitForCompletedStatus(ctx, repoRoot, stateDir, state.RunID)
	if err != nil {
		t.Fatalf("lab did not complete: %v\nlast status:\n%s", err, status)
	}
}

func cleanupKindIntegration(t *testing.T, repoRoot, stateDir, runID string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(stateDir, runID+".yaml")); err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Logf("stat lab run state before cleanup: %v", err)
		return
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cleanupCancel()
	if err := Cleanup(cleanupCtx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: runID}); err != nil {
		t.Logf("Cleanup returned error: %v", err)
	}
}

func waitForCompletedStatus(ctx context.Context, repoRoot, stateDir, runID string) (string, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastStatus string
	var lastErr error
	for {
		status, err := Status(ctx, Options{RepoRoot: repoRoot, StateDir: stateDir, ID: runID})
		if err == nil {
			lastStatus = status
			lastErr = nil
			if statusIndicatesCompleted(status) {
				return status, nil
			}
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			if lastErr != nil {
				return lastStatus, fmt.Errorf("timed out waiting for completion after status error: %w", lastErr)
			}
			return lastStatus, ctx.Err()
		case <-ticker.C:
		}
	}
}

func statusIndicatesCompleted(status string) bool {
	for _, line := range strings.Split(status, "\n") {
		normalized := strings.ReplaceAll(line, "\u2502", " ")
		normalized = strings.ReplaceAll(normalized, "|", " ")
		normalized = strings.ReplaceAll(normalized, ":", " ")
		fields := strings.Fields(normalized)
		for i := 0; i+2 < len(fields); i++ {
			if fields[i] == "lab" && fields[i+1] == "Completed" && fields[i+2] == "True" {
				return true
			}
		}
	}
	return false
}
