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
	Type string `json:"type" example:"like"`
}

type ReactionCountDTO struct {
	Type  string `json:"type" example:"like"`
	Count int64  `json:"count" example:"3"`
}

func allowedReactionType(t string) bool {
	switch t {
	case "like", "love", "laugh", "wow", "sad", "angry":
		return true
	default:
		return false
	}
}

func (s *ReactionService) Upsert(ctx context.Context, userID uuid.UUID, postID uuid.UUID, in ReactInput) error {
	in.Type = strings.TrimSpace(strings.ToLower(in.Type))
	if in.Type == "" {
		return errors.New("type is required")
	}
	if !allowedReactionType(in.Type) {
		return errors.New("reaction type must be one of like, love, laugh, wow, sad, angry")
	}
	post, err := models.GetFeedPostByID(ctx, postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}
	r := &models.Reaction{PostID: postID, UserID: userID, Type: in.Type}
	return models.UpsertReaction(ctx, r)
}

func (s *ReactionService) Remove(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error {
	post, err := models.GetFeedPostByID(ctx, postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}
	return models.RemoveReaction(ctx, postID, userID)
}

func (s *ReactionService) CountByPost(ctx context.Context, postID uuid.UUID) ([]ReactionCountDTO, error) {
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
