package mail

import (
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/markoomisicc/uptime-api/internal/config"
)

type Mailer struct {
	cfg config.SMTPConfig
}

func New(cfg config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Send(to, subject, body string) error {
	from, err := mail.ParseAddress(m.cfg.From)
	if err != nil {
		return fmt.Errorf("parse from address: %w", err)
	}

	headers := []string{
		"From: " + from.String(),
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	message := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

	var auth smtp.Auth
	if m.cfg.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)
	if err := smtp.SendMail(addr, auth, from.Address, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}
