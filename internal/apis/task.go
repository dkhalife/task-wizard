package apis

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	notifications "dkhalife.com/tasks/core/internal/services/notifications"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type LabelReq struct {
	LabelID int `json:"id" binding:"required"`
}

type TaskReq struct {
	ID           string                            `json:"id"`
	Title        string                            `json:"title" binding:"required"`
	NextDueDate  int64                             `json:"next_due_date"`
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
	u, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to fetch tasks",
		})
		return
	}

	tasks, err := h.tRepo.GetTasks(c, u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting tasks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

func (h *TasksAPIHandler) getTask(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to fetch task",
		})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	if currentUser.ID != task.CreatedBy {
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
	currentUser, ok := auth.CurrentUser(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to create tasks",
		})
		return
	}

	var TaskReq TaskReq
	if err := c.ShouldBindJSON(&TaskReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	var dueDate *time.Time

	if TaskReq.NextDueDate != 0 {
		rawDueDate := time.UnixMilli(TaskReq.NextDueDate)
		dueDate = &rawDueDate
	}

	createdTask := &models.Task{
		Title:        TaskReq.Title,
		Frequency:    TaskReq.Frequency,
		NextDueDate:  dueDate,
		CreatedBy:    currentUser.ID,
		IsRolling:    TaskReq.IsRolling,
		IsActive:     true,
		Notification: TaskReq.Notification,
	}
	id, err := h.tRepo.CreateTask(c, createdTask)
	createdTask.ID = id

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error creating task",
		})
		return
	}

	if err := h.lRepo.AssignLabelsToTask(c, createdTask.ID, currentUser.ID, TaskReq.Labels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error adding labels",
		})
		return
	}

	go func() {
		h.nRepo.GenerateNotifications(c, createdTask)
	}()

	c.JSON(http.StatusCreated, gin.H{
		"task": id,
	})
}

func (h *TasksAPIHandler) editTask(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to edit task",
		})
		return
	}

	var TaskReq TaskReq
	if err := c.ShouldBindJSON(&TaskReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	var dueDate *time.Time

	if TaskReq.NextDueDate != 0 {
		rawDueDate := time.UnixMilli(TaskReq.NextDueDate)
		dueDate = &rawDueDate
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	if currentUser.ID != oldTask.CreatedBy {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to edit this task",
		})
		return
	}

	if err := h.lRepo.AssignLabelsToTask(c, taskId, currentUser.ID, TaskReq.Labels); err != nil {
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
		CreatedBy:    currentUser.ID,
		IsRolling:    TaskReq.IsRolling,
		Notification: TaskReq.Notification,
		IsActive:     oldTask.IsActive,
	}

	if err := h.tRepo.UpsertTask(c, updatedTask); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error upserting task",
		})
		return
	}

	go func() {
		h.nRepo.GenerateNotifications(c, updatedTask)
	}()

	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *TasksAPIHandler) deleteTask(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to delete tasks",
		})
		return
	}

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	if err := h.tRepo.IsTaskOwner(c, id, currentUser.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to delete this task",
		})
		return
	}

	if err := h.tRepo.DeleteTask(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error deleting task",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *TasksAPIHandler) skipTask(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to skip tasks",
		})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	nextDueDate, err := tRepo.ScheduleNextDueDate(task, task.NextDueDate.UTC())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error scheduling next due date",
		})
		return
	}

	if err := h.tRepo.CompleteTask(c, task, currentUser.ID, nextDueDate, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error completing task",
		})
		return
	}

	updatedTask, err := h.tRepo.GetTask(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	go func() {
		h.nRepo.GenerateNotifications(c, updatedTask)
	}()

	c.JSON(http.StatusOK, gin.H{
		"task": updatedTask,
	})
}

func (h *TasksAPIHandler) updateDueDate(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to update task due date",
		})
		return
	}

	type DueDateReq struct {
		DueDate int64 `json:"due_date" binding:"required"`
	}

	var dueDateReq DueDateReq
	if err := c.ShouldBindJSON(&dueDateReq); err != nil {
		log.Print(err)
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	if currentUser.ID != task.CreatedBy {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to update this task",
		})
	}

	dueDate := time.UnixMilli(dueDateReq.DueDate)
	task.NextDueDate = &dueDate
	if err := h.tRepo.UpsertTask(c, task); err != nil {
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
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to complete tasks",
		})
		return
	}

	completeTaskID := c.Param("id")
	id, err := strconv.Atoi(completeTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID required",
		})
		return
	}

	type CompleteReq struct {
		CompletedDate int64 `json:"completed_date"`
	}

	var completeReq CompleteReq
	if err := c.ShouldBindJSON(&completeReq); err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	var completedDate time.Time
	rawCompletedDate := completeReq.CompletedDate
	if rawCompletedDate == 0 {
		completedDate = time.Now().UTC()
	} else {
		completedDate = time.UnixMilli(int64(completeReq.CompletedDate))
	}

	task, err := h.tRepo.GetTask(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	var nextDueDate *time.Time
	nextDueDate, err = tRepo.ScheduleNextDueDate(task, completedDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error scheduling next due date: %s", err),
		})
		return
	}

	if err := h.tRepo.CompleteTask(c, task, currentUser.ID, nextDueDate, &completedDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error completing task",
		})
		return
	}

	updatedTask, err := h.tRepo.GetTask(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		})
		return
	}

	go func() {
		h.nRepo.GenerateNotifications(c, updatedTask)
	}()

	c.JSON(http.StatusOK, gin.H{
		"task": updatedTask,
	})
}

func (h *TasksAPIHandler) GetTaskHistory(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Login required to fetch task history",
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

	if err := h.tRepo.IsTaskOwner(c, id, currentUser.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not allowed to view this task's history",
		})
		return
	}

	TaskHistory, err := h.tRepo.GetTaskHistory(c, id)
	if err != nil {
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
		tasksRoutes.GET("/", h.getTasks)
		tasksRoutes.PUT("/", h.editTask)
		tasksRoutes.POST("/", h.createTask)
		tasksRoutes.GET("/:id", h.getTask)
		tasksRoutes.GET("/:id/history", h.GetTaskHistory)
		tasksRoutes.POST("/:id/do", h.completeTask)
		tasksRoutes.POST("/:id/skip", h.skipTask)
		tasksRoutes.PUT("/:id/dueDate", h.updateDueDate)
		tasksRoutes.DELETE("/:id", h.deleteTask)
	}
}
