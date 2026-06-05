package lab

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRunBuildsLoadsAndInstallsChart(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}
	normalizedLabPath, err := normalizeLabPath(labPath)
	if err != nil {
		t.Fatalf("normalizeLabPath returned error: %v", err)
	}

	runner := NewFakeRunner()
	for range 3 {
		runner.QueueResponse(nil, nil)
	}
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, errors.New("not found"))
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, errors.New("not found"))
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)

	chartURI := "oci://example.com/charts/training-platform-assessment"
	stateDir := filepath.Join(repoRoot, ".state")
	state, err := Run(context.Background(), Options{
		LabPath:       labPath,
		RepoRoot:      repoRoot,
		StateDir:      stateDir,
		ID:            "abc123",
		ChartURI:      chartURI,
		ChartVersion:  "1.2.3",
		CheckInterval: 2 * time.Second,
		Runner:        runner,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	validatorTag := "localhost/tpm-lab-pod-image-lab-validator:abc123"
	release := "lab-abc123"
	commandsBeforeHelm := []Command{
		{Name: "docker", Args: []string{"version"}},
		{Name: "helm", Args: []string{"version"}},
		{Name: "kubectl", Args: []string{"version", "--client"}},
		{Name: "kubectl", Args: []string{"config", "current-context"}},
		{Name: "kind", Args: []string{"version"}},
		{Name: "docker", Args: []string{"build", "-t", validatorTag, filepath.Join(normalizedLabPath, "validator")}},
		{Name: "kind", Args: []string{"load", "docker-image", "--name", "local", validatorTag}},
		{Name: "kubectl", Args: []string{"get", "namespace", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"create", "namespace", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"label", "namespace", "lab-abc123-system", "training-platform.coreeng.io/managed-by=tpm", "training-platform.coreeng.io/lab-run-id=abc123", "training-platform.coreeng.io/lab-code=pod-image-lab", "training-platform.coreeng.io/lab-namespace-role=system", "--overwrite"}},
		{Name: "kubectl", Args: []string{"get", "namespace", "lab-abc123-workspace"}},
		{Name: "kubectl", Args: []string{"create", "namespace", "lab-abc123-workspace"}},
		{Name: "kubectl", Args: []string{"label", "namespace", "lab-abc123-workspace", "training-platform.coreeng.io/managed-by=tpm", "training-platform.coreeng.io/lab-run-id=abc123", "training-platform.coreeng.io/lab-code=pod-image-lab", "training-platform.coreeng.io/lab-namespace-role=workspace", "pod-security.kubernetes.io/enforce=restricted", "pod-security.kubernetes.io/audit=restricted", "--overwrite"}},
	}
	if len(runner.Commands) != len(commandsBeforeHelm)+2 {
		t.Fatalf("recorded %d commands, want %d: %#v", len(runner.Commands), len(commandsBeforeHelm)+2, runner.Commands)
	}
	if !reflect.DeepEqual(runner.Commands[:len(commandsBeforeHelm)], commandsBeforeHelm) {
		t.Fatalf("commands before helm = %#v, want %#v", runner.Commands[:len(commandsBeforeHelm)], commandsBeforeHelm)
	}

	showChart := runner.Commands[len(commandsBeforeHelm)]
	wantShowChart := Command{Name: "helm", Args: []string{"show", "chart", chartURI, "--version", "1.2.3"}}
	if !reflect.DeepEqual(showChart, wantShowChart) {
		t.Fatalf("helm show chart command = %#v, want %#v", showChart, wantShowChart)
	}

	helm := runner.Commands[len(runner.Commands)-1]
	if helm.Name != "helm" {
		t.Fatalf("final command name = %q, want helm", helm.Name)
	}
	wantHelmPrefix := []string{"upgrade", "--install", release, chartURI, "--version", "1.2.3", "-n", "lab-abc123-system"}
	if len(helm.Args) < len(wantHelmPrefix) || !reflect.DeepEqual(helm.Args[:len(wantHelmPrefix)], wantHelmPrefix) {
		t.Fatalf("helm args prefix = %#v, want %#v", helm.Args, wantHelmPrefix)
	}
	assertArgsContainAll(t, helm.Args,
		"assessment.instanceID=abc123",
		"assessment.workspaceNS=lab-abc123-workspace",
		"assessment.systemNS=lab-abc123-system",
		"validator.image.repository=localhost/tpm-lab-pod-image-lab-validator",
		"validator.image.tag=abc123",
		"registry.domain=localhost",
		"registry.ingress.enabled=false",
		"registry.registryPassword=local-password",
		"github.repository=local/pod-image-lab",
		"github.accessToken=local-token",
		"validator.extraEnv[0].name=VALIDATOR_CHECK_INTERVAL",
		"validator.extraEnv[0].value=2s",
	)
	assertArgsDoNotContain(t, helm.Args,
		"registry.hostnameOverride=",
		"registry.service.type=",
		"registry.service.nodePort=",
	)

	starterTarball := filepath.Join(repoRoot, ".build", "tpm", "labs", "abc123", "starter-content.tar.gz")
	assertFileExists(t, starterTarball)

	loadedState, err := LoadState(filepath.Join(stateDir, "abc123.yaml"))
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}
	if loadedState.LabPath != normalizedLabPath {
		t.Fatalf("state LabPath = %q, want %q", loadedState.LabPath, normalizedLabPath)
	}
	for name, got := range map[string]string{
		"RunID":              loadedState.RunID,
		"SystemNamespace":    loadedState.SystemNamespace,
		"WorkspaceNamespace": loadedState.WorkspaceNamespace,
		"HelmReleaseName":    loadedState.HelmReleaseName,
		"ValidatorImageTag":  loadedState.ValidatorImageTag,
		"RegistryURL":        loadedState.RegistryURL,
		"RegistryUsername":   loadedState.RegistryUsername,
		"RegistryToken":      loadedState.RegistryToken,
		"ChartURI":           loadedState.ChartURI,
		"ChartVersion":       loadedState.ChartVersion,
	} {
		want := map[string]string{
			"RunID":              "abc123",
			"SystemNamespace":    "lab-abc123-system",
			"WorkspaceNamespace": "lab-abc123-workspace",
			"HelmReleaseName":    release,
			"ValidatorImageTag":  validatorTag,
			"RegistryURL":        "localhost",
			"RegistryUsername":   "workspace",
			"RegistryToken":      "local-password",
			"ChartURI":           chartURI,
			"ChartVersion":       "1.2.3",
		}[name]
		if got != want {
			t.Fatalf("state %s = %q, want %q", name, got, want)
		}
	}
	if state == nil || state.RunID != loadedState.RunID {
		t.Fatalf("returned state = %#v, want saved state", state)
	}
}

func TestRunInstallsLocalChartDirectoryWithoutVersion(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}
	normalizedLabPath, err := normalizeLabPath(labPath)
	if err != nil {
		t.Fatalf("normalizeLabPath returned error: %v", err)
	}

	runner := NewFakeRunner()
	for range 3 {
		runner.QueueResponse(nil, nil)
	}
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, errors.New("not found"))
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, errors.New("not found"))
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)

	chartDir := filepath.Join(repoRoot, "charts", "training-platform-assessment")
	state, err := Run(context.Background(), Options{
		LabPath:  labPath,
		RepoRoot: repoRoot,
		ID:       "abc123",
		ChartDir: chartDir,
		Runner:   runner,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	validatorTag := "localhost/tpm-lab-pod-image-lab-validator:abc123"
	commandsBeforeHelm := []Command{
		{Name: "docker", Args: []string{"version"}},
		{Name: "helm", Args: []string{"version"}},
		{Name: "kubectl", Args: []string{"version", "--client"}},
		{Name: "kubectl", Args: []string{"config", "current-context"}},
		{Name: "kind", Args: []string{"version"}},
		{Name: "docker", Args: []string{"build", "-t", validatorTag, filepath.Join(normalizedLabPath, "validator")}},
		{Name: "kind", Args: []string{"load", "docker-image", "--name", "local", validatorTag}},
		{Name: "kubectl", Args: []string{"get", "namespace", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"create", "namespace", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"label", "namespace", "lab-abc123-system", "training-platform.coreeng.io/managed-by=tpm", "training-platform.coreeng.io/lab-run-id=abc123", "training-platform.coreeng.io/lab-code=pod-image-lab", "training-platform.coreeng.io/lab-namespace-role=system", "--overwrite"}},
		{Name: "kubectl", Args: []string{"get", "namespace", "lab-abc123-workspace"}},
		{Name: "kubectl", Args: []string{"create", "namespace", "lab-abc123-workspace"}},
		{Name: "kubectl", Args: []string{"label", "namespace", "lab-abc123-workspace", "training-platform.coreeng.io/managed-by=tpm", "training-platform.coreeng.io/lab-run-id=abc123", "training-platform.coreeng.io/lab-code=pod-image-lab", "training-platform.coreeng.io/lab-namespace-role=workspace", "pod-security.kubernetes.io/enforce=restricted", "pod-security.kubernetes.io/audit=restricted", "--overwrite"}},
	}
	if len(runner.Commands) != len(commandsBeforeHelm)+1 {
		t.Fatalf("recorded %d commands, want %d: %#v", len(runner.Commands), len(commandsBeforeHelm)+1, runner.Commands)
	}
	if !reflect.DeepEqual(runner.Commands[:len(commandsBeforeHelm)], commandsBeforeHelm) {
		t.Fatalf("commands before helm = %#v, want %#v", runner.Commands[:len(commandsBeforeHelm)], commandsBeforeHelm)
	}

	helm := runner.Commands[len(runner.Commands)-1]
	wantHelmPrefix := []string{"upgrade", "--install", "lab-abc123", chartDir, "-n", "lab-abc123-system"}
	if helm.Name != "helm" || len(helm.Args) < len(wantHelmPrefix) || !reflect.DeepEqual(helm.Args[:len(wantHelmPrefix)], wantHelmPrefix) {
		t.Fatalf("helm command = %#v, want prefix %#v", helm, wantHelmPrefix)
	}
	if containsArg(helm.Args, "--version") {
		t.Fatalf("local chart helm args should not contain --version: %#v", helm.Args)
	}
	if state.ChartDir != chartDir {
		t.Fatalf("state ChartDir = %q, want %q", state.ChartDir, chartDir)
	}
}

func TestRunLogsProgressSteps(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}

	runner := NewFakeRunner()
	for range 3 {
		runner.QueueResponse(nil, nil)
	}
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, errors.New("not found"))
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, errors.New("not found"))
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)

	var logs bytes.Buffer
	_, err := Run(context.Background(), Options{
		LabPath:      labPath,
		RepoRoot:     repoRoot,
		ID:           "abc123",
		ChartURI:     "oci://example.com/charts/training-platform-assessment",
		ChartVersion: "1.2.3",
		Runner:       runner,
		LogWriter:    &logs,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	wantLines := []string{
		"Starting lab pod-image-lab with run id abc123",
		"Checking Docker...",
		"Checking Helm...",
		"Checking kubectl...",
		"Using kubectl context kind-local",
		"Checking kind...",
		"Building validator image localhost/tpm-lab-pod-image-lab-validator:abc123...",
		"Loading validator image into kind cluster local...",
		"Ensuring system namespace lab-abc123-system...",
		"Ensuring workspace namespace lab-abc123-workspace...",
		"Packaging starter content...",
		"Checking lab runtime chart 1.2.3...",
		"Pulling/rendering chart with Helm may take a minute for OCI charts.",
		"Installing lab runtime chart 1.2.3 as release lab-abc123 in lab-abc123-system...",
		"Helm release will appear after chart pull/render succeeds.",
		"Local registry: localhost",
	}
	assertContainsInOrder(t, logs.String(), wantLines)
}

func TestRunRejectsUnsafeCustomIDBeforeCommands(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}

	for _, id := range []string{"Demo", "my_lab", "../escape", "-abc123", "abc123-", strings.Repeat("a", 50)} {
		t.Run(id, func(t *testing.T) {
			runner := NewFakeRunner()
			_, err := Run(context.Background(), Options{
				LabPath:      labPath,
				RepoRoot:     repoRoot,
				ID:           id,
				ChartURI:     "oci://example.com/charts/training-platform-assessment",
				ChartVersion: "1.2.3",
				Runner:       runner,
			})
			if err == nil {
				t.Fatal("Run returned nil error for unsafe custom ID")
			}
			if !strings.Contains(err.Error(), "lab run ID") {
				t.Fatalf("error %q does not mention lab run ID", err.Error())
			}
			if len(runner.Commands) != 0 {
				t.Fatalf("recorded commands before rejecting ID: %#v", runner.Commands)
			}
		})
	}
}

func TestResolveStateRejectsUnsafeCustomID(t *testing.T) {
	repoRoot := t.TempDir()
	stateDir := filepath.Join(repoRoot, ".state")

	_, statePath, err := resolveState(Options{RepoRoot: repoRoot, StateDir: stateDir, ID: "../escape"})
	if err == nil {
		t.Fatal("resolveState returned nil error for unsafe custom ID")
	}
	if statePath != "" {
		t.Fatalf("statePath = %q, want empty", statePath)
	}
	if !strings.Contains(err.Error(), "lab run ID") {
		t.Fatalf("error %q does not mention lab run ID", err.Error())
	}
}

func TestRunExplainsHelmChartAuthorizationFailures(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}

	runner := NewFakeRunner()
	for range 3 {
		runner.QueueResponse(nil, nil)
	}
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, errors.New("not found"))
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, errors.New("not found"))
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	helmAuthErr := errors.New("helm show chart failed: exit status 1\nError: failed to authorize: failed to fetch oauth token: 401 Unauthorized")
	runner.QueueResponse(nil, helmAuthErr)

	_, err := Run(context.Background(), Options{
		LabPath:      labPath,
		RepoRoot:     repoRoot,
		ID:           "abc123",
		ChartURI:     "oci://europe-west2-docker.pkg.dev/example/chart",
		ChartVersion: "1.2.3",
		Runner:       runner,
	})
	if !errors.Is(err, helmAuthErr) {
		t.Fatalf("Run error = %v, want helm auth error", err)
	}
	if !strings.Contains(err.Error(), "helm registry login") || !strings.Contains(err.Error(), "europe-west2-docker.pkg.dev") {
		t.Fatalf("Run error = %q, want helm auth hint", err.Error())
	}
}

func TestRunRequiresExplicitImageAccessForAllowedNonKindContext(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}

	runner := NewFakeRunner()
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse([]byte("docker-desktop\n"), nil)

	_, err := Run(context.Background(), Options{
		LabPath:      labPath,
		RepoRoot:     repoRoot,
		ID:           "abc123",
		ChartURI:     "oci://example.com/charts/training-platform-assessment",
		ChartVersion: "1.2.3",
		AllowNonKind: true,
		Runner:       runner,
	})
	if err == nil {
		t.Fatal("Run returned nil error for non-kind context without explicit image access")
	}
	if !strings.Contains(err.Error(), "non-kind") || !strings.Contains(err.Error(), "image") || !strings.Contains(err.Error(), "accessible") {
		t.Fatalf("error %q does not explain non-kind image accessibility requirement", err.Error())
	}
	for _, command := range runner.Commands {
		if command.Name == "kind" {
			t.Fatalf("recorded kind command %#v for non-kind context", command)
		}
	}
}

func TestRunAllowsNonKindWithExplicitImageAccessWithoutKindCommands(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}

	runner := NewFakeRunner()
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse([]byte("docker-desktop\n"), nil)

	_, err := Run(context.Background(), Options{
		LabPath:               labPath,
		RepoRoot:              repoRoot,
		ID:                    "abc123",
		ChartURI:              "oci://example.com/charts/training-platform-assessment",
		ChartVersion:          "1.2.3",
		AllowNonKind:          true,
		AssumeImageAccessible: true,
		Runner:                runner,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	for _, command := range runner.Commands {
		if command.Name == "kind" {
			t.Fatalf("recorded kind command %#v for non-kind context with explicit image access", command)
		}
	}
}

func TestRunEnsuresNamespacesAndSavesStateBeforeHelm(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}

	helmErr := errors.New("helm failed")
	runner := NewFakeRunner()
	for range 3 {
		runner.QueueResponse(nil, nil)
	}
	runner.QueueResponse([]byte("kind-local\n"), nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, nil)
	runner.QueueResponse(nil, helmErr)

	stateDir := filepath.Join(repoRoot, ".state")
	_, err := Run(context.Background(), Options{
		LabPath:      labPath,
		RepoRoot:     repoRoot,
		StateDir:     stateDir,
		ID:           "abc123",
		ChartURI:     "oci://example.com/charts/training-platform-assessment",
		ChartVersion: "1.2.3",
		Runner:       runner,
	})
	if !errors.Is(err, helmErr) {
		t.Fatalf("Run error = %v, want helmErr", err)
	}

	wantNamespaceCommands := []Command{
		{Name: "kubectl", Args: []string{"get", "namespace", "lab-abc123-system"}},
		{Name: "kubectl", Args: []string{"get", "namespace", "lab-abc123-workspace"}},
	}
	var gotNamespaceCommands []Command
	for _, command := range runner.Commands {
		if command.Name == "kubectl" && len(command.Args) >= 2 && command.Args[0] == "get" && command.Args[1] == "namespace" {
			gotNamespaceCommands = append(gotNamespaceCommands, command)
		}
		if command.Name == "kubectl" && len(command.Args) >= 2 && command.Args[0] == "create" && command.Args[1] == "namespace" {
			t.Fatalf("unexpected namespace create command when namespace exists: %#v", command)
		}
	}
	if !reflect.DeepEqual(gotNamespaceCommands, wantNamespaceCommands) {
		t.Fatalf("namespace get commands = %#v, want %#v", gotNamespaceCommands, wantNamespaceCommands)
	}
	loadedState, err := LoadState(filepath.Join(stateDir, "abc123.yaml"))
	if err != nil {
		t.Fatalf("LoadState returned error after helm failure: %v", err)
	}
	if loadedState.RunID != "abc123" || loadedState.SystemNamespace != "lab-abc123-system" || loadedState.WorkspaceNamespace != "lab-abc123-workspace" {
		t.Fatalf("state after helm failure = %#v", loadedState)
	}
}

func TestCreateStarterTarballIncludesFilesEmptyDirsAndSymlinks(t *testing.T) {
	src := filepath.Join(t.TempDir(), "starter-content")
	writeFile(t, filepath.Join(src, "README.md"), "hello lab")
	if err := os.MkdirAll(filepath.Join(src, "empty-dir"), 0755); err != nil {
		t.Fatalf("create empty dir: %v", err)
	}
	if err := os.Symlink("README.md", filepath.Join(src, "readme-link")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	tarball := filepath.Join(t.TempDir(), "starter-content.tar.gz")
	if err := writeTarGz(src, tarball); err != nil {
		t.Fatalf("writeTarGz returned error: %v", err)
	}

	entries := readTarGzEntries(t, tarball)
	if entries["README.md"].Typeflag != tar.TypeReg {
		t.Fatalf("README.md type = %v, want regular file", entries["README.md"].Typeflag)
	}
	if entries["README.md"].Contents != "hello lab" {
		t.Fatalf("README.md contents = %q, want hello lab", entries["README.md"].Contents)
	}
	if entries["empty-dir"].Typeflag != tar.TypeDir {
		t.Fatalf("empty-dir type = %v, want directory", entries["empty-dir"].Typeflag)
	}
	if entries["readme-link"].Typeflag != tar.TypeSymlink {
		t.Fatalf("readme-link type = %v, want symlink", entries["readme-link"].Typeflag)
	}
	if entries["readme-link"].Linkname != "README.md" {
		t.Fatalf("readme-link target = %q, want README.md", entries["readme-link"].Linkname)
	}
}

type tarEntry struct {
	Typeflag byte
	Linkname string
	Contents string
}

func readTarGzEntries(t *testing.T, path string) map[string]tarEntry {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open tarball: %v", err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("read gzip: %v", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	entries := map[string]tarEntry{}
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("read tar entry: %v", err)
		}
		var contents []byte
		if header.Typeflag == tar.TypeReg {
			contents, err = io.ReadAll(tr)
			if err != nil {
				t.Fatalf("read %s contents: %v", header.Name, err)
			}
		}
		entries[header.Name] = tarEntry{Typeflag: header.Typeflag, Linkname: header.Linkname, Contents: string(contents)}
	}
	return entries
}

func assertArgsContainAll(t *testing.T, args []string, values ...string) {
	t.Helper()
	joined := strings.Join(args, "\x00")
	for _, value := range values {
		if !strings.Contains(joined, value) {
			t.Fatalf("args %#v do not contain %q", args, value)
		}
	}
}

func assertArgsDoNotContain(t *testing.T, args []string, values ...string) {
	t.Helper()
	joined := strings.Join(args, "\x00")
	for _, value := range values {
		if strings.Contains(joined, value) {
			t.Fatalf("args %#v contain unwanted %q", args, value)
		}
	}
}

func containsArg(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}

func assertContainsInOrder(t *testing.T, output string, values []string) {
	t.Helper()
	searchFrom := 0
	for _, value := range values {
		index := strings.Index(output[searchFrom:], value)
		if index == -1 {
			t.Fatalf("output %q does not contain %q after byte %d", output, value, searchFrom)
		}
		searchFrom += index + len(value)
	}
}

func TestRunRejectsNonKindContextByDefault(t *testing.T) {
	repoRoot := t.TempDir()
	labPath := filepath.Join(repoRoot, "labs", "pod-image-lab")
	if err := ScaffoldStandalone(labPath, ScaffoldOptions{Name: "pod-image-lab"}); err != nil {
		t.Fatalf("ScaffoldStandalone returned error: %v", err)
	}
	runner := NewFakeRunner()
	for range 3 {
		runner.QueueResponse(nil, nil)
	}
	runner.QueueResponse([]byte("prod-cluster\n"), nil)

	err := func() error {
		_, err := Run(context.Background(), Options{
			LabPath:      labPath,
			RepoRoot:     repoRoot,
			ID:           "abc123",
			ChartURI:     "oci://example.com/charts/training-platform-assessment",
			ChartVersion: "1.2.3",
			Runner:       runner,
		})
		return err
	}()
	if err == nil {
		t.Fatal("Run returned nil error for non-kind context")
	}
	if !strings.Contains(err.Error(), "kind-") {
		t.Fatalf("error %q does not mention kind- context requirement", err.Error())
	}
	if _, statErr := os.Stat(filepath.Join(repoRoot, ".build", "tpm", "labs", "abc123", "starter-content.tar.gz")); !os.IsNotExist(statErr) {
		t.Fatalf("starter tarball stat error = %v, want not exist", statErr)
	}
}
