package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"pennergame/internal/auth"
	"pennergame/internal/web"
)

func (d *Deps) RegisterForm(w http.ResponseWriter, r *http.Request) {
	data := d.baseData(w, r)
	data["Cities"] = auth.UniCities
	web.Render(w, "register.html", data)
}

func (d *Deps) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		slog.Warn("Verify email called without token")
		http.Error(w, "ungültiger Link", http.StatusBadRequest)
		return
	}

	if err := auth.VerifyEmail(d.DB, token); err != nil {
		if errors.Is(err, auth.ErrInvalidVerifyLink) {
			slog.Info("Email verification attempt with invalid token")
		} else {
			slog.Error("Email verification failed with unexpected error", "error", err)
		}
		data := d.baseData(w, r)
		data["Error"] = "Link ungueltig oder abgelaufen"
		web.Render(w, "verify_confirm.html", data)
		return
	}

	slog.Info("Email verified successfully")
	data := d.baseData(w, r)
	data["Success"] = true
	web.Render(w, "verify_confirm.html", data)
}

func (d *Deps) Register(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")
	city := r.FormValue("city")

	if name == "" || email == "" || len(password) < 8 || !slices.Contains(auth.UniCities, city) {
		slog.Warn("Registration attempt with invalid data", "name", name, "email", email)
		d.renderRegisterError(w, r, name, email, city, "bitte alle Felder ausfuellen (Passwort mind. 8 Zeichen)")
		return
	}

	userID, err := auth.Register(d.DB, name, email, password, city)
	if err != nil {
		if errors.Is(err, auth.ErrEmailTaken) {
			slog.Info("Registration attempt with existing email", "email", email)
			d.renderRegisterError(w, r, name, email, city, "diese E-Mail ist bereits registriert")
			return
		}
		slog.Error("Registration failed with unexpected error", "error", err, "email", email)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	// Send verification email instead of auto-login
	if err := auth.RequestEmailVerification(d.DB, d.Mailer, d.Config.SiteDomain, userID, email, time.Hour); err != nil {
		slog.Error("Failed to send verification email", "error", err, "userID", userID, "email", email)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("User registered, verification email sent", "userID", userID, "email", email)
	data := d.baseData(w, r)
	data["Email"] = email
	web.Render(w, "register_verify.html", data)
}

func (d *Deps) renderRegisterError(w http.ResponseWriter, r *http.Request, name, email, city, msg string) {
	data := d.baseData(w, r)
	data["Name"] = name
	data["Email"] = email
	data["City"] = city
	data["Cities"] = auth.UniCities
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
		if errors.Is(err, auth.ErrUnverifiedEmail) {
			slog.Info("Unverified user login attempt", "email", email)
			data := d.baseData(w, r)
			data["Email"] = email
			data["Error"] = "bitte bestätige zuerst deine E-Mail"
			web.Render(w, "login.html", data)
			return
		}
		if errors.Is(err, auth.ErrInvalidLogin) {
			slog.Info("Failed login attempt", "email", email)
		} else {
			slog.Error("Authentication failed with unexpected error", "error", err, "email", email)
		}
		data := d.baseData(w, r)
		data["Email"] = email
		data["Error"] = "E-Mail oder Passwort falsch"
		web.Render(w, "login.html", data)
		return
	}
	slog.Info("User logged in successfully", "userID", userID, "email", email)
	d.startSession(w, r, userID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (d *Deps) Logout(w http.ResponseWriter, r *http.Request) {
	if rawToken := auth.ReadSessionCookie(r); rawToken != "" {
		err := d.Sessions.Delete(rawToken)
		if err != nil {
			slog.Warn("Failed to delete session", "error", err)
		}
	}
	auth.ClearSessionCookie(w, d.Config.Production)
	slog.Info("User logged out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (d *Deps) startSession(w http.ResponseWriter, r *http.Request, userID int64) {
	rawToken, err := d.Sessions.Create(userID)
	if err != nil {
		slog.Error("Failed to create session", "error", err, "userID", userID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	auth.SetSessionCookie(w, rawToken, d.Config.Production)
	slog.Info("Session started", "userID", userID)
}

func (d *Deps) ResetPasswordRequestForm(w http.ResponseWriter, r *http.Request) {
	web.Render(w, "reset_request.html", d.baseData(w, r))
}

func (d *Deps) ResetPasswordRequest(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	ttl := time.Duration(d.Config.ResetTokenTTLMin) * time.Minute
	if err := auth.RequestPasswordReset(d.DB, d.Mailer, d.Config.SiteDomain, email, ttl); err != nil {
		slog.Error("Password reset request failed", "error", err, "email", email)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	slog.Info("Password reset requested", "email", email)
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
		slog.Warn("Password reset attempt with short password", "token", token)
		data := d.baseData(w, r)
		data["Token"] = token
		data["Error"] = "Passwort muss mind. 8 Zeichen haben"
		web.Render(w, "reset_confirm.html", data)
		return
	}

	if err := auth.ResetPassword(d.DB, token, password); err != nil {
		if errors.Is(err, auth.ErrInvalidResetLink) {
			slog.Info("Password reset attempt with invalid token", "token", token)
		} else {
			slog.Error("Password reset failed with unexpected error", "error", err)
		}
		data := d.baseData(w, r)
		data["Token"] = token
		data["Error"] = "Link ungueltig oder abgelaufen"
		web.Render(w, "reset_confirm.html", data)
		return
	}

	slog.Info("Password reset successful")
	data := d.baseData(w, r)
	data["Done"] = true
	web.Render(w, "reset_confirm.html", data)
}
