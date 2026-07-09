package lab

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ScaffoldOptions struct {
	Name string
}

type ModuleBackedScaffoldOptions struct {
	Chapter          string
	Name             string
	ArtifactRegistry string
}

func ScaffoldStandalone(dir string, opts ScaffoldOptions) error {
	if err := validateLabName(opts.Name); err != nil {
		return err
	}
	if err := ensureEmptyTarget(dir); err != nil {
		return err
	}

	files := runtimeFiles()
	files["lab.yaml"] = labYAML(opts.Name)

	return writeFiles(dir, files)
}

func ScaffoldModuleBacked(dir string, opts ModuleBackedScaffoldOptions) error {
	if err := validateLabName(opts.Chapter); err != nil {
		return fmt.Errorf("chapter: %w", err)
	}
	if err := validateLabName(opts.Name); err != nil {
		return err
	}
	artifactRegistry := strings.TrimSpace(opts.ArtifactRegistry)
	if artifactRegistry == "" {
		artifactRegistry = DefaultArtifactRegistry
	}

	metadataDir := filepath.Join(dir, "module", opts.Chapter, "assessments", opts.Name)
	runtimeDir := filepath.Join(dir, "assessments", opts.Chapter, opts.Name)
	if err := ensureEmptyTarget(metadataDir); err != nil {
		return err
	}
	if err := ensureEmptyTarget(runtimeDir); err != nil {
		return err
	}

	metadataFiles := map[string]string{
		"assessment.yaml":                            moduleBackedAssessmentYAML(opts.Name, artifactRegistry),
		"description.md":                             moduleBackedAssessmentDescription,
		"01-deploy-pod-from-image/challenge.yaml":    moduleBackedChallengeYAML,
		"01-deploy-pod-from-image/description.md":    moduleBackedChallengeDescription,
		"01-deploy-pod-from-image/successMessage.md": moduleBackedChallengeSuccessMessage,
	}
	if err := writeFiles(metadataDir, metadataFiles); err != nil {
		return err
	}
	return writeFiles(runtimeDir, runtimeFiles())
}

func runtimeFiles() map[string]string {
	return map[string]string{
		"starter-content/README.md":  starterREADME,
		"starter-content/Makefile":   labMakefile,
		"starter-content/Dockerfile": labAppDockerfile,
		"starter-content/go.mod":     labAppGoMod,
		"starter-content/main.go":    labAppMain,
		"starter-content/pod.yaml":   labPodYAML,
		"solution/README.md":         solutionREADME,
		"solution/Makefile":          labMakefile,
		"solution/Dockerfile":        labAppDockerfile,
		"solution/go.mod":            labAppGoMod,
		"solution/main.go":           labAppMain,
		"solution/pod.yaml":          labPodYAML,
		"validator/AGENTS.md":        validatorAgents,
		"validator/Dockerfile":       validatorDockerfile,
		"validator/go.mod":           validatorGoMod,
		"validator/main.go":          validatorMain,
	}
}

func writeFiles(dir string, files map[string]string) error {

	for name, contents := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			return fmt.Errorf("create parent for %s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

func validateLabName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("lab name is required")
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("lab name %q must not start or end with '-'", name)
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			continue
		}
		return fmt.Errorf("lab name %q must contain only lowercase letters, numbers, and '-'", name)
	}
	return nil
}

func ensureEmptyTarget(dir string) error {
	entries, err := os.ReadDir(dir)
	if err == nil {
		if len(entries) > 0 {
			return fmt.Errorf("target directory %s is not empty", dir)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("read target directory %s: %w", dir, err)
}

func labYAML(name string) string {
	return fmt.Sprintf(`title: Pod Image Lab
code: %s
timeLimit: 30m
challenges:
  - code: DeployPodFromImage
    title: Deploy a Pod from a built image
    goals:
      - code: PodUsesBuiltImage
        title: Deploy podinfo from your image
`, name)
}

func moduleBackedAssessmentYAML(name, artifactRegistry string) string {
	return fmt.Sprintf(`code: %s
title: Pod Image Lab
timeLimit: 30m
starterImageUri: oci://%s/pod-image-lab-starter
validatorImageUri: oci://%s/pod-image-lab-validator
imageVersion: 0.0.1
`, name, artifactRegistry, artifactRegistry)
}

const moduleBackedChallengeYAML = `code: DeployPodFromImage
title: Deploy a Pod from a built image
estimatedDuration: 30m
goals:
  - code: PodUsesBuiltImage
    title: Deploy podinfo from your image
    description: Complete the lab by building the podinfo image and deploying a Pod that uses it.
`

const moduleBackedAssessmentDescription = `# Pod Image Lab

Complete this lab by building the podinfo image and deploying a Kubernetes Pod from it.
`

const moduleBackedChallengeDescription = `# Deploy a Pod from a built image

Build the podinfo container image, make it available to your cluster, and deploy the Pod manifest.
`

const moduleBackedChallengeSuccessMessage = `Pod image lab complete.
`

const starterREADME = `# Pod Image Lab

Build the podinfo container image and deploy a Pod that uses it.

For a local kind cluster, build and load the image:

` + "```sh" + `
make build
make kind-load
make deploy WORKSPACE_NAMESPACE=<workspace-namespace>
` + "```" + `

For a remote cluster, push the image to a registry you can pull from:

` + "```sh" + `
make build IMAGE=<registry>/podinfo:local
make push REGISTRY=<registry>
make deploy WORKSPACE_NAMESPACE=<workspace-namespace> IMAGE=<registry>/podinfo:local
` + "```" + `
`

const solutionREADME = `# Pod Image Lab Solution

Build and deploy the sample podinfo solution.

For kind:

` + "```sh" + `
make build
make kind-load
make deploy WORKSPACE_NAMESPACE=<workspace-namespace>
` + "```" + `

For a remote cluster:

` + "```sh" + `
make build IMAGE=<registry>/podinfo:local
make push REGISTRY=<registry>
make deploy WORKSPACE_NAMESPACE=<workspace-namespace> IMAGE=<registry>/podinfo:local
` + "```" + `
`

const labMakefile = `IMAGE_NAME ?= podinfo
IMAGE_TAG ?= local
IMAGE ?= $(IMAGE_NAME):$(IMAGE_TAG)
KIND_CLUSTER ?= kind
REGISTRY ?=
WORKSPACE_NAMESPACE ?=

.PHONY: build kind-load push deploy
build:
	docker build -t "$(IMAGE)" .

kind-load: build
	kind load docker-image "$(IMAGE)" --name "$(KIND_CLUSTER)"

push: build
	@test -n "$(REGISTRY)" || (echo "REGISTRY is required" >&2 && exit 1)
	docker tag "$(IMAGE)" "$(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"
	docker push "$(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"

deploy:
	@test -n "$(WORKSPACE_NAMESPACE)" || (echo "WORKSPACE_NAMESPACE is required" >&2 && exit 1)
	sed "s|image: podinfo:local|image: $(IMAGE)|" pod.yaml | kubectl apply -n "$(WORKSPACE_NAMESPACE)" -f -
`

const labAppDockerfile = `FROM golang:1.24 AS builder
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /podinfo ./main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /podinfo /podinfo
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/podinfo"]
`

const labAppGoMod = `module podinfo

go 1.24
`

const labAppMain = `package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "podinfo is running")
	})

	log.Println("podinfo listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
`

const labPodYAML = `apiVersion: v1
kind: Pod
metadata:
  name: podinfo
  labels:
    app: podinfo
spec:
  securityContext:
    runAsNonRoot: true
    seccompProfile:
      type: RuntimeDefault
  containers:
    - name: podinfo
      image: podinfo:local
      securityContext:
        allowPrivilegeEscalation: false
        capabilities:
          drop:
            - ALL
      ports:
        - containerPort: 8080
`

const validatorAgents = `# Validator Agent Notes

Challenge and goal codes are part of the public progress contract. Keep generated condition types aligned with the lab YAML codes.

Validators must publish progress by patching Pod conditions on their own validator Pod. In other words, validator code must patch Pod conditions on the validator Pod rather than only logging or updating local state. The lab runtime reads these Pod conditions to display learner progress.

Learner workspaces are sandboxed namespaces. Starter and solution manifests should assume:

- Pods, Deployments, ReplicaSets, Services, ConfigMaps, Secrets, PersistentVolumeClaims, PodTemplates, and ServiceAccounts are allowed.
- Cluster-scoped resources, RBAC resources, DaemonSets, StatefulSets, Jobs, CronJobs, HPAs, PDBs, Ingresses, NetworkPolicies, and Gateway API resources are blocked.
- LoadBalancer and NodePort Services are blocked by quota. Prefer ClusterIP Services.
- Pod manifests must satisfy restricted Pod Security: run as non-root, set allowPrivilegeEscalation to false, use seccompProfile RuntimeDefault, and drop all capabilities.
- Workspace egress is restricted. Do not depend on external internet access from learner Pods.

Condition prefixes:

- IA_Completed reports overall lab completion.
- IAC_<ChallengeCode> reports challenge progress.
- IAG_<ChallengeCode>_<GoalCode> reports goal progress.

Condition statuses:

- False means the learner still needs to act.
- Unknown means the validator cannot check the cluster state yet.
- True means the matching lab requirement is complete.
`

const validatorDockerfile = `FROM golang:1.24 AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /validator ./main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /validator /validator
USER 65532:65532
ENTRYPOINT ["/validator"]
`

const validatorGoMod = `module pod-image-lab-validator

go 1.24

require (
	k8s.io/api v0.33.0
	k8s.io/apimachinery v0.33.0
	k8s.io/client-go v0.33.0
)
`

const validatorMain = `package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	challengeCode = "DeployPodFromImage"
	goalCode = "PodUsesBuiltImage"
	podName = "podinfo"
	expectedImageName = "podinfo"
	expectedImageTag = "local"
	goalConditionType = "IAG_DeployPodFromImage_PodUsesBuiltImage"
	challengeConditionType = "IAC_DeployPodFromImage"
	completedConditionType = "IA_Completed"
)

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

	for {
		if err := validateOnce(ctx, client, workspaceNS, validatorPodNamespace, validatorPodName); err != nil {
			log.Printf("validation failed: %v", err)
		}
		time.Sleep(interval)
	}
}

func validateOnce(ctx context.Context, client kubernetes.Interface, workspaceNS, validatorPodNamespace, validatorPodName string) error {
	pod, err := client.CoreV1().Pods(workspaceNS).Get(ctx, podName, metav1.GetOptions{})
	status := corev1.ConditionFalse
	reason := "PodMissing"
	message := fmt.Sprintf("Create Pod %s in namespace %s", podName, workspaceNS)
	if kerrors.IsNotFound(err) {
		log.Printf("pod check: %v", err)
	} else if err != nil {
		status = corev1.ConditionUnknown
		reason = "PodCheckFailed"
		message = fmt.Sprintf("Could not check Pod %s in namespace %s", podName, workspaceNS)
		log.Printf("pod check failed: %v", err)
	} else if !podUsesExpectedImage(pod) {
		reason = "PodImageMismatch"
		message = fmt.Sprintf("Pod %s must use image %s:%s", podName, expectedImageName, expectedImageTag)
	} else if !podReady(pod) {
		reason = "PodNotReady"
		message = fmt.Sprintf("Pod %s must be Ready", podName)
	} else {
		status = corev1.ConditionTrue
		reason = "PodReady"
		message = fmt.Sprintf("Pod %s uses the expected image and is Ready", podName)
	}

	conditions := []corev1.PodCondition{
		condition(goalConditionType, status, reason, message),
		condition(challengeConditionType, status, reason, message),
		condition(completedConditionType, status, reason, message),
	}
	return patchPodConditions(ctx, client, validatorPodNamespace, validatorPodName, conditions)
}

func podUsesExpectedImage(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if imageNameMatchesExpected(container.Image) {
			return true
		}
	}
	return false
}

func imageNameMatchesExpected(image string) bool {
	withoutDigest := strings.SplitN(image, "@", 2)[0]
	lastSlash := strings.LastIndex(withoutDigest, "/")
	nameTag := withoutDigest[lastSlash+1:]
	name, tag := splitImageNameTag(nameTag)
	return name == expectedImageName && tag == expectedImageTag
}

func splitImageNameTag(nameTag string) (string, string) {
	colon := strings.LastIndex(nameTag, ":")
	if colon == -1 {
		return nameTag, ""
	}
	return nameTag[:colon], nameTag[colon+1:]
}

func podReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
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
	patch := ` + "`" + `{"status":{"conditions":[` + "`" + `
	for i, condition := range conditions {
		if i > 0 {
			patch += ","
		}
		patch += fmt.Sprintf(` + "`" + `{"type":%q,"status":%q,"lastTransitionTime":%q,"reason":%q,"message":%q}` + "`" + `, condition.Type, condition.Status, condition.LastTransitionTime.Format(time.RFC3339), condition.Reason, condition.Message)
	}
	return patch + ` + "`" + `]}}` + "`" + `
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
`
