# Getting Started with goWebFlow

## Prérequis

- Go 1.22+
- Node.js 18+ (uniquement avec Vite)

---

## 1. Installer gwf

`gwf` est le CLI de goWebFlow. Il génère la structure complète d'un nouveau projet.

```
go install github.com/Florian418/gowebflow/cmd/gwf@latest
```

---

## 2. Créer un projet

```
gwf new monsite                      # layout + Air
gwf new monsite --vite               # layout + Vite + Air
gwf new monsite --no-layout          # pages autonomes + Air
gwf new monsite --no-layout --vite   # pages autonomes + Vite + Air
```

| Flag | Effet |
|---|---|
| _(aucun)_ | Layout partagé (`ui/layouts/base.html`) + Air |
| `--vite` | + Vite pour CSS/JS avec hot reload |
| `--no-layout` | Pages HTML complètes, pas de layout partagé |

---

## 3. Démarrer

```
cd monsite
go mod tidy
air
```

Ouvrir http://localhost:8080.

Installer air si nécessaire :
```
go install github.com/air-verse/air@latest
```

Sans air :
```
go run .
```

---

## 4. Structure générée

### Sans Vite

```
monsite/
├── main.go
├── go.mod
├── .air.toml
├── .gitignore
└── ui/
    ├── layouts/
    │   └── base.html       ← shell HTML commun à toutes les pages
    └── pages/
        └── home.html       ← contenu de chaque page
```

### Avec --vite

```
monsite/
├── main.go
├── go.mod
├── .air.toml
├── .gitignore
└── ui/
    ├── layouts/
    │   └── base.html
    ├── pages/
    │   └── home.html
    ├── src/
    │   └── default/
    │       ├── main.js
    │       └── style.css
    ├── vite.config.js
    └── package.json
```

---

## 5. Ajouter des routes

```go
app.Get("/about", func(w http.ResponseWriter, r *http.Request) error {
    return app.Render(w, "about.html", map[string]any{
        "Title": "À propos",
    })
})
```

Créer la page correspondante :

```html
<!-- ui/pages/about.html -->
{{define "content"}}
<h1>{{ .Title }}</h1>
{{end}}
```

Méthodes disponibles : `Get`, `Post`, `Put`, `Delete`.

---

## 6. Passer des données aux templates

```go
app.Get("/user", func(w http.ResponseWriter, r *http.Request) error {
    return app.Render(w, "user.html", map[string]any{
        "Name":  "Flo",
        "Admin": true,
    })
})
```

```html
{{define "content"}}
<h1>Bonjour {{ .Name }}</h1>
{{ if .Admin }}<p>Vous êtes administrateur</p>{{ end }}
{{end}}
```

---

## 7. Avec Vite

### Workflow de dev

Vite proxie toutes les requêtes vers le serveur Go — ouvrir uniquement `http://localhost:5173`.

**Terminal 1 — Vite :**
```
cd ui
npm install   # une seule fois
npm run dev
```

**Terminal 2 — Go :**
```
air
```

- Modifier un `.go` → air recompile et redémarre le serveur
- Modifier un `.html` → rechargement automatique du navigateur
- Modifier un `.css` / `.js` → HMR Vite sans rechargement de page

### Build de production

```
cd ui
npm run build:all
```

Les assets sont générés dans `ui/dist/default/`.

En production, définir `APP_ENV=production` pour que le serveur serve les fichiers buildés :
```
APP_ENV=production go run .
```

En dev (`APP_ENV` non défini), le serveur pointe vers le dev server Vite.

### Détail de vite.config.js

Le proxy redirige toutes les requêtes vers Go, sauf les assets Vite (`/@`, `/node_modules`, `/src/`).
Le build génère un manifest (`dist/<theme>/.vite/manifest.json`) que goWebFlow lit pour injecter les URLs hachées.

---

## 8. Multi-thème

Chaque thème est un sous-dossier de `ui/src/` avec son propre `main.js`.

### Ajouter un thème

1. Créer `ui/src/<nom>/main.js` (et `style.css`)
2. Ajouter dans `ui/package.json` :
   ```json
   "build:<nom>": "cross-env TPL=<nom> vite build"
   ```
   Et ajouter `&& npm run build:<nom>` à `build:all`.
3. Utiliser `{{ vite "src/<nom>/main.js" }}` dans le layout.

### Changer de thème à chaud

```go
app.Post("/admin/theme", func(w http.ResponseWriter, r *http.Request) error {
    return app.SetTheme(r.FormValue("name"))
})
```

`SetTheme` swaps le manifest actif et le handler de fichiers statiques instantanément, sans redémarrage.

---

## 9. Pages d'erreur

```go
// 404 — page non trouvée
app.NotFound(func(w http.ResponseWriter, r *http.Request) error {
    w.WriteHeader(http.StatusNotFound)
    return app.Render(w, "404.html", r.URL.Path)
})

// 500 masqué en 404 (le client ne sait pas si le serveur a crashé)
app.OnError(http.StatusInternalServerError, func(w http.ResponseWriter, r *http.Request) error {
    w.WriteHeader(http.StatusNotFound)
    return app.Render(w, "404.html", nil)
})

// 403 masqué en 404 (le client ne sait pas si la ressource existe)
app.OnError(http.StatusForbidden, func(w http.ResponseWriter, r *http.Request) error {
    w.WriteHeader(http.StatusNotFound)
    return app.Render(w, "404.html", nil)
})
```

Déclencher un code HTTP spécifique depuis une route avec `httpd.ErrHTTP` :

```go
app.Get("/admin", func(w http.ResponseWriter, r *http.Request) error {
    if !isAdmin(r) {
        return httpd.ErrHTTP(http.StatusForbidden, fmt.Errorf("forbidden"))
    }
    return app.Render(w, "admin.html", nil)
})
```

`ErrHTTP` signale au framework quel handler `OnError` appeler. Le handler décide de ce que le navigateur reçoit — y compris masquer le vrai code HTTP.

---

## 10. Page sans layout

Pour un projet ultra-simple sans shell HTML partagé, omettre `LayoutDir` :

```go
app, err := httpd.New(httpd.Config{
    PageDir: "./ui/pages",
})
```

Chaque page est un fichier HTML complet — **pas de `{{define}}`**, le framework exécute le template directement par son nom de fichier :

```html
<!-- ui/pages/home.html -->
<!DOCTYPE html>
<html lang="fr">
<head><meta charset="UTF-8"><title>Mon site</title></head>
<body>
    <h1>Bienvenue</h1>
</body>
</html>
```

À comparer avec le mode layout où `base.html` doit définir le template nommé `"base"` :

```html
<!-- ui/layouts/base.html -->
{{define "base"}}
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>{{block "title" .}}Mon site{{end}}</title>
</head>
<body>
    {{block "content" .}}{{end}}
</body>
</html>
{{end}}
```

Et chaque page surcharge le bloc `"content"` :

```html
<!-- ui/pages/home.html -->
{{define "content"}}
<h1>{{ .Title }}</h1>
{{end}}
```
