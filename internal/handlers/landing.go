package handlers

import (
	"net/http"
	"strings"

	"pennergame/internal/web"
)

func (d *Deps) Landing(w http.ResponseWriter, r *http.Request) {
	web.Render(w, "landing.html", d.baseData(w, r))
}

func (d *Deps) Waitlist(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" || !strings.Contains(email, "@") {
		http.Error(w, "ungueltige E-Mail-Adresse", http.StatusBadRequest)
		return
	}
	_, err := d.DB.Exec(`INSERT OR IGNORE INTO waitlist (email) VALUES (?)`, email)
	if err != nil {
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	web.RenderPartial(w, "waitlist_success.html", nil)
}
