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

const baseLayout = `{{define "base"}}{{block "content" .}}{{end}}{{end}}`

// newTestApp creates a temporary App with the given layouts and pages for testing.
func newTestApp(t *testing.T, layouts map[string]string, pages map[string]string) *httpd.App {
	t.Helper()
	dir := t.TempDir()
	layoutDir := filepath.Join(dir, "layouts")
	pageDir := filepath.Join(dir, "pages")
	os.MkdirAll(layoutDir, 0755)
	os.MkdirAll(pageDir, 0755)
	for name, content := range layouts {
		os.WriteFile(filepath.Join(layoutDir, name), []byte(content), 0644)
	}
	for name, content := range pages {
		os.WriteFile(filepath.Join(pageDir, name), []byte(content), 0644)
	}
	app, err := httpd.New(httpd.Config{LayoutDir: layoutDir, PageDir: pageDir})
	if err != nil {
		t.Fatalf("newTestApp: %v", err)
	}
	return app
}

func TestNew(t *testing.T) {
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}<h1>Test</h1>{{end}}`},
	)
	if app == nil {
		t.Fatal("expected non-nil App")
	}
}

func TestNewInvalidLayoutDir(t *testing.T) {
	_, err := httpd.New(httpd.Config{LayoutDir: "./nonexistent", PageDir: "./nonexistent"})
	if err == nil {
		t.Fatal("expected error for invalid layout dir, got nil")
	}
}

func TestRender(t *testing.T) {
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}<h1>Bonjour</h1>{{end}}`},
	)

	w := httptest.NewRecorder()
	if err := app.Render(w, "home.html", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(w.Body.String(), "<h1>Bonjour</h1>") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRenderWithData(t *testing.T) {
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}<h1>{{ .Title }}</h1>{{end}}`},
	)

	w := httptest.NewRecorder()
	err := app.Render(w, "home.html", map[string]any{"Title": "goWebFlow"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(w.Body.String(), "<h1>goWebFlow</h1>") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRenderMultiplePages(t *testing.T) {
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{
			"home.html":    `{{define "content"}}home{{end}}`,
			"contact.html": `{{define "content"}}contact{{end}}`,
		},
	)

	for _, tc := range []struct{ page, want string }{
		{"home.html", "home"},
		{"contact.html", "contact"},
	} {
		w := httptest.NewRecorder()
		if err := app.Render(w, tc.page, nil); err != nil {
			t.Fatalf("Render(%s): %v", tc.page, err)
		}
		if !strings.Contains(w.Body.String(), tc.want) {
			t.Errorf("Render(%s): expected %q, got %s", tc.page, tc.want, w.Body.String())
		}
	}
}

func TestGet(t *testing.T) {
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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
	layoutDir := filepath.Join(dir, "layouts")
	pageDir := filepath.Join(dir, "pages")
	os.MkdirAll(layoutDir, 0755)
	os.MkdirAll(pageDir, 0755)
	os.WriteFile(filepath.Join(layoutDir, "base.html"), []byte(
		`{{define "base"}}{{ vite "src/default/main.js" }}{{block "content" .}}{{end}}{{end}}`,
	), 0644)
	os.WriteFile(filepath.Join(pageDir, "home.html"), []byte(
		`{{define "content"}}home{{end}}`,
	), 0644)

	app, err := httpd.New(httpd.Config{
		LayoutDir: layoutDir,
		PageDir:   pageDir,
		StaticDir: dir,
		Dev:       true,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	w := httptest.NewRecorder()
	if err := app.Render(w, "home.html", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(w.Body.String(), "localhost:5173") {
		t.Errorf("expected Vite dev server URL, got: %s", w.Body.String())
	}
}

func TestViteProdMode(t *testing.T) {
	dir := t.TempDir()
	layoutDir := filepath.Join(dir, "layouts")
	pageDir := filepath.Join(dir, "pages")
	os.MkdirAll(layoutDir, 0755)
	os.MkdirAll(pageDir, 0755)

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

	os.WriteFile(filepath.Join(layoutDir, "base.html"), []byte(
		`{{define "base"}}{{ vite "src/main.js" }}{{block "content" .}}{{end}}{{end}}`,
	), 0644)
	os.WriteFile(filepath.Join(pageDir, "home.html"), []byte(
		`{{define "content"}}home{{end}}`,
	), 0644)

	app, err := httpd.New(httpd.Config{
		LayoutDir: layoutDir,
		PageDir:   pageDir,
		StaticDir: dir,
		StaticURL: "/static/",
		Dev:       false,
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
	layoutDir := filepath.Join(dir, "layouts")
	pageDir := filepath.Join(dir, "pages")
	os.MkdirAll(layoutDir, 0755)
	os.MkdirAll(pageDir, 0755)

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
	os.WriteFile(filepath.Join(layoutDir, "base.html"), []byte(
		`{{define "base"}}{{ vite "src/main.js" }}{{block "content" .}}{{end}}{{end}}`,
	), 0644)
	os.WriteFile(filepath.Join(pageDir, "home.html"), []byte(
		`{{define "content"}}home{{end}}`,
	), 0644)

	app, err := httpd.New(httpd.Config{
		LayoutDir: layoutDir,
		PageDir:   pageDir,
		StaticDir: dir,
		StaticURL: "/static/",
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
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

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

func TestStaticCoexistsWithWildcardRoute(t *testing.T) {
	app := newTestApp(t,
		map[string]string{"base.html": baseLayout},
		map[string]string{"home.html": `{{define "content"}}{{end}}`},
	)

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "style.css"), []byte("body{}"), 0644)
	app.Static("/assets/", dir)

	app.Get("/{phase}/{page}", func(w http.ResponseWriter, r *http.Request) error {
		phase := r.PathValue("phase")
		page := r.PathValue("page")
		w.Write([]byte(phase + "/" + page))
		return nil
	})

	for _, tc := range []struct {
		method, path string
		wantCode     int
		wantBody     string
	}{
		{"GET", "/assets/style.css", http.StatusOK, "body{}"},
		{"GET", "/foo/bar", http.StatusOK, "foo/bar"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		if w.Code != tc.wantCode {
			t.Errorf("%s %s: expected %d, got %d", tc.method, tc.path, tc.wantCode, w.Code)
		}
		if !strings.Contains(w.Body.String(), tc.wantBody) {
			t.Errorf("%s %s: expected %q, got %s", tc.method, tc.path, tc.wantBody, w.Body.String())
		}
	}
}

func TestSetThemeInvalidName(t *testing.T) {
	dir := t.TempDir()
	layoutDir := filepath.Join(dir, "layouts")
	pageDir := filepath.Join(dir, "pages")
	os.MkdirAll(layoutDir, 0755)
	os.MkdirAll(pageDir, 0755)

	viteDir := filepath.Join(dir, "default", ".vite")
	os.MkdirAll(viteDir, 0755)
	data, _ := json.Marshal(map[string]any{
		"src/main.js": map[string]any{"file": "assets/main-default.js"},
	})
	os.WriteFile(filepath.Join(viteDir, "manifest.json"), data, 0644)
	os.WriteFile(filepath.Join(layoutDir, "base.html"), []byte(
		`{{define "base"}}{{ vite "src/main.js" }}{{block "content" .}}{{end}}{{end}}`,
	), 0644)
	os.WriteFile(filepath.Join(pageDir, "home.html"), []byte(
		`{{define "content"}}home{{end}}`,
	), 0644)

	app, err := httpd.New(httpd.Config{
		LayoutDir: layoutDir,
		PageDir:   pageDir,
		StaticDir: dir,
		StaticURL: "/static/",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.SetTheme("inexistant"); err == nil {
		t.Fatal("expected error for unknown theme, got nil")
	}
}
