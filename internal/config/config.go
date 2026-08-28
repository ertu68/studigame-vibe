// Package config loads runtime settings from environment variables.
// No config file, no flags library — env vars only, matching a single-VPS deploy.
package config

import "os"

type Config struct {
	Port             string
	DatabasePath     string
	SiteName         string
	SiteDomain       string
	MailFrom         string
	SMTPHost         string
	SMTPPort         string
	SMTPUser         string
	SMTPPass         string
	ResetTokenTTLMin int
	Production       bool
}

func Load() Config {
	env := getenv("APP_ENV", "development")
	return Config{
		Port:             getenv("PORT", "8080"),
		DatabasePath:     getenv("DATABASE_PATH", "pennergame.db"),
		SiteName:         getenv("SITE_NAME", "TODO-Name"),
		SiteDomain:       getenv("SITE_DOMAIN", "localhost:8080"),
		MailFrom:         getenv("MAIL_FROM", "noreply@example.invalid"),
		SMTPHost:         getenv("SMTP_HOST", ""),
		SMTPPort:         getenv("SMTP_PORT", "587"),
		SMTPUser:         getenv("SMTP_USER", ""),
		SMTPPass:         getenv("SMTP_PASS", ""),
		ResetTokenTTLMin: 60,
		Production:       env == "production",
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
