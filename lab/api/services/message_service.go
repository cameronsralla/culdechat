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

const maxMessageLen = 2000

type MessageService struct{}

type SendMessageInput struct {
	Content    string  `json:"content"`
	UserID     *string `json:"user_id"`
	UnitNumber *string `json:"unit_number"`
}

type PeerDTO struct {
	ID                string  `json:"id"`
	UnitNumber        string  `json:"unit_number"`
	Name              *string `json:"name,omitempty"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
	DirectoryOptIn    bool    `json:"directory_opt_in"`
}

type MessageDTO struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type ConversationDTO struct {
	ID          string      `json:"id"`
	Peer        PeerDTO     `json:"peer"`
	LastMessage *MessageDTO `json:"last_message,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type ConversationDetailDTO struct {
	ID       string       `json:"id"`
	Peer     PeerDTO      `json:"peer"`
	Messages []MessageDTO `json:"messages"`
}

type SendMessageResult struct {
	ConversationID string     `json:"conversation_id"`
	Message        MessageDTO `json:"message"`
	Created        bool       `json:"created"`
}

type RecipientDTO struct {
	Kind              string  `json:"kind"` // "user" or "unit"
	ID                *string `json:"id,omitempty"`
	Name              *string `json:"name,omitempty"`
	UnitNumber        string  `json:"unit_number"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
}

func (s *MessageService) ListConversations(ctx context.Context, userID uuid.UUID) ([]ConversationDTO, error) {
	items, err := models.ListConversationsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ConversationDTO, 0, len(items))
	for _, item := range items {
		name := item.PeerName
		photo := item.PeerProfilePictureURL
		if !item.PeerDirectoryOptIn {
			name = nil
			photo = nil
		}
		dto := ConversationDTO{
			ID: item.ID.String(),
			Peer: PeerDTO{
				ID:                item.PeerID.String(),
				UnitNumber:        item.PeerUnit,
				Name:              name,
				ProfilePictureURL: photo,
				DirectoryOptIn:    item.PeerDirectoryOptIn,
			},
			UpdatedAt: item.UpdatedAt,
		}
		if item.LastMessageID != nil && item.LastMessageContent != nil && item.LastMessageSenderID != nil && item.LastMessageAt != nil {
			dto.LastMessage = &MessageDTO{
				ID:             item.LastMessageID.String(),
				ConversationID: item.ID.String(),
				SenderID:       item.LastMessageSenderID.String(),
				Content:        *item.LastMessageContent,
				CreatedAt:      *item.LastMessageAt,
			}
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *MessageService) GetConversation(ctx context.Context, userID, conversationID uuid.UUID, limit int, beforeID *uuid.UUID) (*ConversationDetailDTO, error) {
	c, err := models.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrNotFound
	}
	if c.UserLowID != userID && c.UserHighID != userID {
		return nil, ErrForbidden
	}
	peerID := c.UserHighID
	if c.UserHighID == userID {
		peerID = c.UserLowID
	}
	peer, err := peerDTO(ctx, peerID)
	if err != nil {
		return nil, err
	}
	msgs, err := models.ListMessages(ctx, conversationID, limit, beforeID)
	if err != nil {
		return nil, err
	}
	out := &ConversationDetailDTO{
		ID:       c.ID.String(),
		Peer:     *peer,
		Messages: make([]MessageDTO, 0, len(msgs)),
	}
	for _, m := range msgs {
		out.Messages = append(out.Messages, messageDTO(m))
	}
	return out, nil
}

func (s *MessageService) Send(ctx context.Context, senderID uuid.UUID, in SendMessageInput) (*SendMessageResult, error) {
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return nil, errors.New("content is required")
	}
	if utf8.RuneCountInString(content) > maxMessageLen {
		return nil, errors.New("message is too long")
	}

	recipientID, err := s.resolveRecipient(ctx, senderID, in)
	if err != nil {
		return nil, err
	}
	if recipientID == senderID {
		return nil, errors.New("cannot message yourself")
	}

	recipient, err := models.GetUserByID(ctx, recipientID)
	if err != nil {
		return nil, err
	}
	if recipient == nil || recipient.Status != models.UserStatusActive {
		return nil, ErrNotFound
	}

	created := false
	c, err := models.GetConversationByPair(ctx, senderID, recipientID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		c, err = models.InsertConversation(ctx, senderID, recipientID)
		if err != nil {
			// Race: another request created it — reload.
			existing, getErr := models.GetConversationByPair(ctx, senderID, recipientID)
			if getErr != nil {
				return nil, getErr
			}
			if existing == nil {
				return nil, err
			}
			c = existing
		} else {
			created = true
		}
	}

	m := &models.Message{
		ConversationID: c.ID,
		SenderID:       senderID,
		Content:        content,
	}
	if err := models.InsertMessage(ctx, m); err != nil {
		return nil, err
	}
	if err := models.TouchConversation(ctx, c.ID); err != nil {
		return nil, err
	}

	return &SendMessageResult{
		ConversationID: c.ID.String(),
		Message:        messageDTO(*m),
		Created:        created,
	}, nil
}

func (s *MessageService) SearchRecipients(ctx context.Context, userID uuid.UUID, q string) ([]RecipientDTO, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []RecipientDTO{}, nil
	}
	if utf8.RuneCountInString(q) > 64 {
		return nil, errors.New("search query is too long")
	}
	hits, err := models.SearchMessageRecipients(ctx, userID, q, 20)
	if err != nil {
		return nil, err
	}
	out := make([]RecipientDTO, 0, len(hits))
	for _, h := range hits {
		out = append(out, RecipientDTO{
			Kind:              h.Kind,
			ID:                h.UserID,
			Name:              h.Name,
			UnitNumber:        h.UnitNumber,
			ProfilePictureURL: h.ProfilePictureURL,
		})
	}
	return out, nil
}

func (s *MessageService) resolveRecipient(ctx context.Context, senderID uuid.UUID, in SendMessageInput) (uuid.UUID, error) {
	hasUser := in.UserID != nil && strings.TrimSpace(*in.UserID) != ""
	hasUnit := in.UnitNumber != nil && strings.TrimSpace(*in.UnitNumber) != ""
	if hasUser == hasUnit {
		return uuid.Nil, errors.New("provide either user_id or unit_number")
	}
	if hasUser {
		id, err := uuid.Parse(strings.TrimSpace(*in.UserID))
		if err != nil {
			return uuid.Nil, errors.New("invalid user_id")
		}
		return id, nil
	}
	unit := strings.TrimSpace(*in.UnitNumber)
	u, err := models.GetActiveUserByUnit(ctx, unit)
	if err != nil {
		return uuid.Nil, err
	}
	if u == nil {
		return uuid.Nil, ErrNotFound
	}
	_ = senderID
	return u.ID, nil
}

func peerDTO(ctx context.Context, peerID uuid.UUID) (*PeerDTO, error) {
	u, err := models.GetUserByID(ctx, peerID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	dto := &PeerDTO{
		ID:             u.ID.String(),
		UnitNumber:     u.UnitNumber,
		DirectoryOptIn: u.IsDirectoryOptIn,
	}
	if u.IsDirectoryOptIn {
		if strings.TrimSpace(u.Name) != "" {
			name := u.Name
			dto.Name = &name
		}
		dto.ProfilePictureURL = u.ProfilePictureURL
	}
	return dto, nil
}

func messageDTO(m models.Message) MessageDTO {
	return MessageDTO{
		ID:             m.ID.String(),
		ConversationID: m.ConversationID.String(),
		SenderID:       m.SenderID.String(),
		Content:        m.Content,
		CreatedAt:      m.CreatedAt,
	}
}
