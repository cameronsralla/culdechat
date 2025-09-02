package services

import (
	"context"
	"errors"
	"strings"

	"github.com/cameronsralla/culdechat/models"
	"github.com/google/uuid"
)

type CommentService struct{}

type CreateCommentInput struct {
	PostID  string `json:"post_id"`
	Content string `json:"content"`
}

type CommentDTO struct {
	ID       string `json:"id"`
	PostID   string `json:"post_id"`
	AuthorID string `json:"author_id"`
	Content  string `json:"content"`
}

func (s *CommentService) Create(ctx context.Context, authorID uuid.UUID, in CreateCommentInput) (*CommentDTO, error) {
	in.Content = strings.TrimSpace(in.Content)
	if in.PostID == "" || in.Content == "" {
		return nil, errors.New("post_id and content are required")
	}
	postUUID, err := uuid.Parse(in.PostID)
	if err != nil {
		return nil, errors.New("invalid post_id")
	}
	if err := models.EnsureCommentsTable(ctx); err != nil {
		return nil, err
	}
	c := &models.Comment{PostID: postUUID, AuthorID: authorID, Content: in.Content}
	if err := models.InsertComment(ctx, c); err != nil {
		return nil, err
	}
	return &CommentDTO{ID: c.ID.String(), PostID: c.PostID.String(), AuthorID: c.AuthorID.String(), Content: c.Content}, nil
}

func (s *CommentService) ListByPost(ctx context.Context, postID uuid.UUID) ([]CommentDTO, error) {
	if err := models.EnsureCommentsTable(ctx); err != nil {
		return nil, err
	}
	comments, err := models.ListCommentsByPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	out := make([]CommentDTO, 0, len(comments))
	for _, c := range comments {
		out = append(out, CommentDTO{ID: c.ID.String(), PostID: c.PostID.String(), AuthorID: c.AuthorID.String(), Content: c.Content})
	}
	return out, nil
}
