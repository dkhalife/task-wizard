package grpc

import (
	"context"
	"net/http"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/tasks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskGRPCServer struct {
	UnimplementedTaskServiceServer
	taskService *tasks.TaskService
	getUserID   func(ctx context.Context) (int, error)
}

func NewTaskGRPCServer(taskService *tasks.TaskService, getUserID func(ctx context.Context) (int, error)) *TaskGRPCServer {
	return &TaskGRPCServer{
		taskService: taskService,
		getUserID:   getUserID,
	}
}

func (s *TaskGRPCServer) GetTasks(ctx context.Context, req *Empty) (*TasksResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, result := s.taskService.GetUserTasks(ctx, userID)
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to get tasks")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTasks, ok := data["tasks"].([]*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected tasks format")
	}

	grpcTasks := make([]*Task, len(modelTasks))
	for i, t := range modelTasks {
		grpcTasks[i] = modelTaskToGRPC(t)
	}

	return &TasksResponse{Tasks: grpcTasks}, nil
}

func (s *TaskGRPCServer) GetCompletedTasks(ctx context.Context, req *GetCompletedTasksRequest) (*TasksResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	limit := int(req.GetLimit())
	page := int(req.GetPage())
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	statusCode, result := s.taskService.GetCompletedTasks(ctx, userID, limit, page)
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to get completed tasks")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTasks, ok := data["tasks"].([]*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected tasks format")
	}

	grpcTasks := make([]*Task, len(modelTasks))
	for i, t := range modelTasks {
		grpcTasks[i] = modelTaskToGRPC(t)
	}

	return &TasksResponse{Tasks: grpcTasks}, nil
}

func (s *TaskGRPCServer) GetTask(ctx context.Context, req *GetTaskRequest) (*TaskResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, result := s.taskService.GetTask(ctx, userID, int(req.GetId()))
	if statusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to view this task")
	}
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to get task")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTask, ok := data["task"].(*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected task format")
	}

	return &TaskResponse{Task: modelTaskToGRPC(modelTask)}, nil
}

func (s *TaskGRPCServer) CreateTask(ctx context.Context, req *CreateTaskRequest) (*TaskResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	createReq := models.CreateTaskReq{
		Title:        req.GetTitle(),
		IsRolling:    req.GetIsRolling(),
		Frequency:    grpcFrequencyToModel(req.GetFrequency()),
		Notification: grpcNotificationOptionsToModel(req.GetNotification()),
		Labels:       int32SliceToIntSlice(req.GetLabelIds()),
	}

	if req.NextDueDate != nil {
		createReq.NextDueDate = time.UnixMilli(req.GetNextDueDate()).UTC().Format(time.RFC3339)
	}
	if req.EndDate != nil {
		createReq.EndDate = time.UnixMilli(req.GetEndDate()).UTC().Format(time.RFC3339)
	}

	statusCode, result := s.taskService.CreateTask(ctx, userID, createReq)
	if statusCode == http.StatusBadRequest {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if statusCode != http.StatusCreated {
		return nil, status.Error(codes.Internal, "failed to create task")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	taskID, ok := data["task"].(int)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected task ID format")
	}

	// Get the created task
	statusCode, result = s.taskService.GetTask(ctx, userID, taskID)
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to get created task")
	}

	data, ok = result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTask, ok := data["task"].(*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected task format")
	}

	return &TaskResponse{Task: modelTaskToGRPC(modelTask)}, nil
}

func (s *TaskGRPCServer) UpdateTask(ctx context.Context, req *UpdateTaskRequest) (*TaskResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	updateReq := models.UpdateTaskReq{
		ID:           int(req.GetId()),
		Title:        req.GetTitle(),
		IsRolling:    req.GetIsRolling(),
		Frequency:    grpcFrequencyToModel(req.GetFrequency()),
		Notification: grpcNotificationOptionsToModel(req.GetNotification()),
		Labels:       int32SliceToIntSlice(req.GetLabelIds()),
	}

	if req.NextDueDate != nil {
		updateReq.NextDueDate = time.UnixMilli(req.GetNextDueDate()).UTC().Format(time.RFC3339)
	}
	if req.EndDate != nil {
		updateReq.EndDate = time.UnixMilli(req.GetEndDate()).UTC().Format(time.RFC3339)
	}

	statusCode, _ := s.taskService.EditTask(ctx, userID, updateReq)
	if statusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to edit this task")
	}
	if statusCode == http.StatusBadRequest {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if statusCode != http.StatusNoContent {
		return nil, status.Error(codes.Internal, "failed to update task")
	}

	// Get the updated task
	statusCode, result := s.taskService.GetTask(ctx, userID, int(req.GetId()))
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to get updated task")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTask, ok := data["task"].(*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected task format")
	}

	return &TaskResponse{Task: modelTaskToGRPC(modelTask)}, nil
}

func (s *TaskGRPCServer) DeleteTask(ctx context.Context, req *DeleteTaskRequest) (*Empty, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, _ := s.taskService.DeleteTask(ctx, userID, int(req.GetId()))
	if statusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to delete this task")
	}
	if statusCode != http.StatusNoContent {
		return nil, status.Error(codes.Internal, "failed to delete task")
	}

	return &Empty{}, nil
}

func (s *TaskGRPCServer) CompleteTask(ctx context.Context, req *CompleteTaskRequest) (*TaskResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, result := s.taskService.CompleteTask(ctx, userID, int(req.GetId()), req.GetEndRecurrence())
	if statusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to complete this task")
	}
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to complete task")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTask, ok := data["task"].(*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected task format")
	}

	return &TaskResponse{Task: modelTaskToGRPC(modelTask)}, nil
}

func (s *TaskGRPCServer) UncompleteTask(ctx context.Context, req *UncompleteTaskRequest) (*TaskResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, result := s.taskService.UncompleteTask(ctx, userID, int(req.GetId()))
	if statusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to uncomplete this task")
	}
	if statusCode == http.StatusBadRequest {
		return nil, status.Error(codes.InvalidArgument, "task was not completed already")
	}
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to uncomplete task")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTask, ok := data["task"].(*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected task format")
	}

	return &TaskResponse{Task: modelTaskToGRPC(modelTask)}, nil
}

func (s *TaskGRPCServer) SkipTask(ctx context.Context, req *SkipTaskRequest) (*TaskResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, result := s.taskService.SkipTask(ctx, userID, int(req.GetId()))
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to skip task")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTask, ok := data["task"].(*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected task format")
	}

	return &TaskResponse{Task: modelTaskToGRPC(modelTask)}, nil
}

func (s *TaskGRPCServer) UpdateDueDate(ctx context.Context, req *UpdateDueDateRequest) (*TaskResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	updateReq := models.UpdateDueDateReq{
		DueDate: time.UnixMilli(req.GetDueDate()).UTC().Format(time.RFC3339),
	}

	statusCode, result := s.taskService.UpdateDueDate(ctx, userID, int(req.GetId()), updateReq)
	if statusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to update this task")
	}
	if statusCode == http.StatusBadRequest {
		return nil, status.Error(codes.InvalidArgument, "invalid due date")
	}
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to update due date")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelTask, ok := data["task"].(*models.Task)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected task format")
	}

	return &TaskResponse{Task: modelTaskToGRPC(modelTask)}, nil
}

func (s *TaskGRPCServer) GetTaskHistory(ctx context.Context, req *GetTaskHistoryRequest) (*TaskHistoryResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, result := s.taskService.GetTaskHistory(ctx, userID, int(req.GetId()))
	if statusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to view this task's history")
	}
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to get task history")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	modelHistory, ok := data["history"].([]models.TaskHistory)
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected history format")
	}

	grpcHistory := make([]*TaskHistory, len(modelHistory))
	for i, h := range modelHistory {
		grpcHistory[i] = modelTaskHistoryToGRPC(&h)
	}

	return &TaskHistoryResponse{History: grpcHistory}, nil
}

func modelTaskToGRPC(t *models.Task) *Task {
	if t == nil {
		return nil
	}

	task := &Task{
		Id:           int32(t.ID),
		Title:        t.Title,
		IsRolling:    t.IsRolling,
		Frequency:    modelFrequencyToGRPC(&t.Frequency),
		Notification: modelNotificationOptionsToGRPC(&t.Notification),
		Labels:       modelLabelsToGRPC(t.Labels),
	}

	if t.NextDueDate != nil {
		millis := t.NextDueDate.UnixMilli()
		task.NextDueDate = &millis
	}
	if t.EndDate != nil {
		millis := t.EndDate.UnixMilli()
		task.EndDate = &millis
	}

	return task
}

func modelTaskHistoryToGRPC(h *models.TaskHistory) *TaskHistory {
	if h == nil {
		return nil
	}

	history := &TaskHistory{
		Id:     int32(h.ID),
		TaskId: int32(h.TaskID),
	}

	if h.CompletedDate != nil {
		millis := h.CompletedDate.UnixMilli()
		history.CompletedDate = &millis
	}
	if h.DueDate != nil {
		millis := h.DueDate.UnixMilli()
		history.DueDate = &millis
	}

	return history
}

func modelFrequencyToGRPC(f *models.Frequency) *Frequency {
	if f == nil {
		return nil
	}

	return &Frequency{
		Type:   modelFrequencyTypeToGRPC(f.Type),
		On:     modelRepeatOnToGRPC(f.On),
		Every:  int32(f.Every),
		Unit:   modelIntervalUnitToGRPC(f.Unit),
		Days:   f.Days,
		Months: f.Months,
	}
}

func modelFrequencyTypeToGRPC(t models.FrequencyType) FrequencyType {
	switch t {
	case models.RepeatOnce:
		return FrequencyType_FREQUENCY_TYPE_ONCE
	case models.RepeatDaily:
		return FrequencyType_FREQUENCY_TYPE_DAILY
	case models.RepeatWeekly:
		return FrequencyType_FREQUENCY_TYPE_WEEKLY
	case models.RepeatMonthly:
		return FrequencyType_FREQUENCY_TYPE_MONTHLY
	case models.RepeatYearly:
		return FrequencyType_FREQUENCY_TYPE_YEARLY
	case models.RepeatCustom:
		return FrequencyType_FREQUENCY_TYPE_CUSTOM
	default:
		return FrequencyType_FREQUENCY_TYPE_UNSPECIFIED
	}
}

func modelRepeatOnToGRPC(r models.RepeatOn) RepeatOn {
	switch r {
	case models.Interval:
		return RepeatOn_REPEAT_ON_INTERVAL
	case models.DaysOfTheWeek:
		return RepeatOn_REPEAT_ON_DAYS_OF_THE_WEEK
	case models.DayOfTheMonths:
		return RepeatOn_REPEAT_ON_DAY_OF_THE_MONTHS
	default:
		return RepeatOn_REPEAT_ON_UNSPECIFIED
	}
}

func modelIntervalUnitToGRPC(u models.IntervalUnit) IntervalUnit {
	switch u {
	case models.Hours:
		return IntervalUnit_INTERVAL_UNIT_HOURS
	case models.Days:
		return IntervalUnit_INTERVAL_UNIT_DAYS
	case models.Weeks:
		return IntervalUnit_INTERVAL_UNIT_WEEKS
	case models.Months:
		return IntervalUnit_INTERVAL_UNIT_MONTHS
	case models.Years:
		return IntervalUnit_INTERVAL_UNIT_YEARS
	default:
		return IntervalUnit_INTERVAL_UNIT_UNSPECIFIED
	}
}

func modelNotificationOptionsToGRPC(n *models.NotificationTriggerOptions) *NotificationTriggerOptions {
	if n == nil {
		return nil
	}

	return &NotificationTriggerOptions{
		Enabled: n.Enabled,
		DueDate: n.DueDate,
		PreDue:  n.PreDue,
		Overdue: n.Overdue,
	}
}

func modelLabelsToGRPC(labels []models.Label) []*Label {
	grpcLabels := make([]*Label, len(labels))
	for i, l := range labels {
		grpcLabels[i] = &Label{
			Id:    int32(l.ID),
			Name:  l.Name,
			Color: l.Color,
		}
	}
	return grpcLabels
}

func grpcFrequencyToModel(f *Frequency) models.Frequency {
	if f == nil {
		return models.Frequency{}
	}

	return models.Frequency{
		Type:   grpcFrequencyTypeToModel(f.GetType()),
		On:     grpcRepeatOnToModel(f.GetOn()),
		Every:  int(f.GetEvery()),
		Unit:   grpcIntervalUnitToModel(f.GetUnit()),
		Days:   f.GetDays(),
		Months: f.GetMonths(),
	}
}

func grpcFrequencyTypeToModel(t FrequencyType) models.FrequencyType {
	switch t {
	case FrequencyType_FREQUENCY_TYPE_ONCE:
		return models.RepeatOnce
	case FrequencyType_FREQUENCY_TYPE_DAILY:
		return models.RepeatDaily
	case FrequencyType_FREQUENCY_TYPE_WEEKLY:
		return models.RepeatWeekly
	case FrequencyType_FREQUENCY_TYPE_MONTHLY:
		return models.RepeatMonthly
	case FrequencyType_FREQUENCY_TYPE_YEARLY:
		return models.RepeatYearly
	case FrequencyType_FREQUENCY_TYPE_CUSTOM:
		return models.RepeatCustom
	default:
		return ""
	}
}

func grpcRepeatOnToModel(r RepeatOn) models.RepeatOn {
	switch r {
	case RepeatOn_REPEAT_ON_INTERVAL:
		return models.Interval
	case RepeatOn_REPEAT_ON_DAYS_OF_THE_WEEK:
		return models.DaysOfTheWeek
	case RepeatOn_REPEAT_ON_DAY_OF_THE_MONTHS:
		return models.DayOfTheMonths
	default:
		return ""
	}
}

func grpcIntervalUnitToModel(u IntervalUnit) models.IntervalUnit {
	switch u {
	case IntervalUnit_INTERVAL_UNIT_HOURS:
		return models.Hours
	case IntervalUnit_INTERVAL_UNIT_DAYS:
		return models.Days
	case IntervalUnit_INTERVAL_UNIT_WEEKS:
		return models.Weeks
	case IntervalUnit_INTERVAL_UNIT_MONTHS:
		return models.Months
	case IntervalUnit_INTERVAL_UNIT_YEARS:
		return models.Years
	default:
		return ""
	}
}

func grpcNotificationOptionsToModel(n *NotificationTriggerOptions) models.NotificationTriggerOptions {
	if n == nil {
		return models.NotificationTriggerOptions{}
	}

	return models.NotificationTriggerOptions{
		Enabled: n.GetEnabled(),
		DueDate: n.GetDueDate(),
		PreDue:  n.GetPreDue(),
		Overdue: n.GetOverdue(),
	}
}

func int32SliceToIntSlice(s []int32) []int {
	result := make([]int, len(s))
	for i, v := range s {
		result[i] = int(v)
	}
	return result
}
