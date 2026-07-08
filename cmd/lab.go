package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/coreeng/tpm/pkg/lab"
	"github.com/spf13/cobra"
)

var labCmd = newLabCmd()

func newLabCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "lab",
		Short:        "Work with local labs",
		Long:         "Create, preview, inspect, and clean up local labs.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newLabInitCmd())
	cmd.AddCommand(newLabOutlineCmd())
	cmd.AddCommand(newLabPreviewCmd())
	cmd.AddCommand(newLabListCmd())
	cmd.AddCommand(newLabStatusCmd())
	cmd.AddCommand(newLabCleanupCmd())
	return cmd
}

func newLabInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <lab-path>",
		Short: "Create a standalone lab runtime skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := filepath.Clean(args[0])
			if err := lab.ScaffoldStandalone(target, lab.ScaffoldOptions{Name: filepath.Base(target)}); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Created standalone lab %s\n", target)
			return err
		},
	}
}

type labOutlineOptions struct {
	codes bool
	paths bool
	json  bool
}

func newLabOutlineCmd() *cobra.Command {
	opts := &labOutlineOptions{}
	cmd := &cobra.Command{
		Use:   "outline <lab-path>",
		Short: "Print a lab challenge and goal outline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := lab.Load(args[0])
			if err != nil {
				return err
			}
			if opts.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(loaded)
			}
			return writeLabOutline(cmd, loaded, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.codes, "codes", false, "Show lab, challenge, and goal codes")
	cmd.Flags().BoolVar(&opts.paths, "paths", false, "Show metadata and runtime paths")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Write JSON")
	return cmd
}

func writeLabOutline(cmd *cobra.Command, loaded *lab.Lab, opts *labOutlineOptions) error {
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "%s\n", loaded.Title); err != nil {
		return err
	}
	if opts.codes {
		if _, err := fmt.Fprintf(out, "code: %s\n", loaded.Code); err != nil {
			return err
		}
	}
	if opts.paths {
		if _, err := fmt.Fprintf(out, "metadata: %s\nruntime: %s\n", loaded.MetadataPath, loaded.RuntimePath); err != nil {
			return err
		}
	}
	for i, challenge := range loaded.Challenges {
		if _, err := fmt.Fprintf(out, "%d. %s\n", i+1, challenge.Title); err != nil {
			return err
		}
		if opts.codes {
			if _, err := fmt.Fprintf(out, "   code: %s\n", challenge.Code); err != nil {
				return err
			}
		}
		for j, goal := range challenge.Goals {
			if _, err := fmt.Fprintf(out, "   %d.%d %s\n", i+1, j+1, goal.Title); err != nil {
				return err
			}
			if opts.codes {
				if _, err := fmt.Fprintf(out, "       code: %s\n", goal.Code); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type labPreviewOptions struct {
	addr          string
	noOpenBrowser bool
	watch         bool
}

func newLabPreviewCmd() *cobra.Command {
	previewOpts := &labPreviewOptions{}
	runtimeOpts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "preview <lab-path>",
		Short: "Start a local lab runtime and preview it in a local web UI",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := prepareLabRuntimeOptions(cmd, &runtimeOpts, args[0]); err != nil {
				return err
			}
			return runLabPreview(cmd.Context(), cmd, &runtimeOpts, previewOpts)
		},
	}
	addSharedLabFlags(cmd, &runtimeOpts)
	cmd.Flags().StringVar(&runtimeOpts.ChartDir, "chart-dir", "", "Path to a local lab runtime Helm chart directory")
	cmd.Flags().StringVar(&runtimeOpts.ChartURI, "chart-uri", "", "OCI URI for the lab runtime Helm chart")
	cmd.Flags().StringVar(&runtimeOpts.ChartVersion, "chart-version", "", "Version of the lab runtime Helm chart")
	cmd.Flags().StringVar(&runtimeOpts.ValidatorRegistry, "validator-registry", lab.DefaultArtifactRegistry, "Registry for the locally built validator image")
	cmd.Flags().StringVar(&runtimeOpts.RegistryDomain, "registry-domain", lab.DefaultArtifactRegistry, "Learner registry domain passed to the lab runtime chart")
	cmd.Flags().BoolVar(&runtimeOpts.AssumeImageAccessible, "assume-image-accessible", false, "assume a non-kind cluster can pull the local validator image tag")
	cmd.Flags().DurationVar(&runtimeOpts.CheckInterval, "check-interval", 5*time.Second, "validator check interval")
	cmd.Flags().StringVar(&previewOpts.addr, "addr", "127.0.0.1:0", "Address to listen on")
	cmd.Flags().BoolVar(&previewOpts.noOpenBrowser, "no-open-browser", false, "Do not open the preview URL in the default browser")
	cmd.Flags().BoolVar(&previewOpts.watch, "watch", false, "Reload lab metadata and markdown when source files change")
	return cmd
}

func prepareLabRuntimeOptions(cmd *cobra.Command, opts *lab.Options, labPath string) error {
	chartDir := strings.TrimSpace(opts.ChartDir)
	chartURI := strings.TrimSpace(opts.ChartURI)
	chartURIChanged := cmd.Flags().Changed("chart-uri")
	chartVersion := strings.TrimSpace(opts.ChartVersion)
	if chartDir != "" && chartURI != "" {
		return fmt.Errorf("set either chart-dir or chart-uri, not both")
	}
	if chartURIChanged && chartURI == "" {
		return fmt.Errorf("chart-uri must not be blank")
	}
	if chartDir == "" && chartURI == "" {
		return fmt.Errorf("chart-dir or chart-uri must be set")
	}
	if cmd.Flags().Changed("chart-version") && chartVersion == "" {
		return fmt.Errorf("chart-version must not be blank")
	}
	opts.ChartDir = chartDir
	opts.ChartURI = chartURI
	opts.ChartVersion = chartVersion
	opts.LabPath = labPath
	return nil
}

func runLabPreview(ctx context.Context, cmd *cobra.Command, runtimeOpts *lab.Options, previewOpts *labPreviewOptions) error {
	runtimeOpts.LogWriter = cmd.OutOrStdout()
	if runtimeOpts.Runner == nil {
		runtimeOpts.Runner = lab.ExecRunner{}
	}

	if err := validateLocalPreviewAddress(previewOpts.addr); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", previewOpts.addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
	}()

	state, err := lab.Run(ctx, *runtimeOpts)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Lab %s running\nSystem namespace: %s\nWorkspace namespace: %s\nRegistry URL: %s\nRegistry username: %s\nRegistry token: %s\n", state.RunID, state.SystemNamespace, state.WorkspaceNamespace, state.RegistryURL, state.RegistryUsername, state.RegistryToken); err != nil {
		return err
	}

	var loaded *lab.Lab
	var watchRoots []string
	if !previewOpts.watch {
		loaded, err = lab.Load(state.LabPath)
		if err != nil {
			return err
		}
	} else {
		loadedForWatch, err := lab.Load(state.LabPath)
		if err != nil {
			return err
		}
		watchRoots = labPreviewWatchRoots(loadedForWatch)
	}

	mux := http.NewServeMux()
	if err := registerPreviewHandlers(mux, func() (any, error) {
		current := loaded
		if previewOpts.watch {
			var err error
			current, err = lab.Load(state.LabPath)
			if err != nil {
				return nil, err
			}
		}
		conditions, statusErr := lab.ProgressConditionsForState(ctx, runtimeOpts.Runner, *state)
		return newLabPreviewPage(current, state, conditions, statusErr), nil
	}, watchRoots); err != nil {
		return err
	}

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	url := "http://" + listener.Addr().String()
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Lab preview: %s\n", url); err != nil {
		return err
	}
	if previewOpts.watch {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "watch: reloading lab metadata and markdown when source files change"); err != nil {
			return err
		}
	}
	if !previewOpts.noOpenBrowser {
		_ = openBrowser(url)
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func validateLocalPreviewAddress(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if _, err := net.LookupPort("tcp", port); err != nil {
		return err
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("lab preview address must bind to localhost or a loopback IP because runtime credentials are shown in the local preview; got %q", addr)
}

func labPreviewWatchRoots(loaded *lab.Lab) []string {
	if loaded == nil {
		return nil
	}
	roots := []string{loaded.RootPath}
	if loaded.RuntimePath != "" && loaded.RuntimePath != loaded.RootPath {
		roots = append(roots, loaded.RuntimePath)
	}
	return roots
}

func newLabListCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active local labs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := lab.List(cmd.Context(), opts)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), output)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow listing labs against a kubectl context that is not a kind cluster")
	return cmd
}

func newLabStatusCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print local lab status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := lab.Status(cmd.Context(), opts)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), output)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.ID, "id", "", "lab run ID")
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	return cmd
}

func newLabCleanupCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up a local lab runtime",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lab.Cleanup(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.ID, "id", "", "lab run ID")
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow running against a kubectl context that is not a kind cluster")
	return cmd
}

func addSharedLabFlags(cmd *cobra.Command, opts *lab.Options) {
	cmd.Flags().StringVar(&opts.ID, "id", "", "lab run ID (default generated)")
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow running against a kubectl context that is not a kind cluster")
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		// #nosec G204 -- open is a fixed local executable and url is generated by the local preview server.
		return exec.Command("open", url).Start()
	case "windows":
		// #nosec G204 -- rundll32 is a fixed local executable and url is generated by the local preview server.
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		// #nosec G204 -- xdg-open is a fixed local executable and url is generated by the local preview server.
		return exec.Command("xdg-open", url).Start()
	}
}
