package backend

import (
	"net/http"

	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/utils/middleware"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
)

type Handler struct {
}

func NewHandler(uRepo uRepo.IUserRepo) *Handler {
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
