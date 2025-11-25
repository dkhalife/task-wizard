package apis

import (
	"net/http"
	"strconv"

	"dkhalife.com/tasks/core/config"
	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/models"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/services/users"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	"dkhalife.com/tasks/core/internal/utils/email"
	middleware "dkhalife.com/tasks/core/internal/utils/middleware"
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
)

type UsersAPIHandler struct {
	userRepo    uRepo.IUserRepo
	userService *users.UserService
	nRepo       *nRepo.NotificationRepository
	email       email.IEmailSender
}

func UsersAPI(ur uRepo.IUserRepo, nRepo *nRepo.NotificationRepository, us *users.UserService, email email.IEmailSender, config *config.Config) *UsersAPIHandler {
	return &UsersAPIHandler{
		userRepo:    ur,
		userService: us,
		nRepo:       nRepo,
		email:       email,
	}
}

func (h *UsersAPIHandler) GetUserProfile(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	log := logging.FromContext(c)

	user, err := h.userRepo.GetUser(c, currentIdentity.UserID)
	if err != nil {
		log.Errorf("failed to get user: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user",
		})
		return
	}

	notificationSettings, err := h.nRepo.GetUserNotificationSettings(c, currentIdentity.UserID)
	if err != nil {
		log.Errorf("failed to get notification settings: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get notification settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"display_name":  user.DisplayName,
			"notifications": notificationSettings,
		},
	})
}

func (h *UsersAPIHandler) CreateAppToken(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var req models.CreateAppTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	status, response := h.userService.CreateAppToken(c, currentIdentity.UserID, req)
	c.JSON(status, response)
}

func (h *UsersAPIHandler) GetAllUserToken(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	status, response := h.userService.GetAllAppTokens(c, currentIdentity.UserID)
	c.JSON(status, response)
}

func (h *UsersAPIHandler) DeleteUserToken(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	tokenIDRaw := c.Param("id")
	if tokenIDRaw == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token ID is required",
		})
		return
	}

	tokenID, err := strconv.Atoi(tokenIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token ID",
		})
		return
	}

	status, response := h.userService.DeleteAppToken(c, currentIdentity.UserID, tokenID)
	c.JSON(status, response)
}

func (h *UsersAPIHandler) UpdateNotificationSettings(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var req models.NotificationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	status, response := h.userService.UpdateNotificationSettings(c, currentIdentity.UserID, req)
	c.JSON(status, response)
}

func UserRoutes(router *gin.Engine, h *UsersAPIHandler, auth *authMW.AuthMiddleware, limiter *limiter.Limiter) {
	userRoutes := router.Group("api/v1/users")
	userRoutes.Use(auth.MiddlewareFunc(), middleware.RateLimitMiddleware(limiter))
	{
		userRoutes.GET("/profile", authMW.ScopeMiddleware(models.ApiTokenScopeUserRead), h.GetUserProfile)
		userRoutes.POST("/tokens", authMW.ScopeMiddleware(models.ApiTokenScopeTokensWrite), h.CreateAppToken)
		userRoutes.GET("/tokens", authMW.ScopeMiddleware(models.ApiTokenScopeTokensWrite), h.GetAllUserToken)
		userRoutes.DELETE("/tokens/:id", authMW.ScopeMiddleware(models.ApiTokenScopeTokensWrite), h.DeleteUserToken)
		userRoutes.PUT("/notifications", authMW.ScopeMiddleware(models.ApiTokenScopeUserWrite), h.UpdateNotificationSettings)
	}
}
