package main

import (
	"html/template"
	"net/http"

	"github.com/apex/log"
)

var templates map[string]*template.Template

func render(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	t, ok := templates[tmpl]
	if !ok {
		log.WithField("template", tmpl).Error("template does not exist")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := t.Execute(w, data)
	if err != nil {
		log.WithField("template", tmpl).WithError(err).Error("rendering template")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// parseTemplates returns a map of parsed templates with template names for keys.
func parseTemplates() (map[string]*template.Template, error) {
	// template name to required template files
	paths := map[string][]string{
		"index.html": {"template/base.html", "template/index.html"},
	}
	tmpl := make(map[string]*template.Template, len(paths))
	var err error
	for name, files := range paths {
		tmpl[name], err = template.ParseFiles(files...)
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}
