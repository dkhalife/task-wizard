package apis

import (
	"net/http"

	"dkhalife.com/tasks/core/internal/services/logging"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	middleware "dkhalife.com/tasks/core/internal/utils/middleware"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
)

type LogsAPIHandler struct {
}

func LogsAPI() *LogsAPIHandler {
	return &LogsAPIHandler{}
}

type LogReq struct {
	Message string `json:"message" binding:"required"`
	Route   string `json:"route" binding:"required"`
}

func (h *LogsAPIHandler) Warn(c *gin.Context) {
	var req LogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	currentIdentity := auth.CurrentIdentity(c)
	log := logging.FromContext(c)
	log.Warnf("Route:%s User:%d Message:%s", req.Route, currentIdentity.UserID, req.Message)
	c.Status(http.StatusNoContent)
}

func (h *LogsAPIHandler) Error(c *gin.Context) {
	var req LogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	currentIdentity := auth.CurrentIdentity(c)
	log := logging.FromContext(c)
	log.Errorf("Route:%s User:%d Message:%s", req.Route, currentIdentity.UserID, req.Message)
	c.Status(http.StatusNoContent)
}

func LogRoutes(router *gin.Engine, h *LogsAPIHandler, auth *jwt.GinJWTMiddleware, limiter *limiter.Limiter) {
	logRoutes := router.Group("/api/v1/log")

	logRoutes.Use(auth.MiddlewareFunc(), middleware.RateLimitMiddleware(limiter))
	{
		logRoutes.POST("/warn", h.Warn)
		logRoutes.POST("/error", h.Error)
	}
}
