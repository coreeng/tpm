package cmd

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"time"

	"github.com/coreeng/tpm/pkg/builder"
	"github.com/coreeng/tpm/pkg/module"
	"github.com/spf13/cobra"
)

type modulePreviewOptions struct {
	addr  string
	open  bool
	watch bool
}

func newModulePreviewCmd() *cobra.Command {
	opts := &modulePreviewOptions{}
	cmd := &cobra.Command{
		Use:   "preview <module-path>",
		Short: "Preview a full module in a local web UI",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModulePreview(cmd.Context(), cmd, args[0], opts)
		},
	}
	cmd.Flags().StringVar(&opts.addr, "addr", "127.0.0.1:0", "Address to listen on")
	cmd.Flags().BoolVar(&opts.open, "open", false, "Open the preview URL in the default browser")
	cmd.Flags().BoolVar(&opts.watch, "watch", false, "Reload module metadata and markdown on each browser refresh")
	return cmd
}

func runModulePreview(ctx context.Context, cmd *cobra.Command, modulePath string, opts *modulePreviewOptions) error {
	var loaded *module.BuiltModule
	var err error
	if !opts.watch {
		loaded, err = compilePreviewModule(modulePath)
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		current := loaded
		if opts.watch {
			var err error
			current, err = compilePreviewModule(modulePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err := modulePreviewTemplate.Execute(w, current); err != nil {
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
	fmt.Fprintf(cmd.OutOrStdout(), "Module preview: %s\n", url)
	if opts.watch {
		fmt.Fprintln(cmd.OutOrStdout(), "watch: reloading module metadata and markdown on refresh")
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

func compilePreviewModule(modulePath string) (*module.BuiltModule, error) {
	_, _, built, err := builder.Compile(modulePath, "", "")
	return built, err
}

var modulePreviewTemplate = template.Must(template.New("module-preview").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
:root { color-scheme: light; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; color: #202124; background: #f6f7f9; }
body { margin: 0; }
main { max-width: 1180px; margin: 0 auto; padding: 32px 20px; }
header { margin-bottom: 24px; }
h1 { font-size: 34px; line-height: 1.15; margin: 0 0 8px; }
h2 { font-size: 22px; margin: 0 0 12px; }
h3 { font-size: 16px; margin: 18px 0 8px; }
p, pre { line-height: 1.55; }
pre { white-space: pre-wrap; font: inherit; margin: 0; }
.layout { display: grid; grid-template-columns: 300px 1fr; gap: 20px; align-items: start; }
.panel { background: #fff; border: 1px solid #d8dde6; border-radius: 8px; padding: 18px; }
.chapter-list { display: grid; gap: 8px; }
button { border: 1px solid #c6ccd6; background: #fff; border-radius: 6px; padding: 9px 11px; cursor: pointer; font: inherit; }
.chapter-list button { width: 100%; text-align: left; }
button:hover, button.active { border-color: #1a73e8; background: #eef4ff; }
.item { border-top: 1px solid #e4e7ec; padding: 14px 0; }
.item:first-child { border-top: 0; }
.muted { color: #5f6368; }
.video { display: inline-block; margin-top: 8px; color: #1a73e8; text-decoration: none; }
.quiz-option { display: block; width: auto; min-width: 220px; margin: 8px 0; text-align: left; }
@media (max-width: 760px) { .layout { grid-template-columns: 1fr; } main { padding: 20px 14px; } }
</style>
</head>
<body>
<main>
<header>
<h1>{{.Title}}</h1>
<div class="muted">{{.Level}}</div>
{{if .Description}}<pre>{{.Description}}</pre>{{end}}
</header>
<section class="layout">
<nav class="panel">
<h2>Chapters</h2>
<div class="chapter-list">
{{range $i, $chapter := .Chapters}}
<button type="button" data-index="{{$i}}">{{$chapter.Title}}</button>
{{end}}
</div>
</nav>
<article class="panel" id="chapter"></article>
</section>
</main>
<script>
const chapters = [
{{range .Chapters}}{
  title: {{printf "%q" .Title}},
  description: {{printf "%q" .Description}},
  video: {{printf "%q" .BannerVideo}},
  sections: [
    {{range .Sections}}{ title: {{printf "%q" .Title}}, description: {{printf "%q" .Description}}, video: {{printf "%q" .Video}} },
    {{end}}
  ],
  labs: [
    {{range .Assessments}}{ title: {{printf "%q" .Title}}, description: {{printf "%q" .Description}}, challenges: [
      {{range .Challenges}}{ title: {{printf "%q" .Title}}, description: {{printf "%q" .Description}}, goals: [
        {{range .Goals}}{ title: {{printf "%q" .Title}}, description: {{printf "%q" .Description}} },
        {{end}}
      ] },
      {{end}}
    ] },
    {{end}}
  ],
  quizzes: [
    {{range .MultipleChoiceAssessments}}{ title: {{printf "%q" .Title}}, description: {{printf "%q" .Description}}, questions: [
      {{range .Questions}}{ question: {{printf "%q" .Question}}, type: {{printf "%q" .Type}}, options: [
        {{range .Options}}{ text: {{printf "%q" .Text}}, correct: {{.Correct}} },
        {{end}}
      ] },
      {{end}}
    ] },
    {{end}}
  ]
},
{{end}}
];
const buttons = [...document.querySelectorAll('button[data-index]')];
const detail = document.getElementById('chapter');
function render(index) {
  const chapter = chapters[index];
  buttons.forEach((button) => button.classList.toggle('active', button.dataset.index == index));
  if (!chapter) {
    detail.innerHTML = '<p class="muted">No chapters found.</p>';
    return;
  }
  detail.innerHTML = '<h2>' + escapeHtml(chapter.title) + '</h2>' +
    (chapter.description ? '<pre>' + escapeHtml(chapter.description) + '</pre>' : '') +
    (chapter.video ? '<a class="video" href="' + escapeAttr(chapter.video) + '">Video link</a>' : '') +
    chapter.sections.map(sectionTemplate).join('') +
    chapter.labs.map(labTemplate).join('') +
    chapter.quizzes.map(quizTemplate).join('');
}
function sectionTemplate(section) {
  return '<div class="item"><h3>' + escapeHtml(section.title) + '</h3>' +
    (section.description ? '<pre>' + escapeHtml(section.description) + '</pre>' : '') +
    (section.video ? '<a class="video" href="' + escapeAttr(section.video) + '">Video link</a>' : '') + '</div>';
}
function labTemplate(lab) {
  return '<div class="item"><h3>' + escapeHtml(lab.title) + '</h3>' +
    (lab.description ? '<pre>' + escapeHtml(lab.description) + '</pre>' : '') +
    lab.challenges.map((challenge) => '<h3>' + escapeHtml(challenge.title) + '</h3>' +
      (challenge.description ? '<pre>' + escapeHtml(challenge.description) + '</pre>' : '') +
      challenge.goals.map((goal, index) => '<p><strong>' + (index + 1) + '. ' + escapeHtml(goal.title) + '</strong><br>' + escapeHtml(goal.description) + '</p>').join('')).join('') +
    '</div>';
}
function quizTemplate(quiz) {
  return '<div class="item"><h3>' + escapeHtml(quiz.title) + '</h3>' +
    (quiz.description ? '<p>' + escapeHtml(quiz.description) + '</p>' : '') +
    quiz.questions.map((question) => '<p><strong>' + escapeHtml(question.question) + '</strong></p>' +
      question.options.map((option) => '<button class="quiz-option" type="button">' + escapeHtml(option.text) + '</button>').join('')).join('') +
    '</div>';
}
function escapeHtml(value) {
  return String(value).replace(/[&<>"']/g, (char) => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[char]));
}
function escapeAttr(value) {
  return escapeHtml(value).replace(/"/g, '&quot;');
}
buttons.forEach((button) => button.addEventListener('click', () => render(Number(button.dataset.index))));
render(0);
</script>
</body>
</html>`))
