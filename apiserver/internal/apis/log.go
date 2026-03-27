package apis

import (
	"fmt"
	"net/http"

	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/telemetry"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	middleware "dkhalife.com/tasks/core/internal/utils/middleware"
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
	telemetry.TrackWarning(c, "client_log_warn", "log-handler", req.Message, map[string]string{"route": req.Route})
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
	telemetry.TrackError(c, "client_log_error", "log-handler", fmt.Errorf("%s", req.Message), map[string]string{"route": req.Route})
	c.Status(http.StatusNoContent)
}

func LogRoutes(router *gin.Engine, h *LogsAPIHandler, auth *authMW.AuthMiddleware, limiter *limiter.Limiter) {
	logRoutes := router.Group("/api/v1/log")

	logRoutes.Use(auth.MiddlewareFunc(), middleware.RateLimitMiddleware(limiter))
	{
		logRoutes.POST("/warn", h.Warn)
		logRoutes.POST("/error", h.Error)
	}
}
