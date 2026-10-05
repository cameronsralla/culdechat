// Package users owns the resident roster: profiles, the directory, and the
// admin functions for inviting and managing residents.
package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cameronsralla/culdechat/server/internal/auth"
	"github.com/cameronsralla/culdechat/server/internal/db/dbq"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
	applog "github.com/cameronsralla/culdechat/server/internal/log"
	"github.com/cameronsralla/culdechat/server/internal/mail"
	"github.com/cameronsralla/culdechat/server/internal/settings"
)

type Service struct {
	pool     *pgxpool.Pool
	q        *dbq.Queries
	mail     mail.Sender
	settings *settings.Service
	public   string
}

func NewService(pool *pgxpool.Pool, m mail.Sender, s *settings.Service, publicURL string) *Service {
	return &Service{pool: pool, q: dbq.New(pool), mail: m, settings: s, public: publicURL}
}

// User is the API shape for a resident.
type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	UnitNumber     string    `json:"unit_number"`
	DisplayName    string    `json:"display_name"`
	IsAdmin        bool      `json:"is_admin"`
	Status         string    `json:"status"`
	DirectoryOptIn bool      `json:"directory_opt_in"`
	CreatedAt      time.Time `json:"created_at"`
}

func ToUser(u dbq.User) User {
	return User{
		ID: u.ID, Email: u.Email, UnitNumber: u.UnitNumber, DisplayName: u.DisplayName,
		IsAdmin: u.IsAdmin, Status: u.Status, DirectoryOptIn: u.DirectoryOptIn, CreatedAt: u.CreatedAt,
	}
}

// DirectoryEntry is the resident-visible subset.
type DirectoryEntry struct {
	ID          uuid.UUID `json:"id"`
	UnitNumber  string    `json:"unit_number"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (dbq.User, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return dbq.User{}, httpx.ErrNotFound
	}
	return u, err
}

func (s *Service) Directory(ctx context.Context) ([]DirectoryEntry, error) {
	rows, err := s.q.ListDirectory(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]DirectoryEntry, len(rows))
	for i, r := range rows {
		out[i] = DirectoryEntry{ID: r.ID, UnitNumber: r.UnitNumber, DisplayName: r.DisplayName, Email: r.Email}
	}
	return out, nil
}

type UpdateProfileInput struct {
	DisplayName    string `json:"display_name"`
	DirectoryOptIn bool   `json:"directory_opt_in"`
}

func (in *UpdateProfileInput) Validate() error {
	var f httpx.Fields
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if in.DisplayName == "" {
		f.Add("display_name", "required")
	}
	if len(in.DisplayName) > 80 {
		f.Add("display_name", "max 80 characters")
	}
	return f.Err()
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, in UpdateProfileInput) (dbq.User, error) {
	return s.q.UpdateUserProfile(ctx, dbq.UpdateUserProfileParams{ID: id, DisplayName: in.DisplayName, DirectoryOptIn: in.DirectoryOptIn})
}

// ---- admin ----

func (s *Service) List(ctx context.Context) ([]dbq.User, error) {
	return s.q.ListUsers(ctx)
}

type InviteInput struct {
	Email       string `json:"email"`
	UnitNumber  string `json:"unit_number"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
}

func (in *InviteInput) Validate() error {
	var f httpx.Fields
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.UnitNumber = strings.TrimSpace(in.UnitNumber)
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if !strings.Contains(in.Email, "@") || len(in.Email) > 254 {
		f.Add("email", "must be a valid email")
	}
	if in.UnitNumber == "" || len(in.UnitNumber) > 32 {
		f.Add("unit_number", "required, max 32 characters")
	}
	if len(in.DisplayName) > 80 {
		f.Add("display_name", "max 80 characters")
	}
	return f.Err()
}

// InviteResult is returned to the admin so the passcode can be relayed
// out-of-band (in person / text) as a second factor to the emailed link.
type InviteResult struct {
	User      User      `json:"user"`
	Passcode  string    `json:"passcode"`
	ExpiresAt time.Time `json:"expires_at"`
	// InviteURL is only populated in dev (no SMTP) so testers can complete the flow.
	InviteURL string `json:"invite_url,omitempty"`
}

// Invite creates an invited user and emails them a completion link.
func (s *Service) Invite(ctx context.Context, adminID uuid.UUID, in InviteInput, devMode bool) (InviteResult, error) {
	u, err := s.q.CreateInvitedUser(ctx, dbq.CreateInvitedUserParams{
		Email: in.Email, UnitNumber: in.UnitNumber, DisplayName: in.DisplayName, IsAdmin: in.IsAdmin,
	})
	if isUniqueViolation(err) {
		return InviteResult{}, httpx.ErrConflict.WithMessage("a resident with that email already exists")
	}
	if err != nil {
		return InviteResult{}, err
	}
	return s.issueInvite(ctx, u, adminID, devMode)
}

// Reinvite issues a fresh invite for a still-invited user.
func (s *Service) Reinvite(ctx context.Context, adminID, userID uuid.UUID, devMode bool) (InviteResult, error) {
	u, err := s.Get(ctx, userID)
	if err != nil {
		return InviteResult{}, err
	}
	if u.Status != "invited" {
		return InviteResult{}, httpx.ErrConflict.WithMessage("resident has already completed registration")
	}
	if err := s.q.ExpireOpenInvitesForUser(ctx, userID); err != nil {
		return InviteResult{}, err
	}
	return s.issueInvite(ctx, u, adminID, devMode)
}

func (s *Service) issueInvite(ctx context.Context, u dbq.User, adminID uuid.UUID, devMode bool) (InviteResult, error) {
	token, err := auth.RandomToken(32)
	if err != nil {
		return InviteResult{}, err
	}
	passcode, err := auth.RandomDigits(6)
	if err != nil {
		return InviteResult{}, err
	}
	passHash, err := auth.HashPassword(passcode)
	if err != nil {
		return InviteResult{}, err
	}
	hours := s.settings.Int(ctx, settings.InviteExpiryHours, 168)
	expires := time.Now().Add(time.Duration(hours) * time.Hour)

	if _, err := s.q.CreateInvite(ctx, dbq.CreateInviteParams{
		UserID: u.ID, TokenHash: auth.HashToken(token), PasscodeHash: passHash,
		InvitedBy: &adminID, ExpiresAt: expires,
	}); err != nil {
		return InviteResult{}, err
	}

	community := s.settings.String(ctx, settings.CommunityName, "Cul-de-Chat")
	url := fmt.Sprintf("%s/register?token=%s", s.public, token)
	body := fmt.Sprintf("You've been invited to %s.\n\nOpen this link to set up your account:\n%s\n\nYou'll also need the passcode your community admin gave you. The link expires %s.\n",
		community, url, expires.Format("Jan 2, 2006"))
	if err := s.mail.Send(ctx, mail.Message{To: u.Email, Subject: "Your invitation to " + community, Text: body}); err != nil {
		applog.From(ctx).Error("invite mail failed", "err", err, "user_id", u.ID)
		return InviteResult{}, httpx.ErrInternal.WithMessage("could not send invite email").Wrap(err)
	}

	res := InviteResult{User: ToUser(u), Passcode: passcode, ExpiresAt: expires}
	if devMode {
		res.InviteURL = url
	}
	return res, nil
}

func (s *Service) SetAdmin(ctx context.Context, actor, target uuid.UUID, isAdmin bool) (dbq.User, error) {
	if actor == target && !isAdmin {
		return dbq.User{}, httpx.ErrConflict.WithMessage("you cannot remove your own admin role")
	}
	u, err := s.q.SetUserAdmin(ctx, dbq.SetUserAdminParams{ID: target, IsAdmin: isAdmin})
	if errors.Is(err, pgx.ErrNoRows) {
		return dbq.User{}, httpx.ErrNotFound
	}
	return u, err
}

func (s *Service) SetStatus(ctx context.Context, actor, target uuid.UUID, active bool) (dbq.User, error) {
	if actor == target && !active {
		return dbq.User{}, httpx.ErrConflict.WithMessage("you cannot deactivate yourself")
	}
	status := "inactive"
	if active {
		status = "active"
	}
	u, err := s.q.SetUserStatus(ctx, dbq.SetUserStatusParams{ID: target, Status: status})
	if errors.Is(err, pgx.ErrNoRows) {
		return dbq.User{}, httpx.ErrNotFound
	}
	if err != nil {
		return dbq.User{}, err
	}
	if !active {
		if err := s.q.RevokeAllRefreshTokensForUser(ctx, target); err != nil {
			return dbq.User{}, err
		}
	}
	return u, nil
}

// EnsureBootstrapAdmin creates the first admin from env when the users table is
// empty. It is a no-op once any user exists.
func (s *Service) EnsureBootstrapAdmin(ctx context.Context, email, password, name, unit string) error {
	if email == "" || password == "" {
		return nil
	}
	n, err := s.q.CountUsers(ctx)
	if err != nil || n > 0 {
		return err
	}
	if err := auth.ValidatePassword(password); err != nil {
		return fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD: %w", err)
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.q.CreateActiveUser(ctx, dbq.CreateActiveUserParams{
		Email: email, UnitNumber: unit, DisplayName: name, PasswordHash: &hash, IsAdmin: true,
	})
	if err == nil {
		applog.From(ctx).Info("bootstrap admin created", "email", email)
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
