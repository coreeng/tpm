package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	challengeCode = "HealthChecksAdded"

	deploymentName = "health-app"
	serviceName    = "health-app"
	appPort        = 8080

	readinessPath = "/actuator/health/readiness"
	livenessPath  = "/actuator/health/liveness"

	challengeConditionType = "IAC_HealthChecksAdded"
	completedConditionType = "IA_Completed"
)

// goal couples a goal code with the condition type the runtime reads for it.
type goal struct {
	code          string
	conditionType string
}

var (
	goalAppReady  = goal{"AppDeployedAndReady", "IAG_HealthChecksAdded_AppDeployedAndReady"}
	goalReadiness = goal{"ReadinessProbeWired", "IAG_HealthChecksAdded_ReadinessProbeWired"}
	goalLiveness  = goal{"LivenessProbeWired", "IAG_HealthChecksAdded_LivenessProbeWired"}
)

// result is the outcome of evaluating one goal.
type result struct {
	status  corev1.ConditionStatus
	reason  string
	message string
}

func ok(message string) result {
	return result{corev1.ConditionTrue, "Satisfied", message}
}
func todo(reason, message string) result {
	return result{corev1.ConditionFalse, reason, message}
}
func unknown(reason, message string) result {
	return result{corev1.ConditionUnknown, reason, message}
}

func main() {
	ctx := context.Background()
	workspaceNS := mustEnv("ASSESSMENT_WORKSPACE_NS")
	validatorPodName := mustEnv("POD_NAME")
	validatorPodNamespace := mustEnv("POD_NAMESPACE")
	interval := checkInterval()

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("load cluster config: %v", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("create kubernetes client: %v", err)
	}
	httpClient := &http.Client{Timeout: 3 * time.Second}

	for {
		if err := validateOnce(ctx, client, httpClient, workspaceNS, validatorPodNamespace, validatorPodName); err != nil {
			log.Printf("validation failed: %v", err)
		}
		time.Sleep(interval)
	}
}

func validateOnce(ctx context.Context, client kubernetes.Interface, httpClient *http.Client, workspaceNS, validatorPodNamespace, validatorPodName string) error {
	deployment, getErr := client.AppsV1().Deployments(workspaceNS).Get(ctx, deploymentName, metav1.GetOptions{})

	appReady := evalAppReady(deployment, getErr, workspaceNS)
	readiness := evalProbe(deployment, getErr, httpClient, workspaceNS, "readiness", readinessPath)
	liveness := evalProbe(deployment, getErr, httpClient, workspaceNS, "liveness", livenessPath)

	// Challenge (and overall completion) is satisfied only when every goal is True.
	challenge := aggregate(appReady, readiness, liveness)

	conditions := []corev1.PodCondition{
		condition(goalAppReady.conditionType, appReady.status, appReady.reason, appReady.message),
		condition(goalReadiness.conditionType, readiness.status, readiness.reason, readiness.message),
		condition(goalLiveness.conditionType, liveness.status, liveness.reason, liveness.message),
		condition(challengeConditionType, challenge.status, challenge.reason, challenge.message),
		condition(completedConditionType, challenge.status, challenge.reason, challenge.message),
	}
	return patchPodConditions(ctx, client, validatorPodNamespace, validatorPodName, conditions)
}

func evalAppReady(deployment *appsv1.Deployment, getErr error, ns string) result {
	if kerrors.IsNotFound(getErr) {
		return todo("DeploymentMissing", fmt.Sprintf("Create Deployment %s in namespace %s", deploymentName, ns))
	}
	if getErr != nil {
		return unknown("DeploymentCheckFailed", fmt.Sprintf("Could not check Deployment %s: %v", deploymentName, getErr))
	}
	if deployment.Status.ReadyReplicas < 1 {
		return todo("NotReady", fmt.Sprintf("Deployment %s has no Ready replicas yet", deploymentName))
	}
	return ok(fmt.Sprintf("Deployment %s has a Ready replica", deploymentName))
}

// evalProbe checks that the named probe (readiness/liveness) is both configured on the
// Deployment and actually serving UP from the in-cluster Service.
func evalProbe(deployment *appsv1.Deployment, getErr error, httpClient *http.Client, ns, kind, path string) result {
	if kerrors.IsNotFound(getErr) {
		return todo("DeploymentMissing", fmt.Sprintf("Create Deployment %s in namespace %s", deploymentName, ns))
	}
	if getErr != nil {
		return unknown("DeploymentCheckFailed", fmt.Sprintf("Could not check Deployment %s: %v", deploymentName, getErr))
	}
	if !hasHTTPProbe(deployment, kind, path) {
		return todo("ProbeNotConfigured", fmt.Sprintf("Add a %s probe with HTTP GET %s on port %d to Deployment %s", kind, path, appPort, deploymentName))
	}
	healthy, detail := endpointUp(httpClient, ns, path)
	if !healthy {
		return todo("EndpointNotHealthy", fmt.Sprintf("%s endpoint %s is not serving UP yet: %s", kind, path, detail))
	}
	return ok(fmt.Sprintf("%s probe is configured and %s reports UP", kind, path))
}

// hasHTTPProbe reports whether any container has the given probe configured as an HTTP GET
// to the expected path. kind is "readiness" or "liveness".
func hasHTTPProbe(deployment *appsv1.Deployment, kind, path string) bool {
	for _, c := range deployment.Spec.Template.Spec.Containers {
		probe := c.ReadinessProbe
		if kind == "liveness" {
			probe = c.LivenessProbe
		}
		if probe == nil || probe.HTTPGet == nil {
			continue
		}
		if probe.HTTPGet.Path == path {
			return true
		}
	}
	return false
}

// endpointUp performs an in-cluster GET against the workspace Service and checks for a
// 200 response whose body reports status UP.
func endpointUp(httpClient *http.Client, ns, path string) (bool, string) {
	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d%s", serviceName, ns, appPort, path)
	resp, err := httpClient.Get(url)
	if err != nil {
		return false, fmt.Sprintf("request error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), `"status":"UP"`) {
		return false, "response did not report status UP"
	}
	return true, "UP"
}

func aggregate(results ...result) result {
	for _, r := range results {
		if r.status != corev1.ConditionTrue {
			return result{r.status, "ChallengeIncomplete", "Complete all goals to finish the challenge"}
		}
	}
	return ok("All health-check goals satisfied")
}

func condition(conditionType string, status corev1.ConditionStatus, reason, message string) corev1.PodCondition {
	return corev1.PodCondition{
		Type:               corev1.PodConditionType(conditionType),
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}
}

func patchPodConditions(ctx context.Context, client kubernetes.Interface, namespace, name string, conditions []corev1.PodCondition) error {
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get validator pod: %w", err)
	}
	pod.Status.Conditions = upsertConditions(pod.Status.Conditions, conditions)
	_, err = client.CoreV1().Pods(namespace).Patch(ctx, name, types.StrategicMergePatchType, []byte(statusPatch(pod.Status.Conditions)), metav1.PatchOptions{}, "status")
	if err != nil {
		return fmt.Errorf("patch validator pod status: %w", err)
	}
	return nil
}

func upsertConditions(existing, updates []corev1.PodCondition) []corev1.PodCondition {
	merged := append([]corev1.PodCondition(nil), existing...)
	for _, update := range updates {
		matched := false
		for i := range merged {
			if merged[i].Type == update.Type {
				merged[i] = update
				matched = true
				break
			}
		}
		if !matched {
			merged = append(merged, update)
		}
	}
	return merged
}

func statusPatch(conditions []corev1.PodCondition) string {
	patch := `{"status":{"conditions":[`
	for i, condition := range conditions {
		if i > 0 {
			patch += ","
		}
		patch += fmt.Sprintf(`{"type":%q,"status":%q,"lastTransitionTime":%q,"reason":%q,"message":%q}`, condition.Type, condition.Status, condition.LastTransitionTime.Format(time.RFC3339), condition.Reason, condition.Message)
	}
	return patch + `]}}`
}

func checkInterval() time.Duration {
	value := os.Getenv("VALIDATOR_CHECK_INTERVAL")
	if value == "" {
		return 5 * time.Second
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid VALIDATOR_CHECK_INTERVAL %q: %v; using 5s", value, err)
		return 5 * time.Second
	}
	if parsed <= 0 {
		log.Printf("invalid VALIDATOR_CHECK_INTERVAL %q: must be greater than zero; using 5s", value)
		return 5 * time.Second
	}
	return parsed
}

func mustEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}
