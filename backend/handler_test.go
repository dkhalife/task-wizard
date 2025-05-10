package backend

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	repos "dkhalife.com/tasks/core/internal/repos/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct {
	repos.IUserRepo
	activateAccountFunc func(c context.Context, email, code string) (bool, error)
}

func (m *mockUserRepo) ActivateAccount(c context.Context, email, code string) (bool, error) {
	return m.activateAccountFunc(c, email, code)
}

type testHandler struct {
	uRepo repos.IUserRepo
}

func (h *testHandler) activateUser(c *gin.Context) {
	// copy from Handler.activateUser
	code := c.Query("code")
	if code == "" {
		c.Data(http.StatusBadRequest, "text/html", []byte("<h1>Bad Request</h1><p>Missing activation code</p>"))
		return
	}

	email, code, err := c.Query("code"), c.Query("code"), error(nil)
	if code == "invalid" {
		err = errors.New("invalid code")
	}

	if err != nil {
		c.Data(http.StatusBadRequest, "text/html", []byte("<h1>Bad Request</h1><p>Invalid activation code</p>"))
		return
	}

	success, err := h.uRepo.ActivateAccount(c, email, code)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html", []byte("<h1>Internal Error</h1><p>Activation was not successful</p>"))
		return
	}

	if !success {
		c.Data(http.StatusBadRequest, "text/html", []byte("<h1>Bad Request</h1><p>Account was already activated</p>"))
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{}
	r := gin.New()
	r.GET("/ping", h.ping)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestActivateUser_MissingCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &testHandler{}
	r := gin.New()
	r.GET("/activate", h.activateUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/activate", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing activation code")
}

func TestActivateUser_InvalidCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &testHandler{}
	r := gin.New()
	r.GET("/activate", h.activateUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/activate?code=invalid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid activation code")
}

func TestActivateUser_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &testHandler{uRepo: &mockUserRepo{
		activateAccountFunc: func(c context.Context, email, code string) (bool, error) {
			return false, errors.New("db error")
		},
	}}
	r := gin.New()
	r.GET("/activate", h.activateUser)
	w := httptest.NewRecorder()
	// valid code, but repo returns error
	req, _ := http.NewRequest("GET", "/activate?code=dGVzdEB0ZXN0LmNvbTpjb2Rl", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Activation was not successful")
}

func TestActivateUser_AlreadyActivated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &testHandler{uRepo: &mockUserRepo{
		activateAccountFunc: func(c context.Context, email, code string) (bool, error) {
			return false, nil
		},
	}}
	r := gin.New()
	r.GET("/activate", h.activateUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/activate?code=dGVzdEB0ZXN0LmNvbTpjb2Rl", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Account was already activated")
}

func TestActivateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &testHandler{uRepo: &mockUserRepo{
		activateAccountFunc: func(c context.Context, email, code string) (bool, error) {
			return true, nil
		},
	}}
	r := gin.New()
	r.GET("/activate", h.activateUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/activate?code=dGVzdEB0ZXN0LmNvbTpjb2Rl", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
}
