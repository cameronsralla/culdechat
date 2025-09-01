package services

import (
	"context"
	"errors"
	"strings"

	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/utils"
)

type AuthService struct{}

type RegisterInput struct {
	Email      string `json:"email"`
	UnitNumber string `json:"unit_number"`
	Password   string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  AuthUserDTO `json:"user"`
}

type AuthUserDTO struct {
	ID         string `json:"id"`
	UnitNumber string `json:"unit_number"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResponse, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.UnitNumber = strings.TrimSpace(in.UnitNumber)
	if in.Email == "" || in.UnitNumber == "" || in.Password == "" {
		return nil, errors.New("email, unit_number and password are required")
	}

	// Ensure table exists
	if err := models.EnsureUsersTable(ctx); err != nil {
		return nil, err
	}

	// Check for existing user by email
	existing, err := models.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	hashed, err := utils.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		UnitNumber:       in.UnitNumber,
		Email:            in.Email,
		HashedPassword:   hashed,
		IsDirectoryOptIn: false,
		IsAdmin:          false,
		Status:           "active",
	}
	if err := models.InsertUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := utils.GenerateAccessToken(user.ID.String(), user.UnitNumber)
	if err != nil {
		return nil, err
	}

	utils.Infof("user registered email=%s unit=%s id=%s", user.Email, user.UnitNumber, user.ID)
	return &AuthResponse{
		Token: token,
		User: AuthUserDTO{
			ID:         user.ID.String(),
			UnitNumber: user.UnitNumber,
		},
	}, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResponse, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Email == "" || in.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// Ensure table exists
	if err := models.EnsureUsersTable(ctx); err != nil {
		return nil, err
	}

	u, err := models.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("invalid credentials")
	}
	if !utils.CheckPassword(u.HashedPassword, in.Password) {
		return nil, errors.New("invalid credentials")
	}
	if u.Status != "active" {
		return nil, errors.New("account is not active")
	}

	token, err := utils.GenerateAccessToken(u.ID.String(), u.UnitNumber)
	if err != nil {
		return nil, err
	}

	utils.Infof("user logged in email=%s id=%s", u.Email, u.ID)
	return &AuthResponse{
		Token: token,
		User: AuthUserDTO{
			ID:         u.ID.String(),
			UnitNumber: u.UnitNumber,
		},
	}, nil
}
