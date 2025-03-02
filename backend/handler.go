package backend

import (
	"net/http"

	"dkhalife.com/tasks/core/config"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/utils/auth"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uRepo *uRepo.UserRepository
}

func NewHandler(config *config.Config, uRepo *uRepo.UserRepository) *Handler {
	return &Handler{
		uRepo,
	}
}

func (h *Handler) activateUser(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.Data(http.StatusBadRequest, "text/html", []byte("<h1>Bad Request</h1><p>Missing activation code</p>"))
		return
	}

	email, code, err := auth.DecodeEmailAndCode(code)

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

func Routes(router *gin.Engine, h *Handler) {
	router.GET("/activate", h.activateUser)
}
