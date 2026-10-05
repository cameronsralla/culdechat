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
	UnitID         uuid.UUID `json:"unit_id"`
	IsPrimary      bool      `json:"is_primary"`
	DisplayName    string    `json:"display_name"`
	IsAdmin        bool      `json:"is_admin"`
	Status         string    `json:"status"`
	DirectoryOptIn bool      `json:"directory_opt_in"`
	CreatedAt      time.Time `json:"created_at"`
}

func ToUser(u dbq.Resident) User {
	return User{
		ID: u.ID, Email: u.Email, UnitNumber: u.UnitNumber, UnitID: u.UnitID, IsPrimary: u.IsPrimary,
		DisplayName: u.DisplayName, IsAdmin: u.IsAdmin, Status: u.Status, DirectoryOptIn: u.DirectoryOptIn, CreatedAt: u.CreatedAt,
	}
}

// DirectoryEntry is one People row. A person row is an active resident who
// opted in. A unit row is an occupied unit whose resident stayed hidden:
// no id, name, or email.
type DirectoryEntry struct {
	Kind        string     `json:"kind"`
	ID          *uuid.UUID `json:"id,omitempty"`
	UnitNumber  string     `json:"unit_number"`
	DisplayName string     `json:"display_name,omitempty"`
	Email       string     `json:"email,omitempty"`
	Self        bool       `json:"self,omitempty"`
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (dbq.Resident, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return dbq.Resident{}, httpx.ErrNotFound
	}
	return u, err
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

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, in UpdateProfileInput) (dbq.Resident, error) {
	if err := s.q.UpdateUserProfile(ctx, dbq.UpdateUserProfileParams{ID: id, DisplayName: in.DisplayName, DirectoryOptIn: in.DirectoryOptIn}); err != nil {
		return dbq.Resident{}, err
	}
	return s.Get(ctx, id)
}

// ---- admin ----

func (s *Service) List(ctx context.Context) ([]dbq.Resident, error) {
	return s.q.ListUsers(ctx)
}

type InviteInput struct {
	Email       string    `json:"email"`
	UnitID      uuid.UUID `json:"unit_id"`
	DisplayName string    `json:"display_name"`
	IsAdmin     bool      `json:"is_admin"`
}

func (in *InviteInput) Validate() error {
	var f httpx.Fields
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if !strings.Contains(in.Email, "@") || len(in.Email) > 254 {
		f.Add("email", "must be a valid email")
	}
	if in.UnitID == uuid.Nil {
		f.Add("unit_id", "required")
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
	if _, err := s.q.GetUnit(ctx, in.UnitID); errors.Is(err, pgx.ErrNoRows) {
		return InviteResult{}, httpx.ErrNotFound.WithMessage("that unit does not exist")
	} else if err != nil {
		return InviteResult{}, err
	}
	id, err := s.q.CreateInvitedUser(ctx, dbq.CreateInvitedUserParams{
		Email: in.Email, UnitID: in.UnitID, DisplayName: in.DisplayName, IsAdmin: in.IsAdmin,
	})
	if err != nil {
		return InviteResult{}, inviteConflict(err)
	}
	u, err := s.Get(ctx, id)
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

func (s *Service) issueInvite(ctx context.Context, u dbq.Resident, adminID uuid.UUID, devMode bool) (InviteResult, error) {
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

func (s *Service) SetAdmin(ctx context.Context, actor, target uuid.UUID, isAdmin bool) (dbq.Resident, error) {
	if actor == target && !isAdmin {
		return dbq.Resident{}, httpx.ErrConflict.WithMessage("you cannot remove your own admin role")
	}
	if err := s.q.SetUserAdmin(ctx, dbq.SetUserAdminParams{ID: target, IsAdmin: isAdmin}); err != nil {
		return dbq.Resident{}, err
	}
	return s.Get(ctx, target)
}

func (s *Service) SetStatus(ctx context.Context, actor, target uuid.UUID, active bool) (dbq.Resident, error) {
	if actor == target && !active {
		return dbq.Resident{}, httpx.ErrConflict.WithMessage("you cannot deactivate yourself")
	}
	existing, err := s.Get(ctx, target)
	if err != nil {
		return dbq.Resident{}, err
	}
	if existing.Status == "invited" {
		return dbq.Resident{}, httpx.ErrConflict.WithMessage("invited residents must complete registration; resend the invite instead")
	}
	status := "inactive"
	if active {
		status = "active"
	}
	if err := s.q.SetUserStatus(ctx, dbq.SetUserStatusParams{ID: target, Status: status}); err != nil {
		return dbq.Resident{}, err
	}
	if !active {
		if err := s.q.RevokeAllRefreshTokensForUser(ctx, target); err != nil {
			return dbq.Resident{}, err
		}
	}
	return s.Get(ctx, target)
}

// ResetPasswordResult is shown once to the admin — same out-of-band pattern as invite passcodes.
type ResetPasswordResult struct {
	User              User   `json:"user"`
	TemporaryPassword string `json:"temporary_password"`
}

// ResetPassword sets a new temporary password for a registered resident, revokes
// all their sessions, and returns the plaintext password once for the admin to
// relay out-of-band. Does not email the password.
func (s *Service) ResetPassword(ctx context.Context, actor, target uuid.UUID) (ResetPasswordResult, error) {
	if actor == target {
		return ResetPasswordResult{}, httpx.ErrConflict.WithMessage("use Change password on your profile to update your own password")
	}
	u, err := s.Get(ctx, target)
	if err != nil {
		return ResetPasswordResult{}, err
	}
	if u.Status == "invited" {
		return ResetPasswordResult{}, httpx.ErrConflict.WithMessage("invited residents have not set a password yet; resend the invite instead")
	}
	temp, err := auth.RandomPassword(12)
	if err != nil {
		return ResetPasswordResult{}, err
	}
	hash, err := auth.HashPassword(temp)
	if err != nil {
		return ResetPasswordResult{}, err
	}
	if err := s.q.UpdateUserPassword(ctx, dbq.UpdateUserPasswordParams{ID: target, PasswordHash: &hash}); err != nil {
		return ResetPasswordResult{}, err
	}
	if err := s.q.RevokeAllRefreshTokensForUser(ctx, target); err != nil {
		return ResetPasswordResult{}, err
	}

	community := s.settings.String(ctx, settings.CommunityName, "Cul-de-Chat")
	body := fmt.Sprintf("An admin reset your %s password.\n\nAsk them for the temporary password, then sign in and change it from your profile.\n", community)
	if err := s.mail.Send(ctx, mail.Message{To: u.Email, Subject: "Your " + community + " password was reset", Text: body}); err != nil {
		applog.From(ctx).Error("password reset mail failed", "err", err, "user_id", u.ID)
		// Password is already changed; still return it to the admin.
	}
	return ResetPasswordResult{User: ToUser(u), TemporaryPassword: temp}, nil
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
	number := strings.TrimSpace(unit)
	if number == "" {
		number = "1"
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	created, err := q.CreateUnit(ctx, number)
	if err != nil {
		return err
	}
	if _, err := q.CreateActiveUser(ctx, dbq.CreateActiveUserParams{
		Email: email, UnitID: created.ID, DisplayName: name, PasswordHash: &hash, IsAdmin: true,
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	applog.From(ctx).Info("bootstrap admin created", "email", email)
	return nil
}

func inviteConflict(err error) error {
	switch constraint(err) {
	case "users_email_key":
		return httpx.ErrConflict.WithMessage("a resident with that email already exists")
	case "users_one_primary_per_unit":
		return httpx.ErrConflict.WithMessage("that unit already has a primary resident")
	default:
		return err
	}
}

func constraint(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}
