package handlers

import (
	"net/http"

	"pennergame/internal/web"
)

func (d *Deps) Dashboard(u *user, w http.ResponseWriter, r *http.Request) {
	data := d.baseData(w, r)
	web.Render(w, "dashboard.html", data)
}
