package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Embeds the timezone database: the Alpine image has no system tzdata.
	_ "time/tzdata"

	"studigame/internal/auth"
	"studigame/internal/config"
	"studigame/internal/db"
	"studigame/internal/handlers"
	"studigame/internal/httpx"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	conn, err := db.Open(cfg.DatabasePath)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	deps := &handlers.Deps{
		DB:       conn,
		Sessions: auth.NewSessionStore(conn),
		Mailer:   auth.NewMailer(cfg),
		Config:   cfg,
	}

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start sweeper goroutine for cleaning up expired records
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := db.SweepExpired(ctx, conn); err != nil {
					slog.Error("Sweeper failed", "error", err)
				}
				cancel()
			case <-quit:
				return
			}
		}
	}()

	loginLimiter := httpx.NewIPRateLimiter(10, time.Minute)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Health check endpoint
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Health check", "method", r.Method, "path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /{$}", deps.Landing)
	mux.HandleFunc("POST /waitlist", loginLimiter.Middleware(handlers.RequireCSRF(deps.Waitlist)))

	mux.HandleFunc("GET /register", deps.RegisterForm)
	mux.HandleFunc("POST /register", loginLimiter.Middleware(handlers.RequireCSRF(deps.Register)))
	mux.HandleFunc("GET /verify", deps.VerifyEmail)
	mux.HandleFunc("GET /login", deps.LoginForm)
	mux.HandleFunc("POST /login", loginLimiter.Middleware(handlers.RequireCSRF(deps.Login)))
	mux.HandleFunc("POST /logout", handlers.RequireCSRF(deps.Logout))

	mux.HandleFunc("GET /reset-password", deps.ResetPasswordRequestForm)
	mux.HandleFunc("POST /reset-password", loginLimiter.Middleware(handlers.RequireCSRF(deps.ResetPasswordRequest)))
	mux.HandleFunc("GET /reset-password/confirm", deps.ResetPasswordConfirmForm)
	mux.HandleFunc("POST /reset-password/confirm", handlers.RequireCSRF(deps.ResetPasswordConfirm))

	mux.HandleFunc("GET /dashboard", deps.RequireAuth(deps.Dashboard))
	mux.HandleFunc("GET /character/new", deps.RequireAuth(deps.CharacterNewForm))
	mux.HandleFunc("POST /character/new", handlers.RequireCSRF(deps.RequireAuth(deps.CharacterNew)))

	actionLimiter := httpx.NewIPRateLimiter(30, time.Minute)
	mux.HandleFunc("GET /actions", deps.RequireAuth(deps.Actions))
	mux.HandleFunc("POST /actions/collect", actionLimiter.Middleware(handlers.RequireCSRF(deps.RequireAuth(deps.Collect))))
	mux.HandleFunc("POST /actions/job", actionLimiter.Middleware(handlers.RequireCSRF(deps.RequireAuth(deps.Job))))
	mux.HandleFunc("POST /actions/upgrade-container", actionLimiter.Middleware(handlers.RequireCSRF(deps.RequireAuth(deps.ContainerUpgrade))))

	mux.HandleFunc("GET /skills", deps.RequireAuth(deps.Skills))
	mux.HandleFunc("POST /skills/train", actionLimiter.Middleware(handlers.RequireCSRF(deps.RequireAuth(deps.TrainSkill))))
	mux.HandleFunc("POST /skills/cancel", actionLimiter.Middleware(handlers.RequireCSRF(deps.RequireAuth(deps.CancelSkill))))
	mux.HandleFunc("POST /bafog/toggle", actionLimiter.Middleware(handlers.RequireCSRF(deps.RequireAuth(deps.Bafog))))

	// Create server with timeouts
	addr := ":" + cfg.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      httpx.RecoveryMiddleware(httpx.SecurityHeadersMiddleware(mux)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Server starting", "addr", addr, "production", cfg.Production)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exiting")
}
