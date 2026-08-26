package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"pennergame/internal/auth"
	"pennergame/internal/web"
)

func (d *Deps) RegisterForm(w http.ResponseWriter, r *http.Request) {
	web.Render(w, "register.html", d.baseData(w, r))
}

func (d *Deps) Register(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")

	if name == "" || email == "" || len(password) < 8 {
		d.renderRegisterError(w, r, name, email, "bitte alle Felder ausfuellen (Passwort mind. 8 Zeichen)")
		return
	}

	userID, err := auth.Register(d.DB, name, email, password)
	if err != nil {
		if errors.Is(err, auth.ErrEmailTaken) {
			d.renderRegisterError(w, r, name, email, "diese E-Mail ist bereits registriert")
			return
		}
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	d.startSession(w, r, userID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (d *Deps) renderRegisterError(w http.ResponseWriter, r *http.Request, name, email, msg string) {
	data := d.baseData(w, r)
	data["Name"] = name
	data["Email"] = email
	data["Error"] = msg
	web.Render(w, "register.html", data)
}

func (d *Deps) LoginForm(w http.ResponseWriter, r *http.Request) {
	web.Render(w, "login.html", d.baseData(w, r))
}

func (d *Deps) Login(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")

	userID, err := auth.Authenticate(d.DB, email, password)
	if err != nil {
		data := d.baseData(w, r)
		data["Email"] = email
		data["Error"] = "E-Mail oder Passwort falsch"
		web.Render(w, "login.html", data)
		return
	}
	d.startSession(w, r, userID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (d *Deps) Logout(w http.ResponseWriter, r *http.Request) {
	if rawToken := auth.ReadSessionCookie(r); rawToken != "" {
		d.Sessions.Delete(rawToken)
	}
	auth.ClearSessionCookie(w, d.Config.Production)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (d *Deps) startSession(w http.ResponseWriter, r *http.Request, userID int64) {
	rawToken, err := d.Sessions.Create(userID)
	if err != nil {
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	auth.SetSessionCookie(w, rawToken, d.Config.Production)
}

func (d *Deps) ResetPasswordRequestForm(w http.ResponseWriter, r *http.Request) {
	web.Render(w, "reset_request.html", d.baseData(w, r))
}

func (d *Deps) ResetPasswordRequest(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	ttl := time.Duration(d.Config.ResetTokenTTLMin) * time.Minute
	if err := auth.RequestPasswordReset(d.DB, d.Mailer, d.Config.SiteDomain, email, ttl); err != nil {
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	data := d.baseData(w, r)
	data["Sent"] = true
	web.Render(w, "reset_request.html", data)
}

func (d *Deps) ResetPasswordConfirmForm(w http.ResponseWriter, r *http.Request) {
	data := d.baseData(w, r)
	data["Token"] = r.URL.Query().Get("token")
	web.Render(w, "reset_confirm.html", data)
}

func (d *Deps) ResetPasswordConfirm(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	password := r.FormValue("password")

	if len(password) < 8 {
		data := d.baseData(w, r)
		data["Token"] = token
		data["Error"] = "Passwort muss mind. 8 Zeichen haben"
		web.Render(w, "reset_confirm.html", data)
		return
	}

	if err := auth.ResetPassword(d.DB, token, password); err != nil {
		data := d.baseData(w, r)
		data["Token"] = token
		data["Error"] = "Link ungueltig oder abgelaufen"
		web.Render(w, "reset_confirm.html", data)
		return
	}

	data := d.baseData(w, r)
	data["Done"] = true
	web.Render(w, "reset_confirm.html", data)
}
