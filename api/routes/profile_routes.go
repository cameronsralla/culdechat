package routes

import (
	"net/http"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
)

type profileHandlers struct {
	svc *services.ProfileService
}

// RegisterProfileRoutes registers profile and directory endpoints.
func RegisterProfileRoutes(r gin.IRouter) {
	h := &profileHandlers{svc: &services.ProfileService{}}

	profile := r.Group("/profile", middleware.AuthRequired())
	profile.GET("/me", h.getMe)
	profile.PATCH("/me", h.updateMe)
	profile.POST("/me/photo", h.uploadPhoto)

	r.GET("/directory", middleware.AuthRequired(), h.directory)
	r.GET("/media/profile/:filename", middleware.AuthRequired(), h.media)
}

// getMe godoc
// @Summary Get my profile
// @Tags profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.ProfileDTO
// @Router /profile/me [get]
func (h *profileHandlers) getMe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	out, err := h.svc.Get(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// updateMe godoc
// @Summary Update my profile
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body services.UpdateProfileInput true "Profile fields"
// @Success 200 {object} services.ProfileDTO
// @Router /profile/me [patch]
func (h *profileHandlers) updateMe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var in services.UpdateProfileInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	out, err := h.svc.Update(c.Request.Context(), userID, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// directory godoc
// @Summary Resident directory
// @Description Active users who opted in. Shows name and unit number.
// @Tags profile
// @Produce json
// @Security BearerAuth
// @Success 200 {array} services.DirectoryUserDTO
// @Router /directory [get]
func (h *profileHandlers) directory(c *gin.Context) {
	out, err := h.svc.ListDirectory(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// uploadPhoto godoc
// @Summary Upload profile photo
// @Tags profile
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param photo formData file true "jpeg, png, or webp, max 2MB"
// @Success 200 {object} services.ProfileDTO
// @Router /profile/me/photo [post]
func (h *profileHandlers) uploadPhoto(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "photo file is required"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "could not read photo"})
		return
	}
	defer f.Close()
	out, err := h.svc.UploadPhoto(c.Request.Context(), userID, f)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// media godoc
// @Summary Fetch an uploaded profile photo
// @Tags profile
// @Security BearerAuth
// @Param filename path string true "Filename"
// @Success 200 {file} file
// @Router /media/profile/{filename} [get]
func (h *profileHandlers) media(c *gin.Context) {
	path, err := services.ResolveProfilePhotoPath(c.Param("filename"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.File(path)
}
