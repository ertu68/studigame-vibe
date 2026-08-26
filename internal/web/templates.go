// Package web renders html/template templates embedded at build time.
package web

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed all:templates
var templateFS embed.FS

// Each page is parsed together with the shared layout+partials as its own
// template set, so every page's {{define "content"}} block stays scoped to
// that page instead of overwriting the others in a shared namespace.
var pages = []string{
	"landing.html",
	"register.html",
	"login.html",
	"reset_request.html",
	"reset_confirm.html",
	"dashboard.html",
}

var pageTemplates = mustParsePages()

func mustParsePages() map[string]*template.Template {
	out := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		out[page] = template.Must(template.ParseFS(templateFS,
			"templates/layout.html",
			"templates/partials/*.html",
			"templates/"+page,
		))
	}
	return out
}

func Render(w http.ResponseWriter, page string, data any) {
	t, ok := pageTemplates[page]
	if !ok {
		http.Error(w, "unknown template: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var partialTemplates = template.Must(template.ParseFS(templateFS, "templates/partials/*.html"))

// RenderPartial renders a standalone partial (used for htmx fragment responses).
func RenderPartial(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := partialTemplates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
