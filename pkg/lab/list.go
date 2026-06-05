package lab

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
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
	sort.Slice(rows, func(i, j int) bool { return rows[i].RunID < rows[j].RunID })

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RUN ID\tLAB CODE\tSYSTEM NAMESPACE\tWORKSPACE NAMESPACE")
	for _, lab := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", lab.RunID, lab.LabCode, lab.SystemNamespace, lab.WorkspaceNamespace)
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
	SystemNamespace    string
	WorkspaceNamespace string
}
