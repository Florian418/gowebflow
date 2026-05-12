# Migration guide

## v0.3.0 — CLI de scaffolding `gwf` (aucun breaking change)

### Ce qui change

Ajout du binaire `gwf` dans `cmd/gwf`. L'API du framework est inchangée.

```
go install github.com/Florian418/gowebflow/cmd/gwf@latest

gwf new monsite          # projet de base (Go + Air)
gwf new monsite --vite   # avec Vite + Air
```

### Projets existants

Aucune migration nécessaire. Le CLI ne sert qu'à créer de nouveaux projets.

---

## v0.2.0 — Layout+block rendering (breaking)

### Ce qui change

Le système de rendu passe d'un glob unique à un modèle layout+page.
L'ancienne approche (tous les `.html` dans un seul dossier, un seul `*template.Template` global)
était incompatible avec `{{define "content"}}` dès qu'on avait plusieurs pages.

---

### 1. Config — champs renommés

```go
// Avant (v0.1.0)
httpd.Config{
    TemplateDir: "./ui/html",
}

// Après (v0.2.0)
httpd.Config{
    LayoutDir: "./ui/layouts",
    PageDir:   "./ui/pages",
}
```

`TemplateDir` est supprimé. Remplace-le par `LayoutDir` + `PageDir`.

---

### 2. Structure des dossiers à créer

```
ui/
  layouts/        ← nouveau (était : ui/html/ à plat)
    base.html
  pages/          ← nouveau
    home.html
    contact.html
    404.html
  assets/
```

---

### 3. Format des templates

**Layout (`ui/layouts/base.html`)** — doit définir `"base"` et déclarer le block `"content"` :

```html
{{define "base"}}
<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8">
  <title>{{ .Title }}</title>
  {{ vite "src/main.js" }}
</head>
<body>
  {{block "content" .}}{{end}}
</body>
</html>
{{end}}
```

**Page (`ui/pages/home.html`)** — doit surcharger `"content"` :

```html
{{define "content"}}
<h1>{{ .Title }}</h1>
<p>Bienvenue</p>
{{end}}
```

> Avant (v0.1.0), les templates étaient des fichiers HTML complets nommés
> directement (ex: `home.html` contenait tout le `<!DOCTYPE html>`).
> Maintenant chaque page ne contient que son bloc de contenu.

---

### 4. Appel à `Render` — inchangé

```go
// Identique avant et après
return app.Render(w, "home.html", data)
```

Le nom passé est le fichier dans `PageDir`. Le layout `"base"` est toujours exécuté automatiquement.

---

### 5. Checklist de migration par projet

- [ ] Remplacer `TemplateDir` par `LayoutDir` + `PageDir` dans `httpd.New()`
- [ ] Créer `ui/layouts/base.html` avec `{{define "base"}}` et `{{block "content" .}}`
- [ ] Créer `ui/pages/` et y déplacer les anciennes pages
- [ ] Dans chaque page : supprimer le HTML shell, ne garder que `{{define "content"}}...{{end}}`
- [ ] Vérifier que `app.Render(w, "home.html", data)` pointe bien sur un fichier dans `PageDir`
- [ ] Lancer `go build ./...` et vérifier qu'il n'y a plus de `unknown field TemplateDir`
