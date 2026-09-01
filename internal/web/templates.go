// Package web renders html/template templates embedded at build time.
package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
)

//go:embed all:templates all:components
var templateFS embed.FS

// Each page is parsed together with the shared layout+partials as its own
// template set, so every page's {{define "content"}} block stays scoped to
// that page instead of overwriting the others in a shared namespace.
var pages = []string{
	"landing.html",
	"register.html",
	"register_verify.html",
	"verify_confirm.html",
	"login.html",
	"reset_request.html",
	"reset_confirm.html",
	"character_new.html",
	"dashboard.html",
	"actions.html",
	"skills.html",
	"klausur.html",
	"shop.html",
}

var pageTemplates = mustParsePages()

func mustParsePages() map[string]*template.Template {
	out := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		tmpl := template.New("").Funcs(template.FuncMap{
			"dict":    dict,
			"cents":   FormatCents,
			"icon":    icon,
			"percent": percentOf,
		})
		out[page] = template.Must(tmpl.ParseFS(templateFS,
			"templates/layout.html",
			"templates/partials/*.html",
			"components/*.html",
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

var partialTemplates = template.Must(template.New("").Funcs(template.FuncMap{
	"dict":  dict,
	"cents": FormatCents,
	"icon":  icon,
}).ParseFS(templateFS, "templates/partials/*.html", "components/*.html"))

// RenderPartial renders a standalone partial (used for htmx fragment responses).
func RenderPartial(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := partialTemplates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// dict creates a map from a list of key-value pairs.
func dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict: odd number of arguments: %d", len(values))
	}

	result := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: key argument %d must be a string", i)
		}
		result[key] = values[i+1]
	}
	return result, nil
}

// FormatCents renders cents as German currency, e.g. 1234 -> "12,34 €".
// It is the single formatter for money, used by Go handlers and templates.
func FormatCents(cents int) string {
	return fmt.Sprintf("%d,%02d \u20ac", cents/100, cents%100)
}

// percentOf is part/whole as a percentage for progress bars, e.g. skill level
// 3 of 20 -> 15. Guards against a zero whole so a template never divides by 0.
func percentOf(part, whole int) int {
	if whole <= 0 {
		return 0
	}
	return part * 100 / whole
}
