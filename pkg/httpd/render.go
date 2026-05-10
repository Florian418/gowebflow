package httpd

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// LoadTemplates parses all .html files in LayoutDir and caches them.
// It registers the "vite" template function for asset injection.
// It is called automatically by New, and on every request in Dev mode.
func (a *App) LoadTemplates() error {
	funcMap := template.FuncMap{
		"vite": func(entry string) template.HTML {
			a.mu.RLock()
			v := a.vite
			a.mu.RUnlock()
			if v == nil {
				return ""
			}
			return template.HTML(v.tag(entry))
		},
	}

	if a.config.LayoutDir == "" {
		a.mu.Lock()
		a.templates = template.New("").Funcs(funcMap)
		a.mu.Unlock()
		return nil
	}

	pattern := filepath.Join(a.config.LayoutDir, "*.html")
	tpl, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.templates = tpl
	a.mu.Unlock()
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
	a.mu.RLock()
	base := a.templates
	a.mu.RUnlock()

	tpl, err := base.Clone()
	if err != nil {
		return err
	}
	pageFile := filepath.Join(a.config.PageDir, name)
	if _, err = tpl.ParseFiles(pageFile); err != nil {
		return err
	}
	templateName := "base"
	if a.config.LayoutDir == "" {
		templateName = filepath.Base(name)
	}
	return tpl.ExecuteTemplate(w, templateName, data)
}
