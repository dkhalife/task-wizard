package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	authUtils "dkhalife.com/tasks/core/internal/utils/auth"
	oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type AuthMiddleware struct {
	enabled  bool
	keySet   oidc.KeySet
	issuer   string
	audience string
	userRepo uRepo.IUserRepo
	secret   string
}

type accessTokenClaims struct {
	Issuer            string `json:"iss"`
	Audience          string `json:"aud"`
	ExpiresAt         int64  `json:"exp"`
	TenantID          string `json:"tid"`
	ObjectID          string `json:"oid"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
}

func NewAuthMiddleware(cfg *config.Config, userRepo uRepo.IUserRepo) (*AuthMiddleware, error) {
	m := &AuthMiddleware{
		enabled:  cfg.Entra.Enabled,
		userRepo: userRepo,
		secret:   cfg.Jwt.Secret,
	}

	if !cfg.Entra.Enabled {
		return m, nil
	}

	issuer := cfg.Entra.Issuer
	if issuer == "" {
		issuer = "https://login.microsoftonline.com/" + cfg.Entra.TenantID + "/v2.0"
	}
	m.issuer = issuer
	m.audience = cfg.Entra.Audience

	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %s", err.Error())
	}

	var providerClaims struct {
		JWKSURL string `json:"jwks_uri"`
	}
	if err := provider.Claims(&providerClaims); err != nil {
		return nil, fmt.Errorf("failed to extract JWKS URI: %s", err.Error())
	}

	m.keySet = oidc.NewRemoteKeySet(context.Background(), providerClaims.JWKSURL)

	return m, nil
}

func (m *AuthMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := m.authenticate(c)
		if err != nil {
			if strings.HasPrefix(c.Request.URL.Path, "/dav") {
				c.Header("WWW-Authenticate", `Basic realm="Task Wizard"`)
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Set(authUtils.IdentityKey, identity)
		c.Next()
	}
}

func (m *AuthMiddleware) authenticate(c *gin.Context) (*models.SignedInIdentity, error) {
	if !m.enabled {
		return m.bypassAuth(c.Request.Context())
	}

	token := extractBearerToken(c)
	if token == "" {
		return nil, fmt.Errorf("missing authorization token")
	}

	if identity, err := m.verifyAppToken(c, token); err == nil {
		return identity, nil
	}

	return m.verifyAccessToken(c.Request.Context(), token)
}

func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}

func (m *AuthMiddleware) verifyAppToken(c *gin.Context, rawToken string) (*models.SignedInIdentity, error) {
	log := logging.FromContext(c)

	token, err := jwt.Parse(rawToken, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(m.secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid app token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != string(models.IdentityTypeApp) {
		return nil, fmt.Errorf("not an app token")
	}

	userIDRaw, ok := claims[authUtils.IdentityKey].(string)
	if !ok {
		return nil, fmt.Errorf("missing user ID")
	}

	userID, err := strconv.Atoi(userIDRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	var tokenID int
	if raw, ok := claims[authUtils.AppTokenKey].(string); ok {
		tokenID, _ = strconv.Atoi(raw)
	}

	if tokenID == 0 {
		return nil, fmt.Errorf("missing token ID")
	}

	if _, err := m.userRepo.GetAppTokenByID(c.Request.Context(), tokenID); err != nil {
		log.Errorw("failed to find app token", "err", err)
		return nil, fmt.Errorf("app token not found")
	}

	scopesRaw, ok := claims["scopes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing scopes")
	}

	var scopes []models.ApiTokenScope
	for _, scope := range scopesRaw {
		if s, ok := scope.(string); ok {
			scopes = append(scopes, models.ApiTokenScope(s))
		}
	}

	return &models.SignedInIdentity{
		UserID:  userID,
		TokenID: tokenID,
		Type:    models.IdentityTypeApp,
		Scopes:  scopes,
	}, nil
}

func (m *AuthMiddleware) verifyAccessToken(ctx context.Context, rawToken string) (*models.SignedInIdentity, error) {
	payload, err := m.keySet.VerifySignature(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("token signature verification failed: %s", err.Error())
	}

	var claims accessTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse token claims: %s", err.Error())
	}

	if claims.Issuer != m.issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	if claims.Audience != m.audience {
		return nil, fmt.Errorf("invalid token audience")
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token has expired")
	}

	if claims.TenantID == "" || claims.ObjectID == "" {
		return nil, fmt.Errorf("missing tid or oid in token claims")
	}

	user, err := m.userRepo.EnsureUser(ctx, claims.TenantID, claims.ObjectID, claims.Name, claims.PreferredUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user identity: %s", err.Error())
	}

	return &models.SignedInIdentity{
		UserID:  user.ID,
		TokenID: 0,
		Type:    models.IdentityTypeUser,
		Scopes:  models.AllUserScopes(),
	}, nil
}

func (m *AuthMiddleware) bypassAuth(ctx context.Context) (*models.SignedInIdentity, error) {
	user, err := m.userRepo.EnsureUser(ctx, "dev-directory", "dev-object", "Dev User", "dev@localhost")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure dev user: %s", err.Error())
	}

	return &models.SignedInIdentity{
		UserID:  user.ID,
		TokenID: 0,
		Type:    models.IdentityTypeUser,
		Scopes:  models.AllUserScopes(),
	}, nil
}

// VerifyWSToken validates a token for WebSocket connections.
func (m *AuthMiddleware) VerifyWSToken(ctx context.Context, rawToken string) (*models.SignedInIdentity, error) {
	if !m.enabled {
		return m.bypassAuth(ctx)
	}

	return m.verifyAccessToken(ctx, rawToken)
}

func ScopeMiddleware(requiredScope models.ApiTokenScope) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentIdentity := authUtils.CurrentIdentity(c)

		if slices.Contains(currentIdentity.Scopes, requiredScope) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Missing required scope: " + requiredScope,
		})
	}
}
