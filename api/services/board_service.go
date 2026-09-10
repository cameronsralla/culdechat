package services

import (
	"context"
	"errors"
	"strings"

	"github.com/cameronsralla/culdechat/models"
	"github.com/google/uuid"
)

type BoardService struct{}

type CreateBoardInput struct {
	Name        string  `json:"name" example:"Book Club"`
	Description *string `json:"description" example:"Let's read and discuss!"`
}

type BoardDTO struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	CreatorID       string  `json:"creator_id"`
	SubscriberCount int64   `json:"subscriber_count"`
	IsSubscribed    bool    `json:"is_subscribed"`
}

type SubscribeResponse struct {
	Subscribed bool   `json:"subscribed"`
	Message    string `json:"message"`
}

func (s *BoardService) Create(ctx context.Context, creatorID uuid.UUID, in CreateBoardInput) (*BoardDTO, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	if in.Description != nil {
		trimmed := strings.TrimSpace(*in.Description)
		in.Description = &trimmed
	}
	b := &models.Board{CreatorID: creatorID, Name: in.Name, Description: in.Description}
	if err := models.InsertBoard(ctx, b); err != nil {
		return nil, err
	}
	if err := models.Subscribe(ctx, creatorID, b.ID); err != nil {
		return nil, err
	}
	return &BoardDTO{
		ID:              b.ID.String(),
		Name:            b.Name,
		Description:     b.Description,
		CreatorID:       b.CreatorID.String(),
		SubscriberCount: 1,
		IsSubscribed:    true,
	}, nil
}

func (s *BoardService) List(ctx context.Context, userID uuid.UUID) ([]BoardDTO, error) {
	boards, err := models.ListBoardsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]BoardDTO, 0, len(boards))
	for _, b := range boards {
		out = append(out, BoardDTO{
			ID:              b.ID.String(),
			Name:            b.Name,
			Description:     b.Description,
			CreatorID:       b.CreatorID.String(),
			SubscriberCount: b.SubscriberCount,
			IsSubscribed:    b.IsSubscribed,
		})
	}
	return out, nil
}

func (s *BoardService) Get(ctx context.Context, userID, boardID uuid.UUID) (*BoardDTO, error) {
	b, err := models.GetBoardForUser(ctx, userID, boardID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, ErrNotFound
	}
	return &BoardDTO{
		ID:              b.ID.String(),
		Name:            b.Name,
		Description:     b.Description,
		CreatorID:       b.CreatorID.String(),
		SubscriberCount: b.SubscriberCount,
		IsSubscribed:    b.IsSubscribed,
	}, nil
}

func (s *BoardService) ToggleSubscribe(ctx context.Context, userID uuid.UUID, boardID uuid.UUID) (*SubscribeResponse, error) {
	board, err := models.GetBoardByID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board == nil {
		return nil, ErrNotFound
	}
	subscribed, err := models.IsSubscribed(ctx, userID, boardID)
	if err != nil {
		return nil, err
	}
	if subscribed {
		if err := models.Unsubscribe(ctx, userID, boardID); err != nil {
			return nil, err
		}
		return &SubscribeResponse{Subscribed: false, Message: "Successfully unsubscribed from the board."}, nil
	}
	if err := models.Subscribe(ctx, userID, boardID); err != nil {
		return nil, err
	}
	return &SubscribeResponse{Subscribed: true, Message: "Successfully subscribed to the board."}, nil
}
