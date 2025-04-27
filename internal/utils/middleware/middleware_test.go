package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
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
