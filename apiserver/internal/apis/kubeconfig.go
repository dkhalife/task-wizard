package apis

import (
	"net/http"
	"strconv"

	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/models"
	kubeconfigService "dkhalife.com/tasks/core/internal/services/kubeconfig"
	"dkhalife.com/tasks/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type KubeconfigAPIHandler struct {
	service *kubeconfigService.KubeconfigService
}

func KubeconfigAPI(service *kubeconfigService.KubeconfigService) *KubeconfigAPIHandler {
	return &KubeconfigAPIHandler{
		service: service,
	}
}

// importKubeconfig handles importing a kubeconfig YAML
func (h *KubeconfigAPIHandler) importKubeconfig(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var req models.ImportKubeconfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	status, response := h.service.ImportKubeconfig(c, currentIdentity.UserID, req.KubeconfigYAML)
	c.JSON(status, response)
}

// listContexts handles listing all contexts for the current user
func (h *KubeconfigAPIHandler) listContexts(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	status, response := h.service.ListContexts(c, currentIdentity.UserID)
	c.JSON(status, response)
}

// getContext handles retrieving a specific context
func (h *KubeconfigAPIHandler) getContext(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	contextID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid context ID",
		})
		return
	}

	status, response := h.service.GetContext(c, currentIdentity.UserID, contextID)
	c.JSON(status, response)
}

// setActiveContext handles setting a context as active
func (h *KubeconfigAPIHandler) setActiveContext(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	var req models.SetActiveContextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	status, response := h.service.SetActiveContext(c, currentIdentity.UserID, req.ContextID)
	c.JSON(status, response)
}

// deleteContext handles deleting a context
func (h *KubeconfigAPIHandler) deleteContext(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	contextID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid context ID",
		})
		return
	}

	status, response := h.service.DeleteContext(c, currentIdentity.UserID, contextID)
	c.JSON(status, response)
}

// getActiveContext handles retrieving the active context
func (h *KubeconfigAPIHandler) getActiveContext(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	status, response := h.service.GetActiveContext(c, currentIdentity.UserID)
	c.JSON(status, response)
}

// KubeconfigRoutes registers the kubeconfig routes
func KubeconfigRoutes(r *gin.Engine, h *KubeconfigAPIHandler, authGate *jwt.GinJWTMiddleware) {
	kubeconfigRoutes := r.Group("api/v1/kubeconfig")
	kubeconfigRoutes.Use(authGate.MiddlewareFunc())
	{
		kubeconfigRoutes.POST("/import", authMW.ScopeMiddleware(models.ApiTokenScopeUserWrite), h.importKubeconfig)
		kubeconfigRoutes.GET("", authMW.ScopeMiddleware(models.ApiTokenScopeUserRead), h.listContexts)
		kubeconfigRoutes.GET("/active", authMW.ScopeMiddleware(models.ApiTokenScopeUserRead), h.getActiveContext)
		kubeconfigRoutes.GET("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeUserRead), h.getContext)
		kubeconfigRoutes.PUT("/active", authMW.ScopeMiddleware(models.ApiTokenScopeUserWrite), h.setActiveContext)
		kubeconfigRoutes.DELETE("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeUserWrite), h.deleteContext)
	}
}
