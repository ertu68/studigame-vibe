package handlers

import (
	"log/slog"
	"net/http"

	"pennergame/internal/web"
)

func (d *Deps) Dashboard(u *user, w http.ResponseWriter, r *http.Request) {
	slog.Info("Dashboard accessed", "userID", u.ID, "userName", u.Name)
	data := d.baseData(w, r)
	web.Render(w, "dashboard.html", data)
}
