package main

import (
	"log"
	"net/http"
	"time"

	"pennergame/internal/auth"
	"pennergame/internal/config"
	"pennergame/internal/db"
	"pennergame/internal/handlers"
	"pennergame/internal/httpx"
)

func main() {
	cfg := config.Load()

	conn, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer conn.Close()

	deps := &handlers.Deps{
		DB:       conn,
		Sessions: auth.NewSessionStore(conn),
		Mailer:   auth.ConsoleMailer{},
		Config:   cfg,
	}

	loginLimiter := httpx.NewIPRateLimiter(10, time.Minute)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /{$}", deps.Landing)
	mux.HandleFunc("POST /waitlist", deps.Waitlist)

	mux.HandleFunc("GET /register", deps.RegisterForm)
	mux.HandleFunc("POST /register", loginLimiter.Middleware(handlers.RequireCSRF(deps.Register)))
	mux.HandleFunc("GET /login", deps.LoginForm)
	mux.HandleFunc("POST /login", loginLimiter.Middleware(handlers.RequireCSRF(deps.Login)))
	mux.HandleFunc("POST /logout", handlers.RequireCSRF(deps.Logout))

	mux.HandleFunc("GET /reset-password", deps.ResetPasswordRequestForm)
	mux.HandleFunc("POST /reset-password", loginLimiter.Middleware(handlers.RequireCSRF(deps.ResetPasswordRequest)))
	mux.HandleFunc("GET /reset-password/confirm", deps.ResetPasswordConfirmForm)
	mux.HandleFunc("POST /reset-password/confirm", handlers.RequireCSRF(deps.ResetPasswordConfirm))

	mux.HandleFunc("GET /dashboard", deps.RequireAuth(deps.Dashboard))

	addr := ":" + cfg.Port
	log.Printf("%s listening on %s (production=%v)", cfg.SiteName, addr, cfg.Production)
	log.Fatal(http.ListenAndServe(addr, mux))
}
