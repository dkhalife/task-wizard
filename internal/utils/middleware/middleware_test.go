package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			RateLimit:  1,
			RatePeriod: time.Second,
		},
	}
	limiter := NewRateLimiter(cfg)

	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRequestLogger(t *testing.T) {
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
