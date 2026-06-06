package lab

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestScaffoldStandaloneLab(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "config-map-lab")

	if err := ScaffoldStandalone(dir, ScaffoldOptions{Name: "config-map-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}

	for _, name := range []string{
		"lab.yaml",
		"starter-content/README.md",
		"starter-content/Makefile",
		"starter-content/Dockerfile",
		"starter-content/go.mod",
		"starter-content/main.go",
		"starter-content/pod.yaml",
		"solution/README.md",
		"solution/Makefile",
		"solution/Dockerfile",
		"solution/go.mod",
		"solution/main.go",
		"solution/pod.yaml",
		"validator/AGENTS.md",
		"validator/Dockerfile",
		"validator/go.mod",
		"validator/main.go",
	} {
		assertFileExists(t, filepath.Join(dir, name))
	}

	labYAML := readFile(t, filepath.Join(dir, "lab.yaml"))
	var metadata Lab
	if err := yaml.Unmarshal([]byte(labYAML), &metadata); err != nil {
		t.Fatalf("parse lab.yaml: %v", err)
	}
	if metadata.Code != "config-map-lab" {
		t.Fatalf("lab code = %q, want config-map-lab", metadata.Code)
	}
	if len(metadata.Challenges) != 1 {
		t.Fatalf("len(challenges) = %d, want 1", len(metadata.Challenges))
	}
	if metadata.Challenges[0].Code != "DeployPodFromImage" {
		t.Fatalf("challenge code = %q, want DeployPodFromImage", metadata.Challenges[0].Code)
	}
	if len(metadata.Challenges[0].Goals) != 1 {
		t.Fatalf("len(goals) = %d, want 1", len(metadata.Challenges[0].Goals))
	}
	if metadata.Challenges[0].Goals[0].Code != "PodUsesBuiltImage" {
		t.Fatalf("goal code = %q, want PodUsesBuiltImage", metadata.Challenges[0].Goals[0].Code)
	}

	validatorMain := readFile(t, filepath.Join(dir, "validator/main.go"))
	for _, want := range []string{
		`challengeCode`,
		`"DeployPodFromImage"`,
		`goalCode`,
		`"PodUsesBuiltImage"`,
		`podName`,
		`"podinfo"`,
		`expectedImageName = "podinfo"`,
		`goalConditionType`,
		`"IAG_DeployPodFromImage_PodUsesBuiltImage"`,
		`completedConditionType`,
		`"IA_Completed"`,
		`kerrors.IsNotFound(err)`,
		`corev1.ConditionUnknown`,
		`ASSESSMENT_WORKSPACE_NS`,
		`POD_NAME`,
		`POD_NAMESPACE`,
	} {
		if !strings.Contains(validatorMain, want) {
			t.Fatalf("validator/main.go does not contain %q", want)
		}
	}
	for _, reject := range []string{`goalPrefix+goalCode`, `assessmentPrefix+"Complete"`} {
		if strings.Contains(validatorMain, reject) {
			t.Fatalf("validator/main.go contains obsolete condition expression %q", reject)
		}
	}
	for _, want := range []string{
		`validatorPodName := mustEnv("POD_NAME")`,
		`validatorPodNamespace := mustEnv("POD_NAMESPACE")`,
		`interval := checkInterval()`,
		`client.CoreV1().Pods(workspaceNS).Get(ctx, podName, metav1.GetOptions{})`,
		`patchPodConditions(ctx, client, validatorPodNamespace, validatorPodName, conditions)`,
		`imageNameMatchesExpected(container.Image)`,
		`expectedImageTag = "local"`,
		`withoutDigest := strings.SplitN(image, "@", 2)[0]`,
		`lastSlash := strings.LastIndex(withoutDigest, "/")`,
		`name, tag := splitImageNameTag(nameTag)`,
		`return name == expectedImageName && tag == expectedImageTag`,
		`func splitImageNameTag(nameTag string) (string, string)`,
		`func checkInterval() time.Duration`,
		`value := os.Getenv("VALIDATOR_CHECK_INTERVAL")`,
		`parsed, err := time.ParseDuration(value)`,
		`func podReady(pod *corev1.Pod) bool`,
	} {
		if !strings.Contains(validatorMain, want) {
			t.Fatalf("validator/main.go does not contain %q", want)
		}
	}
	for _, reject := range []string{`strings.Contains(container.Image, expectedImageName)`, `time.Sleep(5 * time.Second)`, `func podReadyOrRunning`, `pod.Status.Phase == corev1.PodRunning`, `must be Running or Ready`} {
		if strings.Contains(validatorMain, reject) {
			t.Fatalf("validator/main.go contains obsolete validation expression %q", reject)
		}
	}

	dockerfile := readFile(t, filepath.Join(dir, "validator/Dockerfile"))
	if strings.Contains(dockerfile, "go.sum") {
		t.Fatalf("validator/Dockerfile references go.sum, but scaffold does not create validator/go.sum")
	}
	if !strings.Contains(dockerfile, "COPY go.mod ./") {
		t.Fatalf("validator/Dockerfile does not copy go.mod before downloading dependencies")
	}
	if !strings.Contains(dockerfile, "RUN go mod tidy") {
		t.Fatalf("validator/Dockerfile does not generate go.sum before building")
	}

	runGeneratedValidatorCompile(t, filepath.Join(dir, "validator"))

	solution := readFile(t, filepath.Join(dir, "solution/pod.yaml"))
	for _, want := range []string{
		"kind: Pod",
		"name: podinfo",
		"image: podinfo:local",
		"runAsNonRoot: true",
		"allowPrivilegeEscalation: false",
		"type: RuntimeDefault",
		"drop:",
		"- ALL",
	} {
		if !strings.Contains(solution, want) {
			t.Fatalf("solution/pod.yaml does not contain %q", want)
		}
	}
	starterMain := readFile(t, filepath.Join(dir, "starter-content/main.go"))
	for _, want := range []string{"package main", "podinfo", "http.ListenAndServe"} {
		if !strings.Contains(starterMain, want) {
			t.Fatalf("starter-content/main.go does not contain %q", want)
		}
	}
	for _, name := range []string{"starter-content/Dockerfile", "solution/Dockerfile"} {
		dockerfile := readFile(t, filepath.Join(dir, name))
		if !strings.Contains(dockerfile, "USER 65532:65532") {
			t.Fatalf("%s does not use a numeric non-root user", name)
		}
		if strings.Contains(dockerfile, "USER nonroot:nonroot") {
			t.Fatalf("%s uses a non-numeric user that kubelet cannot verify with runAsNonRoot", name)
		}
	}
	agents := readFile(t, filepath.Join(dir, "validator/AGENTS.md"))
	for _, want := range []string{"Challenge and goal codes", "patch Pod conditions", "validator Pod", "sandboxed namespaces", "LoadBalancer and NodePort Services are blocked", "restricted Pod Security", "Workspace egress is restricted", "IA_Completed", "IAC_<ChallengeCode>", "IAG_<ChallengeCode>_<GoalCode>", "False", "Unknown", "True"} {
		if !strings.Contains(agents, want) {
			t.Fatalf("validator/AGENTS.md does not contain %q", want)
		}
	}
	solutionREADME := readFile(t, filepath.Join(dir, "solution/README.md"))
	for _, want := range []string{"make build", "make kind-load", "make push", "make deploy", "WORKSPACE_NAMESPACE"} {
		if !strings.Contains(solutionREADME, want) {
			t.Fatalf("solution/README.md does not contain %q", want)
		}
	}

	for _, name := range []string{"starter-content/Makefile", "solution/Makefile"} {
		makefile := readFile(t, filepath.Join(dir, name))
		for _, want := range []string{"IMAGE_NAME", "IMAGE_TAG", "IMAGE", "KIND_CLUSTER", "REGISTRY", "kind-load", "push", "deploy"} {
			if !strings.Contains(makefile, want) {
				t.Fatalf("%s does not contain %q", name, want)
			}
		}
		for _, reject := range []string{"REGISTRY_USERNAME", "REGISTRY_TOKEN", "docker login"} {
			if strings.Contains(makefile, reject) {
				t.Fatalf("%s contains registry-auth-specific scaffold text %q", name, reject)
			}
		}
	}

	for _, name := range []string{
		"lab.yaml",
		"starter-content/README.md",
		"starter-content/Makefile",
		"starter-content/Dockerfile",
		"starter-content/go.mod",
		"starter-content/main.go",
		"starter-content/pod.yaml",
		"solution/README.md",
		"solution/Makefile",
		"solution/Dockerfile",
		"solution/go.mod",
		"solution/main.go",
		"solution/pod.yaml",
		"validator/AGENTS.md",
		"validator/Dockerfile",
		"validator/go.mod",
		"validator/main.go",
	} {
		contents := readFile(t, filepath.Join(dir, name))
		assertASCII(t, name, contents)
	}

	for _, name := range []string{
		"lab.yaml",
		"starter-content/README.md",
		"starter-content/Makefile",
		"starter-content/Dockerfile",
		"starter-content/go.mod",
		"starter-content/main.go",
		"starter-content/pod.yaml",
		"solution/README.md",
		"solution/Makefile",
		"solution/Dockerfile",
		"solution/go.mod",
		"solution/main.go",
		"solution/pod.yaml",
		"validator/AGENTS.md",
	} {
		contents := strings.ToLower(readFile(t, filepath.Join(dir, name)))
		if strings.Contains(contents, "assessment") {
			t.Fatalf("%s contains user-facing assessment wording", name)
		}
	}

	nonEmpty := filepath.Join(t.TempDir(), "existing")
	if err := os.MkdirAll(nonEmpty, 0755); err != nil {
		t.Fatalf("create non-empty dir: %v", err)
	}
	writeFile(t, filepath.Join(nonEmpty, "keep.txt"), "do not overwrite")
	if err := ScaffoldStandalone(nonEmpty, ScaffoldOptions{Name: "config-map-lab"}); err == nil {
		t.Fatal("ScaffoldStandalone returned nil error for non-empty target")
	}

	unsafeDir := filepath.Join(t.TempDir(), "unsafe")
	if err := ScaffoldStandalone(unsafeDir, ScaffoldOptions{Name: "bad:name"}); err == nil {
		t.Fatal("ScaffoldStandalone returned nil error for unsafe lab name")
	}
}

func TestScaffoldModuleBackedLab(t *testing.T) {
	dir := t.TempDir()

	if err := ScaffoldModuleBacked(dir, ModuleBackedScaffoldOptions{
		Chapter: "01-config-maps",
		Name:    "01-config-map-lab",
	}); err != nil {
		t.Fatalf("ScaffoldModuleBacked returned error: %v", err)
	}

	for _, name := range []string{
		"module/01-config-maps/assessments/01-config-map-lab/assessment.yaml",
		"module/01-config-maps/assessments/01-config-map-lab/description.md",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/challenge.yaml",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/description.md",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/successMessage.md",
		"assessments/01-config-maps/01-config-map-lab/starter-content/README.md",
		"assessments/01-config-maps/01-config-map-lab/starter-content/Makefile",
		"assessments/01-config-maps/01-config-map-lab/starter-content/Dockerfile",
		"assessments/01-config-maps/01-config-map-lab/starter-content/go.mod",
		"assessments/01-config-maps/01-config-map-lab/starter-content/main.go",
		"assessments/01-config-maps/01-config-map-lab/starter-content/pod.yaml",
		"assessments/01-config-maps/01-config-map-lab/solution/README.md",
		"assessments/01-config-maps/01-config-map-lab/solution/Makefile",
		"assessments/01-config-maps/01-config-map-lab/solution/Dockerfile",
		"assessments/01-config-maps/01-config-map-lab/solution/go.mod",
		"assessments/01-config-maps/01-config-map-lab/solution/main.go",
		"assessments/01-config-maps/01-config-map-lab/solution/pod.yaml",
		"assessments/01-config-maps/01-config-map-lab/validator/AGENTS.md",
		"assessments/01-config-maps/01-config-map-lab/validator/Dockerfile",
		"assessments/01-config-maps/01-config-map-lab/validator/go.mod",
		"assessments/01-config-maps/01-config-map-lab/validator/main.go",
	} {
		assertFileExists(t, filepath.Join(dir, name))
	}

	assessment := readFile(t, filepath.Join(dir, "module/01-config-maps/assessments/01-config-map-lab/assessment.yaml"))
	for _, want := range []string{
		"title: Pod Image Lab",
		"timeLimit: 30m",
		"starterImageUri: oci://localhost/pod-image-lab-starter",
		"validatorImageUri: oci://localhost/pod-image-lab-validator",
		"imageVersion: 0.0.1",
	} {
		if !strings.Contains(assessment, want) {
			t.Fatalf("assessment.yaml does not contain %q", want)
		}
	}

	challenge := readFile(t, filepath.Join(dir, "module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/challenge.yaml"))
	for _, want := range []string{
		"code: DeployPodFromImage",
		"code: PodUsesBuiltImage",
		"estimatedDuration: 30m",
	} {
		if !strings.Contains(challenge, want) {
			t.Fatalf("challenge.yaml does not contain %q", want)
		}
	}
	for _, reject := range []string{
		"\ndescription:",
		"\nsuccessMessage:",
	} {
		if strings.Contains("\n"+challenge, reject) {
			t.Fatalf("challenge.yaml contains unsupported top-level source challenge property %q", strings.TrimSpace(reject))
		}
	}

	moduleBackedTextFiles := []string{
		"module/01-config-maps/assessments/01-config-map-lab/assessment.yaml",
		"module/01-config-maps/assessments/01-config-map-lab/description.md",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/challenge.yaml",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/description.md",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/successMessage.md",
		"assessments/01-config-maps/01-config-map-lab/starter-content/README.md",
		"assessments/01-config-maps/01-config-map-lab/starter-content/Makefile",
		"assessments/01-config-maps/01-config-map-lab/starter-content/Dockerfile",
		"assessments/01-config-maps/01-config-map-lab/starter-content/go.mod",
		"assessments/01-config-maps/01-config-map-lab/starter-content/main.go",
		"assessments/01-config-maps/01-config-map-lab/starter-content/pod.yaml",
		"assessments/01-config-maps/01-config-map-lab/solution/README.md",
		"assessments/01-config-maps/01-config-map-lab/solution/Makefile",
		"assessments/01-config-maps/01-config-map-lab/solution/Dockerfile",
		"assessments/01-config-maps/01-config-map-lab/solution/go.mod",
		"assessments/01-config-maps/01-config-map-lab/solution/main.go",
		"assessments/01-config-maps/01-config-map-lab/solution/pod.yaml",
		"assessments/01-config-maps/01-config-map-lab/validator/AGENTS.md",
		"assessments/01-config-maps/01-config-map-lab/validator/Dockerfile",
		"assessments/01-config-maps/01-config-map-lab/validator/go.mod",
		"assessments/01-config-maps/01-config-map-lab/validator/main.go",
	}
	for _, name := range moduleBackedTextFiles {
		contents := readFile(t, filepath.Join(dir, name))
		assertASCII(t, name, contents)
	}

	for _, name := range []string{
		"module/01-config-maps/assessments/01-config-map-lab/assessment.yaml",
		"module/01-config-maps/assessments/01-config-map-lab/description.md",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/challenge.yaml",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/description.md",
		"module/01-config-maps/assessments/01-config-map-lab/01-deploy-pod-from-image/successMessage.md",
		"assessments/01-config-maps/01-config-map-lab/starter-content/README.md",
		"assessments/01-config-maps/01-config-map-lab/solution/README.md",
	} {
		contents := readFile(t, filepath.Join(dir, name))
		if strings.Contains(strings.ToLower(contents), "lab") == false {
			t.Fatalf("%s does not use lab terminology", name)
		}
		if strings.Contains(strings.ToLower(contents), "assessment") {
			t.Fatalf("%s contains user-facing assessment wording", name)
		}
	}

	nonEmptyRuntime := filepath.Join(t.TempDir(), "module-backed")
	writeFile(t, filepath.Join(nonEmptyRuntime, "assessments/01-config-maps/01-config-map-lab/keep.txt"), "do not overwrite")
	if err := ScaffoldModuleBacked(nonEmptyRuntime, ModuleBackedScaffoldOptions{Chapter: "01-config-maps", Name: "01-config-map-lab"}); err == nil {
		t.Fatal("ScaffoldModuleBacked returned nil error for non-empty runtime target")
	}
}

func TestScaffoldModuleBackedLabUsesCustomArtifactRegistry(t *testing.T) {
	dir := t.TempDir()

	if err := ScaffoldModuleBacked(dir, ModuleBackedScaffoldOptions{
		Chapter:          "01-pod-images",
		Name:             "01-pod-image-lab",
		ArtifactRegistry: "registry.example.com/training",
	}); err != nil {
		t.Fatalf("ScaffoldModuleBacked returned error: %v", err)
	}

	assessment := readFile(t, filepath.Join(dir, "module/01-pod-images/assessments/01-pod-image-lab/assessment.yaml"))
	for _, want := range []string{
		"starterImageUri: oci://registry.example.com/training/pod-image-lab-starter",
		"validatorImageUri: oci://registry.example.com/training/pod-image-lab-validator",
	} {
		if !strings.Contains(assessment, want) {
			t.Fatalf("assessment.yaml does not contain %q", want)
		}
	}
}

func runGeneratedValidatorCompile(t *testing.T, validatorDir string) {
	t.Helper()
	writeFile(t, filepath.Join(validatorDir, "main_test.go"), generatedValidatorBehaviorTest)

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = validatorDir
	tidy.Env = os.Environ()
	if output, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("generated validator dependencies cannot be resolved: %v\n%s", err, output)
	}

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = validatorDir
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated validator does not compile: %v\n%s", err, output)
	}
}

const generatedValidatorBehaviorTest = `package main

import (
	"testing"
	"time"
)

func TestImageNameMatchesExpected(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  bool
	}{
		{name: "local image", image: "podinfo:local", want: true},
		{name: "registry image", image: "localhost:30500/podinfo:local", want: true},
		{name: "wrong tag", image: "ghcr.io/stefanprodan/podinfo:latest", want: false},
		{name: "spoofed name", image: "evil/podinfo-spoof:local", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := imageNameMatchesExpected(tt.image); got != tt.want {
				t.Fatalf("imageNameMatchesExpected(%q) = %v, want %v", tt.image, got, tt.want)
			}
		})
	}
}

func TestSplitImageNameTag(t *testing.T) {
	nameTag := "podinfo:local"
	name, tag := splitImageNameTag(nameTag)
	if name != "podinfo" || tag != "local" {
		t.Fatalf("splitImageNameTag(%q) = %q, %q; want podinfo, local", nameTag, name, tag)
	}
}

func TestCheckInterval(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "blank defaults", value: "", want: 5 * time.Second},
		{name: "valid duration", value: "250ms", want: 250 * time.Millisecond},
		{name: "invalid defaults", value: "not-a-duration", want: 5 * time.Second},
		{name: "zero defaults", value: "0s", want: 5 * time.Second},
		{name: "negative defaults", value: "-1s", want: 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VALIDATOR_CHECK_INTERVAL", tt.value)
			if got := checkInterval(); got != tt.want {
				t.Fatalf("checkInterval() = %s, want %s", got, tt.want)
			}
		})
	}
}
`

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(contents)
}

func assertASCII(t *testing.T, name, contents string) {
	t.Helper()
	for i, r := range contents {
		if r > 127 {
			t.Fatalf("%s contains non-ASCII rune %q at byte %d", name, r, i)
		}
	}
}
