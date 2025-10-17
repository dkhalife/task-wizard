package apis

import (
	"fmt"
	"net/http"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	middleware "dkhalife.com/tasks/core/internal/utils/middleware"
	"dkhalife.com/tasks/core/internal/utils/oauth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
)

type OAuthAPIHandler struct {
	userRepo      uRepo.IUserRepo
	oauthProvider *oauth.OAuthProvider
	cfg           *config.Config
	jwtMiddleware *jwt.GinJWTMiddleware
}

func OAuthAPI(
	ur uRepo.IUserRepo,
	cfg *config.Config,
	jwtMiddleware *jwt.GinJWTMiddleware,
) (*OAuthAPIHandler, error) {
	provider, err := oauth.NewOAuthProvider(&cfg.OAuth)
	if err != nil {
		return nil, err
	}

	return &OAuthAPIHandler{
		userRepo:      ur,
		oauthProvider: provider,
		cfg:           cfg,
		jwtMiddleware: jwtMiddleware,
	}, nil
}

// getOAuthConfig returns OAuth configuration for the frontend
func (h *OAuthAPIHandler) getOAuthConfig(c *gin.Context) {
	if !h.cfg.OAuth.Enabled {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "OAuth not configured",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":      h.cfg.OAuth.Enabled,
		"client_id":    h.cfg.OAuth.ClientID,
		"authorize_url": h.cfg.OAuth.AuthorizeURL,
		"scope":        h.cfg.OAuth.Scope,
		"redirect_url": h.cfg.OAuth.RedirectURL,
	})
}

// initiateOAuth starts the OAuth flow by generating a state and returning the authorization URL
func (h *OAuthAPIHandler) initiateOAuth(c *gin.Context) {
	if h.oauthProvider == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "OAuth not configured",
		})
		return
	}

	state, err := oauth.GenerateStateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate state token",
		})
		return
	}

	// Store state in session or cookie for validation
	// For now, we'll return it and expect the client to send it back
	authURL := h.oauthProvider.GetAuthorizationURL(state)

	c.JSON(http.StatusOK, gin.H{
		"authorization_url": authURL,
		"state":             state,
	})
}

// callbackOAuth handles the OAuth callback
func (h *OAuthAPIHandler) callbackOAuth(c *gin.Context) {
	if h.oauthProvider == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "OAuth not configured",
		})
		return
	}

	log := logging.FromContext(c)

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization code not provided",
		})
		return
	}

	// Exchange code for token
	token, err := h.oauthProvider.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		log.Errorf("Failed to exchange code: %s", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed to exchange authorization code",
		})
		return
	}

	// Extract ID token if available
	idToken, ok := token.Extra("id_token").(string)
	if !ok || idToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "ID token not found in response",
		})
		return
	}

	// Validate the ID token
	claims, err := h.oauthProvider.ValidateToken(c.Request.Context(), idToken)
	if err != nil {
		log.Errorf("Failed to validate token: %s", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed to validate token",
		})
		return
	}

	// Check if user exists, if not create one
	email := claims.Email
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email not found in token claims",
		})
		return
	}

	user, err := h.userRepo.FindByEmail(c.Request.Context(), email)
	if err != nil {
		// User doesn't exist, create one
		displayName := claims.Name
		if displayName == "" {
			displayName = claims.PreferredUsername
		}
		if displayName == "" {
			displayName = email
		}

		user = &models.User{
			Email:       email,
			DisplayName: displayName,
			Password:    "", // No password for OAuth users
			Disabled:    false,
		}

		if err := h.userRepo.CreateUser(c.Request.Context(), user); err != nil {
			log.Errorf("Failed to create user: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create user",
			})
			return
		}

		// Refetch user to get ID
		user, err = h.userRepo.FindByEmail(c.Request.Context(), email)
		if err != nil {
			log.Errorf("Failed to fetch newly created user: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user",
			})
			return
		}
	}

	// Check if user is disabled
	if user.Disabled {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User account is disabled",
		})
		return
	}

	// Generate JWT token for the user
	tokenString, expire, err := h.jwtMiddleware.TokenGenerator(map[string]interface{}{
		auth.IdentityKey: fmt.Sprintf("%d", user.ID),
		"type":           "user",
		"scopes":         []string{"task:read", "task:write", "label:read", "label:write", "user:read", "user:write", "token:write"},
	})
	if err != nil {
		log.Errorf("Failed to generate JWT: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate session token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"expiration": expire.Format(time.RFC3339),
	})
}

func OAuthRoutes(router *gin.Engine, h *OAuthAPIHandler, limiter *limiter.Limiter) {
	if h == nil || h.oauthProvider == nil {
		// OAuth not configured, skip routes
		return
	}

	oauthRoutes := router.Group("api/v1/oauth")
	oauthRoutes.Use(middleware.RateLimitMiddleware(limiter))
	{
		oauthRoutes.GET("/config", h.getOAuthConfig)
		oauthRoutes.GET("/authorize", h.initiateOAuth)
		oauthRoutes.GET("/callback", h.callbackOAuth)
	}
}
