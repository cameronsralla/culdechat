package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cameronsralla/culdechat/models"
	"github.com/google/uuid"
)

const (
	defaultFeedLimit = 20
	maxFeedLimit     = 50
	snippetRunes     = 180
	maxTitleLen      = 200
	maxContentLen    = 10000
)

type PostService struct{}

type CreatePostInput struct {
	Title    string `json:"title" example:"New book for September!"`
	Content  string `json:"content" example:"We'll be reading The Midnight Library."`
	PostType string `json:"post_type" example:"standard"`
	IsPinned *bool  `json:"is_pinned"`
}

type AuthorDTO struct {
	ID         string  `json:"id"`
	UnitNumber string  `json:"unit_number"`
	Name       *string `json:"name,omitempty"`
}

type BoardRefDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FeedPostDTO struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Snippet       string      `json:"snippet"`
	Author        AuthorDTO   `json:"author"`
	Board         BoardRefDTO `json:"board"`
	CommentCount  int64       `json:"comment_count"`
	ReactionCount int64       `json:"reaction_count"`
	PostType      string      `json:"post_type"`
	IsPinned      bool        `json:"is_pinned"`
	MyReaction    string      `json:"my_reaction,omitempty"`
	CreatedAt     string      `json:"created_at"`
}

type FeedPageDTO struct {
	Posts          []FeedPostDTO `json:"posts"`
	NextPageCursor *string       `json:"next_page_cursor"`
}

type CreatedPostDTO struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID string `json:"author_id"`
	BoardID  string `json:"board_id"`
	PostType string `json:"post_type"`
	IsPinned bool   `json:"is_pinned"`
}

type PostDetailDTO struct {
	ID               string             `json:"id"`
	Title            string             `json:"title"`
	Content          string             `json:"content"`
	Author           AuthorDTO          `json:"author"`
	Board            BoardRefDTO        `json:"board"`
	PostType         string             `json:"post_type"`
	IsPinned         bool               `json:"is_pinned"`
	CommentsDisabled bool               `json:"comments_disabled"`
	Comments         []CommentDTO       `json:"comments"`
	Reactions        []ReactionCountDTO `json:"reactions"`
	MyReaction       string             `json:"my_reaction,omitempty"`
	CreatedAt        string             `json:"created_at"`
}

type UpdatePostInput struct {
	Title    *string `json:"title"`
	Content  *string `json:"content"`
	IsPinned *bool   `json:"is_pinned"`
}

type feedCursor struct {
	Pinned    bool      `json:"p"`
	CreatedAt time.Time `json:"t"`
	ID        uuid.UUID `json:"id"`
}

func (s *PostService) Create(ctx context.Context, authorID uuid.UUID, boardID uuid.UUID, in CreatePostInput, isAdmin bool) (*CreatedPostDTO, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Content = strings.TrimSpace(in.Content)
	if in.Title == "" || in.Content == "" {
		return nil, errors.New("title and content are required")
	}
	if utf8.RuneCountInString(in.Title) > maxTitleLen {
		return nil, errors.New("title is too long")
	}
	if utf8.RuneCountInString(in.Content) > maxContentLen {
		return nil, errors.New("content is too long")
	}

	board, err := models.GetBoardByID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board == nil {
		return nil, ErrNotFound
	}
	if !isAdmin {
		ok, err := models.IsSubscribed(ctx, authorID, boardID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrNotSubscribed
		}
	}

	postType, err := normalizePostType(in.PostType)
	if err != nil {
		return nil, err
	}
	if postType == models.PostTypeBulletin && !isAdmin {
		return nil, errors.New("only admins can create bulletin posts")
	}

	pinned := false
	if postType == models.PostTypeBulletin {
		pinned = true
	} else if in.IsPinned != nil && *in.IsPinned {
		if !isAdmin {
			return nil, errors.New("only admins can pin posts")
		}
		pinned = true
	}

	post := &models.Post{
		BoardID:  boardID,
		AuthorID: authorID,
		Title:    in.Title,
		Content:  in.Content,
		PostType: postType,
		IsPinned: pinned,
	}
	if err := models.InsertPost(ctx, post); err != nil {
		return nil, err
	}
	return &CreatedPostDTO{
		ID:       post.ID.String(),
		Title:    post.Title,
		Content:  post.Content,
		AuthorID: post.AuthorID.String(),
		BoardID:  post.BoardID.String(),
		PostType: post.PostType,
		IsPinned: post.IsPinned,
	}, nil
}

func (s *PostService) ListFeed(ctx context.Context, boardID *uuid.UUID, limit int, cursorStr string, userID uuid.UUID) (*FeedPageDTO, error) {
	if limit <= 0 {
		limit = defaultFeedLimit
	}
	if limit > maxFeedLimit {
		limit = maxFeedLimit
	}

	var pinned *bool
	var created *time.Time
	var id *uuid.UUID
	if strings.TrimSpace(cursorStr) != "" {
		cur, err := decodeFeedCursor(cursorStr)
		if err != nil {
			return nil, errors.New("invalid cursor")
		}
		pinned = &cur.Pinned
		created = &cur.CreatedAt
		id = &cur.ID
	}

	rows, err := models.ListFeed(ctx, boardID, limit+1, pinned, created, id)
	if err != nil {
		return nil, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	page := &FeedPageDTO{Posts: make([]FeedPostDTO, 0, len(rows))}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	mine, err := models.GetReactionTypesForPosts(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		dto := toFeedPostDTO(row)
		dto.MyReaction = mine[row.ID]
		page.Posts = append(page.Posts, dto)
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		cur := encodeFeedCursor(feedCursor{Pinned: last.IsPinned, CreatedAt: last.CreatedAt, ID: last.ID})
		page.NextPageCursor = &cur
	}
	return page, nil
}

func (s *PostService) Get(ctx context.Context, postID, userID uuid.UUID) (*PostDetailDTO, error) {
	post, err := models.GetFeedPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrNotFound
	}

	comments, err := models.ListCommentsByPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	commentDTOs := make([]CommentDTO, 0, len(comments))
	for _, c := range comments {
		commentDTOs = append(commentDTOs, toCommentDTO(c))
	}

	counts, err := models.CountReactionsByPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	reactions := make([]ReactionCountDTO, 0, len(counts))
	for _, rc := range counts {
		reactions = append(reactions, ReactionCountDTO{Type: rc.Type, Count: rc.Count})
	}

	disabled := post.PostType == models.PostTypeBulletin
	if disabled {
		commentDTOs = []CommentDTO{}
	}

	mine, err := models.GetReactionType(ctx, postID, userID)
	if err != nil {
		return nil, err
	}

	return &PostDetailDTO{
		ID:               post.ID.String(),
		Title:            post.Title,
		Content:          post.Content,
		Author:           toAuthorDTO(post),
		Board:            BoardRefDTO{ID: post.BoardID.String(), Name: post.BoardName},
		PostType:         post.PostType,
		IsPinned:         post.IsPinned,
		CommentsDisabled: disabled,
		Comments:         commentDTOs,
		Reactions:        reactions,
		MyReaction:       mine,
		CreatedAt:        post.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *PostService) Update(ctx context.Context, userID, postID uuid.UUID, in UpdatePostInput, isAdmin bool) (*CreatedPostDTO, error) {
	post, err := models.GetFeedPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrNotFound
	}
	if post.AuthorID != userID && !isAdmin {
		return nil, ErrForbidden
	}

	title := post.Title
	content := post.Content
	pinned := post.IsPinned
	if in.Title != nil {
		title = strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, errors.New("title is required")
		}
		if utf8.RuneCountInString(title) > maxTitleLen {
			return nil, errors.New("title is too long")
		}
	}
	if in.Content != nil {
		content = strings.TrimSpace(*in.Content)
		if content == "" {
			return nil, errors.New("content is required")
		}
		if utf8.RuneCountInString(content) > maxContentLen {
			return nil, errors.New("content is too long")
		}
	}
	if in.IsPinned != nil {
		if !isAdmin {
			return nil, errors.New("only admins can pin posts")
		}
		if post.PostType == models.PostTypeBulletin && !*in.IsPinned {
			return nil, errors.New("bulletin posts stay pinned")
		}
		pinned = *in.IsPinned
	}

	updated := &models.Post{ID: post.ID, Title: title, Content: content, IsPinned: pinned, PostType: post.PostType, BoardID: post.BoardID, AuthorID: post.AuthorID}
	if err := models.UpdatePost(ctx, updated); err != nil {
		return nil, err
	}
	return &CreatedPostDTO{
		ID:       updated.ID.String(),
		Title:    updated.Title,
		Content:  updated.Content,
		AuthorID: post.AuthorID.String(),
		BoardID:  post.BoardID.String(),
		PostType: post.PostType,
		IsPinned: updated.IsPinned,
	}, nil
}

func (s *PostService) Delete(ctx context.Context, userID, postID uuid.UUID, isAdmin bool) error {
	post, err := models.GetFeedPostByID(ctx, postID)
	if err != nil {
		return err
	}
	if post == nil {
		return ErrNotFound
	}
	if post.AuthorID != userID && !isAdmin {
		return ErrForbidden
	}
	return models.DeletePost(ctx, postID)
}

func toFeedPostDTO(row models.FeedPost) FeedPostDTO {
	return FeedPostDTO{
		ID:            row.ID.String(),
		Title:         row.Title,
		Snippet:       snippet(row.Content),
		Author:        toAuthorDTO(&row),
		Board:         BoardRefDTO{ID: row.BoardID.String(), Name: row.BoardName},
		CommentCount:  row.CommentCount,
		ReactionCount: row.ReactionCount,
		PostType:      row.PostType,
		IsPinned:      row.IsPinned,
		CreatedAt:     row.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toAuthorDTO(row *models.FeedPost) AuthorDTO {
	return AuthorDTO{
		ID:         row.AuthorID.String(),
		UnitNumber: row.AuthorUnit,
		Name:       row.AuthorName,
	}
}

func snippet(content string) string {
	if utf8.RuneCountInString(content) <= snippetRunes {
		return content
	}
	runes := []rune(content)
	return string(runes[:snippetRunes]) + "…"
}

func normalizePostType(raw string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		v = models.PostTypeStandard
	}
	switch v {
	case models.PostTypeStandard, models.PostTypeBulletin:
		return v, nil
	default:
		return "", errors.New("post_type must be standard or bulletin")
	}
}

func encodeFeedCursor(c feedCursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeFeedCursor(s string) (feedCursor, error) {
	var c feedCursor
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(b, &c)
	return c, err
}
