package services

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

type recordingMailer struct {
	configured bool
	sent       []MailMessage
	err        error
}

func (r *recordingMailer) Configured() bool { return r.configured }

func (r *recordingMailer) Send(msg MailMessage) error {
	r.sent = append(r.sent, msg)
	return r.err
}

func TestInviteMailContainsTokenAndPasscode(t *testing.T) {
	t.Setenv("CULDECHAT_COMMUNITY_NAME", "Oak Court")
	t.Setenv("CULDECHAT_APP_URL", "culdechat://")
	exp := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	msg := InviteMail("a@b.com", "tok-123", "ab34cd56ef", exp)
	if msg.To != "a@b.com" {
		t.Fatalf("to=%q", msg.To)
	}
	if !strings.Contains(msg.Subject, "Oak Court") {
		t.Fatalf("subject=%q", msg.Subject)
	}
	for _, want := range []string{"tok-123", "ab34cd56ef", "a@b.com", "culdechat://register?token=tok-123"} {
		if !strings.Contains(msg.Body, want) {
			t.Fatalf("body missing %q:\n%s", want, msg.Body)
		}
	}
}

func TestSMTPMailerConfigured(t *testing.T) {
	os.Unsetenv("SMTP_HOST")
	if (SMTPMailer{}).Configured() {
		t.Fatal("expected false")
	}
	t.Setenv("SMTP_HOST", "mailpit")
	if !(SMTPMailer{}).Configured() {
		t.Fatal("expected true")
	}
}

func TestDeliverInviteWithoutSMTP(t *testing.T) {
	os.Unsetenv("SMTP_HOST")
	svc := &AuthService{}
	resp := svc.deliverInvite("a@b.com", "tok", "passcode12", time.Now())
	if resp.EmailSent {
		t.Fatal("expected email_sent false")
	}
	if resp.RegistrationToken != "tok" || resp.Passcode != "passcode12" {
		t.Fatalf("secrets should still be returned to the admin: %+v", resp)
	}
}

func TestDeliverInviteUsesMailer(t *testing.T) {
	rec := &recordingMailer{configured: true}
	svc := &AuthService{Mail: rec}
	exp := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	resp := svc.deliverInvite("a@b.com", "tok", "passcode12", exp)
	if !resp.EmailSent || len(rec.sent) != 1 {
		t.Fatalf("sent=%d resp=%+v", len(rec.sent), resp)
	}
	if rec.sent[0].To != "a@b.com" || !strings.Contains(rec.sent[0].Body, "passcode12") {
		t.Fatalf("unexpected message: %+v", rec.sent[0])
	}
}

func TestDeliverInviteMailerError(t *testing.T) {
	rec := &recordingMailer{configured: true, err: errors.New("boom")}
	svc := &AuthService{Mail: rec}
	resp := svc.deliverInvite("a@b.com", "tok", "passcode12", time.Now())
	if resp.EmailSent || resp.EmailError == "" {
		t.Fatalf("expected send failure, got %+v", resp)
	}
}
