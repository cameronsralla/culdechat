package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type messageHandlers struct {
	svc *services.MessageService
}

// RegisterMessageRoutes registers direct-message endpoints.
func RegisterMessageRoutes(r gin.IRouter) {
	h := &messageHandlers{svc: &services.MessageService{}}
	auth := r.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/messages/conversations", h.listConversations)
	auth.GET("/messages/conversations/:id", h.getConversation)
	auth.POST("/messages", h.send)
	auth.GET("/messages/recipients", h.searchRecipients)
}

// listConversations godoc
// @Summary List DM conversations
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Success 200 {array} services.ConversationDTO
// @Failure 401 {object} ErrorResponse
// @Router /messages/conversations [get]
func (h *messageHandlers) listConversations(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	out, err := h.svc.ListConversations(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		out = []services.ConversationDTO{}
	}
	c.JSON(http.StatusOK, out)
}

// getConversation godoc
// @Summary Get a DM conversation and recent messages
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Param id path string true "Conversation ID"
// @Param limit query int false "Max messages (default 50)"
// @Param before query string false "Message ID — return older messages"
// @Success 200 {object} services.ConversationDetailDTO
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /messages/conversations/{id} [get]
func (h *messageHandlers) getConversation(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var beforeID *uuid.UUID
	if raw := c.Query("before"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid before"})
			return
		}
		beforeID = &parsed
	}
	out, err := h.svc.GetConversation(c.Request.Context(), userID, id, queryLimit(c), beforeID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// send godoc
// @Summary Send a direct message (creates the thread on first send)
// @Tags messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body services.SendMessageInput true "content plus user_id or unit_number"
// @Success 201 {object} services.SendMessageResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /messages [post]
func (h *messageHandlers) send(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var in services.SendMessageInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid body"})
		return
	}
	out, err := h.svc.Send(c.Request.Context(), userID, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// searchRecipients godoc
// @Summary Search people and units to start a DM
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Param q query string true "Name or unit search"
// @Success 200 {array} services.RecipientDTO
// @Failure 401 {object} ErrorResponse
// @Router /messages/recipients [get]
func (h *messageHandlers) searchRecipients(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	out, err := h.svc.SearchRecipients(c.Request.Context(), userID, c.Query("q"))
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		out = []services.RecipientDTO{}
	}
	c.JSON(http.StatusOK, out)
}
