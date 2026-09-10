package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
)

type postHandlers struct {
	posts     *services.PostService
	comments  *services.CommentService
	reactions *services.ReactionService
}

// RegisterPostRoutes registers feed, post detail, comments, and reactions.
func RegisterPostRoutes(r gin.IRouter) {
	h := &postHandlers{
		posts:     &services.PostService{},
		comments:  &services.CommentService{},
		reactions: &services.ReactionService{},
	}
	grp := r.Group("/posts", middleware.AuthRequired())
	grp.GET("", h.feed)
	grp.GET("/:postId", h.get)
	grp.PATCH("/:postId", h.update)
	grp.DELETE("/:postId", h.delete)
	grp.POST("/:postId/comments", h.createComment)
	grp.GET("/:postId/comments", h.listComments)
	grp.PATCH("/:postId/comments/:commentId", h.updateComment)
	grp.DELETE("/:postId/comments/:commentId", h.deleteComment)
	grp.PUT("/:postId/reactions", h.react)
	grp.DELETE("/:postId/reactions", h.unreact)
}

// feed godoc
// @Summary General feed
// @Description Paginated posts from all boards. Pinned bulletin posts appear first.
// @Tags posts
// @Produce json
// @Security BearerAuth
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size (default 20, max 50)"
// @Success 200 {object} services.FeedPageDTO
// @Router /posts [get]
func (h *postHandlers) feed(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	out, err := h.posts.ListFeed(c.Request.Context(), nil, queryLimit(c), c.Query("cursor"), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// get godoc
// @Summary Get a post
// @Description Full post including comments and reaction counts. Bulletin posts have comments disabled.
// @Tags posts
// @Produce json
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Success 200 {object} services.PostDetailDTO
// @Failure 404 {object} ErrorResponse
// @Router /posts/{postId} [get]
func (h *postHandlers) get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	postID, ok := parseUUIDParam(c, "postId")
	if !ok {
		return
	}
	out, err := h.posts.Get(c.Request.Context(), postID, userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// update godoc
// @Summary Edit a post
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Param body body services.UpdatePostInput true "Fields"
// @Success 200 {object} services.CreatedPostDTO
// @Router /posts/{postId} [patch]
func (h *postHandlers) update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	postID, ok := parseUUIDParam(c, "postId")
	if !ok {
		return
	}
	var in services.UpdatePostInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.posts.Update(c.Request.Context(), userID, postID, in, currentIsAdmin(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// delete godoc
// @Summary Delete a post
// @Tags posts
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Success 204 {string} string "No Content"
// @Router /posts/{postId} [delete]
func (h *postHandlers) delete(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	postID, ok := parseUUIDParam(c, "postId")
	if !ok {
		return
	}
	if err := h.posts.Delete(c.Request.Context(), userID, postID, currentIsAdmin(c)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// createComment godoc
// @Summary Comment on a post
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Param body body services.CreateCommentInput true "Comment"
// @Success 201 {object} services.CreatedCommentDTO
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /posts/{postId}/comments [post]
func (h *postHandlers) createComment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	postID, ok := parseUUIDParam(c, "postId")
	if !ok {
		return
	}
	var in services.CreateCommentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.comments.Create(c.Request.Context(), userID, postID, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// listComments godoc
// @Summary List comments on a post
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Success 200 {array} services.CommentDTO
// @Router /posts/{postId}/comments [get]
func (h *postHandlers) listComments(c *gin.Context) {
	postID, ok := parseUUIDParam(c, "postId")
	if !ok {
		return
	}
	out, err := h.comments.ListByPost(c.Request.Context(), postID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// updateComment godoc
// @Summary Edit a comment
// @Tags comments
// @Accept json
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Param commentId path string true "Comment ID"
// @Param body body services.CreateCommentInput true "Comment"
// @Success 200 {object} services.CreatedCommentDTO
// @Router /posts/{postId}/comments/{commentId} [patch]
func (h *postHandlers) updateComment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	commentID, ok := parseUUIDParam(c, "commentId")
	if !ok {
		return
	}
	var in services.CreateCommentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.comments.Update(c.Request.Context(), userID, commentID, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// deleteComment godoc
// @Summary Delete a comment
// @Tags comments
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Param commentId path string true "Comment ID"
// @Success 204 {string} string "No Content"
// @Router /posts/{postId}/comments/{commentId} [delete]
func (h *postHandlers) deleteComment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	commentID, ok := parseUUIDParam(c, "commentId")
	if !ok {
		return
	}
	if err := h.comments.Delete(c.Request.Context(), userID, commentID, currentIsAdmin(c)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// react godoc
// @Summary Set a reaction
// @Description One reaction per user per post. Allowed types: like, love, laugh, wow, sad, angry.
// @Tags reactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Param body body services.ReactInput true "Reaction"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorResponse
// @Router /posts/{postId}/reactions [put]
func (h *postHandlers) react(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	postID, ok := parseUUIDParam(c, "postId")
	if !ok {
		return
	}
	var in services.ReactInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.reactions.Upsert(c.Request.Context(), userID, postID, in); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// unreact godoc
// @Summary Remove your reaction
// @Tags reactions
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Success 204 {string} string "No Content"
// @Router /posts/{postId}/reactions [delete]
func (h *postHandlers) unreact(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	postID, ok := parseUUIDParam(c, "postId")
	if !ok {
		return
	}
	if err := h.reactions.Remove(c.Request.Context(), userID, postID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
