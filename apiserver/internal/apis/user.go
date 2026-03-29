package apis

import (
	"net/http"

	"dkhalife.com/tasks/core/config"
	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/models"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/services/users"
	"dkhalife.com/tasks/core/internal/telemetry"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	middleware "dkhalife.com/tasks/core/internal/utils/middleware"
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
)

type UsersAPIHandler struct {
	userRepo    uRepo.IUserRepo
	userService *users.UserService
	nRepo       *nRepo.NotificationRepository
	cfg         *config.Config
}

func UsersAPI(ur uRepo.IUserRepo, nRepo *nRepo.NotificationRepository, us *users.UserService, cfg *config.Config) *UsersAPIHandler {
	return &UsersAPIHandler{
		userRepo:    ur,
		userService: us,
		nRepo:       nRepo,
		cfg:         cfg,
	}
}

func (h *UsersAPIHandler) GetAuthConfig(c *gin.Context) {
	if !h.cfg.Entra.Enabled {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":   true,
		"tenant_id": h.cfg.Entra.TenantID,
		"client_id": h.cfg.Entra.ClientID,
		"audience":  h.cfg.Entra.Audience,
	})
}

func (h *UsersAPIHandler) GetUserProfile(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	log := logging.FromContext(c)

	user, err := h.userRepo.GetUser(c, currentIdentity.UserID)
	if err != nil {
		log.Errorf("failed to get user: %s", err.Error())
		telemetry.TrackError(c, "user_profile_get_failed", "user-handler", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user profile",
		})
		return
	}

	notificationSettings, err := h.nRepo.GetUserNotificationSettings(c, currentIdentity.UserID)
	if err != nil {
		log.Errorf("failed to get notification settings: %s", err.Error())
		telemetry.TrackError(c, "user_profile_get_failed", "user-handler", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get notification settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"notifications":         notificationSettings,
			"deletion_requested_at": user.DeletionRequestedAt,
		},
	})
}

func (h *UsersAPIHandler) UpdateNotificationSettings(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var req models.NotificationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.TrackWarning(c, "user_bind_failed", "user-handler", err.Error(), nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	status, response := h.userService.UpdateNotificationSettings(c, currentIdentity.UserID, req)
	c.JSON(status, response)
}

func (h *UsersAPIHandler) RequestDeletion(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	status, response := h.userService.RequestDeletion(c, currentIdentity.UserID)
	c.JSON(status, response)
}

func (h *UsersAPIHandler) CancelDeletion(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	status, response := h.userService.CancelDeletion(c, currentIdentity.UserID)
	c.JSON(status, response)
}

func UserRoutes(router *gin.Engine, h *UsersAPIHandler, authMiddleware *authMW.AuthMiddleware, limiter *limiter.Limiter) {
	userRoutes := router.Group("api/v1/users")
	userRoutes.Use(authMiddleware.MiddlewareFunc(), middleware.RateLimitMiddleware(limiter))
	{
		userRoutes.GET("/profile", authMW.ScopeMiddleware(models.ApiTokenScopeUserRead), h.GetUserProfile)
		userRoutes.PUT("/notifications", authMW.ScopeMiddleware(models.ApiTokenScopeUserWrite), middleware.DeletionGuardMiddleware(), h.UpdateNotificationSettings)
		userRoutes.POST("/deletion", authMW.ScopeMiddleware(models.ApiTokenScopeUserWrite), h.RequestDeletion)
		userRoutes.DELETE("/deletion", authMW.ScopeMiddleware(models.ApiTokenScopeUserWrite), h.CancelDeletion)
	}

	authRoutes := router.Group("api/v1/auth")
	authRoutes.Use(middleware.RateLimitMiddleware(limiter))
	{
		authRoutes.GET("/config", h.GetAuthConfig)
	}
}
