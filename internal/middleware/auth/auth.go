package auth

import (
	"net/http"
	"time"

	"donetick.com/core/config"
	uModel "donetick.com/core/internal/models/user"
	uRepo "donetick.com/core/internal/repos/user"
	"donetick.com/core/internal/services/logging"
	auth "donetick.com/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type signIn struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func NewAuthMiddleware(cfg *config.Config, userRepo *uRepo.UserRepository) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte(cfg.Jwt.Secret),
		Timeout:     cfg.Jwt.SessionTime,
		MaxRefresh:  cfg.Jwt.MaxRefresh, // 7 days as long as their token is valid they can refresh it
		IdentityKey: auth.IdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if u, ok := data.(*uModel.User); ok {
				return jwt.MapClaims{
					auth.IdentityKey: u.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			username, ok := claims[auth.IdentityKey].(string)
			if !ok {
				return nil
			}
			user, err := userRepo.GetUserByUsername(c.Request.Context(), username)
			if err != nil {
				return nil
			}
			return user
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			provider := c.Value("auth_provider")
			switch provider {
			case nil:
				var req signIn
				if err := c.ShouldBindJSON(&req); err != nil {
					return "", jwt.ErrMissingLoginValues
				}

				user, err := userRepo.GetUserByUsername(c.Request.Context(), req.Username)
				if err != nil || user.Disabled {
					return nil, jwt.ErrFailedAuthentication
				}
				err = auth.Matches(user.Password, req.Password)
				if err != nil {
					if err != bcrypt.ErrMismatchedHashAndPassword {
						logging.FromContext(c).Warnw("middleware.jwt.Authenticator found unknown error when matches password", "err", err)
					}
					return nil, jwt.ErrFailedAuthentication
				}
				return &uModel.User{
					ID:        user.ID,
					Username:  user.Username,
					Password:  "",
					CreatedAt: user.CreatedAt,
					UpdatedAt: user.UpdatedAt,
					Disabled:  user.Disabled,
				}, nil

			default:
				return nil, jwt.ErrFailedAuthentication
			}
		},

		Authorizator: func(data interface{}, c *gin.Context) bool {
			if _, ok := data.(*uModel.User); ok {
				return true
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			logging.FromContext(c).Info("middleware.jwt.Unauthorized", "code", code, "message", message)
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"code":   code,
				"token":  token,
				"expire": expire,
			})
		},
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
}
