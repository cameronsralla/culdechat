package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

// ErrSMTPNotConfigured is returned by SMTPMailer.Send when SMTP_HOST is empty.
var ErrSMTPNotConfigured = errors.New("smtp not configured")

// MailMessage is a plain-text outbound email.
type MailMessage struct {
	To      string
	Subject string
	Body    string
}

// Mailer sends mail through the instance's community mailbox.
type Mailer interface {
	Configured() bool
	Send(msg MailMessage) error
}

// SMTPMailer delivers MailMessage values over SMTP.
type SMTPMailer struct{}

// Configured is true when SMTP_HOST is set.
func (SMTPMailer) Configured() bool {
	return strings.TrimSpace(os.Getenv("SMTP_HOST")) != ""
}

// Send delivers one plain-text message. Recipients are a single To address.
func (m SMTPMailer) Send(msg MailMessage) error {
	if !m.Configured() {
		return ErrSMTPNotConfigured
	}
	to := strings.TrimSpace(msg.To)
	subject := strings.TrimSpace(msg.Subject)
	if to == "" {
		return errors.New("mail recipient is required")
	}
	if subject == "" {
		return errors.New("mail subject is required")
	}
	return sendSMTP(to, subject, msg.Body)
}

// CommunityName is the display name used in outbound mail (CULDECHAT_COMMUNITY_NAME).
func CommunityName() string {
	name := strings.TrimSpace(os.Getenv("CULDECHAT_COMMUNITY_NAME"))
	if name == "" {
		return "Cul-de-Chat"
	}
	return name
}

func sendSMTP(to, subject, body string) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if host == "" {
		return ErrSMTPNotConfigured
	}
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	if port == "" {
		port = "587"
	}
	fromRaw := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := strings.ReplaceAll(os.Getenv("SMTP_PASS"), " ", "")
	if fromRaw == "" {
		fromRaw = user
	}
	if fromRaw == "" {
		return errors.New("SMTP_FROM or SMTP_USER is required")
	}
	fromAddr, err := mail.ParseAddress(fromRaw)
	if err != nil {
		return fmt.Errorf("SMTP_FROM: %w", err)
	}
	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("recipient: %w", err)
	}

	payload := strings.Join([]string{
		"From: " + fromAddr.String(),
		"To: " + toAddr.Address,
		"Subject: " + sanitizeHeader(subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	if shouldStartTLS(user, port) {
		cfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		if truthy(os.Getenv("SMTP_TLS_SKIP_VERIFY")) {
			cfg.InsecureSkipVerify = true
		}
		if err := client.StartTLS(cfg); err != nil {
			return err
		}
	}
	if user != "" {
		if err := client.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
			return err
		}
	}
	if err := client.Mail(fromAddr.Address); err != nil {
		return err
	}
	if err := client.Rcpt(toAddr.Address); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, payload); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func shouldStartTLS(user, port string) bool {
	if v := strings.TrimSpace(os.Getenv("SMTP_STARTTLS")); v != "" {
		return truthy(v)
	}
	if user != "" {
		return true
	}
	n, err := strconv.Atoi(port)
	return err == nil && n == 587
}

func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func sanitizeHeader(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, s)
}
