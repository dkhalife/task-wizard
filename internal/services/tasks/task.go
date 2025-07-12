package tasks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/services/notifications"
	"dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskService struct {
	t        *tRepo.TaskRepository
	ws       *ws.WSServer
	notifier *notifications.Notifier
	n        *nRepo.NotificationRepository
	l        *lRepo.LabelRepository
}

func NewTaskService(t *tRepo.TaskRepository, ws *ws.WSServer, notifier *notifications.Notifier, n *nRepo.NotificationRepository, l *lRepo.LabelRepository) *TaskService {
	return &TaskService{
		t:        t,
		ws:       ws,
		notifier: notifier,
		n:        n,
		l:        l,
	}
}

func (s *TaskService) GetUserTasks(ctx context.Context, userID int) (int, interface{}) {
	log := logging.FromContext(ctx)
	tasks, err := s.t.GetTasks(ctx, userID)
	if err != nil {
		log.Errorf("error getting tasks: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting tasks",
		}
	}

	return http.StatusOK, gin.H{
		"tasks": tasks,
	}
}

func (s *TaskService) GetCompletedTasks(ctx context.Context, userID, limit, page int) (int, interface{}) {
	log := logging.FromContext(ctx)
	offset := (page - 1) * limit

	tasks, err := s.t.GetCompletedTasks(ctx, userID, limit, offset)
	if err != nil {
		log.Errorf("error getting completed tasks: %s", err.Error())
		return http.StatusInternalServerError, gin.H{}
	}

	return http.StatusOK, gin.H{
		"tasks": tasks,
	}
}

func (s *TaskService) GetTask(ctx context.Context, userID, taskID int) (int, interface{}) {
	log := logging.FromContext(ctx)

	task, err := s.t.GetTask(ctx, taskID)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		}
	}

	if userID != task.CreatedBy {
		return http.StatusForbidden, gin.H{
			"error": "You are not allowed to view this task",
		}
	}

	return http.StatusOK, gin.H{
		"task": task,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, userID int, req models.CreateTaskReq) (int, interface{}) {
	log := logging.FromContext(ctx)

	var dueDate *time.Time
	if req.NextDueDate != "" {
		rawDueDate, err := time.Parse(time.RFC3339, req.NextDueDate)
		if err != nil {
			log.Errorf("error parsing due date: %s", err.Error())
			return http.StatusBadRequest, gin.H{
				"error": "Due date must be in UTC format",
			}
		}

		rawDueDate = rawDueDate.UTC()
		dueDate = &rawDueDate
	}

	var endDate *time.Time
	if req.EndDate != "" {
		rawEndDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			log.Errorf("error parsing end date: %s", err.Error())
			return http.StatusBadRequest, gin.H{
				"error": "End date must be in UTC format",
			}
		}

		rawEndDate = rawEndDate.UTC()
		endDate = &rawEndDate
	}

	createdTask := &models.Task{
		Title:        req.Title,
		Frequency:    req.Frequency,
		NextDueDate:  dueDate,
		EndDate:      endDate,
		CreatedBy:    userID,
		IsRolling:    req.IsRolling,
		IsActive:     true,
		Notification: req.Notification,
	}

	id, err := s.t.CreateTask(ctx, createdTask)
	createdTask.ID = id

	if err != nil {
		log.Errorf("error creating task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error creating task",
		}
	}

	if err := s.l.AssignLabelsToTask(ctx, createdTask.ID, userID, req.Labels); err != nil {
		log.Errorf("error assigning labels to task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error adding labels",
		}
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := logging.ContextWithLogger(context.Background(), logger)
		s.n.GenerateNotifications(ctx, task)
	}(createdTask, log)

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "task_created",
		Data:   createdTask,
	})

	return http.StatusCreated, gin.H{
		"task": id,
	}
}

func (s *TaskService) EditTask(ctx context.Context, userID int, req models.UpdateTaskReq) (int, interface{}) {
	log := logging.FromContext(ctx)

	var dueDate *time.Time
	if req.NextDueDate != "" {
		rawDueDate, err := time.Parse(time.RFC3339, req.NextDueDate)
		if err != nil {
			log.Errorf("error parsing due date: %s", err.Error())
			return http.StatusBadRequest, gin.H{
				"error": "Due date must be in UTC format",
			}
		}

		rawDueDate = rawDueDate.UTC()
		dueDate = &rawDueDate
	}

	var endDate *time.Time
	if req.EndDate != "" {
		rawEndDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			log.Errorf("error parsing end date: %s", err.Error())
			return http.StatusBadRequest, gin.H{
				"error": "End date must be in UTC format",
			}
		}

		rawEndDate = rawEndDate.UTC()
		endDate = &rawEndDate
	}

	taskId, err := strconv.Atoi(req.ID)
	if err != nil {
		return http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		}
	}
	oldTask, err := s.t.GetTask(ctx, taskId)

	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		}
	}

	if userID != oldTask.CreatedBy {
		return http.StatusForbidden, gin.H{
			"error": "You are not allowed to edit this task",
		}
	}

	if err := s.l.AssignLabelsToTask(ctx, taskId, userID, req.Labels); err != nil {
		log.Errorf("error assigning labels to task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error adding labels",
		}
	}

	updatedTask := &models.Task{
		ID:           taskId,
		Title:        req.Title,
		Frequency:    req.Frequency,
		NextDueDate:  dueDate,
		EndDate:      endDate,
		CreatedBy:    userID,
		IsRolling:    req.IsRolling,
		Notification: req.Notification,
		IsActive:     oldTask.IsActive,
	}

	if err := s.t.UpsertTask(ctx, updatedTask); err != nil {
		log.Errorf("error upserting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error upserting task",
		}
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := logging.ContextWithLogger(context.Background(), logger)
		s.n.GenerateNotifications(ctx, task)
	}(updatedTask, log)

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "task_updated",
		Data:   updatedTask,
	})

	return http.StatusNoContent, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, userID, taskID int) (int, interface{}) {
	log := logging.FromContext(ctx)

	if err := s.t.IsTaskOwner(ctx, taskID, userID); err != nil {
		return http.StatusForbidden, gin.H{
			"error": "You are not allowed to delete this task",
		}
	}

	if err := s.t.DeleteTask(ctx, taskID); err != nil {
		log.Errorf("error deleting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error deleting task",
		}
	}

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "task_deleted",
		Data: gin.H{
			"id": taskID,
		},
	})

	return http.StatusNoContent, nil
}

func (s *TaskService) SkipTask(ctx context.Context, userID, taskID int) (int, interface{}) {
	log := logging.FromContext(ctx)
	task, err := s.t.GetTask(ctx, taskID)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		}
	}

	nextDueDate, err := tRepo.ScheduleNextDueDate(task, task.NextDueDate.UTC())
	if err != nil {
		log.Errorf("error scheduling next due date: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error scheduling next due date",
		}
	}

	if err := s.t.CompleteTask(ctx, task, userID, nextDueDate, nil); err != nil {
		log.Errorf("error completing task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error completing task",
		}
	}

	updatedTask, err := s.t.GetTask(ctx, taskID)
	if err != nil {
		log.Errorf("error getting updated task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting updated task",
		}
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := logging.ContextWithLogger(context.Background(), logger)
		s.n.GenerateNotifications(ctx, task)
	}(updatedTask, log)

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "task_skipped",
		Data:   updatedTask,
	})

	return http.StatusOK, gin.H{
		"task": updatedTask,
	}
}

func (s *TaskService) UpdateDueDate(ctx context.Context, userID, taskID int, req models.UpdateDueDateReq) (int, interface{}) {
	log := logging.FromContext(ctx)

	task, err := s.t.GetTask(ctx, taskID)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		}
	}

	if userID != task.CreatedBy {
		return http.StatusForbidden, gin.H{
			"error": "You are not allowed to update this task",
		}
	}

	if req.DueDate != "" {
		rawDueDate, err := time.Parse(time.RFC3339, req.DueDate)
		if err != nil {
			log.Errorf("error parsing due date: %s", err.Error())
			return http.StatusBadRequest, gin.H{
				"error": "Due date must be in UTC format",
			}
		}

		rawDueDate = rawDueDate.UTC()
		task.NextDueDate = &rawDueDate
	}

	if err := s.t.UpsertTask(ctx, task); err != nil {
		log.Errorf("error updating due date: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error updating due date",
		}
	}

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "task_updated",
		Data:   task,
	})

	return http.StatusOK, gin.H{
		"task": task,
	}
}

func (s *TaskService) CompleteTask(ctx context.Context, userID, taskID int) (int, interface{}) {
	log := logging.FromContext(ctx)

	task, err := s.t.GetTask(ctx, taskID)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		}
	}

	var completedDate time.Time = time.Now().UTC()
	var nextDueDate *time.Time
	nextDueDate, err = tRepo.ScheduleNextDueDate(task, completedDate)
	if err != nil {
		log.Errorf("error scheduling next due date: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error scheduling next due date: %s", err),
		}
	}

	if err := s.t.CompleteTask(ctx, task, userID, nextDueDate, &completedDate); err != nil {
		log.Errorf("error completing task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error completing task",
		}
	}

	updatedTask, err := s.t.GetTask(ctx, taskID)
	if err != nil {
		log.Errorf("error getting updated task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting updated task",
		}
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := logging.ContextWithLogger(context.Background(), logger)
		s.n.GenerateNotifications(ctx, task)
	}(updatedTask, log)

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "task_completed",
		Data:   updatedTask,
	})

	return http.StatusOK, gin.H{
		"task": updatedTask,
	}
}

func (s *TaskService) UncompleteTask(ctx context.Context, userID, taskID int) (int, interface{}) {
	log := logging.FromContext(ctx)
	task, err := s.t.GetTask(ctx, taskID)
	if err != nil {
		log.Errorf("error getting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting task",
		}
	}

	if userID != task.CreatedBy {
		return http.StatusForbidden, gin.H{
			"error": "You are not allowed to update this task",
		}
	}

	if err := s.t.UncompleteTask(ctx, taskID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusBadRequest, gin.H{
				"error": "Task was not completed already",
			}
		}

		log.Errorf("error uncompleting task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error uncompleting task",
		}
	}

	updatedTask, err := s.t.GetTask(ctx, taskID)
	if err != nil {
		log.Errorf("error getting updated task: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting updated task",
		}
	}

	go func(task *models.Task, logger *zap.SugaredLogger) {
		ctx := logging.ContextWithLogger(context.Background(), logger)
		s.n.GenerateNotifications(ctx, task)
	}(updatedTask, log)

	s.ws.BroadcastToUser(userID, ws.WSResponse{
		Action: "task_uncompleted",
		Data:   updatedTask,
	})

	return http.StatusOK, gin.H{
		"task": updatedTask,
	}
}

func (s *TaskService) GetTaskHistory(ctx context.Context, userID, taskID int) (int, interface{}) {
	log := logging.FromContext(ctx)

	if err := s.t.IsTaskOwner(ctx, taskID, userID); err != nil {
		return http.StatusForbidden, gin.H{
			"error": "You are not allowed to view this task's history",
		}
	}

	TaskHistory, err := s.t.GetTaskHistory(ctx, taskID)
	if err != nil {
		log.Errorf("error getting task history: %s", err.Error())
		return http.StatusInternalServerError, gin.H{
			"error": "Error getting task history",
		}
	}

	return http.StatusOK, gin.H{
		"history": TaskHistory,
	}
}
