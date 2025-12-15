package grpc

import (
	"context"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/users"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGRPCServer struct {
	UnimplementedUserServiceServer
	userService *users.UserService
	userRepo    uRepo.IUserRepo
	nRepo       *nRepo.NotificationRepository
	getUserID   func(ctx context.Context) (int, error)
}

func NewUserGRPCServer(
	userService *users.UserService,
	userRepo uRepo.IUserRepo,
	nRepo *nRepo.NotificationRepository,
	getUserID func(ctx context.Context) (int, error),
) *UserGRPCServer {
	return &UserGRPCServer{
		userService: userService,
		userRepo:    userRepo,
		nRepo:       nRepo,
		getUserID:   getUserID,
	}
}

func (s *UserGRPCServer) GetProfile(ctx context.Context, req *Empty) (*UserProfileResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	notificationSettings, err := s.nRepo.GetUserNotificationSettings(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notification settings")
	}

	return &UserProfileResponse{
		User: &UserProfile{
			DisplayName:          user.DisplayName,
			NotificationSettings: modelNotificationSettingsToGRPC(notificationSettings),
		},
	}, nil
}

func (s *UserGRPCServer) UpdateNotificationSettings(ctx context.Context, req *UpdateNotificationSettingsRequest) (*Empty, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	updateReq := models.NotificationUpdateRequest{
		Provider: grpcNotificationProviderToModel(req.GetProvider()),
		Triggers: grpcNotificationOptionsToModel(req.GetTriggers()),
	}

	statusCode, _ := s.userService.UpdateNotificationSettings(ctx, userID, updateReq)
	if statusCode != http.StatusNoContent {
		return nil, status.Error(codes.Internal, "failed to update notification settings")
	}

	return &Empty{}, nil
}

func (s *UserGRPCServer) ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*Empty, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	newPassword := req.GetNewPassword()
	if newPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "new password is required")
	}

	err = s.userRepo.UpdatePasswordByUserId(ctx, userID, newPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	return &Empty{}, nil
}

func modelNotificationSettingsToGRPC(s *models.NotificationSettings) *NotificationSettings {
	if s == nil {
		return nil
	}

	return &NotificationSettings{
		Provider: modelNotificationProviderToGRPC(&s.Provider),
		Triggers: modelNotificationOptionsToGRPC(&s.Triggers),
	}
}

func modelNotificationProviderToGRPC(p *models.NotificationProvider) *NotificationProvider {
	if p == nil {
		return nil
	}

	return &NotificationProvider{
		Provider: modelNotificationProviderTypeToGRPC(p.Provider),
		Url:      p.URL,
		Method:   p.Method,
		Token:    p.Token,
	}
}

func modelNotificationProviderTypeToGRPC(t models.NotificationProviderType) NotificationProviderType {
	switch t {
	case models.NotificationProviderNone:
		return NotificationProviderType_NOTIFICATION_PROVIDER_NONE
	case models.NotificationProviderWebhook:
		return NotificationProviderType_NOTIFICATION_PROVIDER_WEBHOOK
	case models.NotificationProviderGotify:
		return NotificationProviderType_NOTIFICATION_PROVIDER_GOTIFY
	default:
		return NotificationProviderType_NOTIFICATION_PROVIDER_NONE
	}
}

func grpcNotificationProviderToModel(p *NotificationProvider) models.NotificationProvider {
	if p == nil {
		return models.NotificationProvider{}
	}

	return models.NotificationProvider{
		Provider: grpcNotificationProviderTypeToModel(p.GetProvider()),
		URL:      p.GetUrl(),
		Method:   p.GetMethod(),
		Token:    p.GetToken(),
	}
}

func grpcNotificationProviderTypeToModel(t NotificationProviderType) models.NotificationProviderType {
	switch t {
	case NotificationProviderType_NOTIFICATION_PROVIDER_NONE:
		return models.NotificationProviderNone
	case NotificationProviderType_NOTIFICATION_PROVIDER_WEBHOOK:
		return models.NotificationProviderWebhook
	case NotificationProviderType_NOTIFICATION_PROVIDER_GOTIFY:
		return models.NotificationProviderGotify
	default:
		return models.NotificationProviderNone
	}
}
