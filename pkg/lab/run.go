package lab

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	LabelManagedBy     = "training-platform.coreeng.io/managed-by"
	LabelLabRunID      = "training-platform.coreeng.io/lab-run-id"
	LabelLabCode       = "training-platform.coreeng.io/lab-code"
	LabelNamespaceRole = "training-platform.coreeng.io/lab-namespace-role"

	LabelManagedByValue = "tpm"
	NamespaceRoleSystem = "system"
	NamespaceRoleWork   = "workspace"

	DefaultArtifactRegistry      = "localhost"
	PodSecurityEnforceLabel      = "pod-security.kubernetes.io/enforce"
	PodSecurityAuditLabel        = "pod-security.kubernetes.io/audit"
	PodSecurityRestrictedValue   = "restricted"
	DefaultLocalRegistryUsername = "workspace"
	DefaultLocalRegistryPassword = "local-password"
)

type Options struct {
	LabPath       string
	RepoRoot      string
	StateDir      string
	ID            string
	ChartDir      string
	ChartURI      string
	ChartVersion  string
	CheckInterval time.Duration
	AllowNonKind  bool
	// ValidatorRegistry is the registry used for the locally built validator image.
	ValidatorRegistry string
	// RegistryDomain is passed to the lab runtime chart for the learner registry.
	RegistryDomain string
	// AssumeImageAccessible confirms non-kind clusters can pull the local validator image tag.
	AssumeImageAccessible bool
	Runner                Runner
	LogWriter             io.Writer
}

func Run(ctx context.Context, opts Options) (*RunState, error) {
	runner := opts.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	if opts.ID == "" {
		opts.ID = uuid.NewString()
	}
	if err := validateRunID(opts.ID); err != nil {
		return nil, err
	}
	if opts.RepoRoot == "" {
		opts.RepoRoot = "."
	}
	if opts.StateDir == "" {
		opts.StateDir = StateDir(opts.RepoRoot)
	}
	if opts.CheckInterval == 0 {
		opts.CheckInterval = 5 * time.Second
	}
	validatorRegistry := strings.TrimSpace(opts.ValidatorRegistry)
	if validatorRegistry == "" {
		validatorRegistry = DefaultArtifactRegistry
	}
	registryDomain := strings.TrimSpace(opts.RegistryDomain)
	if registryDomain == "" {
		registryDomain = DefaultArtifactRegistry
	}
	labPath, err := normalizeLabPath(opts.LabPath)
	if err != nil {
		return nil, err
	}
	lab, err := Load(labPath)
	if err != nil {
		return nil, err
	}
	names := NewRunNames(opts.ID, lab.Code)
	release := "lab-" + opts.ID
	validatorRepo := strings.TrimSuffix(validatorRegistry, "/") + "/tpm-lab-" + lab.Code + "-validator"
	validatorTag := validatorRepo + ":" + opts.ID
	registryURL := registryDomain
	logStep(opts, "Starting lab %s with run id %s", lab.Code, opts.ID)

	for _, check := range []Command{
		{Name: "docker", Args: []string{"version"}},
		{Name: "helm", Args: []string{"version"}},
		{Name: "kubectl", Args: []string{"version", "--client"}},
	} {
		logStep(opts, "Checking %s...", displayCommandName(check.Name))
		if err := runner.Run(ctx, check.Name, check.Args...); err != nil {
			return nil, fmt.Errorf("check %s %s: %w", check.Name, strings.Join(check.Args, " "), err)
		}
	}

	contextOutput, err := runner.Output(ctx, "kubectl", "config", "current-context")
	if err != nil {
		return nil, fmt.Errorf("check kubectl current context: %w", err)
	}
	currentContext := strings.TrimSpace(string(contextOutput))
	logStep(opts, "Using kubectl context %s", currentContext)
	kindCluster, isKindContext := strings.CutPrefix(currentContext, "kind-")
	if !opts.AllowNonKind && !isKindContext {
		return nil, fmt.Errorf("current kubectl context %q does not start with kind-; use a kind cluster or allow non-kind contexts", currentContext)
	}
	if opts.AllowNonKind && !isKindContext && !opts.AssumeImageAccessible {
		return nil, fmt.Errorf("non-kind clusters need the validator image already accessible to the cluster; set AssumeImageAccessible when %q can pull %q", currentContext, validatorTag)
	}
	if isKindContext {
		logStep(opts, "Checking kind...")
		if err := runner.Run(ctx, "kind", "version"); err != nil {
			return nil, fmt.Errorf("check kind version: %w", err)
		}
	}

	logStep(opts, "Building validator image %s...", validatorTag)
	if err := runner.Run(ctx, "docker", "build", "-t", validatorTag, lab.ValidatorPath); err != nil {
		return nil, fmt.Errorf("build lab validator image: %w", err)
	}
	if isKindContext {
		logStep(opts, "Loading validator image into kind cluster %s...", kindCluster)
		if err := runner.Run(ctx, "kind", "load", "docker-image", "--name", kindCluster, validatorTag); err != nil {
			return nil, fmt.Errorf("load lab validator image into kind: %w", err)
		}
	}
	logStep(opts, "Ensuring system namespace %s...", names.SystemNamespace)
	if err := ensureNamespace(ctx, runner, names.SystemNamespace, labNamespaceLabels(opts.ID, lab.Code, NamespaceRoleSystem)); err != nil {
		return nil, fmt.Errorf("ensure lab system namespace: %w", err)
	}
	logStep(opts, "Ensuring workspace namespace %s...", names.WorkspaceNamespace)
	if err := ensureNamespace(ctx, runner, names.WorkspaceNamespace, workspaceNamespaceLabels(opts.ID, lab.Code)); err != nil {
		return nil, fmt.Errorf("ensure lab workspace namespace: %w", err)
	}

	state := RunState{
		LabPath:            labPath,
		RunID:              opts.ID,
		SystemNamespace:    names.SystemNamespace,
		WorkspaceNamespace: names.WorkspaceNamespace,
		ValidatorImageTag:  validatorTag,
		RegistryURL:        registryURL,
		RegistryUsername:   DefaultLocalRegistryUsername,
		RegistryToken:      DefaultLocalRegistryPassword,
		ChartDir:           opts.ChartDir,
		HelmReleaseName:    release,
		ChartURI:           opts.ChartURI,
		ChartVersion:       opts.ChartVersion,
		CreatedAt:          time.Now().UTC(),
	}
	if err := SaveState(opts.StateDir, state); err != nil {
		return nil, err
	}

	starterTarball := filepath.Join(opts.StateDir, opts.ID, "starter-content.tar.gz")
	logStep(opts, "Packaging starter content...")
	if err := writeTarGz(lab.StarterPath, starterTarball); err != nil {
		return nil, fmt.Errorf("package lab starter content: %w", err)
	}

	chartRef := opts.ChartURI
	if opts.ChartDir != "" {
		chartRef = opts.ChartDir
	}
	helmArgs := []string{
		"upgrade", "--install", release, chartRef,
	}
	if opts.ChartURI != "" {
		helmArgs = append(helmArgs, "--version", opts.ChartVersion)
	}
	helmArgs = append(helmArgs,
		"-n", names.SystemNamespace,
		"--set", "assessment.instanceID="+opts.ID,
		"--set", "assessment.workspaceNS="+names.WorkspaceNamespace,
		"--set", "assessment.systemNS="+names.SystemNamespace,
		"--set", "validator.image.repository="+validatorRepo,
		"--set", "validator.image.tag="+opts.ID,
		"--set", "registry.domain="+registryDomain,
		"--set", "registry.ingress.enabled=false",
		"--set", "registry.registryPassword="+DefaultLocalRegistryPassword,
		"--set", "github.repository=local/"+lab.Code,
		"--set", "github.accessToken=local-token",
		"--set", "validator.extraEnv[0].name=VALIDATOR_CHECK_INTERVAL",
		"--set", "validator.extraEnv[0].value="+opts.CheckInterval.String(),
	)
	if opts.ChartURI != "" {
		logStep(opts, "Checking lab runtime chart %s...", opts.ChartVersion)
		logStep(opts, "Pulling/rendering chart with Helm may take a minute for OCI charts.")
		if err := runner.Run(ctx, "helm", "show", "chart", opts.ChartURI, "--version", opts.ChartVersion); err != nil {
			return nil, explainHelmChartAccessError(opts.ChartURI, err)
		}
	}
	chartDisplay := chartRef
	if opts.ChartURI != "" {
		chartDisplay = opts.ChartVersion
	}
	logStep(opts, "Installing lab runtime chart %s as release %s in %s...", chartDisplay, release, names.SystemNamespace)
	logStep(opts, "Helm release will appear after chart pull/render succeeds.")
	if err := runner.Run(ctx, "helm", helmArgs...); err != nil {
		return nil, fmt.Errorf("install lab runtime chart: %w", err)
	}
	logStep(opts, "Local registry: %s", registryURL)
	if err := SaveState(opts.StateDir, state); err != nil {
		return nil, err
	}
	return &state, nil
}

func explainHelmChartAccessError(chartURI string, err error) error {
	message := err.Error()
	if strings.Contains(message, "401 Unauthorized") || strings.Contains(message, "failed to authorize") {
		registry := helmOCIRegistry(chartURI)
		if registry != "" {
			return fmt.Errorf("check lab runtime chart access: %w\nHelm could not authenticate to %s. Run: gcloud auth print-access-token | helm registry login -u oauth2accesstoken --password-stdin %s", err, registry, registry)
		}
		return fmt.Errorf("check lab runtime chart access: %w\nHelm could not authenticate to the chart registry. Run helm registry login for the chart registry and try again", err)
	}
	return fmt.Errorf("check lab runtime chart access: %w", err)
}

func helmOCIRegistry(chartURI string) string {
	withoutScheme, ok := strings.CutPrefix(chartURI, "oci://")
	if !ok {
		return ""
	}
	registry, _, _ := strings.Cut(withoutScheme, "/")
	return registry
}

func logStep(opts Options, format string, args ...any) {
	if opts.LogWriter == nil {
		return
	}
	fmt.Fprintf(opts.LogWriter, format+"\n", args...)
}

func displayCommandName(name string) string {
	switch name {
	case "docker":
		return "Docker"
	case "helm":
		return "Helm"
	case "kubectl":
		return "kubectl"
	default:
		return name
	}
}

func ensureNamespace(ctx context.Context, runner Runner, namespace string, labels []string) error {
	if err := runner.Run(ctx, "kubectl", "get", "namespace", namespace); err == nil {
		return labelNamespace(ctx, runner, namespace, labels)
	}
	if err := runner.Run(ctx, "kubectl", "create", "namespace", namespace); err != nil {
		return err
	}
	return labelNamespace(ctx, runner, namespace, labels)
}

func labNamespaceLabels(runID, labCode, role string) []string {
	return []string{
		LabelManagedBy + "=" + LabelManagedByValue,
		LabelLabRunID + "=" + runID,
		LabelLabCode + "=" + labCode,
		LabelNamespaceRole + "=" + role,
	}
}

func workspaceNamespaceLabels(runID, labCode string) []string {
	return append(labNamespaceLabels(runID, labCode, NamespaceRoleWork),
		PodSecurityEnforceLabel+"="+PodSecurityRestrictedValue,
		PodSecurityAuditLabel+"="+PodSecurityRestrictedValue,
	)
}

func labelNamespace(ctx context.Context, runner Runner, namespace string, labels []string) error {
	args := append([]string{"label", "namespace", namespace}, labels...)
	args = append(args, "--overwrite")
	return runner.Run(ctx, "kubectl", args...)
}

func writeTarGz(srcDir, dstPath string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	file, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	return filepath.WalkDir(srcDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == srcDir {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		linkTarget := ""
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		header, err := tar.FileInfoHeader(info, linkTarget)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if entry.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tw, input)
		closeErr := input.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}
