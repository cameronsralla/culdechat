package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/cameronsralla/culdechat/models"
	"github.com/google/uuid"
)

type ProfileService struct{}

type UpdateProfileInput struct {
	Name              *string `json:"name"`
	ProfilePictureURL *string `json:"profile_picture_url"`
	DirectoryOptIn    *bool   `json:"directory_opt_in"`
}

type ProfileDTO struct {
	ID                string  `json:"id"`
	Email             string  `json:"email"`
	Name              string  `json:"name"`
	UnitNumber        string  `json:"unit_number"`
	ProfilePictureURL *string `json:"profile_picture_url"`
	DirectoryOptIn    bool    `json:"directory_opt_in"`
	IsAdmin           bool    `json:"is_admin"`
}

func (s *ProfileService) Get(ctx context.Context, userID uuid.UUID) (*ProfileDTO, error) {
	u, err := models.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return &ProfileDTO{
		ID:                u.ID.String(),
		Email:             u.Email,
		Name:              u.Name,
		UnitNumber:        u.UnitNumber,
		ProfilePictureURL: u.ProfilePictureURL,
		DirectoryOptIn:    u.IsDirectoryOptIn,
		IsAdmin:           u.IsAdmin,
	}, nil
}

func (s *ProfileService) Update(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*ProfileDTO, error) {
	u, err := models.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, errors.New("name cannot be empty")
		}
		if utf8.RuneCountInString(name) > maxNameLen {
			return nil, errors.New("name is too long")
		}
		u.Name = name
	}
	if in.ProfilePictureURL != nil {
		cleaned, err := validateProfilePictureURL(*in.ProfilePictureURL)
		if err != nil {
			return nil, err
		}
		if cleaned == "" {
			u.ProfilePictureURL = nil
		} else {
			u.ProfilePictureURL = &cleaned
		}
	}
	if in.DirectoryOptIn != nil {
		u.IsDirectoryOptIn = *in.DirectoryOptIn
	}
	if err := models.UpdateUser(ctx, u); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID)
}

func (s *ProfileService) UploadPhoto(ctx context.Context, userID uuid.UUID, r io.Reader) (*ProfileDTO, error) {
	path, err := SaveProfilePhoto(userID, r)
	if err != nil {
		return nil, err
	}
	return s.Update(ctx, userID, UpdateProfileInput{ProfilePictureURL: &path})
}

type DirectoryUserDTO struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
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
		out = append(out, DirectoryUserDTO{
			ID:                u.ID,
			Name:              u.Name,
			UnitNumber:        u.UnitNumber,
			ProfilePictureURL: u.ProfilePictureURL,
		})
	}
	return out, nil
}
