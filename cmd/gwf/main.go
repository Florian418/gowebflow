package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const gowebflowVersion = "v0.2.1"

type genFile struct {
	path    string
	content string
}

func main() {
	if len(os.Args) < 3 || os.Args[1] != "new" {
		fmt.Fprintln(os.Stderr, "Usage: gwf new <project-name> [--vite] [--no-layout]")
		os.Exit(1)
	}

	sub := flag.NewFlagSet("new", flag.ExitOnError)
	withVite := sub.Bool("vite", false, "add Vite + Air scaffolding")
	noLayout := sub.Bool("no-layout", false, "pages are standalone HTML, no shared layout")
	sub.Parse(os.Args[3:])

	if err := scaffoldAt(".", os.Args[2], *withVite, *noLayout); err != nil {
		fmt.Fprintf(os.Stderr, "gwf: %v\n", err)
		os.Exit(1)
	}
}

func scaffoldAt(baseDir, name string, withVite, noLayout bool) error {
	root := filepath.Join(baseDir, name)

	if _, err := os.Stat(root); !os.IsNotExist(err) {
		return fmt.Errorf("directory %q already exists", name)
	}

	dirs := []string{
		filepath.Join(root, "cmd", "server"),
		filepath.Join(root, "ui", "pages"),
	}
	if !noLayout {
		dirs = append(dirs, filepath.Join(root, "ui", "layouts"))
	}
	if withVite {
		dirs = append(dirs, filepath.Join(root, "ui", "src", "default"))
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create dir: %w", err)
		}
	}

	for _, f := range buildFiles(name, withVite, noLayout) {
		path := filepath.Join(root, f.path)
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	printSuccess(name, withVite)
	return nil
}

func buildFiles(name string, withVite, noLayout bool) []genFile {
	files := []genFile{
		{"go.mod", goMod(name)},
		{"cmd/server/main.go", mainGo(withVite, noLayout)},
		{"ui/pages/home.html", homePage(withVite, noLayout)},
		{".air.toml", airToml()},
		{".gitignore", gitignore(withVite)},
	}
	if !noLayout {
		files = append(files, genFile{"ui/layouts/base.html", baseLayout(withVite)})
	}
	if withVite {
		files = append(files,
			genFile{"ui/vite.config.js", viteConfig()},
			genFile{"ui/package.json", packageJSON(name)},
			genFile{"ui/src/default/main.js", mainJS()},
			genFile{"ui/src/default/style.css", styleCSS()},
		)
	}
	return files
}

func goMod(name string) string {
	return fmt.Sprintf("module %s\n\ngo 1.22\n\nrequire github.com/Florian418/gowebflow %s\n", name, gowebflowVersion)
}

func mainGo(withVite, noLayout bool) string {
	imports := `"log"
	"net/http"`
	if withVite {
		imports += `
	"os"`
	}

	var config string
	if !noLayout {
		config += "\n\t\tLayoutDir: \"./ui/layouts\","
	}
	config += "\n\t\tPageDir:   \"./ui/pages\","
	if withVite {
		config += `
		StaticDir:   "./ui/dist",
		StaticURL:   "/static/",
		ActiveTheme: "default",
		Dev:         os.Getenv("APP_ENV") != "production",`
	}

	return fmt.Sprintf(`package main

import (
	%s

	"github.com/Florian418/gowebflow/pkg/httpd"
)

func main() {
	app, err := httpd.New(httpd.Config{%s
	})
	if err != nil {
		log.Fatal(err)
	}

	app.Get("/{$}", func(w http.ResponseWriter, r *http.Request) error {
		return app.Render(w, "home.html", map[string]any{
			"Title": "Accueil",
		})
	})

	log.Println("Serveur démarré sur http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
`, imports, config)
}

func baseLayout(withVite bool) string {
	viteTag := ""
	if withVite {
		viteTag = "\n    {{ vite \"src/default/main.js\" }}"
	}
	return fmt.Sprintf(`{{define "base"}}
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}Mon site{{end}}</title>%s
</head>
<body>
    {{block "content" .}}{{end}}
</body>
</html>
{{end}}
`, viteTag)
}

func homePage(withVite, noLayout bool) string {
	if noLayout {
		viteTag := ""
		if withVite {
			viteTag = "\n    {{ vite \"src/default/main.js\" }}"
		}
		return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Accueil</title>%s
</head>
<body>
    <h1>Bienvenue</h1>
    <p>Bienvenue sur votre nouveau site goWebFlow.</p>
</body>
</html>
`, viteTag)
	}
	return `{{define "content"}}
<h1>{{ .Title }}</h1>
<p>Bienvenue sur votre nouveau site goWebFlow.</p>
{{end}}
`
}

func airToml() string {
	bin := "tmp/main"
	if runtime.GOOS == "windows" {
		bin = "tmp/main.exe"
	}
	return fmt.Sprintf(`root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./%s ./cmd/server"
  bin = "%s"
  include_ext = ["go", "html"]
  exclude_dir = ["tmp", "node_modules", "ui/dist"]
  delay = 1000

[screen]
  clear_on_rebuild = false
`, bin, bin)
}

func gitignore(withVite bool) string {
	s := "tmp/\n"
	if withVite {
		s += "ui/dist/\nui/node_modules/\n"
	}
	return s
}

func viteConfig() string {
	bt := "`"
	return `import { defineConfig } from 'vite'

const tpl = process.env.TPL || 'default'

export default defineConfig(({ command }) => ({
  base: command === 'build' ? '/static/' : '/',
  plugins: [
    {
      name: 'watch-go-templates',
      handleHotUpdate({ file, server }) {
        if (file.endsWith('.html')) {
          server.ws.send({ type: 'full-reload' })
        }
      }
    }
  ],
  server: {
    proxy: {
      '/': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        bypass(req) {
          const u = req.url ?? ''
          if (u.startsWith('/@') || u.startsWith('/node_modules') || u.startsWith('/src/')) {
            return u
          }
        }
      }
    }
  },
  build: {
    manifest: true,
    outDir: ` + bt + `dist/${tpl}` + bt + `,
    rollupOptions: {
      input: ` + bt + `src/${tpl}/main.js` + bt + `
    }
  }
}))
`
}

func packageJSON(name string) string {
	return fmt.Sprintf(`{
  "name": %q,
  "private": true,
  "version": "0.0.0",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "build:default": "cross-env TPL=default vite build",
    "build:all": "npm run build:default"
  },
  "devDependencies": {
    "cross-env": "^7.0.3",
    "vite": "^5.0.0"
  }
}
`, name)
}

func mainJS() string {
	return "import './style.css'\n"
}

func styleCSS() string {
	return `* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: system-ui, sans-serif;
  line-height: 1.6;
  padding: 2rem;
  color: #1a1a1a;
}

h1 {
  margin-bottom: 1rem;
}
`
}

func printSuccess(name string, withVite bool) {
	if withVite {
		fmt.Printf(`
Projet "%s" créé avec Vite !

  cd %s
  go mod tidy

  Terminal 1 — Vite :
    cd ui && npm install && npm run dev

  Terminal 2 — Go (air) :
    go install github.com/air-verse/air@latest
    air

  → http://localhost:5173
`, name, name)
		return
	}
	fmt.Printf(`
Projet "%s" créé !

  cd %s
  go mod tidy

  Démarrer avec air (hot reload) :
    go install github.com/air-verse/air@latest
    air

  Ou directement :
    go run .

  → http://localhost:8080
`, name, name)
}
