package cmd

import (
	"embed"
	"html/template"
)

//go:embed templates/*.tmpl
var previewTemplateFiles embed.FS

func mustPreviewTemplate(name string) *template.Template {
	return template.Must(template.New(name).Funcs(template.FuncMap{
		"add1": func(value int) int {
			return value + 1
		},
	}).ParseFS(previewTemplateFiles, "templates/"+name))
}
