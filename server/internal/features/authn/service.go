// Package authn is the session feature: login, refresh, logout, invite
// completion, and password changes. Token primitives live in internal/auth.
package authn

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cameronsralla/culdechat/server/internal/auth"
	"github.com/cameronsralla/culdechat/server/internal/db/dbq"
	"github.com/cameronsralla/culdechat/server/internal/features/users"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
	applog "github.com/cameronsralla/culdechat/server/internal/log"
)

type Service struct {
	pool       *pgxpool.Pool
	q          *dbq.Queries
	tokens     *auth.Tokens
	refreshTTL time.Duration
}

func NewService(pool *pgxpool.Pool, tokens *auth.Tokens, refreshTTL time.Duration) *Service {
	return &Service{pool: pool, q: dbq.New(pool), tokens: tokens, refreshTTL: refreshTTL}
}

// Session is what every successful auth call returns.
type Session struct {
	AccessToken  string     `json:"access_token"`
	ExpiresIn    int        `json:"expires_in"`
	RefreshToken string     `json:"refresh_token"`
	User         users.User `json:"user"`
}

// Client describes the device for refresh-token bookkeeping.
type Client struct {
	UserAgent string
	IP        string
}

var errBadCreds = httpx.ErrUnauthorized.WithMessage("incorrect email or password")

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (in *LoginInput) Validate() error {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Email == "" || in.Password == "" {
		return errors.New("email and password are required")
	}
	return nil
}

func (s *Service) Login(ctx context.Context, in LoginInput, c Client) (Session, error) {
	u, err := s.q.GetUserByEmail(ctx, in.Email)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && u.PasswordHash == nil) {
		auth.VerifyDummy()
		return Session{}, errBadCreds
	}
	if err != nil {
		return Session{}, err
	}
	if !auth.VerifyPassword(*u.PasswordHash, in.Password) {
		return Session{}, errBadCreds
	}
	if u.Status != "active" {
		return Session{}, httpx.ErrForbidden.WithMessage("this account is not active")
	}
	return s.issue(ctx, u, uuid.New(), c)
}

// Refresh rotates the refresh token. Presenting a token that was already
// rotated is treated as theft: the whole family is revoked.
func (s *Service) Refresh(ctx context.Context, raw string, c Client) (Session, error) {
	rt, err := s.q.GetRefreshTokenByHash(ctx, auth.HashToken(raw))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, httpx.ErrUnauthorized.WithMessage("invalid refresh token")
	}
	if err != nil {
		return Session{}, err
	}
	if rt.RevokedAt != nil {
		applog.From(ctx).Warn("refresh token reuse detected; revoking family", "user_id", rt.UserID, "family", rt.FamilyID)
		_ = s.q.RevokeRefreshFamily(ctx, rt.FamilyID)
		return Session{}, httpx.ErrUnauthorized.WithMessage("session revoked, please sign in again")
	}
	if time.Now().After(rt.ExpiresAt) {
		return Session{}, httpx.ErrUnauthorized.WithMessage("session expired, please sign in again")
	}
	u, err := s.q.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return Session{}, err
	}
	if u.Status != "active" {
		_ = s.q.RevokeRefreshFamily(ctx, rt.FamilyID)
		return Session{}, httpx.ErrForbidden.WithMessage("this account is not active")
	}
	if err := s.q.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return Session{}, err
	}
	return s.issue(ctx, u, rt.FamilyID, c)
}

// Logout revokes the presented refresh token's family (this device).
func (s *Service) Logout(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	rt, err := s.q.GetRefreshTokenByHash(ctx, auth.HashToken(raw))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.q.RevokeRefreshFamily(ctx, rt.FamilyID)
}

// LogoutAll revokes every session for the user.
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.q.RevokeAllRefreshTokensForUser(ctx, userID)
}

type CompleteInviteInput struct {
	Token       string `json:"token"`
	Passcode    string `json:"passcode"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (in *CompleteInviteInput) Validate() error {
	var f httpx.Fields
	in.Token = strings.TrimSpace(in.Token)
	in.Passcode = strings.TrimSpace(in.Passcode)
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if in.Token == "" {
		f.Add("token", "required")
	}
	if in.Passcode == "" {
		f.Add("passcode", "required")
	}
	if err := auth.ValidatePassword(in.Password); err != nil {
		f.Add("password", err.Error())
	}
	if in.DisplayName == "" || len(in.DisplayName) > 80 {
		f.Add("display_name", "required, max 80 characters")
	}
	return f.Err()
}

var errBadInvite = httpx.ErrUnauthorized.WithMessage("invalid or expired invitation")

// CompleteInvite activates an invited user and signs them in.
func (s *Service) CompleteInvite(ctx context.Context, in CompleteInviteInput, c Client) (Session, error) {
	inv, err := s.q.GetInviteByTokenHash(ctx, auth.HashToken(in.Token))
	if errors.Is(err, pgx.ErrNoRows) {
		auth.VerifyDummy()
		return Session{}, errBadInvite
	}
	if err != nil {
		return Session{}, err
	}
	if inv.ConsumedAt != nil || time.Now().After(inv.ExpiresAt) {
		return Session{}, errBadInvite
	}
	if !auth.VerifyPassword(inv.PasscodeHash, in.Passcode) {
		return Session{}, httpx.ErrUnauthorized.WithMessage("incorrect passcode")
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return Session{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	u, err := q.ActivateUser(ctx, dbq.ActivateUserParams{ID: inv.UserID, PasswordHash: &hash, DisplayName: in.DisplayName})
	if err != nil {
		return Session{}, err
	}
	if err := q.ConsumeInvite(ctx, inv.ID); err != nil {
		return Session{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, err
	}
	return s.issue(ctx, u, uuid.New(), c)
}

// PeekInvite returns the invitee's email/unit so the registration page can
// show who the invite is for. It does not require the passcode.
func (s *Service) PeekInvite(ctx context.Context, token string) (users.User, error) {
	inv, err := s.q.GetInviteByTokenHash(ctx, auth.HashToken(token))
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (inv.ConsumedAt != nil || time.Now().After(inv.ExpiresAt))) {
		return users.User{}, errBadInvite
	}
	if err != nil {
		return users.User{}, err
	}
	u, err := s.q.GetUserByID(ctx, inv.UserID)
	if err != nil {
		return users.User{}, err
	}
	out := users.ToUser(u)
	out.ID = uuid.Nil // not useful pre-auth
	return out, nil
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (in *ChangePasswordInput) Validate() error {
	if in.CurrentPassword == "" {
		return errors.New("current_password: required")
	}
	if err := auth.ValidatePassword(in.NewPassword); err != nil {
		return errors.New("new_password: " + err.Error())
	}
	return nil
}

// ChangePassword updates the password and signs out all other sessions,
// returning a fresh session for this device.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, in ChangePasswordInput, c Client) (Session, error) {
	u, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		return Session{}, err
	}
	if u.PasswordHash == nil || !auth.VerifyPassword(*u.PasswordHash, in.CurrentPassword) {
		return Session{}, httpx.ErrUnauthorized.WithMessage("current password is incorrect")
	}
	hash, err := auth.HashPassword(in.NewPassword)
	if err != nil {
		return Session{}, err
	}
	if err := s.q.UpdateUserPassword(ctx, dbq.UpdateUserPasswordParams{ID: userID, PasswordHash: &hash}); err != nil {
		return Session{}, err
	}
	if err := s.q.RevokeAllRefreshTokensForUser(ctx, userID); err != nil {
		return Session{}, err
	}
	return s.issue(ctx, u, uuid.New(), c)
}

func (s *Service) issue(ctx context.Context, u dbq.User, family uuid.UUID, c Client) (Session, error) {
	now := time.Now()
	access, err := s.tokens.Issue(u.ID, u.IsAdmin, now)
	if err != nil {
		return Session{}, err
	}
	raw, err := auth.RandomToken(32)
	if err != nil {
		return Session{}, err
	}
	if _, err := s.q.CreateRefreshToken(ctx, dbq.CreateRefreshTokenParams{
		UserID: u.ID, TokenHash: auth.HashToken(raw), FamilyID: family,
		ExpiresAt: now.Add(s.refreshTTL), UserAgent: truncate(c.UserAgent, 256), Ip: c.IP,
	}); err != nil {
		return Session{}, err
	}
	return Session{
		AccessToken:  access,
		ExpiresIn:    int(s.tokens.TTL().Seconds()),
		RefreshToken: raw,
		User:         users.ToUser(u),
	}, nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
