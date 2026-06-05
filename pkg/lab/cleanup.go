package lab

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

func Cleanup(ctx context.Context, opts Options) error {
	runner := opts.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	state, statePath, err := resolveState(opts)
	if err != nil {
		return err
	}
	if err := requireSafeCleanupContext(ctx, runner, opts.AllowNonKind); err != nil {
		return err
	}

	var errs []error
	if err := runner.Run(ctx, "helm", "uninstall", state.HelmReleaseName, "-n", state.SystemNamespace); err != nil {
		if !isCleanupNotFoundError(err) {
			errs = append(errs, fmt.Errorf("uninstall lab runtime chart: %w", err))
		}
	}
	for _, namespace := range []string{state.SystemNamespace, state.WorkspaceNamespace} {
		if err := runner.Run(ctx, "kubectl", "delete", "namespace", namespace); err != nil {
			if !isCleanupNotFoundError(err) {
				errs = append(errs, fmt.Errorf("delete lab namespace %s: %w", namespace, err))
			}
		}
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove lab run state: %w", err)
	}
	return nil
}

func isCleanupNotFoundError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not found") || strings.Contains(message, "notfound")
}

func requireSafeCleanupContext(ctx context.Context, runner Runner, allowNonKind bool) error {
	contextOutput, err := runner.Output(ctx, "kubectl", "config", "current-context")
	if err != nil {
		return fmt.Errorf("check kubectl current context: %w", err)
	}
	currentContext := strings.TrimSpace(string(contextOutput))
	if !allowNonKind && !strings.HasPrefix(currentContext, "kind-") {
		return fmt.Errorf("current kubectl context %q does not start with kind-; use a kind cluster or allow non-kind contexts", currentContext)
	}
	return nil
}
