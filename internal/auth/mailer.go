package auth

import (
	"fmt"
	"log"
	"net/smtp"

	"studigame/internal/config"
)

// Mailer abstracts outbound email so implementations can be swapped via env vars.
type Mailer interface {
	Send(to, subject, body string) error
}

// ConsoleMailer logs instead of sending — dev fallback when SMTP not configured.
type ConsoleMailer struct{}

func (ConsoleMailer) Send(to, subject, body string) error {
	log.Printf("[mailer:console] to=%s subject=%q body=%q", to, subject, body)
	return nil
}

// SMTPMailer sends email via SMTP+TLS.
type SMTPMailer struct {
	host string
	port string
	user string
	pass string
	from string
}

func (m SMTPMailer) Send(to, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.from, to, subject, body)
	addr := m.host + ":" + m.port
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}

// NewMailer returns a Mailer based on configuration.
// If SMTP_HOST is empty, falls back to console logging (development).
func NewMailer(cfg config.Config) Mailer {
	if cfg.SMTPHost == "" {
		return ConsoleMailer{}
	}
	return SMTPMailer{
		host: cfg.SMTPHost,
		port: cfg.SMTPPort,
		user: cfg.SMTPUser,
		pass: cfg.SMTPPass,
		from: cfg.MailFrom,
	}
}
