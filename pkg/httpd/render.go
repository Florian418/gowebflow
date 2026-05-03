package httpd

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// LoadTemplates parses all .html files in the configured TemplateDir.
// It registers the "vite" template function for asset injection.
// It is called automatically by New, and on every request in Dev mode.
func (a *App) LoadTemplates() error {
	pattern := filepath.Join(a.config.TemplateDir, "*.html")

	tpl, err := template.New("").Funcs(template.FuncMap{
		"vite": func(entry string) template.HTML {
			a.mu.RLock()
			v := a.vite
			a.mu.RUnlock()
			if v == nil {
				return ""
			}
			return template.HTML(v.tag(entry))
		},
	}).ParseGlob(pattern)
	if err != nil {
		return err
	}

	a.templates = tpl
	return nil
}

// Render executes the named template and writes the result to w.
// data contains the variables available in the template (can be nil).
// In Dev mode, templates are reloaded on every request for hot reload.
func (a *App) Render(w http.ResponseWriter, name string, data any) error {
	if a.config.Dev {
		if err := a.LoadTemplates(); err != nil {
			return err
		}
	}
	return a.templates.ExecuteTemplate(w, name, data)
}
