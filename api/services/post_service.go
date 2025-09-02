package services

import (
	"context"
	"errors"
	"strings"

	"github.com/cameronsralla/culdechat/models"
	"github.com/google/uuid"
)

type PostService struct{}

type CreatePostInput struct {
	BoardID  string `json:"board_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Bulletin bool   `json:"bulletin"`
}

type PostDTO struct {
	ID         string `json:"id"`
	BoardID    string `json:"board_id"`
	AuthorID   string `json:"author_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	IsBulletin bool   `json:"is_bulletin"`
}

func (s *PostService) Create(ctx context.Context, authorID uuid.UUID, in CreatePostInput, isAdmin bool) (*PostDTO, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Content = strings.TrimSpace(in.Content)
	if in.BoardID == "" || in.Title == "" || in.Content == "" {
		return nil, errors.New("board_id, title and content are required")
	}
	boardUUID, err := uuid.Parse(in.BoardID)
	if err != nil {
		return nil, errors.New("invalid board_id")
	}
	if err := models.EnsureBoardsTable(ctx); err != nil {
		return nil, err
	}
	if err := models.EnsurePostsTable(ctx); err != nil {
		return nil, err
	}
	// Allow bulletin creation only for admins
	if in.Bulletin && !isAdmin {
		return nil, errors.New("only admins can create bulletin posts")
	}
	post := &models.Post{
		BoardID:    boardUUID,
		AuthorID:   authorID,
		Title:      in.Title,
		Content:    in.Content,
		IsBulletin: in.Bulletin,
	}
	if err := models.InsertPost(ctx, post); err != nil {
		return nil, err
	}
	return &PostDTO{
		ID:         post.ID.String(),
		BoardID:    post.BoardID.String(),
		AuthorID:   post.AuthorID.String(),
		Title:      post.Title,
		Content:    post.Content,
		IsBulletin: post.IsBulletin,
	}, nil
}

func (s *PostService) ListByBoard(ctx context.Context, boardID uuid.UUID) ([]PostDTO, error) {
	if err := models.EnsurePostsTable(ctx); err != nil {
		return nil, err
	}
	posts, err := models.ListPostsByBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	out := make([]PostDTO, 0, len(posts))
	for _, p := range posts {
		out = append(out, PostDTO{
			ID:         p.ID.String(),
			BoardID:    p.BoardID.String(),
			AuthorID:   p.AuthorID.String(),
			Title:      p.Title,
			Content:    p.Content,
			IsBulletin: p.IsBulletin,
		})
	}
	return out, nil
}

func (s *PostService) ListBulletins(ctx context.Context) ([]PostDTO, error) {
	if err := models.EnsurePostsTable(ctx); err != nil {
		return nil, err
	}
	posts, err := models.ListBulletins(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PostDTO, 0, len(posts))
	for _, p := range posts {
		out = append(out, PostDTO{
			ID:         p.ID.String(),
			BoardID:    p.BoardID.String(),
			AuthorID:   p.AuthorID.String(),
			Title:      p.Title,
			Content:    p.Content,
			IsBulletin: p.IsBulletin,
		})
	}
	return out, nil
}
