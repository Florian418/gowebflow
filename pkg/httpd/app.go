package httpd

import (
	"html/template"
	"net/http"
	"path/filepath"
	"sync"
)

// Config holds the configuration for a goWebFlow application.
type Config struct {
	TemplateDir string // path to the folder containing HTML templates
	StaticDir   string // path to the Vite build output folder (e.g. ./ui/dist)
	StaticURL   string // URL prefix used to serve static assets (e.g. /static/)
	ActiveTheme string // nouveau — sous-dossier de StaticDir (ex: "default")
	Dev         bool   // enables Vite dev server mode and template hot reload
}

// App is the central object of the framework.
// It holds the router, the template engine and the configuration.
type App struct {
	config      Config
	mux         *http.ServeMux
	templates   *template.Template
	vite        *viteManifest
	mu          sync.RWMutex
	activeTheme string
	errorHandlers map[int]HandlerFunc
}

// New creates and returns a new App with the given configuration.
// Returns an error if the templates cannot be loaded.
func New(cfg Config) (*App, error) {
	a := &App{
		config:      cfg,
		mux:         http.NewServeMux(),
		activeTheme: cfg.ActiveTheme,
		errorHandlers: map[int]HandlerFunc{
			http.StatusNotFound: func(w http.ResponseWriter, r *http.Request) error {
				http.NotFound(w, r)
				return nil
			},
			http.StatusInternalServerError: func(w http.ResponseWriter, r *http.Request) error {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil
			},
		},
	}
	if a.activeTheme == "" {
		a.activeTheme = "default"
	}

	a.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := a.errorHandlers[http.StatusNotFound](w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	if cfg.StaticDir != "" {
		if cfg.Dev {
			a.vite = &viteManifest{dev: true}
		} else {
			themeDir := filepath.Join(cfg.StaticDir, a.activeTheme)
			var err error
			a.vite, err = loadManifest(themeDir, cfg.StaticURL)
			if err != nil {
				return nil, err
			}
			// handler enregistré une fois, lit activeTheme à chaque requête
			a.mux.Handle(cfg.StaticURL, http.StripPrefix(cfg.StaticURL,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a.mu.RLock()
					theme := a.activeTheme
					a.mu.RUnlock()
					dir := filepath.Join(a.config.StaticDir, theme)
					http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
				})))
		}
	}

	if err := a.LoadTemplates(); err != nil {
		return nil, err
	}

	return a, nil
}

// ServeHTTP implements the http.Handler interface.
// This allows passing the App directly to http.ListenAndServe.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

// Listen starts the HTTP server on the given address (e.g. ":8080").
func (a *App) Listen(addr string) error {
	return http.ListenAndServe(addr, a)
}

// Static serves files from dir at the given URL prefix.
// e.g. app.Static("/fonts/", "./ui/public/fonts")
func (a *App) Static(prefix, dir string) {
	a.mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
}

// NotFound registers a custom handler called when no route matches.
// Sugar pour OnError(http.StatusNotFound, h).
func (a *App) NotFound(h HandlerFunc) {
	a.OnError(http.StatusNotFound, h)
}

// OnError registers a custom handler for the given HTTP status code.
// Called automatically for 404 (no route matched) and 500 (handler returned an error).
// For other codes (403, etc.), trigger them from a route with httpd.ErrHTTP.
func (a *App) OnError(code int, h HandlerFunc) {
	a.errorHandlers[code] = h
}

// SetTheme swaps the active Vite theme at runtime without restarting the server.
// It reloads the manifest from StaticDir/<name>/.vite/manifest.json.
func (a *App) SetTheme(name string) error {
	themeDir := filepath.Join(a.config.StaticDir, name)
	v, err := loadManifest(themeDir, a.config.StaticURL)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.activeTheme = name
	a.vite = v
	return nil
}
