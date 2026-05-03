package httpd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// manifestEntry represents a single entry in the Vite manifest.json file.
type manifestEntry struct {
	File string   `json:"file"` // hashed output filename (e.g. assets/main-BqX4QSYP.js)
	CSS  []string `json:"css"`  // associated CSS files generated from this entry
}

// viteManifest holds the parsed Vite manifest and serves asset HTML tags.
type viteManifest struct {
	entries   map[string]manifestEntry
	staticURL string
	dev       bool
}

// loadManifest reads and parses the Vite manifest.json from the given staticDir.
// Vite 5+ places the manifest at <staticDir>/.vite/manifest.json.
func loadManifest(staticDir, staticURL string) (*viteManifest, error) {
	path := filepath.Join(staticDir, ".vite", "manifest.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries map[string]manifestEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return &viteManifest{entries: entries, staticURL: staticURL}, nil
}

// tag generates the HTML script and link tags for a given Vite entry point.
// In Dev mode, it points to the Vite dev server (localhost:5173).
// In production mode, it reads hashed filenames from the manifest.
func (v *viteManifest) tag(entry string) string {
	if v.dev {
		return `<script type="module" src="http://localhost:5173/@vite/client"></script>` + "\n" +
			`<script type="module" src="http://localhost:5173/` + entry + `"></script>`
	}

	e, ok := v.entries[entry]
	if !ok {
		return ""
	}

	var b strings.Builder
	for _, css := range e.CSS {
		b.WriteString(`<link rel="stylesheet" href="` + v.staticURL + css + `">` + "\n")
	}
	b.WriteString(`<script type="module" src="` + v.staticURL + e.File + `"></script>`)
	return b.String()
}
