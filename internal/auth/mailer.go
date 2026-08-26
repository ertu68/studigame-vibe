package auth

import "log"

// Mailer abstracts outbound email so a real provider (SES, Postmark, ...) can
// be wired in later via env vars without touching call sites.
type Mailer interface {
	Send(to, subject, body string) error
}

// ConsoleMailer logs instead of sending — placeholder until SMTP/API creds exist.
type ConsoleMailer struct{}

func (ConsoleMailer) Send(to, subject, body string) error {
	log.Printf("[mailer:console] to=%s subject=%q body=%q", to, subject, body)
	return nil
}
