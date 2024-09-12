package lib

import (
	"io/fs"
	"net/http"
	"text/template"
)

type Engine struct {
	views        fs.FS
	baseFolder   string
	baseTemplate string
}

func NewEngine(views fs.FS, baseFolder string, baseTemplate string) *Engine {
	return &Engine{views: views, baseFolder: baseFolder, baseTemplate: baseTemplate}
}

func (e *Engine) ExecuteContent(w http.ResponseWriter, templateName string, data any) error {
	t, err := template.ParseFS(e.views, e.baseFolder+"/"+e.baseTemplate, e.baseFolder+"/"+templateName)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, e.baseTemplate, data)
}
