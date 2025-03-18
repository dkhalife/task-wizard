package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/logging"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

const (
	XRequestIdKey = "X-Request-ID"
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
			panic(err)
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

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)

		defer func() {
			if ctx.Err() == context.DeadlineExceeded {
				c.AbortWithStatus(http.StatusGatewayTimeout)
			}
			cancel()
		}()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		log := logging.FromContext(c)
		log.Infof("IP:%s UA:%q Route:%s Status:%d\n",
			c.ClientIP(), c.Request.UserAgent(), c.Request.URL.Path, c.Writer.Status())
	}
}
