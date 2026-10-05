// Package messages is the resident-to-resident note. A listed neighbor gets
// an open thread. A hidden unit gets a request they can accept or decline.
package messages

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cameronsralla/culdechat/server/internal/db/dbq"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
)

type Service struct {
	pool *pgxpool.Pool
	q    *dbq.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, q: dbq.New(pool)}
}

type SendInput struct {
	Content        string     `json:"content"`
	UserID         *uuid.UUID `json:"user_id"`
	UnitNumber     string     `json:"unit_number"`
	ConversationID *uuid.UUID `json:"conversation_id"`
}

func (in *SendInput) Validate() error {
	var f httpx.Fields
	n := 0
	if in.UserID != nil {
		n++
	}
	in.UnitNumber = strings.TrimSpace(in.UnitNumber)
	if in.UnitNumber != "" {
		n++
	}
	if in.ConversationID != nil {
		n++
	}
	if n != 1 {
		f.Add("target", "send user_id, unit_number, or conversation_id")
	}
	in.Content = strings.TrimSpace(in.Content)
	if in.Content == "" {
		f.Add("content", "required")
	}
	if len(in.Content) > 2000 {
		f.Add("content", "max 2000 characters")
	}
	return f.Err()
}

type Peer struct {
	ID          *uuid.UUID `json:"id,omitempty"`
	UnitNumber  string     `json:"unit_number"`
	DisplayName string     `json:"display_name,omitempty"`
	Listed      bool       `json:"listed"`
}

type MessageView struct {
	ID        uuid.UUID `json:"id"`
	Mine      bool      `json:"mine"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversationView struct {
	ID            uuid.UUID     `json:"id"`
	Status        string        `json:"status"`
	Incoming      bool          `json:"incoming"`
	RequestedByMe bool          `json:"requested_by_me"`
	Peer          Peer          `json:"peer"`
	LastMessage   string        `json:"last_message,omitempty"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Messages      []MessageView `json:"messages,omitempty"`
}

func (s *Service) List(ctx context.Context, me uuid.UUID) ([]ConversationView, error) {
	rows, err := s.q.ListConversationsForUser(ctx, me)
	if err != nil {
		return nil, err
	}
	out := make([]ConversationView, 0, len(rows))
	for _, row := range rows {
		view, err := s.present(ctx, s.q, me, conversationFromList(row), row.LastMessage, false)
		if err != nil {
			return nil, err
		}
		out = append(out, view)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, me, id uuid.UUID) (ConversationView, error) {
	conv, err := s.loadMine(ctx, s.q, me, id)
	if err != nil {
		return ConversationView{}, err
	}
	msgs, err := s.q.ListDirectMessages(ctx, conv.ID)
	if err != nil {
		return ConversationView{}, err
	}
	last := ""
	if n := len(msgs); n > 0 {
		last = msgs[n-1].Body
	}
	view, err := s.present(ctx, s.q, me, conv, last, false)
	if err != nil {
		return ConversationView{}, err
	}
	view.Messages = views(me, msgs)
	return view, nil
}

func (s *Service) Send(ctx context.Context, me uuid.UUID, in SendInput) (ConversationView, error) {
	if in.ConversationID != nil {
		return s.reply(ctx, me, *in.ConversationID, in.Content)
	}
	target, hidden, err := s.resolve(ctx, me, in)
	if err != nil {
		return ConversationView{}, err
	}
	return s.deliver(ctx, me, target, hidden, in.Content)
}

func (s *Service) Accept(ctx context.Context, me, id uuid.UUID) (ConversationView, error) {
	return s.decide(ctx, me, id, "open")
}

func (s *Service) Decline(ctx context.Context, me, id uuid.UUID) (ConversationView, error) {
	return s.decide(ctx, me, id, "declined")
}

func (s *Service) reply(ctx context.Context, me, id uuid.UUID, body string) (ConversationView, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ConversationView{}, err
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	conv, err := s.loadMine(ctx, q, me, id)
	if err != nil {
		return ConversationView{}, err
	}
	if conv.Status != "open" {
		if conv.Status == "pending" {
			return ConversationView{}, httpx.ErrConflict.WithMessage("accept the message before replying")
		}
		return ConversationView{}, httpx.ErrConflict.WithMessage("this conversation is closed")
	}
	if err := s.append(ctx, q, conv.ID, me, body); err != nil {
		return ConversationView{}, err
	}
	conv, err = q.GetConversation(ctx, conv.ID)
	if err != nil {
		return ConversationView{}, err
	}
	view, err := s.present(ctx, q, me, conv, body, true)
	if err != nil {
		return ConversationView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ConversationView{}, err
	}
	return view, nil
}

func (s *Service) deliver(ctx context.Context, me, target uuid.UUID, hidden bool, body string) (ConversationView, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ConversationView{}, err
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	conv, err := s.ensure(ctx, q, me, target, hidden)
	if err != nil {
		return ConversationView{}, err
	}
	if err := s.append(ctx, q, conv.ID, me, body); err != nil {
		return ConversationView{}, err
	}
	conv, err = q.GetConversation(ctx, conv.ID)
	if err != nil {
		return ConversationView{}, err
	}
	view, err := s.present(ctx, q, me, conv, body, true)
	if err != nil {
		return ConversationView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ConversationView{}, err
	}
	return view, nil
}

func (s *Service) ensure(ctx context.Context, q *dbq.Queries, me, target uuid.UUID, hidden bool) (dbq.Conversation, error) {
	low, high := pair(me, target)
	conv, err := q.FindConversation(ctx, dbq.FindConversationParams{UserLow: low, UserHigh: high})
	if errors.Is(err, pgx.ErrNoRows) {
		status := "open"
		if hidden {
			status = "pending"
		}
		conv, err = q.InsertConversation(ctx, dbq.InsertConversationParams{
			UserLow: low, UserHigh: high, Status: status, RequestedBy: me,
		})
		if isUnique(err) {
			conv, err = q.FindConversation(ctx, dbq.FindConversationParams{UserLow: low, UserHigh: high})
			if err != nil {
				return dbq.Conversation{}, err
			}
			return s.advance(ctx, q, conv, me, hidden)
		}
		return conv, err
	}
	if err != nil {
		return dbq.Conversation{}, err
	}
	return s.advance(ctx, q, conv, me, hidden)
}

func (s *Service) advance(ctx context.Context, q *dbq.Queries, conv dbq.Conversation, me uuid.UUID, hidden bool) (dbq.Conversation, error) {
	switch conv.Status {
	case "open":
		return conv, nil
	case "pending":
		return dbq.Conversation{}, httpx.ErrConflict.WithMessage("waiting for them to accept")
	case "declined":
		status := "pending"
		if !hidden {
			status = "open"
		}
		return q.SetConversationState(ctx, dbq.SetConversationStateParams{ID: conv.ID, Status: status, RequestedBy: me})
	default:
		return dbq.Conversation{}, httpx.ErrConflict
	}
}

func (s *Service) decide(ctx context.Context, me, id uuid.UUID, status string) (ConversationView, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ConversationView{}, err
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	conv, err := s.loadMine(ctx, q, me, id)
	if err != nil {
		return ConversationView{}, err
	}
	if conv.Status != "pending" {
		return ConversationView{}, httpx.ErrConflict.WithMessage("nothing to respond to")
	}
	if conv.RequestedBy == me {
		return ConversationView{}, httpx.ErrConflict.WithMessage("waiting for them to accept")
	}
	conv, err = q.SetConversationState(ctx, dbq.SetConversationStateParams{ID: conv.ID, Status: status, RequestedBy: conv.RequestedBy})
	if err != nil {
		return ConversationView{}, err
	}
	view, err := s.present(ctx, q, me, conv, "", true)
	if err != nil {
		return ConversationView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ConversationView{}, err
	}
	return view, nil
}

func (s *Service) resolve(ctx context.Context, me uuid.UUID, in SendInput) (uuid.UUID, bool, error) {
	if in.UserID != nil {
		u, err := s.q.GetUserContact(ctx, *in.UserID)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && u.Status != "active") {
			return uuid.Nil, false, httpx.ErrNotFound.WithMessage("no one to message")
		}
		if err != nil {
			return uuid.Nil, false, err
		}
		if u.ID == me {
			return uuid.Nil, false, httpx.ErrConflict.WithMessage("you can't message yourself")
		}
		return u.ID, !u.DirectoryOptIn, nil
	}
	id, err := s.q.ActiveHiddenByUnit(ctx, in.UnitNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, httpx.ErrNotFound.WithMessage("no one to message at that unit")
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	if id == me {
		return uuid.Nil, false, httpx.ErrConflict.WithMessage("that's your unit")
	}
	return id, true, nil
}

func (s *Service) append(ctx context.Context, q *dbq.Queries, conversation, sender uuid.UUID, body string) error {
	if _, err := q.InsertDirectMessage(ctx, dbq.InsertDirectMessageParams{
		ConversationID: conversation, SenderID: sender, Body: body,
	}); err != nil {
		return err
	}
	return q.TouchConversation(ctx, conversation)
}

func (s *Service) loadMine(ctx context.Context, q *dbq.Queries, me, id uuid.UUID) (dbq.Conversation, error) {
	conv, err := q.GetConversation(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && conv.UserLow != me && conv.UserHigh != me) {
		return dbq.Conversation{}, httpx.ErrNotFound
	}
	return conv, err
}

func (s *Service) present(ctx context.Context, q *dbq.Queries, me uuid.UUID, conv dbq.Conversation, last string, withMessages bool) (ConversationView, error) {
	contact, err := q.GetUserContact(ctx, otherID(conv, me))
	if errors.Is(err, pgx.ErrNoRows) {
		return ConversationView{}, httpx.ErrNotFound
	}
	if err != nil {
		return ConversationView{}, err
	}
	view := ConversationView{
		ID:            conv.ID,
		Status:        conv.Status,
		Incoming:      conv.Status == "pending" && conv.RequestedBy != me,
		RequestedByMe: conv.RequestedBy == me,
		Peer:          peerFrom(contact),
		LastMessage:   last,
		UpdatedAt:     conv.UpdatedAt,
	}
	if withMessages {
		msgs, err := q.ListDirectMessages(ctx, conv.ID)
		if err != nil {
			return ConversationView{}, err
		}
		view.Messages = views(me, msgs)
		if last == "" && len(view.Messages) > 0 {
			view.LastMessage = view.Messages[len(view.Messages)-1].Body
		}
	}
	return view, nil
}

func peerFrom(u dbq.GetUserContactRow) Peer {
	p := Peer{UnitNumber: u.UnitNumber, Listed: u.DirectoryOptIn && u.Status == "active"}
	if p.Listed {
		id := u.ID
		p.ID = &id
		p.DisplayName = u.DisplayName
	}
	return p
}

func views(me uuid.UUID, msgs []dbq.DirectMessage) []MessageView {
	out := make([]MessageView, len(msgs))
	for i, m := range msgs {
		out[i] = MessageView{ID: m.ID, Mine: m.SenderID == me, Body: m.Body, CreatedAt: m.CreatedAt}
	}
	return out
}

func conversationFromList(row dbq.ListConversationsForUserRow) dbq.Conversation {
	return dbq.Conversation{
		ID: row.ID, UserLow: row.UserLow, UserHigh: row.UserHigh, Status: row.Status,
		RequestedBy: row.RequestedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func otherID(c dbq.Conversation, me uuid.UUID) uuid.UUID {
	if c.UserLow == me {
		return c.UserHigh
	}
	return c.UserLow
}

func pair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if bytes.Compare(a[:], b[:]) < 0 {
		return a, b
	}
	return b, a
}

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
