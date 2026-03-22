package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/ulule/limiter/v3"
)

type MiddlewareTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func (s *MiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *MiddlewareTestSuite) TestRateLimitMiddleware() {
	cfg := &config.Config{
		Server: config.ServerConfig{
			RateLimit:  1,
			RatePeriod: time.Hour,
		},
	}
	limiter := NewRateLimiter(cfg)

	s.router.Use(RateLimitMiddleware(limiter))
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusTooManyRequests, w.Code)
}

// failingStore is a limiter store that always returns an error.
type failingStore struct{}

func (f *failingStore) Get(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store failure")
}

func (f *failingStore) Peek(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store failure")
}

func (f *failingStore) Reset(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store failure")
}

func (f *failingStore) Increment(ctx context.Context, key string, count int64, rate limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store failure")
}

func (s *MiddlewareTestSuite) TestRateLimitMiddlewareStoreFailure() {
	cfg := &config.Config{
		Server: config.ServerConfig{
			RateLimit:  1,
			RatePeriod: time.Hour,
		},
	}
	store := &failingStore{}
	limit := limiter.New(store, limiter.Rate{Period: cfg.Server.RatePeriod, Limit: int64(cfg.Server.RateLimit)})

	s.router.Use(RateLimitMiddleware(limit))
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *MiddlewareTestSuite) TestSecurityHeadersAddsHSTS() {
	cfg := &config.Config{
		Server: config.ServerConfig{
			HostName: "example.com",
			Port:     443,
		},
	}

	s.router.Use(SecurityHeaders(cfg))
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("max-age=31536000; includeSubDomains; preload", w.Header().Get("Strict-Transport-Security"))
}

func (s *MiddlewareTestSuite) TestSecurityHeadersRedirectsHTTP() {
	cfg := &config.Config{
		Server: config.ServerConfig{
			HostName: "example.com",
			Port:     443,
		},
	}

	s.router.Use(SecurityHeaders(cfg))
	s.router.GET("/path", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/path?q=1", nil)
	req.Header.Set("X-Forwarded-Proto", "http")
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusMovedPermanently, w.Code)
	s.Equal("https://example.com/path?q=1", w.Header().Get("Location"))
}

func (s *MiddlewareTestSuite) TestSecurityHeadersRedirectsHTTPNonStandardPort() {
	cfg := &config.Config{
		Server: config.ServerConfig{
			HostName: "example.com",
			Port:     8443,
		},
	}

	s.router.Use(SecurityHeaders(cfg))
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-Proto", "http")
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusMovedPermanently, w.Code)
	s.Equal("https://example.com:8443/", w.Header().Get("Location"))
}

func (s *MiddlewareTestSuite) TestSecurityHeadersNoRedirectForHTTPS() {
	cfg := &config.Config{
		Server: config.ServerConfig{
			HostName: "example.com",
			Port:     443,
		},
	}

	s.router.Use(SecurityHeaders(cfg))
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("max-age=31536000; includeSubDomains; preload", w.Header().Get("Strict-Transport-Security"))
}
