# goWebFlow

A minimal web framework in pure Go. No dependencies outside the standard library.

## Quick start

```go
import "git.euflow.fr/flo/gowebflow/pkg/httpd"

app, err := httpd.New(httpd.Config{
    TemplateDir: "./ui/html",
})
if err != nil {
    log.Fatal(err)
}

app.Get("/", func(w http.ResponseWriter, r *http.Request) error {
    return app.Render(w, "home.html", nil)
})

log.Fatal(app.Listen(":8080"))
```

## Routes

```go
app.Get("/", handler)
app.Post("/contact", handler)
app.Put("/user", handler)
app.Delete("/user", handler)
```

## Error pages

```go
app.NotFound(func(w http.ResponseWriter, r *http.Request) error {
    w.WriteHeader(http.StatusNotFound)
    return app.Render(w, "404.html", nil)
})

app.OnError(http.StatusInternalServerError, func(w http.ResponseWriter, r *http.Request) error {
    w.WriteHeader(http.StatusNotFound) // masquer le 500 en 404 (sécurité)
    return app.Render(w, "404.html", nil)
})
```

Return a specific HTTP status from a route with `httpd.ErrHTTP`:

```go
app.Get("/admin", func(w http.ResponseWriter, r *http.Request) error {
    if !isAdmin(r) {
        return httpd.ErrHTTP(http.StatusForbidden, fmt.Errorf("forbidden"))
    }
    return app.Render(w, "admin.html", nil)
})
```

## Static files

```go
app.Static("/fonts/", "./ui/public/fonts")
```

## Passing data to templates

```go
app.Get("/", func(w http.ResponseWriter, r *http.Request) error {
    return app.Render(w, "home.html", map[string]any{
        "Title": "Accueil",
        "User":  "Flo",
    })
})
```

```html
<h1>{{ .Title }}</h1>
<p>Bonjour {{ .User }}</p>
```

## Vite integration

```go
app, err := httpd.New(httpd.Config{
    TemplateDir: "./ui/html",
    StaticDir:   "./ui/dist",   // root of all themes
    StaticURL:   "/static/",
    ActiveTheme: "default",     // serves from ui/dist/default/
    Dev:         false,         // true in development
})
```

```html
<head>
    {{ vite "src/default/main.js" }}
</head>
```

## Switching theme at runtime

```go
// Switch theme without restarting the server
app.Post("/admin/theme", func(w http.ResponseWriter, r *http.Request) error {
    return app.SetTheme(r.FormValue("name"))
})
```

See [GETTING_STARTED.md](GETTING_STARTED.md) for the full multi-theme Vite setup.

## Config

| Field         | Type   | Default     | Description                                              |
|---------------|--------|-------------|----------------------------------------------------------|
| `TemplateDir` | string | —           | Path to the HTML templates folder                        |
| `StaticDir`   | string | —           | Path to the Vite dist root folder (e.g. `./ui/dist`)     |
| `StaticURL`   | string | —           | URL prefix for static assets (e.g. `/static/`)           |
| `ActiveTheme` | string | `"default"` | Vite template subdirectory to serve (e.g. `"tpl_matrix"`) |
| `Dev`         | bool   | `false`     | Enables Vite dev server mode + template hot reload       |

## Run tests

```
go test ./pkg/httpd/
```

## Local development with another project

```
go work init ./gowebflow ./monprojet
```

See [GETTING_STARTED.md](GETTING_STARTED.md) for a full setup guide.
