package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"studigame/internal/httpx"
	"studigame/internal/web"
)

func (d *Deps) Landing(w http.ResponseWriter, r *http.Request) {
	slog.Info("Landing page accessed", "method", r.Method, "path", r.URL.Path)
	web.Render(w, "landing.html", d.baseData(w, r))
}

func (d *Deps) Waitlist(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" || !strings.Contains(email, "@") {
		slog.Warn("Invalid email for waitlist", "email", email, "ip", httpx.ClientIP(r))
		http.Error(w, "ungueltige E-Mail-Adresse", http.StatusBadRequest)
		return
	}
	_, err := d.DB.Exec(`INSERT OR IGNORE INTO waitlist (email) VALUES (?)`, email)
	if err != nil {
		slog.Error("Failed to add to waitlist", "error", err, "email", email)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	slog.Info("Added to waitlist", "email", email)
	web.RenderPartial(w, "waitlist_success.html", nil)
}
