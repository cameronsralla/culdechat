package services

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cameronsralla/culdechat/models"
	"github.com/google/uuid"
)

const maxCommentLen = 2000

type CommentService struct{}

type CreateCommentInput struct {
	Content string `json:"content" example:"I have one you can use!"`
}

type CommentDTO struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Author    AuthorDTO `json:"author"`
	CreatedAt string    `json:"created_at,omitempty"`
}

type CreatedCommentDTO struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	AuthorID string `json:"author_id"`
}

func (s *CommentService) Create(ctx context.Context, authorID uuid.UUID, postID uuid.UUID, in CreateCommentInput) (*CreatedCommentDTO, error) {
	in.Content = strings.TrimSpace(in.Content)
	if in.Content == "" {
		return nil, errors.New("content is required")
	}
	post, err := models.GetFeedPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrNotFound
	}
	if post.PostType == models.PostTypeBulletin {
		return nil, ErrCommentsDisabled
	}
	if utf8.RuneCountInString(in.Content) > maxCommentLen {
		return nil, errors.New("content is too long")
	}
	c := &models.Comment{PostID: postID, AuthorID: authorID, Content: in.Content}
	if err := models.InsertComment(ctx, c); err != nil {
		return nil, err
	}
	return &CreatedCommentDTO{ID: c.ID.String(), Content: c.Content, AuthorID: c.AuthorID.String()}, nil
}

func (s *CommentService) ListByPost(ctx context.Context, postID uuid.UUID) ([]CommentDTO, error) {
	post, err := models.GetFeedPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrNotFound
	}
	if post.PostType == models.PostTypeBulletin {
		return []CommentDTO{}, nil
	}
	comments, err := models.ListCommentsByPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	out := make([]CommentDTO, 0, len(comments))
	for _, c := range comments {
		out = append(out, toCommentDTO(c))
	}
	return out, nil
}

func (s *CommentService) Update(ctx context.Context, userID, commentID uuid.UUID, in CreateCommentInput) (*CreatedCommentDTO, error) {
	in.Content = strings.TrimSpace(in.Content)
	if in.Content == "" {
		return nil, errors.New("content is required")
	}
	if utf8.RuneCountInString(in.Content) > maxCommentLen {
		return nil, errors.New("content is too long")
	}
	cmt, err := models.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if cmt == nil {
		return nil, ErrNotFound
	}
	if cmt.AuthorID != userID {
		return nil, ErrForbidden
	}
	post, err := models.GetFeedPostByID(ctx, cmt.PostID)
	if err != nil {
		return nil, err
	}
	if post != nil && post.PostType == models.PostTypeBulletin {
		return nil, ErrCommentsDisabled
	}
	if err := models.UpdateCommentContent(ctx, commentID, in.Content); err != nil {
		return nil, err
	}
	return &CreatedCommentDTO{ID: cmt.ID.String(), Content: in.Content, AuthorID: cmt.AuthorID.String()}, nil
}

func (s *CommentService) Delete(ctx context.Context, userID, commentID uuid.UUID, isAdmin bool) error {
	cmt, err := models.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}
	if cmt == nil {
		return ErrNotFound
	}
	if cmt.AuthorID != userID && !isAdmin {
		return ErrForbidden
	}
	return models.DeleteComment(ctx, commentID)
}

func toCommentDTO(c models.CommentWithAuthor) CommentDTO {
	created := ""
	if !c.CreatedAt.IsZero() {
		created = c.CreatedAt.UTC().Format(time.RFC3339)
	}
	return CommentDTO{
		ID:      c.ID.String(),
		Content: c.Content,
		Author: AuthorDTO{
			ID:         c.AuthorID.String(),
			UnitNumber: c.AuthorUnit,
			Name:       c.AuthorName,
		},
		CreatedAt: created,
	}
}
