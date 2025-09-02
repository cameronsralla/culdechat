package services

import (
	"context"
	"errors"
	"strings"

	"github.com/cameronsralla/culdechat/models"
)

type BoardService struct{}

type CreateBoardInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type BoardDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (s *BoardService) Create(ctx context.Context, in CreateBoardInput) (*BoardDTO, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	if err := models.EnsureBoardsTable(ctx); err != nil {
		return nil, err
	}
	b := &models.Board{Name: in.Name, Description: in.Description}
	if err := models.InsertBoard(ctx, b); err != nil {
		return nil, err
	}
	return &BoardDTO{ID: b.ID.String(), Name: b.Name, Description: b.Description}, nil
}

func (s *BoardService) List(ctx context.Context) ([]BoardDTO, error) {
	if err := models.EnsureBoardsTable(ctx); err != nil {
		return nil, err
	}
	boards, err := models.ListBoards(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]BoardDTO, 0, len(boards))
	for _, b := range boards {
		out = append(out, BoardDTO{ID: b.ID.String(), Name: b.Name, Description: b.Description})
	}
	return out, nil
}
