package lab

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func Status(ctx context.Context, opts Options) (string, error) {
	runner := opts.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	state, _, err := resolveState(opts)
	if err != nil {
		return "", err
	}

	pods, err := validatorPods(ctx, runner, *state)
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return renderFailedProvisioningStatus(state), nil
	}
	if len(pods.Items) != 1 {
		return "", fmt.Errorf("expected one lab validator pod in namespace %q, found %d", state.SystemNamespace, len(pods.Items))
	}
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].Metadata.Name < pods.Items[j].Metadata.Name
	})

	var b strings.Builder
	fmt.Fprintf(&b, "Lab %s status\n", state.RunID)
	for _, pod := range pods.Items {
		fmt.Fprintf(&b, "Pod %s\n", pod.Metadata.Name)
		var labConditions []kubectlPodCondition
		var podConditions []kubectlPodCondition
		for _, condition := range pod.Status.Conditions {
			if isLabCondition(condition.Type) {
				labConditions = append(labConditions, condition)
				continue
			}
			podConditions = append(podConditions, condition)
		}
		if err := renderLabProgressTable(&b, labConditions); err != nil {
			return "", err
		}
		renderPodConditions(&b, podConditions)
	}
	return b.String(), nil
}

type ProgressCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

func ProgressConditionsForState(ctx context.Context, runner Runner, state RunState) ([]ProgressCondition, error) {
	if runner == nil {
		runner = ExecRunner{}
	}
	pods, err := validatorPods(ctx, runner, state)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, nil
	}
	if len(pods.Items) != 1 {
		return nil, fmt.Errorf("expected one lab validator pod in namespace %q, found %d", state.SystemNamespace, len(pods.Items))
	}

	conditions := make([]ProgressCondition, 0, len(pods.Items[0].Status.Conditions))
	for _, condition := range pods.Items[0].Status.Conditions {
		if !isLabCondition(condition.Type) {
			continue
		}
		conditions = append(conditions, ProgressCondition(condition))
	}
	sort.Slice(conditions, func(i, j int) bool {
		return conditionSortKey(conditions[i].Type) < conditionSortKey(conditions[j].Type)
	})
	return conditions, nil
}

func validatorPods(ctx context.Context, runner Runner, state RunState) (kubectlPods, error) {
	output, err := runner.Output(ctx, "kubectl", "get", "pods", "-n", state.SystemNamespace, "-l", "app.kubernetes.io/component=validator", "-o", "json")
	if err != nil {
		return kubectlPods{}, fmt.Errorf("get lab validator pods: %w", err)
	}

	var pods kubectlPods
	if err := json.Unmarshal(output, &pods); err != nil {
		return kubectlPods{}, fmt.Errorf("parse lab validator pods: %w", err)
	}
	return pods, nil
}

func renderLabProgressTable(b *strings.Builder, conditions []kubectlPodCondition) error {
	if len(conditions) == 0 {
		return nil
	}
	b.WriteString("Lab Progress")
	b.WriteByte('\n')
	sort.Slice(conditions, func(i, j int) bool {
		return conditionSortKey(conditions[i].Type) < conditionSortKey(conditions[j].Type)
	})
	table := tablewriter.NewWriter(b)
	table.Header("TYPE", "NAME", "STATUS", "REASON", "MESSAGE")
	for _, condition := range conditions {
		conditionType, name := splitConditionLabel(renderConditionLabel(condition.Type))
		if err := table.Append([]string{conditionType, name, condition.Status, condition.Reason, condition.Message}); err != nil {
			return fmt.Errorf("render lab progress table row: %w", err)
		}
	}
	if err := table.Render(); err != nil {
		return fmt.Errorf("render lab progress table: %w", err)
	}
	return nil
}

func renderPodConditions(b *strings.Builder, conditions []kubectlPodCondition) {
	if len(conditions) == 0 {
		return
	}
	b.WriteString("Validator Pod Conditions")
	b.WriteByte('\n')
	sort.Slice(conditions, func(i, j int) bool {
		return conditionSortKey(conditions[i].Type) < conditionSortKey(conditions[j].Type)
	})
	for _, condition := range conditions {
		label := renderConditionLabel(condition.Type)
		fmt.Fprintf(b, "%s: %s", label, condition.Status)
		if condition.Reason != "" {
			fmt.Fprintf(b, " (%s)", condition.Reason)
		}
		if condition.Message != "" {
			fmt.Fprintf(b, " %s", condition.Message)
		}
		b.WriteByte('\n')
	}
}

func splitConditionLabel(label string) (string, string) {
	conditionType, name, ok := strings.Cut(label, " ")
	if !ok {
		return label, ""
	}
	return conditionType, name
}

func isLabCondition(conditionType string) bool {
	return strings.HasPrefix(conditionType, "IA_") || strings.HasPrefix(conditionType, "IAC_") || strings.HasPrefix(conditionType, "IAG_")
}

func renderFailedProvisioningStatus(state *RunState) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Lab %s status\n", state.RunID)
	b.WriteString("Provisioning: Failed\n")
	fmt.Fprintf(&b, "No validator pod found in system namespace %s.\n", state.SystemNamespace)
	b.WriteString("The lab runtime chart likely failed to install or was uninstalled before the validator started.\n")
	b.WriteString("Check Helm with:\n")
	fmt.Fprintf(&b, "helm status %s -n %s\n", state.HelmReleaseName, state.SystemNamespace)
	return b.String()
}

type kubectlPods struct {
	Items []kubectlPod `json:"items"`
}

type kubectlPod struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Status struct {
		Conditions []kubectlPodCondition `json:"conditions"`
	} `json:"status"`
}

type kubectlPodCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func conditionSortKey(conditionType string) string {
	switch {
	case conditionType == "IA_Ready":
		return "0_" + conditionType
	case strings.HasPrefix(conditionType, "IAC_"):
		return "1_" + conditionType
	case strings.HasPrefix(conditionType, "IAG_"):
		return "2_" + conditionType
	case conditionType == "IA_Completed":
		return "3_" + conditionType
	case strings.HasPrefix(conditionType, "IA_"):
		return "4_" + conditionType
	default:
		return "5_" + conditionType
	}
}

func renderConditionLabel(conditionType string) string {
	if name, ok := strings.CutPrefix(conditionType, "IA_"); ok {
		return "lab " + name
	}
	if name, ok := strings.CutPrefix(conditionType, "IAC_"); ok {
		return "challenge " + name
	}
	if name, ok := strings.CutPrefix(conditionType, "IAG_"); ok {
		challenge, goal, ok := strings.Cut(name, "_")
		if ok {
			return "goal " + challenge + "/" + goal
		}
		return "goal " + name
	}
	return conditionType
}
