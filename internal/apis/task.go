package apis

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/models"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	"dkhalife.com/tasks/core/internal/services/logging"
	notifications "dkhalife.com/tasks/core/internal/services/notifications"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type LabelReq struct {
	LabelID int `json:"id" binding:"required"`
}

type TaskReq struct {
	ID           string                            `json:"id"`
	Title        string                            `json:"title" binding:"required"`
	NextDueDate  string                            `json:"next_due_date"`
	EndDate      string                            `json:"end_date"`
	IsRolling    bool                              `json:"is_rolling"`
	Frequency    models.Frequency                  `json:"frequency"`
	Notification models.NotificationTriggerOptions `json:"notification"`
	Labels       []int                             `json:"labels"`
}

type TasksAPIHandler struct {
	tRepo    *tRepo.TaskRepository
	notifier *notifications.Notifier
	nRepo    *nRepo.NotificationRepository
	lRepo    *lRepo.LabelRepository
}

func TasksAPI(cr *tRepo.TaskRepository, nt *notifications.Notifier,
	nRepo *nRepo.NotificationRepository, lRepo *lRepo.LabelRepository) *TasksAPIHandler {
	return &TasksAPIHandler{
		tRepo:    cr,
		notifier: nt,
		nRepo:    nRepo,
		lRepo:    lRepo,
	}
}

func (h *TasksAPIHandler) getTasks(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)
	tasks, err := h.tRepo.GetTasks(c, currentIdentity.UserID)
	if err != nil {
		log.Errorf("error getting tasks: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting tasks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

func (h *TasksAPIHandler) getCompletedTasks(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)

	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	offset := (page - 1) * limit

	tasks, err := h.tRepo.GetCompletedTasks(c, currentIdentity.UserID, limit, offset)
	if err != nil {
		log.Errorf("error getting completed tasks: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

func (h *TasksAPIHandler) getTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)
	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	task, err := h.tRepo.GetTask(c, id)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	if currentIdentity.UserID != task.CreatedBy {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to view this task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": task,
	})
}

func (h *TasksAPIHandler) createTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)
	var TaskReq TaskReq
	if err := c.ShouldBindJSON(&TaskReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	var dueDate *time.Time
	if TaskReq.NextDueDate != "" {
		rawDueDate, err := time.Parse(time.RFC3339, TaskReq.NextDueDate)
		if err != nil {
			log.Errorf("error parsing due date: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Due date must be in UTC format",
			})
			return
		}

		rawDueDate = rawDueDate.UTC()
		dueDate = &rawDueDate
	}

	var endDate *time.Time
	if TaskReq.EndDate != "" {
		rawEndDate, err := time.Parse(time.RFC3339, TaskReq.EndDate)
		if err != nil {
			log.Errorf("error parsing end date: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "End date must be in UTC format",
			})
			return
		}

		rawEndDate = rawEndDate.UTC()
		endDate = &rawEndDate
	}

	createdTask := &models.Task{
		Title:        TaskReq.Title,
		Frequency:    TaskReq.Frequency,
		NextDueDate:  dueDate,
		EndDate:      endDate,
		CreatedBy:    currentIdentity.UserID,
		IsRolling:    TaskReq.IsRolling,
		IsActive:     true,
		Notification: TaskReq.Notification,
	}
	id, err := h.tRepo.CreateTask(c, createdTask)
	createdTask.ID = id

	if err != nil {
		log.Errorf("error creating task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error creating task",
		})
		return
	}

	if err := h.lRepo.AssignLabelsToTask(c, createdTask.ID, currentIdentity.UserID, TaskReq.Labels); err != nil {
		log.Errorf("error assigning labels to task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error adding labels",
		})
		return
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := context.WithValue(context.Background(), "logger", logger)
		h.nRepo.GenerateNotifications(ctx, task)
	}(createdTask, log)

	c.JSON(http.StatusCreated, gin.H{
		"task": id,
	})
}

func (h *TasksAPIHandler) editTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)

	var TaskReq TaskReq
	if err := c.ShouldBindJSON(&TaskReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	var dueDate *time.Time
	if TaskReq.NextDueDate != "" {
		rawDueDate, err := time.Parse(time.RFC3339, TaskReq.NextDueDate)
		if err != nil {
			log.Errorf("error parsing due date: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Due date must be in UTC format",
			})
			return
		}

		rawDueDate = rawDueDate.UTC()
		dueDate = &rawDueDate
	}

	var endDate *time.Time
	if TaskReq.EndDate != "" {
		rawEndDate, err := time.Parse(time.RFC3339, TaskReq.EndDate)
		if err != nil {
			log.Errorf("error parsing end date: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "End date must be in UTC format",
			})
			return
		}

		rawEndDate = rawEndDate.UTC()
		endDate = &rawEndDate
	}

	taskId, err := strconv.Atoi(TaskReq.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}
	oldTask, err := h.tRepo.GetTask(c, taskId)

	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	if currentIdentity.UserID != oldTask.CreatedBy {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to edit this task",
		})
		return
	}

	if err := h.lRepo.AssignLabelsToTask(c, taskId, currentIdentity.UserID, TaskReq.Labels); err != nil {
		log.Errorf("error assigning labels to task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error adding labels",
		})
		return
	}

	updatedTask := &models.Task{
		ID:           taskId,
		Title:        TaskReq.Title,
		Frequency:    TaskReq.Frequency,
		NextDueDate:  dueDate,
		EndDate:      endDate,
		CreatedBy:    currentIdentity.UserID,
		IsRolling:    TaskReq.IsRolling,
		Notification: TaskReq.Notification,
		IsActive:     oldTask.IsActive,
	}

	if err := h.tRepo.UpsertTask(c, updatedTask); err != nil {
		log.Errorf("error upserting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error upserting task",
		})
		return
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := context.WithValue(context.Background(), "logger", logger)
		h.nRepo.GenerateNotifications(ctx, task)
	}(updatedTask, log)

	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *TasksAPIHandler) deleteTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)
	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	if err := h.tRepo.IsTaskOwner(c, id, currentIdentity.UserID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to delete this task",
		})
		return
	}

	if err := h.tRepo.DeleteTask(c, id); err != nil {
		log.Errorf("error deleting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error deleting task",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *TasksAPIHandler) skipTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)

	log := logging.FromContext(c)
	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	task, err := h.tRepo.GetTask(c, id)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	nextDueDate, err := tRepo.ScheduleNextDueDate(task, task.NextDueDate.UTC())
	if err != nil {
		log.Errorf("error scheduling next due date: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error scheduling next due date",
		})
		return
	}

	if err := h.tRepo.CompleteTask(c, task, currentIdentity.UserID, nextDueDate, nil); err != nil {
		log.Errorf("error completing task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error completing task",
		})
		return
	}

	updatedTask, err := h.tRepo.GetTask(c, id)
	if err != nil {
		log.Errorf("error getting updated task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting updated task",
		})
		return
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := context.WithValue(context.Background(), "logger", logger)
		h.nRepo.GenerateNotifications(ctx, task)
	}(updatedTask, log)

	c.JSON(http.StatusOK, gin.H{
		"task": updatedTask,
	})
}

func (h *TasksAPIHandler) updateDueDate(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	log := logging.FromContext(c)

	type DueDateReq struct {
		DueDate string `json:"due_date" binding:"required"`
	}

	var dueDateReq DueDateReq
	if err := c.ShouldBindJSON(&dueDateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID required",
		})
		return
	}

	task, err := h.tRepo.GetTask(c, id)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	if currentIdentity.UserID != task.CreatedBy {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to update this task",
		})
	}

	if dueDateReq.DueDate != "" {
		rawDueDate, err := time.Parse(time.RFC3339, dueDateReq.DueDate)
		if err != nil {
			log.Errorf("error parsing due date: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Due date must be in UTC format",
			})
			return
		}

		rawDueDate = rawDueDate.UTC()
		task.NextDueDate = &rawDueDate
	}

	if err := h.tRepo.UpsertTask(c, task); err != nil {
		log.Errorf("error updating due date: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error updating due date",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": task,
	})
}

func (h *TasksAPIHandler) completeTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	log := logging.FromContext(c)

	completeTaskID := c.Param("id")
	id, err := strconv.Atoi(completeTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID required",
		})
		return
	}

	task, err := h.tRepo.GetTask(c, id)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	var completedDate time.Time = time.Now().UTC()
	var nextDueDate *time.Time
	nextDueDate, err = tRepo.ScheduleNextDueDate(task, completedDate)
	if err != nil {
		log.Errorf("error scheduling next due date: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error scheduling next due date: %s", err),
		})
		return
	}

	if err := h.tRepo.CompleteTask(c, task, currentIdentity.UserID, nextDueDate, &completedDate); err != nil {
		log.Errorf("error completing task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error completing task",
		})
		return
	}

	updatedTask, err := h.tRepo.GetTask(c, id)
	if err != nil {
		log.Errorf("error getting updated task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting updated task",
		})
		return
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := context.WithValue(context.Background(), "logger", logger)
		h.nRepo.GenerateNotifications(ctx, task)
	}(updatedTask, log)

	c.JSON(http.StatusOK, gin.H{
		"task": updatedTask,
	})
}

func (h *TasksAPIHandler) uncompleteTask(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	log := logging.FromContext(c)

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	task, err := h.tRepo.GetTask(c, id)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	if currentIdentity.UserID != task.CreatedBy {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to update this task",
		})
		return
	}

	if err := h.tRepo.UncompleteTask(c, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task was not completed already"})
			return
		}
		log.Errorf("error uncompleting task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error uncompleting task",
		})
		return
	}

	updatedTask, err := h.tRepo.GetTask(c, id)
	if err != nil {
		log.Errorf("error getting updated task: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting updated task",
		})
		return
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := context.WithValue(context.Background(), "logger", logger)
		h.nRepo.GenerateNotifications(ctx, task)
	}(updatedTask, log)

	c.JSON(http.StatusOK, gin.H{
		"task": updatedTask,
	})
}

func (h *TasksAPIHandler) GetTaskHistory(c *gin.Context) {
	currentIdentity := auth.CurrentIdentity(c)
	log := logging.FromContext(c)

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID required",
		})
		return
	}

	if err := h.tRepo.IsTaskOwner(c, id, currentIdentity.UserID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to view this task's history",
		})
		return
	}

	TaskHistory, err := h.tRepo.GetTaskHistory(c, id)
	if err != nil {
		log.Errorf("error getting task history: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": TaskHistory,
	})
}

func TaskRoutes(router *gin.Engine, h *TasksAPIHandler, auth *jwt.GinJWTMiddleware) {
	tasksRoutes := router.Group("api/v1/tasks")
	tasksRoutes.Use(auth.MiddlewareFunc())
	{
		tasksRoutes.GET("/", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.getTasks)
		tasksRoutes.GET("/completed", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.getCompletedTasks)
		tasksRoutes.PUT("/", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.editTask)
		tasksRoutes.POST("/", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.createTask)
		tasksRoutes.GET("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.getTask)
		tasksRoutes.GET("/:id/history", authMW.ScopeMiddleware(models.ApiTokenScopeTaskRead), h.GetTaskHistory)
		tasksRoutes.POST("/:id/do", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.completeTask)
		tasksRoutes.POST("/:id/undo", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.uncompleteTask)
		tasksRoutes.POST("/:id/skip", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.skipTask)
		tasksRoutes.PUT("/:id/dueDate", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.updateDueDate)
		tasksRoutes.DELETE("/:id", authMW.ScopeMiddleware(models.ApiTokenScopeTaskWrite), h.deleteTask)
	}
}
