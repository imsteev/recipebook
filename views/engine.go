package views

import (
	"embed"
	"fmt"
	"net/http"
	"text/template"
)

//go:embed *.html
var allViews embed.FS

type Engine struct {
	baseTemplate string
}

func NewEngine(baseTemplate string) *Engine {
	return &Engine{baseTemplate: baseTemplate}
}

func (e *Engine) ExecuteContent(w http.ResponseWriter, templateName string, data any) error {
	t, err := template.ParseFS(allViews, e.baseTemplate, templateName)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}
	return t.ExecuteTemplate(w, "base", data)
}
