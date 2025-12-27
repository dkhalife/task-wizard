package apis

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	models "dkhalife.com/tasks/core/internal/models"
	cRepo "dkhalife.com/tasks/core/internal/repos/caldav"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/utils/auth"
	"dkhalife.com/tasks/core/internal/utils/caldav"
	middleware "dkhalife.com/tasks/core/internal/utils/middleware"
	"dkhalife.com/tasks/core/internal/ws"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CalDAVAPIHandler struct {
	tRepo *tRepo.TaskRepository
	cRepo *cRepo.CalDavRepository
	nRepo *nRepo.NotificationRepository
	ws    *ws.WSServer
}

func CalDAVAPI(tRepo *tRepo.TaskRepository, cRepo *cRepo.CalDavRepository, nRepo *nRepo.NotificationRepository, wsServer *ws.WSServer) *CalDAVAPIHandler {
	return &CalDAVAPIHandler{
		tRepo: tRepo,
		cRepo: cRepo,
		nRepo: nRepo,
		ws:    wsServer,
	}
}

func (h *CalDAVAPIHandler) handleHead(c *gin.Context) {
	log := logging.FromContext(c)
	log.Debugf("CalDAV HEAD request for %s", c.Request.URL.Path)

	urlPath := c.Request.URL.Path
	if strings.HasSuffix(urlPath, ".ics") {
		filename := path.Base(urlPath)
		taskID, err := strconv.Atoi(strings.TrimSuffix(filename, ".ics"))

		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		task, err := h.tRepo.GetTask(c, taskID)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if task.CreatedBy != auth.CurrentIdentity(c).UserID {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}

	c.Status(http.StatusOK)
}

func (h *CalDAVAPIHandler) handlePropfindTask(c *gin.Context, taskID int) {
	log := logging.FromContext(c)

	response, taskOwner, err := h.cRepo.PropfindTask(c, taskID)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	currentIdentity := auth.CurrentIdentity(c)
	if currentIdentity == nil || taskOwner != currentIdentity.UserID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	response.Responses[0].Href = c.Request.URL.Path

	data, err := caldav.BuildXmlResponse(response)
	if err != nil {
		log.Errorf("Error encoding XML: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, "application/xml; charset=utf-8", data)
}

func (h *CalDAVAPIHandler) handlePropfind(c *gin.Context) {
	log := logging.FromContext(c)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf("Error reading request body: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Debugf("PROPFIND request body: %s", string(body))

	if strings.HasSuffix(c.Request.URL.Path, ".ics") {
		filename := path.Base(c.Request.URL.Path)
		taskID, err := strconv.Atoi(strings.TrimSuffix(filename, ".ics"))
		if err != nil {
			log.Infof("Invalid task ID: %s", err.Error())
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		h.handlePropfindTask(c, taskID)
		return
	}

	currentIdentity := auth.CurrentIdentity(c)
	if currentIdentity == nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	response, err := h.cRepo.PropfindUserTasks(c, currentIdentity.UserID)
	if err != nil {
		log.Errorf("Error getting tasks: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	data, err := caldav.BuildXmlResponse(response)
	if err != nil {
		log.Errorf("Error encoding XML: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusMultiStatus, "application/xml; charset=utf-8", data)
}

func (h *CalDAVAPIHandler) handleGet(c *gin.Context) {
	log := logging.FromContext(c)

	urlPath := c.Request.URL.Path
	filename := path.Base(urlPath)

	if !strings.HasSuffix(c.Request.URL.Path, ".ics") {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	log.Debugf("CalDAV GET request for %s", filename)
	taskID, err := strconv.Atoi(strings.TrimSuffix(filename, ".ics"))

	if err != nil {
		log.Errorf("Invalid task ID: %s", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vtodoContent, taskOwner, err := h.cRepo.GetTask(c, taskID)
	if err != nil {
		log.Errorf("Error getting task: %s", err.Error())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	currentIdentity := auth.CurrentIdentity(c)
	if currentIdentity == nil || taskOwner != currentIdentity.UserID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Data(http.StatusOK, "text/calendar; charset=utf-8", []byte(vtodoContent))
}

func (h *CalDAVAPIHandler) handleReport(c *gin.Context) {
	log := logging.FromContext(c)
	log.Debugf("CalDAV REPORT request for %s", c.Request.URL.Path)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf("Error reading request body: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Debugf("REPORT request body: %s", string(body))
	var report models.CalendarMultiget
	err = xml.Unmarshal(body, &report)
	if err != nil {
		log.Errorf("Error parsing REPORT request: %s", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	response, err := h.cRepo.MultiGet(c, report)
	if err != nil {
		log.Errorf("Error getting tasks: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	data, err := caldav.BuildXmlResponse(response)
	if err != nil {
		log.Errorf("Error encoding XML: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusMultiStatus, "application/xml; charset=utf-8", data)
}

func (h *CalDAVAPIHandler) handlePut(c *gin.Context) {
	log := logging.FromContext(c)

	if !strings.HasSuffix(c.Request.URL.Path, ".ics") {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf("Error reading request body: %s", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	filename := path.Base(c.Request.URL.Path)
	taskID, err := strconv.Atoi(strings.TrimSuffix(filename, ".ics"))
	if err != nil {
		log.Infof("Invalid task ID: %s", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	title, due, err := caldav.ParseVTODO(string(body))
	if err != nil {
		log.Errorf("Error parsing VTODO: %s", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	currentIdentity := auth.CurrentIdentity(c)
	if currentIdentity == nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	task, err := h.tRepo.GetTask(c, taskID)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if task.CreatedBy != currentIdentity.UserID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if err := h.cRepo.UpdateTask(c, taskID, title, due); err != nil {
		log.Errorf("Error updating task: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Get the updated task to broadcast via WebSocket
	updatedTask, err := h.tRepo.GetTask(c, taskID)
	if err != nil {
		log.Errorf("Error getting updated task: %s", err.Error())
		// Don't fail the request if we can't broadcast
	} else {
		// Regenerate notifications for the updated task
		go func(task *models.Task, logger *zap.SugaredLogger) {
			ctx := logging.ContextWithLogger(context.Background(), logger)
			h.nRepo.GenerateNotifications(ctx, task)
		}(updatedTask, log)

		// Broadcast the update to all connected clients for this user
		h.ws.BroadcastToUser(currentIdentity.UserID, ws.WSResponse{
			Action: "task_updated",
			Data:   updatedTask,
		})
	}

	c.Status(http.StatusNoContent)
}

func (h *CalDAVAPIHandler) handleRootRedirect(c *gin.Context) {
	log := logging.FromContext(c)
	log.Infof("Redirecting CalDAV %s request from / to /dav/tasks/", c.Request.Method)

	// Save the request body for forwarding
	var bodyBytes []byte
	if c.Request.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(c.Request.Body)
		if err != nil {
			log.Errorf("Error reading request body: %s", err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	c.Request.URL.Path = "/dav/tasks/"

	switch c.Request.Method {
	case "PROPFIND":
		h.handlePropfind(c)
	case "REPORT":
		h.handleReport(c)
	default:
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	}
}

func CalDAVRoutes(router *gin.Engine, h *CalDAVAPIHandler, auth *jwt.GinJWTMiddleware) {
	davRoutes := router.Group("dav")
	davRoutes.Use(middleware.BasicAuthToJWTAdapter())
	davRoutes.Use(auth.MiddlewareFunc())
	{
		davRoutes.HEAD("/tasks/*path", authMW.ScopeMiddleware(models.ApiTokenScopeDavRead), h.handleHead)
		davRoutes.Handle("PROPFIND", "/tasks/*path", authMW.ScopeMiddleware(models.ApiTokenScopeDavRead), h.handlePropfind)
		davRoutes.Handle("REPORT", "/tasks/*path", authMW.ScopeMiddleware(models.ApiTokenScopeDavRead), h.handleReport)
		davRoutes.GET("/tasks/*path", authMW.ScopeMiddleware(models.ApiTokenScopeDavRead), h.handleGet)
		davRoutes.PUT("/tasks/*path", authMW.ScopeMiddleware(models.ApiTokenScopeDavWrite), h.handlePut)
	}

	router.Handle("PROPFIND", "/", h.handleRootRedirect)
	router.Handle("REPORT", "/", h.handleRootRedirect)
}
