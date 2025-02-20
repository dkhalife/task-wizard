package user

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"net/http"
	"time"

	"dkhalife.com/tasks/core/config"
	nModel "dkhalife.com/tasks/core/internal/models/notifier"
	uModel "dkhalife.com/tasks/core/internal/models/user"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	"dkhalife.com/tasks/core/internal/utils/email"
	middleware "dkhalife.com/tasks/core/internal/utils/middleware"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
)

type Handler struct {
	userRepo *uRepo.UserRepository
	nRepo    *nRepo.NotificationRepository
	jwtAuth  *jwt.GinJWTMiddleware
	email    *email.EmailSender
}

func NewHandler(ur *uRepo.UserRepository, nRepo *nRepo.NotificationRepository, jwtAuth *jwt.GinJWTMiddleware, email *email.EmailSender, config *config.Config) *Handler {
	return &Handler{
		userRepo: ur,
		nRepo:    nRepo,
		jwtAuth:  jwtAuth,
		email:    email,
	}
}

func (h *Handler) signUp(c *gin.Context) {
	type SignUpReq struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=8,max=45"`
		DisplayName string `json:"displayName" binding:"required"`
	}

	var signupReq SignUpReq
	if err := c.BindJSON(&signupReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	password, err := auth.EncodePassword(signupReq.Password)
	signupReq.DisplayName = html.EscapeString(signupReq.DisplayName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Encoding password failed",
		})
		return
	}

	if err = h.userRepo.CreateUser(c, &uModel.User{
		Password:    password,
		DisplayName: signupReq.DisplayName,
		Email:       signupReq.Email,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Error creating user, an account with this email already exists",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func (h *Handler) GetUserProfile(c *gin.Context) {
	user, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to access user profile",
		})
		return
	}

	notificationSettings, err := h.nRepo.GetUserNotificationSettings(c, user.ID)
	if err != nil {
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

func (h *Handler) resetPassword(c *gin.Context) {
	log := logging.FromContext(c)
	type ResetPasswordReq struct {
		Email string `json:"email" binding:"required,email"`
	}

	var req ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	_, err := h.userRepo.FindByEmail(c, req.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		log.Error("account.handler.resetPassword failed to find user")
		return
	}

	token, err := auth.GenerateEmailResetToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to generate reset token",
		})
		return
	}

	err = h.userRepo.SetPasswordResetToken(c, req.Email, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to set reset token",
		})
		return
	}

	err = h.email.SendResetPasswordEmail(c, req.Email, token)
	if err != nil {
		log.Errorw("account.handler.resetPassword failed to send email", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to send email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) updateUserPassword(c *gin.Context) {
	logger := logging.FromContext(c)

	code := c.Query("c")

	email, code, err := email.DecodeEmailAndCode(code)
	if err != nil {
		logger.Errorw("account.handler.verify failed to decode email and code", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reset code",
		})
		return

	}

	type RequestBody struct {
		Password string `json:"password" binding:"required,min=8,max=32"`
	}

	var body RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Errorw("user.handler.resetAccountPassword failed to bind", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "New password was not provided",
		})
		return

	}

	password, err := auth.EncodePassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	err = h.userRepo.UpdatePasswordByToken(c.Request.Context(), email, code, password)
	if err != nil {
		logger.Errorw("account.handler.resetAccountPassword failed to reset password", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to reset password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) CreateLongLivedToken(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to create tokens",
		})
		return
	}

	type TokenRequest struct {
		Name string `json:"name" binding:"required"`
	}

	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate random token",
		})
		return
	}

	timestamp := time.Now().Unix()
	hashInput := fmt.Sprintf("%d:%d:%x", currentUser.ID, timestamp, randomBytes)
	hash := sha256.Sum256([]byte(hashInput))

	token := hex.EncodeToString(hash[:])

	tokenModel, err := h.userRepo.StoreAPIToken(c, currentUser.ID, req.Name, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to store token",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": gin.H{
			"id":        tokenModel.ID,
			"name":      tokenModel.Name,
			"token":     tokenModel.Token,
			"createdAt": tokenModel.CreatedAt,
		},
	})
}

func (h *Handler) GetAllUserToken(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to fetch tokens",
		})
		return
	}

	tokens, err := h.userRepo.GetAllUserTokens(c, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user tokens",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
	})
}

func (h *Handler) DeleteUserToken(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to delete tokens",
		})
		return
	}

	tokenID := c.Param("id")

	err := h.userRepo.DeleteAPIToken(c, currentUser.ID, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete the token",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *Handler) UpdateNotificationSettings(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to update notification target",
		})
		return
	}

	type Request struct {
		Provider nModel.NotificationProvider       `json:"provider" binding:"required"`
		Triggers nModel.NotificationTriggerOptions `json:"triggers" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	err := h.userRepo.UpdateNotificationSettings(c, currentUser.ID, req.Provider, req.Triggers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update notification target",
		})
		return
	}

	// TODO: Reschedule all notifications for this user

	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *Handler) updateUserPasswordLoggedInOnly(c *gin.Context) {
	logger := logging.FromContext(c)

	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to updated password",
		})
		return
	}

	type RequestBody struct {
		Password string `json:"password" binding:"required,min=8,max=32"`
	}

	var body RequestBody

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Errorw("user.handler.resetAccountPassword failed to bind", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	password, err := auth.EncodePassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to encode password",
		})
		return
	}

	err = h.userRepo.UpdatePasswordByUserId(c.Request.Context(), currentUser.ID, password)
	if err != nil {
		logger.Errorw("account.handler.resetAccountPassword failed to reset password", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to reset password",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func Routes(router *gin.Engine, h *Handler, auth *jwt.GinJWTMiddleware, limiter *limiter.Limiter) {
	userRoutes := router.Group("api/v1/users")
	userRoutes.Use(auth.MiddlewareFunc(), middleware.RateLimitMiddleware(limiter))
	{
		userRoutes.GET("/profile", h.GetUserProfile)
		userRoutes.POST("/tokens", h.CreateLongLivedToken)
		userRoutes.GET("/tokens", h.GetAllUserToken)
		userRoutes.DELETE("/tokens/:id", h.DeleteUserToken)
		userRoutes.PUT("/notifications", h.UpdateNotificationSettings)
		userRoutes.PUT("change_password", h.updateUserPasswordLoggedInOnly)
	}

	authRoutes := router.Group("api/v1/auth")
	authRoutes.Use(middleware.RateLimitMiddleware(limiter))
	{
		authRoutes.POST("/", h.signUp)
		authRoutes.POST("login", auth.LoginHandler)
		authRoutes.GET("refresh", auth.RefreshHandler)
		authRoutes.POST("reset", h.resetPassword)
		authRoutes.POST("password", h.updateUserPassword)
	}
}
