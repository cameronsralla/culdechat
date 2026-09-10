package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
)

const (
	inviteTTL        = 7 * 24 * time.Hour
	refreshTTL       = 30 * 24 * time.Hour
	maxNameLen       = 80
	passcodeAlphabet = "abcdefghijkmnpqrstuvwxyz23456789"
	passcodeLen      = 10
)

type AuthService struct {
	Mail Mailer
}

type RegisterInput struct {
	Email      string `json:"email" example:"new.resident@example.com"`
	UnitNumber string `json:"unit_number" example:"101"`
}

type CompleteRegistrationInput struct {
	Token    string `json:"token" example:"invite-token"`
	Passcode string `json:"passcode" example:"a3k9wm2p7x"`
	Password string `json:"password" example:"a-strong-password"`
	Name     string `json:"name" example:"Alex Rivera"`
}

type LoginInput struct {
	Email    string `json:"email" example:"resident@example.com"`
	Password string `json:"password" example:"user_password"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutInput struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type RegisterResponse struct {
	Message           string `json:"message"`
	Email             string `json:"email"`
	RegistrationToken string `json:"registration_token"`
	Passcode          string `json:"passcode"`
	InviteExpiresAt   string `json:"invite_expires_at"`
	EmailSent         bool   `json:"email_sent"`
	EmailError        string `json:"email_error,omitempty"`
}

type AuthResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         AuthUserDTO `json:"user"`
}

type AuthUserDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UnitNumber string `json:"unit_number"`
	IsAdmin    bool   `json:"is_admin"`
}

type MeDTO struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	UnitNumber string `json:"unit_number"`
	Status     string `json:"status"`
	IsAdmin    bool   `json:"is_admin"`
}

type AdminUserDTO struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	UnitNumber string `json:"unit_number"`
	Status     string `json:"status"`
	IsAdmin    bool   `json:"is_admin"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*RegisterResponse, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.UnitNumber = strings.TrimSpace(in.UnitNumber)
	if in.Email == "" || in.UnitNumber == "" {
		return nil, errors.New("email and unit_number are required")
	}

	existing, err := models.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Status == models.UserStatusPending {
			return s.refreshInvite(ctx, existing)
		}
		return nil, errors.New("user already exists")
	}

	placeholder, err := utils.HashPassword(uuid.NewString())
	if err != nil {
		return nil, err
	}
	passcode, passcodeHash, err := generatePasscode()
	if err != nil {
		return nil, err
	}
	token := uuid.NewString()
	tokenHash := utils.HashOpaque(token)
	expires := time.Now().Add(inviteTTL)

	user := &models.User{
		UnitNumber:      in.UnitNumber,
		Email:           in.Email,
		Name:            "",
		HashedPassword:  placeholder,
		IsAdmin:         false,
		Status:          models.UserStatusPending,
		InviteToken:     &tokenHash,
		PasscodeHash:    &passcodeHash,
		InviteExpiresAt: &expires,
	}
	if err := models.InsertUser(ctx, user); err != nil {
		return nil, err
	}

	utils.Infof("user invited unit=%s id=%s", user.UnitNumber, user.ID)
	return s.deliverInvite(user.Email, token, passcode, expires), nil
}

func (s *AuthService) refreshInvite(ctx context.Context, existing *models.User) (*RegisterResponse, error) {
	passcode, passcodeHash, err := generatePasscode()
	if err != nil {
		return nil, err
	}
	token := uuid.NewString()
	tokenHash := utils.HashOpaque(token)
	expires := time.Now().Add(inviteTTL)
	existing.InviteToken = &tokenHash
	existing.PasscodeHash = &passcodeHash
	existing.InviteExpiresAt = &expires
	if err := models.UpdateUser(ctx, existing); err != nil {
		return nil, err
	}
	return s.deliverInvite(existing.Email, token, passcode, expires), nil
}

func (s *AuthService) CompleteRegistration(ctx context.Context, in CompleteRegistrationInput) (*AuthResponse, error) {
	in.Token = strings.TrimSpace(in.Token)
	in.Passcode = strings.TrimSpace(in.Passcode)
	in.Name = strings.TrimSpace(in.Name)
	if in.Token == "" || in.Passcode == "" || in.Password == "" || in.Name == "" {
		return nil, errors.New("token, passcode, password, and name are required")
	}
	if utf8.RuneCountInString(in.Name) > maxNameLen {
		return nil, errors.New("name is too long")
	}
	if err := utils.ValidatePassword(in.Password); err != nil {
		return nil, err
	}

	u, err := models.GetUserByInviteTokenHash(ctx, utils.HashOpaque(in.Token))
	if err != nil {
		return nil, err
	}
	if u == nil || u.Status != models.UserStatusPending {
		return nil, errors.New("invalid or expired registration token")
	}
	if u.InviteExpiresAt != nil && time.Now().After(*u.InviteExpiresAt) {
		return nil, errors.New("invalid or expired registration token")
	}
	if u.PasscodeHash == nil || !utils.CheckPassword(*u.PasscodeHash, in.Passcode) {
		return nil, errors.New("invalid passcode")
	}

	hashed, err := utils.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	u.Name = in.Name
	u.HashedPassword = hashed
	u.Status = models.UserStatusActive
	u.InviteToken = nil
	u.PasscodeHash = nil
	u.InviteExpiresAt = nil
	if err := models.UpdateUser(ctx, u); err != nil {
		return nil, err
	}

	return issueAuth(ctx, u)
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResponse, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Email == "" || in.Password == "" {
		return nil, errors.New("email and password are required")
	}

	u, err := models.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}
	if !utils.CheckPassword(u.HashedPassword, in.Password) {
		return nil, ErrInvalidCredentials
	}
	if u.Status != models.UserStatusActive {
		return nil, ErrInactiveAccount
	}

	utils.Infof("user logged in id=%s", u.ID)
	return issueAuth(ctx, u)
}

func (s *AuthService) Refresh(ctx context.Context, in RefreshInput) (*AuthResponse, error) {
	in.RefreshToken = strings.TrimSpace(in.RefreshToken)
	if in.RefreshToken == "" {
		return nil, errors.New("refresh_token is required")
	}
	hash := utils.HashOpaque(in.RefreshToken)
	userID, err := models.LookupRefreshToken(ctx, hash)
	if err != nil {
		return nil, ErrUnauthorized
	}
	u, err := models.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil || u.Status != models.UserStatusActive {
		return nil, ErrInactiveAccount
	}
	if err := models.RevokeRefreshToken(ctx, hash); err != nil {
		return nil, err
	}
	return issueAuth(ctx, u)
}

func (s *AuthService) Logout(ctx context.Context, in LogoutInput) error {
	in.RefreshToken = strings.TrimSpace(in.RefreshToken)
	if in.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	return models.RevokeRefreshToken(ctx, utils.HashOpaque(in.RefreshToken))
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return models.RevokeRefreshTokensForUser(ctx, userID)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, in ChangePasswordInput) error {
	if err := utils.ValidatePassword(in.NewPassword); err != nil {
		return err
	}
	u, err := models.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrNotFound
	}
	if !utils.CheckPassword(u.HashedPassword, in.CurrentPassword) {
		return ErrInvalidCredentials
	}
	hashed, err := utils.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	u.HashedPassword = hashed
	if err := models.UpdateUser(ctx, u); err != nil {
		return err
	}
	return models.RevokeRefreshTokensForUser(ctx, userID)
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*MeDTO, error) {
	u, err := models.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return &MeDTO{
		ID:         u.ID.String(),
		Email:      u.Email,
		Name:       u.Name,
		UnitNumber: u.UnitNumber,
		Status:     u.Status,
		IsAdmin:    u.IsAdmin,
	}, nil
}

func (s *AuthService) ListUsers(ctx context.Context) ([]AdminUserDTO, error) {
	users, err := models.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AdminUserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, AdminUserDTO{
			ID:         u.ID.String(),
			Email:      u.Email,
			Name:       u.Name,
			UnitNumber: u.UnitNumber,
			Status:     u.Status,
			IsAdmin:    u.IsAdmin,
		})
	}
	return out, nil
}

func (s *AuthService) Offboard(ctx context.Context, userID uuid.UUID) error {
	u, err := models.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrNotFound
	}
	if u.IsAdmin {
		return errors.New("cannot offboard an admin")
	}
	if err := models.RevokeRefreshTokensForUser(ctx, userID); err != nil {
		return err
	}
	return models.SoftDeleteUser(ctx, userID)
}

func issueAuth(ctx context.Context, u *models.User) (*AuthResponse, error) {
	token, err := utils.GenerateAccessToken(u.ID.String(), u.UnitNumber, u.IsAdmin)
	if err != nil {
		return nil, err
	}
	raw, err := randomToken()
	if err != nil {
		return nil, err
	}
	if err := models.InsertRefreshToken(ctx, u.ID, utils.HashOpaque(raw), time.Now().Add(refreshTTL)); err != nil {
		return nil, err
	}
	return &AuthResponse{
		Token:        token,
		RefreshToken: raw,
		User: AuthUserDTO{
			ID:         u.ID.String(),
			Name:       u.Name,
			UnitNumber: u.UnitNumber,
			IsAdmin:    u.IsAdmin,
		},
	}, nil
}

func generatePasscode() (plain string, hashed string, err error) {
	buf := make([]byte, passcodeLen)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	alpha := []byte(passcodeAlphabet)
	for i := range buf {
		buf[i] = alpha[int(buf[i])%len(alpha)]
	}
	plain = string(buf)
	hashed, err = utils.HashPassword(plain)
	return plain, hashed, err
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
