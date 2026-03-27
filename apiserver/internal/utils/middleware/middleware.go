package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/telemetry"
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
			telemetry.TrackError(c, "rate_limit_store_error", "rate-limiter", err, nil)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if context.Reached {
			telemetry.TrackWarning(c, "rate_limited", "rate-limiter", "Too many requests", nil)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "Too many requests"})
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))
		c.Next()
	}
}

func effectiveScheme(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		scheme := forwarded
		if i := strings.IndexByte(scheme, ','); i >= 0 {
			scheme = scheme[:i]
		}
		return strings.ToLower(strings.TrimSpace(scheme))
	}

	if c.Request.TLS != nil {
		return "https"
	}

	return "http"
}

func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	hostName := cfg.Server.HostName
	port := cfg.Server.Port

	return func(c *gin.Context) {
		scheme := effectiveScheme(c)
		if scheme == "http" {
			target := fmt.Sprintf("https://%s", hostName)
			if port != 443 {
				target = fmt.Sprintf("%s:%d", target, port)
			}
			target = fmt.Sprintf("%s%s", target, c.Request.URL.RequestURI())
			c.Redirect(http.StatusMovedPermanently, target)
			c.Abort()
			return
		}

		if scheme == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

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

func TelemetryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("DNT") == "1" {
			telemetry.SetDNT(c)
		}

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		telemetry.TrackEvent(c, "http_request", "http-middleware", map[string]string{
			"method":    c.Request.Method,
			"route":     c.Request.URL.Path,
			"status":    strconv.Itoa(c.Writer.Status()),
			"duration":  duration.String(),
			"client_ip": c.ClientIP(),
		})
	}
}
