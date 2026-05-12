package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("fichier manquant : %s", path)
	}
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("fichier ne devrait pas exister : %s", path)
	}
}

func mustContain(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("impossible de lire %s : %v", path, err)
	}
	if !strings.Contains(string(data), want) {
		t.Errorf("%s : attendu %q, non trouvé\ncontenu :\n%s", path, want, data)
	}
}

func mustNotContain(t *testing.T, path, unwanted string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("impossible de lire %s : %v", path, err)
	}
	if strings.Contains(string(data), unwanted) {
		t.Errorf("%s : %q ne devrait pas être présent", path, unwanted)
	}
}

func TestScaffoldBasic(t *testing.T) {
	base := t.TempDir()

	if err := scaffoldAt(base, "monsite", false, false); err != nil {
		t.Fatalf("scaffoldAt: %v", err)
	}

	root := filepath.Join(base, "monsite")

	for _, f := range []string{
		"go.mod",
		filepath.Join("cmd", "server", "main.go"),
		".air.toml",
		".gitignore",
		filepath.Join("ui", "layouts", "base.html"),
		filepath.Join("ui", "pages", "home.html"),
	} {
		mustExist(t, filepath.Join(root, f))
	}

	for _, f := range []string{
		filepath.Join("ui", "vite.config.js"),
		filepath.Join("ui", "package.json"),
		filepath.Join("ui", "src", "default", "main.js"),
	} {
		mustNotExist(t, filepath.Join(root, f))
	}

	mustContain(t, filepath.Join(root, "go.mod"), "module monsite")
	mustContain(t, filepath.Join(root, "go.mod"), "require github.com/Florian418/gowebflow")

	mainGo := filepath.Join(root, "cmd", "server", "main.go")
	mustContain(t, mainGo, "package main")
	mustContain(t, mainGo, "LayoutDir")
	mustNotContain(t, mainGo, `os.Getenv`)

	mustContain(t, filepath.Join(root, "ui", "layouts", "base.html"), `{{define "base"}}`)
	mustNotContain(t, filepath.Join(root, "ui", "layouts", "base.html"), "vite")

	mustContain(t, filepath.Join(root, "ui", "pages", "home.html"), `{{define "content"}}`)
}

func TestScaffoldVite(t *testing.T) {
	base := t.TempDir()

	if err := scaffoldAt(base, "monsite", true, false); err != nil {
		t.Fatalf("scaffoldAt: %v", err)
	}

	root := filepath.Join(base, "monsite")

	for _, f := range []string{
		"go.mod",
		filepath.Join("cmd", "server", "main.go"),
		".air.toml",
		".gitignore",
		filepath.Join("ui", "layouts", "base.html"),
		filepath.Join("ui", "pages", "home.html"),
		filepath.Join("ui", "vite.config.js"),
		filepath.Join("ui", "package.json"),
		filepath.Join("ui", "src", "default", "main.js"),
		filepath.Join("ui", "src", "default", "style.css"),
	} {
		mustExist(t, filepath.Join(root, f))
	}

	mustContain(t, filepath.Join(root, "go.mod"), "module monsite")
	mustContain(t, filepath.Join(root, "cmd", "server", "main.go"), `os.Getenv("APP_ENV")`)
	mustContain(t, filepath.Join(root, "cmd", "server", "main.go"), "LayoutDir")

	mustContain(t, filepath.Join(root, "ui", "layouts", "base.html"), `{{define "base"}}`)
	mustContain(t, filepath.Join(root, "ui", "layouts", "base.html"), "vite")

	mustContain(t, filepath.Join(root, "ui", "pages", "home.html"), `{{define "content"}}`)
	mustContain(t, filepath.Join(root, "ui", "package.json"), `"name": "monsite"`)
	mustContain(t, filepath.Join(root, ".gitignore"), "ui/dist/")
}

func TestScaffoldNoLayout(t *testing.T) {
	base := t.TempDir()

	if err := scaffoldAt(base, "monsite", false, true); err != nil {
		t.Fatalf("scaffoldAt: %v", err)
	}

	root := filepath.Join(base, "monsite")

	for _, f := range []string{
		"go.mod",
		filepath.Join("cmd", "server", "main.go"),
		".air.toml",
		".gitignore",
		filepath.Join("ui", "pages", "home.html"),
	} {
		mustExist(t, filepath.Join(root, f))
	}

	// pas de layouts/
	mustNotExist(t, filepath.Join(root, "ui", "layouts"))

	// main.go sans LayoutDir
	mustNotContain(t, filepath.Join(root, "cmd", "server", "main.go"), "LayoutDir")

	// home.html est une page HTML complète, pas un bloc {{define}}
	mustContain(t, filepath.Join(root, "ui", "pages", "home.html"), "<!DOCTYPE html>")
	mustNotContain(t, filepath.Join(root, "ui", "pages", "home.html"), `{{define "content"}}`)
}

func TestScaffoldNoLayoutVite(t *testing.T) {
	base := t.TempDir()

	if err := scaffoldAt(base, "monsite", true, true); err != nil {
		t.Fatalf("scaffoldAt: %v", err)
	}

	root := filepath.Join(base, "monsite")

	// pas de layouts/
	mustNotExist(t, filepath.Join(root, "ui", "layouts"))

	// fichiers Vite présents
	for _, f := range []string{
		filepath.Join("ui", "vite.config.js"),
		filepath.Join("ui", "package.json"),
		filepath.Join("ui", "src", "default", "main.js"),
	} {
		mustExist(t, filepath.Join(root, f))
	}

	// home.html : HTML complet avec tag vite, sans {{define}}
	mustContain(t, filepath.Join(root, "ui", "pages", "home.html"), "<!DOCTYPE html>")
	mustContain(t, filepath.Join(root, "ui", "pages", "home.html"), "vite")
	mustNotContain(t, filepath.Join(root, "ui", "pages", "home.html"), `{{define "content"}}`)

	// main.go sans LayoutDir, avec APP_ENV
	mustNotContain(t, filepath.Join(root, "cmd", "server", "main.go"), "LayoutDir")
	mustContain(t, filepath.Join(root, "cmd", "server", "main.go"), `os.Getenv("APP_ENV")`)
}

func TestScaffoldModuleName(t *testing.T) {
	base := t.TempDir()

	if err := scaffoldAt(base, "myapp", false, false); err != nil {
		t.Fatalf("scaffoldAt: %v", err)
	}

	root := filepath.Join(base, "myapp")
	mustContain(t, filepath.Join(root, "go.mod"), "module myapp")
	mustExist(t, filepath.Join(root, "cmd", "server", "main.go"))
}

func TestScaffoldDirAlreadyExists(t *testing.T) {
	base := t.TempDir()

	if err := os.Mkdir(filepath.Join(base, "existe"), 0755); err != nil {
		t.Fatal(err)
	}

	err := scaffoldAt(base, "existe", false, false)
	if err == nil {
		t.Fatal("attendu une erreur, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("message d'erreur inattendu : %v", err)
	}
}
