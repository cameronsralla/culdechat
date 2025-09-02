package services

import (
	"context"
	"errors"
	"strings"

	"github.com/cameronsralla/culdechat/models"
	"github.com/google/uuid"
)

type ReactionService struct{}

type ReactInput struct {
	PostID string `json:"post_id"`
	Type   string `json:"type"`
}

type ReactionCountDTO struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

func (s *ReactionService) Upsert(ctx context.Context, userID uuid.UUID, in ReactInput) error {
	in.Type = strings.TrimSpace(strings.ToLower(in.Type))
	if in.PostID == "" || in.Type == "" {
		return errors.New("post_id and type are required")
	}
	postUUID, err := uuid.Parse(in.PostID)
	if err != nil {
		return errors.New("invalid post_id")
	}
	if err := models.EnsureReactionsTable(ctx); err != nil {
		return err
	}
	r := &models.Reaction{PostID: postUUID, UserID: userID, Type: in.Type}
	return models.UpsertReaction(ctx, r)
}

func (s *ReactionService) Remove(ctx context.Context, userID uuid.UUID, postIDStr string) error {
	postUUID, err := uuid.Parse(postIDStr)
	if err != nil {
		return errors.New("invalid post_id")
	}
	if err := models.EnsureReactionsTable(ctx); err != nil {
		return err
	}
	return models.RemoveReaction(ctx, postUUID, userID)
}

func (s *ReactionService) CountByPost(ctx context.Context, postID uuid.UUID) ([]ReactionCountDTO, error) {
	if err := models.EnsureReactionsTable(ctx); err != nil {
		return nil, err
	}
	counts, err := models.CountReactionsByPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	out := make([]ReactionCountDTO, 0, len(counts))
	for _, c := range counts {
		out = append(out, ReactionCountDTO{Type: c.Type, Count: c.Count})
	}
	return out, nil
}
