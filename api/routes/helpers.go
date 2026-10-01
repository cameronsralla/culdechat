package routes

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cameronsralla/culdechat/middleware"
	"github.com/cameronsralla/culdechat/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/cameronsralla/culdechat/docs"
)

// ErrorResponse is the standard error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}

func writeError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, services.ErrCommentsDisabled), errors.Is(err, services.ErrForbidden), errors.Is(err, services.ErrNotSubscribed):
		status = http.StatusForbidden
	case errors.Is(err, services.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, services.ErrInvalidCredentials), errors.Is(err, services.ErrUnauthorized), errors.Is(err, services.ErrInactiveAccount):
		status = http.StatusUnauthorized
	case strings.Contains(err.Error(), "only admins"), strings.Contains(err.Error(), "cannot offboard"):
		status = http.StatusForbidden
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				c.JSON(http.StatusConflict, ErrorResponse{Error: "already exists"})
				return
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "could not save"})
			return
		}
	}
	c.JSON(status, ErrorResponse{Error: err.Error()})
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid " + name})
		return uuid.Nil, false
	}
	return id, true
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user id in token"})
		return uuid.Nil, false
	}
	return id, true
}

func currentIsAdmin(c *gin.Context) bool {
	v, _ := c.Get("is_admin")
	ok, _ := v.(bool)
	return ok
}

func queryLimit(c *gin.Context) int {
	raw := c.Query("limit")
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return n
}

// NewRouter constructs the gin.Engine with all routes and middleware registered.
func NewRouter() *gin.Engine {
	router := gin.New()
	middleware.ApplyTrustedProxies(router)
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.LimitBodySize())
	router.Use(middleware.RequestLogger())

	api := router.Group("/api")

	api.GET("/health", health)
	if strings.EqualFold(os.Getenv("CULDECHAT_DOCS"), "true") {
		api.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	RegisterAuthRoutes(api)
	RegisterBoardRoutes(api)
	RegisterPostRoutes(api)
	RegisterProfileRoutes(api)
	RegisterAdminRoutes(api)
	RegisterMessageRoutes(api)

	return router
}

// health godoc
// @Summary Health check
// @Tags meta
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func loginLimiter() gin.HandlerFunc {
	return middleware.RateLimit(5, 15*time.Minute)
}

func inviteCompleteLimiter() gin.HandlerFunc {
	return middleware.RateLimit(10, 15*time.Minute)
}
