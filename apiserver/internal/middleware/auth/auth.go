package auth

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	repos "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	TenantID     string
	Audience     string
	Verifier     *oidc.IDTokenVerifier
	Registration bool
	uRepo        repos.IUserRepo
}

func NewAuthMiddleware(cfg *config.Config, uRepo repos.IUserRepo) *AuthMiddleware {
	issuer := "https://login.microsoftonline.com/" + cfg.Entra.TenantID + "/v2.0"

	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		panic("Failed to create OIDC provider: " + err.Error())
	}

	return &AuthMiddleware{
		TenantID: cfg.Entra.TenantID,
		Audience: cfg.Entra.Audience,
		Verifier: provider.Verifier(&oidc.Config{
			ClientID: cfg.Entra.Audience, // Must match "aud" in token
		}),
		Registration: cfg.Server.Registration,
		uRepo:        uRepo,
	}
}

func (m *AuthMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logging.FromContext(c.Request.Context())

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		rawToken := parts[1]

		// Verify token signature + issuer + audience
		idToken, err := m.Verifier.Verify(c, rawToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		// Extract claims
		var claims map[string]interface{}
		if err := idToken.Claims(&claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to parse claims"})
			return
		}

		user, err := m.uRepo.FindByEmail(c, claims["preferred_username"].(string))
		if err != nil {
			if !m.Registration {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				return
			}

			err = m.uRepo.CreateUser(c, &models.User{
				Email:       claims["preferred_username"].(string),
				DisplayName: claims["name"].(string),
			})

			if err != nil {
				logger.Error("Failed to register user: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}

			user, err = m.uRepo.FindByEmail(c, claims["preferred_username"].(string))
			if err != nil {
				logger.Error("Failed to retrieve newly created user: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
				return
			}
		}

		claimsStr := claims["scp"].(string)
		scopes := auth.ConvertStringArrayToScopes(strings.Split(claimsStr, " "))

		c.Set(auth.IdentityKey, &models.SignedInIdentity{
			UserID: user.ID,
			Type:   "user",
			Scopes: scopes,
		})

		c.Next()
	}
}

func ScopeMiddleware(requiredScope models.ApiTokenScope) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentIdentity := auth.CurrentIdentity(c)

		if slices.Contains(currentIdentity.Scopes, requiredScope) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Missing required scope: " + requiredScope,
		})
	}
}
