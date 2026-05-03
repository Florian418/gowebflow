package httpd_test

import (
	"encoding/json"
	"fmt"
	"git.euflow.fr/flo/gowebflow/pkg/httpd"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestApp creates a temporary App with the given templates for testing.
func newTestApp(t *testing.T, templates map[string]string) *httpd.App {
	t.Helper()
	dir := t.TempDir()
	for name, content := range templates {
		os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	}
	app, err := httpd.New(httpd.Config{TemplateDir: dir})
	if err != nil {
		t.Fatalf("newTestApp: %v", err)
	}
	return app
}

func TestNew(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Test</h1>",
	})
	if app == nil {
		t.Fatal("expected non-nil App")
	}
}

func TestNewInvalidTemplateDir(t *testing.T) {
	_, err := httpd.New(httpd.Config{TemplateDir: "./nonexistent"})
	if err == nil {
		t.Fatal("expected error for invalid template dir, got nil")
	}
}

func TestRender(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Bonjour</h1>",
	})

	w := httptest.NewRecorder()
	if err := app.Render(w, "home.html", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(w.Body.String(), "<h1>Bonjour</h1>") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRenderWithData(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>{{ .Title }}</h1>",
	})

	w := httptest.NewRecorder()
	err := app.Render(w, "home.html", map[string]any{"Title": "goWebFlow"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(w.Body.String(), "<h1>goWebFlow</h1>") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestGet(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.Get("/hello", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("hello"))
		return nil
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "hello" {
		t.Errorf("expected 'hello', got %s", w.Body.String())
	}
}

func TestPost(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.Post("/submit", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("submitted"))
		return nil
	})

	req := httptest.NewRequest("POST", "/submit", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "submitted" {
		t.Errorf("expected 'submitted', got %s", w.Body.String())
	}
}

func TestPut(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.Put("/update", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("updated"))
		return nil
	})

	req := httptest.NewRequest("PUT", "/update", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDelete(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.Delete("/remove", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("deleted"))
		return nil
	})

	req := httptest.NewRequest("DELETE", "/remove", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestWrapHandlesError(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.Get("/fail", func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("something went wrong")
	})

	req := httptest.NewRequest("GET", "/fail", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestNotFound(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.Get("/", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("home"))
		return nil
	})

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRootDoesNotCatchAll(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.Get("/", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("home"))
		return nil
	})

	req := httptest.NewRequest("GET", "/other", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("/ should not catch /other, expected 404, got %d", w.Code)
	}
}

func TestCustomNotFound(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.NotFound(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("custom 404"))
		return nil
	})

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "custom 404") {
		t.Errorf("expected custom body, got: %s", w.Body.String())
	}
}

func TestCustomError500(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	app.OnError(http.StatusInternalServerError, func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("custom 500"))
		return nil
	})

	app.Get("/fail", func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("boom")
	})

	req := httptest.NewRequest("GET", "/fail", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "custom 500") {
		t.Errorf("expected custom body, got: %s", w.Body.String())
	}
}

func TestErrHTTPTriggersOnError(t *testing.T) {
	app := newTestApp(t, map[string]string{
		"home.html": "<h1>Home</h1>",
	})

	// sécurité : on répond 404 même quand c'est un 403
	app.OnError(http.StatusForbidden, func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return nil
	})

	app.Get("/admin", func(w http.ResponseWriter, r *http.Request) error {
		return httpd.ErrHTTP(http.StatusForbidden, fmt.Errorf("forbidden"))
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 (masked 403), got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "not found") {
		t.Errorf("expected custom body, got: %s", w.Body.String())
	}
}

func TestViteDevMode(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "home.html"), []byte(`{{ vite "src/default/main.js" }}`), 0644)

	app, err := httpd.New(httpd.Config{
		TemplateDir: dir,
		StaticDir:   dir,
		Dev:         true,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	w := httptest.NewRecorder()
	if err := app.Render(w, "home.html", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "localhost:5173") {
		t.Errorf("expected Vite dev server URL, got: %s", body)
	}
}

func TestViteProdMode(t *testing.T) {
	dir := t.TempDir()

	// create a fake Vite manifest under the default theme subfolder
	viteDir := filepath.Join(dir, "default", ".vite")
	os.MkdirAll(viteDir, 0755)
	manifest := map[string]any{
		"src/main.js": map[string]any{
			"file": "assets/main-abc123.js",
			"css":  []string{"assets/main-def456.css"},
		},
	}
	data, _ := json.Marshal(manifest)
	os.WriteFile(filepath.Join(viteDir, "manifest.json"), data, 0644)

	os.WriteFile(filepath.Join(dir, "home.html"), []byte(`{{ vite "src/main.js" }}`), 0644)

	app, err := httpd.New(httpd.Config{
		TemplateDir: dir,
		StaticDir:   dir,
		StaticURL:   "/static/",
		Dev:         false,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	w := httptest.NewRecorder()
	if err := app.Render(w, "home.html", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "main-abc123.js") {
		t.Errorf("expected hashed JS filename, got: %s", body)
	}
	if !strings.Contains(body, "main-def456.css") {
		t.Errorf("expected hashed CSS filename, got: %s", body)
	}
}

func TestSetTheme(t *testing.T) {
	dir := t.TempDir()

	writeThemeManifest := func(theme, jsFile string) {
		viteDir := filepath.Join(dir, theme, ".vite")
		os.MkdirAll(viteDir, 0755)
		data, _ := json.Marshal(map[string]any{
			"src/main.js": map[string]any{"file": jsFile},
		})
		os.WriteFile(filepath.Join(viteDir, "manifest.json"), data, 0644)
	}

	writeThemeManifest("default", "assets/main-default.js")
	writeThemeManifest("autre", "assets/main-autre.js")
	os.WriteFile(filepath.Join(dir, "home.html"), []byte(`{{ vite "src/main.js" }}`), 0644)

	app, err := httpd.New(httpd.Config{
		TemplateDir: dir,
		StaticDir:   dir,
		StaticURL:   "/static/",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.SetTheme("autre"); err != nil {
		t.Fatalf("SetTheme: %v", err)
	}

	w := httptest.NewRecorder()
	if err := app.Render(w, "home.html", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(w.Body.String(), "main-autre.js") {
		t.Errorf("expected theme 'autre' manifest, got: %s", w.Body.String())
	}
}

func TestStatic(t *testing.T) {
	app := newTestApp(t, map[string]string{"home.html": "<h1>ok</h1>"})

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("zion"), 0644)
	app.Static("/files/", dir)

	req := httptest.NewRequest("GET", "/files/hello.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "zion" {
		t.Errorf("expected 'zion', got %s", w.Body.String())
	}
}

func TestSetThemeInvalidName(t *testing.T) {
	dir := t.TempDir()

	viteDir := filepath.Join(dir, "default", ".vite")
	os.MkdirAll(viteDir, 0755)
	data, _ := json.Marshal(map[string]any{
		"src/main.js": map[string]any{"file": "assets/main-default.js"},
	})
	os.WriteFile(filepath.Join(viteDir, "manifest.json"), data, 0644)
	os.WriteFile(filepath.Join(dir, "home.html"), []byte(`{{ vite "src/main.js" }}`), 0644)

	app, err := httpd.New(httpd.Config{
		TemplateDir: dir,
		StaticDir:   dir,
		StaticURL:   "/static/",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.SetTheme("inexistant"); err == nil {
		t.Fatal("expected error for unknown theme, got nil")
	}
}
