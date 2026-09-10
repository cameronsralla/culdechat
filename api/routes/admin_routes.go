package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
)

type adminHandlers struct {
	auth *services.AuthService
}

// RegisterAdminRoutes registers business-admin endpoints.
func RegisterAdminRoutes(r gin.IRouter) {
	h := &adminHandlers{auth: &services.AuthService{}}
	grp := r.Group("/admin", middleware.AdminRequired())
	grp.GET("/users", h.listUsers)
	grp.POST("/users/:userId/offboard", h.offboard)
}

// listUsers godoc
// @Summary List residents
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} services.AdminUserDTO
// @Router /admin/users [get]
func (h *adminHandlers) listUsers(c *gin.Context) {
	out, err := h.auth.ListUsers(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// offboard godoc
// @Summary Offboard a resident
// @Tags admin
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Success 204 {string} string "No Content"
// @Router /admin/users/{userId}/offboard [post]
func (h *adminHandlers) offboard(c *gin.Context) {
	userID, ok := parseUUIDParam(c, "userId")
	if !ok {
		return
	}
	if err := h.auth.Offboard(c.Request.Context(), userID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
