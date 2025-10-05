package middleware

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/logging"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func NewRateLimiter(cfg *config.Config) *limiter.Limiter {
	store := memory.NewStore()

	rate := limiter.Rate{
		Period: cfg.Server.RatePeriod,
		Limit:  int64(cfg.Server.RateLimit),
	}

	return limiter.New(store, rate)
}

func RateLimitMiddleware(limiter *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use the IP as the key, which is the client IP.
		context, err := limiter.Get(c.Request.Context(), c.ClientIP())
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if context.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "Too many requests"})
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		log := logging.FromContext(c)
		log.Infof("IP:%s Method:%s UA:%q Route:%s Status:%d",
			c.ClientIP(), c.Request.Method, c.Request.UserAgent(), c.Request.URL.Path, c.Writer.Status())
	}
}

func BasicAuthToJWTAdapter() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if strings.HasPrefix(authHeader, "Basic ") {
			b64 := strings.TrimPrefix(authHeader, "Basic ")
			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err == nil {
				parts := strings.SplitN(string(decoded), ":", 2)
				token := parts[0]

				if len(parts) == 2 {
					token = parts[1]
				}

				c.Request.Header.Set("Authorization", "Bearer "+token)
			}
		}

		c.Next()
	}
}
