# goWebFlow

[![CI](https://github.com/Florian418/gowebflow/actions/workflows/ci.yml/badge.svg)](https://github.com/Florian418/gowebflow/actions/workflows/ci.yml)

Framework web minimal en Go pur. Aucune dépendance hors bibliothèque standard.

## Philosophie

Les handlers retournent une `error` — le framework l'intercepte, la logge, et appelle le bon handler d'erreur. Tu décides ce que le navigateur reçoit.

Ça veut dire que masquer un 500 ou un 403 en 404 est un choix explicite. Ne pas révéler si une ressource est interdite ou si le serveur a crashé est une décision de sécurité — goWebFlow t'en donne le contrôle.

## Démarrage rapide

```
go install github.com/Florian418/gowebflow/cmd/gwf@latest

gwf new monsite                 # layout + Air
gwf new monsite --vite          # layout + Vite + Air
gwf new monsite --no-layout     # pages autonomes + Air
gwf new monsite --no-layout --vite  # pages autonomes + Vite + Air

cd monsite && go mod tidy && air
```

→ [GETTING_STARTED.md](GETTING_STARTED.md) pour le guide complet.

---

## Routes

```go
app.Get("/", handler)
app.Post("/contact", handler)
app.Put("/user", handler)
app.Delete("/user", handler)
```

## Render

```go
return app.Render(w, "home.html", map[string]any{
    "Title": "Accueil",
    "User":  "Flo",
})
```

```html
{{define "content"}}
<h1>{{ .Title }}</h1>
{{end}}
```

## Pages d'erreur

```go
app.NotFound(func(w http.ResponseWriter, r *http.Request) error {
    w.WriteHeader(http.StatusNotFound)
    return app.Render(w, "404.html", nil)
})

app.OnError(http.StatusInternalServerError, func(w http.ResponseWriter, r *http.Request) error {
    w.WriteHeader(http.StatusNotFound)
    return app.Render(w, "404.html", nil)
})
```

Déclencher un code depuis une route :

```go
return httpd.ErrHTTP(http.StatusForbidden, fmt.Errorf("forbidden"))
```

## Vite

```go
app, err := httpd.New(httpd.Config{
    LayoutDir:   "./ui/layouts",
    PageDir:     "./ui/pages",
    StaticDir:   "./ui/dist",
    StaticURL:   "/static/",
    ActiveTheme: "default",
    Dev:         os.Getenv("APP_ENV") != "production",
})
```

```html
<head>{{ vite "src/default/main.js" }}</head>
```

## Changer de thème à chaud

```go
app.Post("/admin/theme", func(w http.ResponseWriter, r *http.Request) error {
    return app.SetTheme(r.FormValue("name"))
})
```

## Config

| Field         | Type   | Default       | Description                                          |
|---------------|--------|---------------|------------------------------------------------------|
| `LayoutDir`   | string | —             | Dossier des layouts (ex: `./ui/layouts`)             |
| `PageDir`     | string | —             | Dossier des pages (ex: `./ui/pages`)                 |
| `StaticDir`   | string | —             | Racine du dist Vite (ex: `./ui/dist`)                |
| `StaticURL`   | string | —             | Préfixe URL des assets (ex: `/static/`)              |
| `ActiveTheme` | string | `"default"`   | Sous-dossier Vite actif                              |
| `Dev`         | bool   | `false`       | Mode dev Vite + rechargement templates               |

## Tests

```
go test ./pkg/httpd/
```
