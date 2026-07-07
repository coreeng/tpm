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
		Long:         "Create, preview, start, inspect, and stop local labs.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newLabInitCmd())
	cmd.AddCommand(newLabOutlineCmd())
	cmd.AddCommand(newLabPreviewCmd())
	cmd.AddCommand(newLabStartCmd())
	cmd.AddCommand(newLabListCmd())
	cmd.AddCommand(newLabStatusCmd())
	cmd.AddCommand(newLabStopCmd())
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
	fmt.Fprintf(out, "%s\n", loaded.Title)
	if opts.codes {
		fmt.Fprintf(out, "code: %s\n", loaded.Code)
	}
	if opts.paths {
		fmt.Fprintf(out, "metadata: %s\nruntime: %s\n", loaded.MetadataPath, loaded.RuntimePath)
	}
	for i, challenge := range loaded.Challenges {
		fmt.Fprintf(out, "%d. %s\n", i+1, challenge.Title)
		if opts.codes {
			fmt.Fprintf(out, "   code: %s\n", challenge.Code)
		}
		for j, goal := range challenge.Goals {
			fmt.Fprintf(out, "   %d.%d %s\n", i+1, j+1, goal.Title)
			if opts.codes {
				fmt.Fprintf(out, "       code: %s\n", goal.Code)
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

func newLabPreviewCmd() *cobra.Command {
	opts := &labPreviewOptions{}
	cmd := &cobra.Command{
		Use:   "preview <lab-path>",
		Short: "Preview lab text, challenges, and goals in a local web UI",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLabPreview(cmd.Context(), cmd, args[0], opts)
		},
	}
	cmd.Flags().StringVar(&opts.addr, "addr", "127.0.0.1:0", "Address to listen on")
	cmd.Flags().BoolVar(&opts.open, "open", false, "Open the preview URL in the default browser")
	cmd.Flags().BoolVar(&opts.watch, "watch", false, "Reload lab metadata and markdown on each browser refresh")
	return cmd
}

func runLabPreview(ctx context.Context, cmd *cobra.Command, labPath string, opts *labPreviewOptions) error {
	var loaded *lab.Lab
	var err error
	if !opts.watch {
		loaded, err = lab.Load(labPath)
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		current := loaded
		if opts.watch {
			var err error
			current, err = lab.Load(labPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err := labPreviewTemplate.Execute(w, current); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	listener, err := net.Listen("tcp", opts.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{Handler: mux}
	errCh := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	url := "http://" + listener.Addr().String()
	fmt.Fprintf(cmd.OutOrStdout(), "Lab preview: %s\n", url)
	if opts.watch {
		fmt.Fprintln(cmd.OutOrStdout(), "watch: reloading lab metadata and markdown on refresh")
	}
	if opts.open {
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

func newLabStartCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "start <lab-path>",
		Short: "Start a local lab runtime",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			opts.LabPath = args[0]
			opts.LogWriter = cmd.OutOrStdout()
			state, err := lab.Run(cmd.Context(), opts)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Lab %s running\nSystem namespace: %s\nWorkspace namespace: %s\nRegistry URL: %s\nRegistry username: %s\nRegistry token: %s\n", state.RunID, state.SystemNamespace, state.WorkspaceNamespace, state.RegistryURL, state.RegistryUsername, state.RegistryToken)
			return err
		},
	}
	addSharedLabFlags(cmd, &opts)
	cmd.Flags().StringVar(&opts.ChartDir, "chart-dir", "", "Path to a local lab runtime Helm chart directory")
	cmd.Flags().StringVar(&opts.ChartURI, "chart-uri", "", "OCI URI for the lab runtime Helm chart")
	cmd.Flags().StringVar(&opts.ChartVersion, "chart-version", "", "Version of the lab runtime Helm chart")
	cmd.Flags().StringVar(&opts.ValidatorRegistry, "validator-registry", lab.DefaultArtifactRegistry, "Registry for the locally built validator image")
	cmd.Flags().StringVar(&opts.RegistryDomain, "registry-domain", lab.DefaultArtifactRegistry, "Learner registry domain passed to the lab runtime chart")
	cmd.Flags().BoolVar(&opts.AssumeImageAccessible, "assume-image-accessible", false, "assume a non-kind cluster can pull the local validator image tag")
	cmd.Flags().DurationVar(&opts.CheckInterval, "check-interval", 5*time.Second, "validator check interval")
	return cmd
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

func newLabStopCmd() *cobra.Command {
	opts := lab.Options{}
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a local lab runtime",
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
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
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
