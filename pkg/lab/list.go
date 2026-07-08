package lab

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

func List(ctx context.Context, opts Options) (string, error) {
	runner := opts.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	if err := requireSafeCleanupContext(ctx, runner, opts.AllowNonKind); err != nil {
		return "", err
	}

	output, err := runner.Output(ctx, "kubectl", "get", "namespaces", "-l", LabelManagedBy+"="+LabelManagedByValue, "-o", "json")
	if err != nil {
		return "", fmt.Errorf("list lab namespaces: %w", err)
	}

	var namespaces kubectlNamespaces
	if err := json.Unmarshal(output, &namespaces); err != nil {
		return "", fmt.Errorf("parse lab namespaces: %w", err)
	}
	if len(namespaces.Items) == 0 {
		return "No active labs found\n", nil
	}

	labs := map[string]*listedLab{}
	for _, namespace := range namespaces.Items {
		labels := namespace.Metadata.Labels
		runID := labels[LabelLabRunID]
		if runID == "" {
			continue
		}
		lab := labs[runID]
		if lab == nil {
			lab = &listedLab{RunID: runID, LabCode: labels[LabelLabCode], SystemNamespace: "-", WorkspaceNamespace: "-"}
			labs[runID] = lab
		}
		if lab.LabCode == "" {
			lab.LabCode = labels[LabelLabCode]
		}
		switch labels[LabelNamespaceRole] {
		case NamespaceRoleSystem:
			lab.SystemNamespace = namespace.Metadata.Name
		case NamespaceRoleWork:
			lab.WorkspaceNamespace = namespace.Metadata.Name
		}
	}

	rows := make([]listedLab, 0, len(labs))
	for _, lab := range labs {
		if lab.LabCode == "" {
			lab.LabCode = "-"
		}
		rows = append(rows, *lab)
	}
	if len(rows) == 0 {
		return "No active labs found\n", nil
	}
	opts.resolveStateDir()
	for i := range rows {
		// A missing, unreadable, or corrupt state file just means we cannot show
		// a created time for this lab; skip it rather than failing the whole list.
		state, err := LoadState(filepath.Join(opts.StateDir, rows[i].RunID+".yaml"))
		if err != nil {
			continue
		}
		rows[i].CreatedAt = state.CreatedAt
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CreatedAt.IsZero() != rows[j].CreatedAt.IsZero() {
			return !rows[i].CreatedAt.IsZero()
		}
		if !rows[i].CreatedAt.Equal(rows[j].CreatedAt) {
			return rows[i].CreatedAt.After(rows[j].CreatedAt)
		}
		return rows[i].RunID < rows[j].RunID
	})

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "RUN ID\tLAB CODE\tCREATED\tSYSTEM NAMESPACE\tWORKSPACE NAMESPACE"); err != nil {
		return "", fmt.Errorf("render lab list: %w", err)
	}
	for _, lab := range rows {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", lab.RunID, lab.LabCode, formatCreatedAt(lab.CreatedAt), lab.SystemNamespace, lab.WorkspaceNamespace); err != nil {
			return "", fmt.Errorf("render lab list: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("render lab list: %w", err)
	}
	return b.String(), nil
}

type kubectlNamespaces struct {
	Items []kubectlNamespace `json:"items"`
}

type kubectlNamespace struct {
	Metadata struct {
		Name   string            `json:"name"`
		Labels map[string]string `json:"labels"`
	} `json:"metadata"`
}

type listedLab struct {
	RunID              string
	LabCode            string
	CreatedAt          time.Time
	SystemNamespace    string
	WorkspaceNamespace string
}

func formatCreatedAt(createdAt time.Time) string {
	if createdAt.IsZero() {
		return "-"
	}
	return createdAt.UTC().Format(time.RFC3339)
}
