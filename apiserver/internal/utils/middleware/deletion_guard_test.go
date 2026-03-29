package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"dkhalife.com/tasks/core/internal/models"
	authUtils "dkhalife.com/tasks/core/internal/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type DeletionGuardTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func TestDeletionGuardTestSuite(t *testing.T) {
	suite.Run(t, new(DeletionGuardTestSuite))
}

func (s *DeletionGuardTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *DeletionGuardTestSuite) injectIdentity(pendingDeletion bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(authUtils.IdentityKey, &models.SignedInIdentity{
			UserID:          1,
			Type:            models.IdentityTypeUser,
			Scopes:          models.AllUserScopes(),
			PendingDeletion: pendingDeletion,
		})
		c.Next()
	}
}

func (s *DeletionGuardTestSuite) TestWriteBlockedWhenPendingDeletion() {
	s.router.Use(s.injectIdentity(true), DeletionGuardMiddleware())
	s.router.POST("/api/v1/tasks/", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(method, "/api/v1/tasks/", nil)
		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusForbidden, w.Code, "expected 403 for %s when pending deletion", method)
	}
}

func (s *DeletionGuardTestSuite) TestReadAllowedWhenPendingDeletion() {
	s.router.Use(s.injectIdentity(true), DeletionGuardMiddleware())
	s.router.GET("/api/v1/tasks/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tasks/", nil)
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
}

func (s *DeletionGuardTestSuite) TestWriteAllowedWhenNotPendingDeletion() {
	s.router.Use(s.injectIdentity(false), DeletionGuardMiddleware())
	s.router.POST("/api/v1/tasks/", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/tasks/", nil)
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusCreated, w.Code)
}

func (s *DeletionGuardTestSuite) TestDeletionEndpointExemptedFromBlock() {
	s.router.Use(s.injectIdentity(true), DeletionGuardMiddleware())
	s.router.POST("/api/v1/users/deletion", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	s.router.DELETE("/api/v1/users/deletion", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, method := range []string{http.MethodPost, http.MethodDelete} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(method, "/api/v1/users/deletion", nil)
		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusNoContent, w.Code, "deletion endpoint should be exempt for method %s", method)
	}
}

func (s *DeletionGuardTestSuite) TestNoIdentityAllowsPassThrough() {
	s.router.Use(DeletionGuardMiddleware())
	s.router.POST("/api/v1/tasks/", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/tasks/", nil)
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusCreated, w.Code)
}
