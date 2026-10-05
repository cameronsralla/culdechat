package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
)

type boardHandlers struct {
	boards *services.BoardService
	posts  *services.PostService
}

// RegisterBoardRoutes registers board related endpoints under /boards.
func RegisterBoardRoutes(r gin.IRouter) {
	h := &boardHandlers{
		boards: &services.BoardService{},
		posts:  &services.PostService{},
	}
	grp := r.Group("/boards", middleware.AuthRequired())
	grp.GET("", h.list)
	grp.POST("", h.create)
	grp.GET("/:boardId", h.get)
	grp.POST("/:boardId/subscribe", h.subscribe)
	grp.GET("/:boardId/posts", h.listPosts)
	grp.POST("/:boardId/posts", h.createPost)
}

// list godoc
// @Summary List boards
// @Tags boards
// @Produce json
// @Security BearerAuth
// @Success 200 {array} services.BoardDTO
// @Failure 401 {object} ErrorResponse
// @Router /boards [get]
func (h *boardHandlers) list(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	out, err := h.boards.List(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// create godoc
// @Summary Create a board
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body services.CreateBoardInput true "Board"
// @Success 201 {object} services.BoardDTO
// @Failure 400 {object} ErrorResponse
// @Router /boards [post]
func (h *boardHandlers) create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var in services.CreateBoardInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.boards.Create(c.Request.Context(), userID, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// get godoc
// @Summary Get a board
// @Tags boards
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Success 200 {object} services.BoardDTO
// @Router /boards/{boardId} [get]
func (h *boardHandlers) get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	boardID, ok := parseUUIDParam(c, "boardId")
	if !ok {
		return
	}
	out, err := h.boards.Get(c.Request.Context(), userID, boardID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// subscribe godoc
// @Summary Subscribe or unsubscribe
// @Description Toggles the current user's subscription to the board.
// @Tags boards
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Success 200 {object} services.SubscribeResponse
// @Failure 404 {object} ErrorResponse
// @Router /boards/{boardId}/subscribe [post]
func (h *boardHandlers) subscribe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	boardID, ok := parseUUIDParam(c, "boardId")
	if !ok {
		return
	}
	out, err := h.boards.ToggleSubscribe(c.Request.Context(), userID, boardID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// listPosts godoc
// @Summary List posts on a board
// @Tags posts
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size (default 20, max 50)"
// @Success 200 {object} services.FeedPageDTO
// @Router /boards/{boardId}/posts [get]
func (h *boardHandlers) listPosts(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	boardID, ok := parseUUIDParam(c, "boardId")
	if !ok {
		return
	}
	out, err := h.posts.ListFeed(c.Request.Context(), &boardID, queryLimit(c), c.Query("cursor"), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// createPost godoc
// @Summary Create a post on a board
// @Description Admins may set post_type=bulletin (auto-pinned, comments disabled) or is_pinned=true.
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param body body services.CreatePostInput true "Post"
// @Success 201 {object} services.CreatedPostDTO
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /boards/{boardId}/posts [post]
func (h *boardHandlers) createPost(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	boardID, ok := parseUUIDParam(c, "boardId")
	if !ok {
		return
	}
	var in services.CreatePostInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.posts.Create(c.Request.Context(), userID, boardID, in, currentIsAdmin(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}
