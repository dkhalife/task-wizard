package backend

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"taskwiz.app/core/internal/utils/middleware"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ping(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain", []byte("pong"))
}

func Routes(router *gin.Engine, h *Handler, limiter *limiter.Limiter) {
	backendRoutes := router.Group("/")
	backendRoutes.Use(middleware.RateLimitMiddleware(limiter))
	{
		backendRoutes.GET("/ping", h.ping)
	}
}
