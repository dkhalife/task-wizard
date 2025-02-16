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
	jwtAuth  *jwt.GinJWTMiddleware
	email    *email.EmailSender
}

func NewHandler(ur *uRepo.UserRepository, jwtAuth *jwt.GinJWTMiddleware, email *email.EmailSender, config *config.Config) *Handler {
	return &Handler{
		userRepo: ur,
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
		c.JSON(400, gin.H{
			"error": "Invalid request",
		})
		return
	}
	password, err := auth.EncodePassword(signupReq.Password)
	signupReq.DisplayName = html.EscapeString(signupReq.DisplayName)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error encoding password",
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
		c.JSON(500, gin.H{
			"error": "Error creating user, email already exists",
		})
		return
	}

	c.JSON(201, gin.H{})
}

func (h *Handler) GetUserProfile(c *gin.Context) {
	user, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting user",
		})
		return
	}
	c.JSON(200, gin.H{
		"user": gin.H{
			"display_name":      user.DisplayName,
			"notification_type": user.NotificationType,
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
			"error": "Invalid request",
		})
		return
	}
	_, err := h.userRepo.FindByEmail(c, req.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		log.Error("account.handler.resetPassword failed to find user")
		return
	}
	// generate a random password:
	token, err := auth.GenerateEmailResetToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to generate token",
		})
		return
	}

	err = h.userRepo.SetPasswordResetToken(c, req.Email, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to generate password",
		})
		return
	}
	// send an email to the user with the new password:
	err = h.email.SendResetPasswordEmail(c, req.Email, token)
	if err != nil {
		log.Errorw("account.handler.resetPassword failed to send email", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to send email",
		})
		return
	}

	// send an email to the user with the new password:
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) updateUserPassword(c *gin.Context) {
	logger := logging.FromContext(c)
	// read the code from query param:
	code := c.Query("c")
	email, code, err := email.DecodeEmailAndCode(code)
	if err != nil {
		logger.Errorw("account.handler.verify failed to decode email and code", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid code",
		})
		return

	}
	// read password from body:
	type RequestBody struct {
		Password string `json:"password" binding:"required,min=8,max=32"`
	}
	var body RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Errorw("user.handler.resetAccountPassword failed to bind", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return

	}
	password, err := auth.EncodePassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to process password",
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}
	type TokenRequest struct {
		Name string `json:"name" binding:"required"`
	}
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Step 1: Generate a secure random number
	randomBytes := make([]byte, 16) // 128 bits are enough for strong randomness
	_, err := rand.Read(randomBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate random part of the token"})
		return
	}

	timestamp := time.Now().Unix()
	hashInput := fmt.Sprintf("%s:%d:%x", currentUser.ID, timestamp, randomBytes)
	hash := sha256.Sum256([]byte(hashInput))

	token := hex.EncodeToString(hash[:])

	tokenModel, err := h.userRepo.StoreAPIToken(c, currentUser.ID, req.Name, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store the token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": gin.H{
		"id":        tokenModel.ID,
		"name":      tokenModel.Name,
		"token":     tokenModel.Token,
		"createdAt": tokenModel.CreatedAt,
	}})
}

func (h *Handler) GetAllUserToken(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	tokens, err := h.userRepo.GetAllUserTokens(c, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})

}

func (h *Handler) DeleteUserToken(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	tokenID := c.Param("id")

	err := h.userRepo.DeleteAPIToken(c, currentUser.ID, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) UpdateNotificationTarget(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	type Request struct {
		Type nModel.NotificationType `json:"type" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.userRepo.UpdateNotificationTarget(c, currentUser.ID, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification target"})
		return
	}

	err = h.userRepo.UpdateNotificationTargetForAllNotifications(c, currentUser.ID, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification target for all notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) updateUserPasswordLoggedInOnly(c *gin.Context) {
	logger := logging.FromContext(c)
	type RequestBody struct {
		Password string `json:"password" binding:"required,min=8,max=32"`
	}
	var body RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Errorw("user.handler.resetAccountPassword failed to bind", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	password, err := auth.EncodePassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to process password",
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
	c.JSON(http.StatusOK, gin.H{})
}

func Routes(router *gin.Engine, h *Handler, auth *jwt.GinJWTMiddleware, limiter *limiter.Limiter) {
	userRoutes := router.Group("api/v1/users")
	userRoutes.Use(auth.MiddlewareFunc(), middleware.RateLimitMiddleware(limiter))
	{
		userRoutes.GET("/profile", h.GetUserProfile)
		userRoutes.POST("/tokens", h.CreateLongLivedToken)
		userRoutes.GET("/tokens", h.GetAllUserToken)
		userRoutes.DELETE("/tokens/:id", h.DeleteUserToken)
		userRoutes.PUT("/targets", h.UpdateNotificationTarget)
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
