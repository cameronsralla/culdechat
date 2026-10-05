package services

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cameronsralla/culdechat/utils"
)

// InviteMail builds the resident invite message. Other outbound mail should
// similarly construct a MailMessage and call Mailer.Send.
func InviteMail(to, token, passcode string, expires time.Time) MailMessage {
	name := CommunityName()
	var b strings.Builder
	fmt.Fprintf(&b, "You have been invited to %s (%s).\n\n", name, to)
	fmt.Fprintf(&b, "Finish registration with this token and passcode:\n")
	fmt.Fprintf(&b, "  Token: %s\n", token)
	fmt.Fprintf(&b, "  Passcode: %s\n", passcode)
	fmt.Fprintf(&b, "  Expires: %s\n", expires.UTC().Format(time.RFC3339))
	if appURL := strings.TrimSpace(os.Getenv("CULDECHAT_APP_URL")); appURL != "" {
		fmt.Fprintf(&b, "\nOr open: %s/register?token=%s\n", strings.TrimSuffix(appURL, "/"), token)
	}
	fmt.Fprintf(&b, "\nIf you did not expect this, ignore the email.\n")
	return MailMessage{
		To:      to,
		Subject: "Your " + name + " invite",
		Body:    b.String(),
	}
}

func (s *AuthService) mailer() Mailer {
	if s.Mail != nil {
		return s.Mail
	}
	return SMTPMailer{}
}

func (s *AuthService) deliverInvite(email, token, passcode string, expires time.Time) *RegisterResponse {
	resp := &RegisterResponse{
		Email:             email,
		RegistrationToken: token,
		Passcode:          passcode,
		InviteExpiresAt:   expires.UTC().Format(time.RFC3339),
	}
	mailer := s.mailer()
	if !mailer.Configured() {
		resp.Message = fmt.Sprintf("Invite created for %s (email not configured; share the token and passcode)", email)
		return resp
	}
	if err := mailer.Send(InviteMail(email, token, passcode, expires)); err != nil {
		utils.Errorf("invite email failed to=%s: %v", email, err)
		resp.EmailError = "could not send invite email"
		resp.Message = fmt.Sprintf("Invite created for %s, but email failed to send", email)
		return resp
	}
	resp.EmailSent = true
	resp.Message = fmt.Sprintf("Invite emailed to %s", email)
	return resp
}
