package auth

import (
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
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type signIn struct {
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func NewAuthMiddleware(cfg *config.Config, userRepo uRepo.IUserRepo) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "Task Wizard",
		Key:         []byte(cfg.Jwt.Secret),
		Timeout:     cfg.Jwt.SessionTime,
		MaxRefresh:  cfg.Jwt.MaxRefresh, // 7 days as long as their token is valid they can refresh it
		IdentityKey: auth.IdentityKey,
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var req signIn
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			user, err := userRepo.FindByEmail(c.Request.Context(), req.Email)
			if err != nil || user.Disabled {
				return nil, jwt.ErrFailedAuthentication
			}

			err = auth.Matches(user.Password, req.Password)
			if err != nil {
				if err != bcrypt.ErrMismatchedHashAndPassword {
					logging.FromContext(c).Errorf("found unknown error when matches password", "err", err)
				}
				return nil, jwt.ErrFailedAuthentication
			}

			return user, nil
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if u, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					auth.IdentityKey: fmt.Sprintf("%d", u.ID),
					"type":           "user",
					"scopes":         []string{"task:read", "task:write", "label:read", "label:write", "user:read", "user:write", "token:write"},
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)

			userIDRaw, ok := claims[auth.IdentityKey].(string)
			if !ok {
				logging.FromContext(c).Errorw("failed to extract ID from claims")
				return nil
			}
			userID, err := strconv.Atoi(userIDRaw)
			if err != nil {
				return nil
			}

			var tokenID = 0
			appTokenIDRaw, ok := claims[auth.AppTokenKey].(string)
			if ok {
				tokenID, err = strconv.Atoi(appTokenIDRaw)
				if err != nil {
					tokenID = 0
				}
			}

			scopesRaw, ok := claims["scopes"].([]interface{})
			if !ok {
				return nil
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
				Type:    models.IdentityType(claims["type"].(string)),
				Scopes:  scopes,
			}
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if identity, ok := data.(*models.SignedInIdentity); ok {
				if identity.Type == models.IdentityTypeUser {
					// Check that the user still exists
					_, err := userRepo.GetUser(c.Request.Context(), identity.UserID)
					if err != nil {
						logging.FromContext(c).Errorw("failed to find user", "err", err)
						return false
					}
				} else if identity.Type == models.IdentityTypeApp {
					// An app token id must be present
					if identity.TokenID == 0 {
						logging.FromContext(c).Errorw("app token ID is nil")
						return false
					}

					// Check that the app token still exists
					_, err := userRepo.GetAppTokenByID(c.Request.Context(), identity.TokenID)
					if err != nil {
						logging.FromContext(c).Errorw("failed to find app token", "err", err)
						return false
					}
				}

				return true
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			if strings.HasPrefix(c.Request.URL.Path, "/dav") {
				c.Header("WWW-Authenticate", `Basic realm="Task Wizard"`)
			}

			c.JSON(code, gin.H{
				"error": message,
			})
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"token":      token,
				"expiration": expire,
			})
		},
		RefreshResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"token":      token,
				"expiration": expire,
			})
		},
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
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
