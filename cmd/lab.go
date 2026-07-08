package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
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
	addr  string
	open  bool
	watch bool
}

type labPreviewPage struct {
	*lab.Lab
	State *lab.RunState
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
	cmd.Flags().BoolVar(&previewOpts.open, "open", false, "Open the preview URL in the default browser")
	cmd.Flags().BoolVar(&previewOpts.watch, "watch", false, "Reload lab metadata and markdown on each browser refresh")
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
	if !previewOpts.watch {
		loaded, err = lab.Load(state.LabPath)
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		current := loaded
		if previewOpts.watch {
			var err error
			current, err = lab.Load(state.LabPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err := labPreviewTemplate.Execute(w, labPreviewPage{Lab: current, State: state}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

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
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "watch: reloading lab metadata and markdown on refresh"); err != nil {
			return err
		}
	}
	if previewOpts.open {
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
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow listing labs against a kubectl context that is not kind-")
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
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow running against a kubectl context that is not kind-")
	return cmd
}

func addSharedLabFlags(cmd *cobra.Command, opts *lab.Options) {
	cmd.Flags().StringVar(&opts.ID, "id", "", "lab run ID (default generated)")
	cmd.Flags().StringVar(&opts.StateDir, "state-dir", "", "lab state directory (default "+lab.DefaultStateDirHelp+")")
	cmd.Flags().BoolVar(&opts.AllowNonKind, "allow-non-kind", false, "allow running against a kubectl context that is not kind-")
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

var labPreviewTemplate = template.Must(template.New("lab-preview").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
:root { color-scheme: light; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; color: #202124; background: #f6f7f9; }
body { margin: 0; }
main { max-width: 1120px; margin: 0 auto; padding: 32px 20px; }
header { margin-bottom: 24px; }
h1 { font-size: 32px; line-height: 1.15; margin: 0 0 8px; }
h2 { font-size: 20px; margin: 0 0 12px; }
p { line-height: 1.55; }
.meta { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px; }
.meta span { border: 1px solid #d8dde6; border-radius: 6px; background: #fff; padding: 6px 8px; font-size: 13px; }
.layout { display: grid; grid-template-columns: 320px 1fr; gap: 20px; align-items: start; }
.panel { background: #fff; border: 1px solid #d8dde6; border-radius: 8px; padding: 18px; }
.challenge-list { display: grid; gap: 8px; }
button { width: 100%; text-align: left; border: 1px solid #c6ccd6; background: #fff; border-radius: 6px; padding: 10px 12px; cursor: pointer; font: inherit; }
button:hover, button.active { border-color: #1a73e8; background: #eef4ff; }
.goal { border-top: 1px solid #e4e7ec; padding: 12px 0; }
.goal:first-child { border-top: 0; }
.muted { color: #5f6368; }
pre { white-space: pre-wrap; font: inherit; }
@media (max-width: 760px) { .layout { grid-template-columns: 1fr; } main { padding: 20px 14px; } }
</style>
</head>
<body>
<main>
<header>
<h1>{{.Title}}</h1>
{{if .TimeLimit}}<div class="muted">{{.TimeLimit}}</div>{{end}}
{{if .Description}}<pre>{{.Description}}</pre>{{end}}
{{if .State}}<div class="meta">
<span>Run {{.State.RunID}}</span>
<span>System {{.State.SystemNamespace}}</span>
<span>Workspace {{.State.WorkspaceNamespace}}</span>
<span>Registry {{.State.RegistryURL}}</span>
</div>{{end}}
</header>
<section class="layout">
<nav class="panel">
<h2>Challenges</h2>
<div class="challenge-list">
{{range $i, $challenge := .Challenges}}
<button type="button" data-index="{{$i}}">{{$challenge.Title}}</button>
{{end}}
</div>
</nav>
<article class="panel" id="challenge"></article>
</section>
</main>
<script>
const challenges = [
{{range .Challenges}}{
  title: {{printf "%q" .Title}},
  description: {{printf "%q" .Description}},
  successMessage: {{printf "%q" .SuccessMessage}},
  goals: [
    {{range .Goals}}{ title: {{printf "%q" .Title}}, description: {{printf "%q" .Description}} },
    {{end}}
  ]
},
{{end}}
];
const buttons = [...document.querySelectorAll('button[data-index]')];
const detail = document.getElementById('challenge');
function render(index) {
  const challenge = challenges[index];
  buttons.forEach((button) => button.classList.toggle('active', button.dataset.index == index));
  if (!challenge) {
    detail.innerHTML = '<p class="muted">No challenges found.</p>';
    return;
  }
  detail.innerHTML = '<h2>' + escapeHtml(challenge.title) + '</h2>' +
    (challenge.description ? '<pre>' + escapeHtml(challenge.description) + '</pre>' : '') +
    challenge.goals.map((goal, goalIndex) => '<div class="goal"><strong>' + (goalIndex + 1) + '. ' + escapeHtml(goal.title) + '</strong>' + (goal.description ? '<p>' + escapeHtml(goal.description) + '</p>' : '') + '</div>').join('') +
    (challenge.successMessage ? '<p class="muted">' + escapeHtml(challenge.successMessage) + '</p>' : '');
}
function escapeHtml(value) {
  return String(value).replace(/[&<>"']/g, (char) => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[char]));
}
buttons.forEach((button) => button.addEventListener('click', () => render(Number(button.dataset.index))));
render(0);
</script>
</body>
</html>`))
