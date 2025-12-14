package auth

import (
	"context"
	"net/http"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"honnef.co/go/tools/config"
)

type AuthMiddleware struct {
	TenantID       string
	Audience       string
	Verifier       *oidc.IDTokenVerifier
	Registration   bool
	uRepo          repos.IUserRepo
	AppTokenSecret string
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
		Registration:   cfg.Server.Registration,
		uRepo:          uRepo,
		AppTokenSecret: cfg.Jwt.Secret,
	}
}

func (m *AuthMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes := []models.ApiTokenScope{models.ReadScope, models.WriteScope}

		c.Set(auth.IdentityKey, &models.SignedInIdentity{
			UserID: 1,
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
