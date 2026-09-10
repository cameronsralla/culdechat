package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
)

type authHandlers struct {
	svc *services.AuthService
}

// RegisterAuthRoutes registers authentication-related routes under /auth.
func RegisterAuthRoutes(r gin.IRouter) {
	h := &authHandlers{svc: &services.AuthService{}}
	auth := r.Group("/auth")

	auth.POST("/register", middleware.AdminRequired(), h.register)
	auth.POST("/complete-registration", inviteCompleteLimiter(), h.completeRegistration)
	auth.POST("/login", loginLimiter(), h.login)
	auth.POST("/refresh", inviteCompleteLimiter(), h.refresh)
	auth.POST("/logout", h.logout)
	auth.POST("/logout-all", middleware.AuthRequired(), h.logoutAll)
	auth.POST("/change-password", middleware.AuthRequired(), h.changePassword)
	auth.GET("/me", middleware.AuthRequired(), h.me)
}

// register godoc
// @Summary Invite a resident
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body services.RegisterInput true "Resident email and unit"
// @Success 201 {object} services.RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/register [post]
func (h *authHandlers) register(c *gin.Context) {
	var in services.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.svc.Register(c.Request.Context(), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// completeRegistration godoc
// @Summary Complete resident registration
// @Tags auth
// @Accept json
// @Produce json
// @Param body body services.CompleteRegistrationInput true "Invite completion"
// @Success 200 {object} services.AuthResponse
// @Router /auth/complete-registration [post]
func (h *authHandlers) completeRegistration(c *gin.Context) {
	var in services.CompleteRegistrationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.svc.CompleteRegistration(c.Request.Context(), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// login godoc
// @Summary Log in
// @Tags auth
// @Accept json
// @Produce json
// @Param body body services.LoginInput true "Credentials"
// @Success 200 {object} services.AuthResponse
// @Router /auth/login [post]
func (h *authHandlers) login(c *gin.Context) {
	var in services.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.svc.Login(c.Request.Context(), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// refresh godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body services.RefreshInput true "Refresh token"
// @Success 200 {object} services.AuthResponse
// @Router /auth/refresh [post]
func (h *authHandlers) refresh(c *gin.Context) {
	var in services.RefreshInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.svc.Refresh(c.Request.Context(), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// logout godoc
// @Summary Revoke a refresh token
// @Tags auth
// @Accept json
// @Param body body services.LogoutInput true "Refresh token"
// @Success 204 {string} string "No Content"
// @Router /auth/logout [post]
func (h *authHandlers) logout(c *gin.Context) {
	var in services.LogoutInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.svc.Logout(c.Request.Context(), in); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// logoutAll godoc
// @Summary Revoke all refresh tokens for the current user
// @Tags auth
// @Security BearerAuth
// @Success 204 {string} string "No Content"
// @Router /auth/logout-all [post]
func (h *authHandlers) logoutAll(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	if err := h.svc.LogoutAll(c.Request.Context(), userID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// changePassword godoc
// @Summary Change password
// @Tags auth
// @Accept json
// @Security BearerAuth
// @Param body body services.ChangePasswordInput true "Passwords"
// @Success 204 {string} string "No Content"
// @Router /auth/change-password [post]
func (h *authHandlers) changePassword(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var in services.ChangePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), userID, in); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// me godoc
// @Summary Current user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.MeDTO
// @Router /auth/me [get]
func (h *authHandlers) me(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	out, err := h.svc.Me(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
