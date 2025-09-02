package services

import (
	"context"
	"errors"

	"github.com/cameronsralla/culdechat/models"
	"github.com/google/uuid"
)

type ProfileService struct{}

type UpdateProfileInput struct {
	ProfilePictureURL *string `json:"profile_picture_url"`
	DirectoryOptIn    *bool   `json:"directory_opt_in"`
}

type ProfileDTO struct {
	ID                string  `json:"id"`
	Email             string  `json:"email"`
	UnitNumber        string  `json:"unit_number"`
	ProfilePictureURL *string `json:"profile_picture_url"`
	DirectoryOptIn    bool    `json:"directory_opt_in"`
}

func (s *ProfileService) Get(ctx context.Context, userID uuid.UUID) (*ProfileDTO, error) {
	u, err := models.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("user not found")
	}
	return &ProfileDTO{
		ID:                u.ID.String(),
		Email:             u.Email,
		UnitNumber:        u.UnitNumber,
		ProfilePictureURL: u.ProfilePictureURL,
		DirectoryOptIn:    u.IsDirectoryOptIn,
	}, nil
}

func (s *ProfileService) Update(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*ProfileDTO, error) {
	u, err := models.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("user not found")
	}
	if in.ProfilePictureURL != nil {
		u.ProfilePictureURL = in.ProfilePictureURL
	}
	if in.DirectoryOptIn != nil {
		u.IsDirectoryOptIn = *in.DirectoryOptIn
	}
	if err := models.UpdateUser(ctx, u); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID)
}

type DirectoryUserDTO struct {
	ID                string  `json:"id"`
	UnitNumber        string  `json:"unit_number"`
	ProfilePictureURL *string `json:"profile_picture_url"`
}

func (s *ProfileService) ListDirectory(ctx context.Context) ([]DirectoryUserDTO, error) {
	users, err := models.ListDirectoryUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]DirectoryUserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, DirectoryUserDTO{ID: u.ID, UnitNumber: u.UnitNumber, ProfilePictureURL: u.ProfilePictureURL})
	}
	return out, nil
}
