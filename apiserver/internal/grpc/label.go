package grpc

import (
	"context"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/labels"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LabelGRPCServer struct {
	UnimplementedLabelServiceServer
	labelService *labels.LabelService
	getUserID    func(ctx context.Context) (int, error)
}

func NewLabelGRPCServer(labelService *labels.LabelService, getUserID func(ctx context.Context) (int, error)) *LabelGRPCServer {
	return &LabelGRPCServer{
		labelService: labelService,
		getUserID:    getUserID,
	}
}

func (s *LabelGRPCServer) GetLabels(ctx context.Context, req *Empty) (*LabelsResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, result := s.labelService.GetUserLabels(ctx, userID)
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to get labels")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	labelResponses, ok := data["labels"].([]map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected labels format")
	}

	grpcLabels := make([]*Label, len(labelResponses))
	for i, l := range labelResponses {
		id, _ := l["id"].(int)
		name, _ := l["name"].(string)
		color, _ := l["color"].(string)
		grpcLabels[i] = &Label{
			Id:    int32(id),
			Name:  name,
			Color: color,
		}
	}

	return &LabelsResponse{Labels: grpcLabels}, nil
}

func (s *LabelGRPCServer) CreateLabel(ctx context.Context, req *CreateLabelRequest) (*LabelResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	createReq := models.CreateLabelReq{
		Name:  req.GetName(),
		Color: req.GetColor(),
	}

	statusCode, result := s.labelService.CreateLabel(ctx, userID, createReq)
	if statusCode != http.StatusCreated {
		return nil, status.Error(codes.Internal, "failed to create label")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	labelData, ok := data["label"].(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected label format")
	}

	id, _ := labelData["id"].(int)
	name, _ := labelData["name"].(string)
	color, _ := labelData["color"].(string)

	return &LabelResponse{
		Label: &Label{
			Id:    int32(id),
			Name:  name,
			Color: color,
		},
	}, nil
}

func (s *LabelGRPCServer) UpdateLabel(ctx context.Context, req *UpdateLabelRequest) (*LabelResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	updateReq := models.UpdateLabelReq{
		ID: int(req.GetId()),
		CreateLabelReq: models.CreateLabelReq{
			Name:  req.GetName(),
			Color: req.GetColor(),
		},
	}

	statusCode, result := s.labelService.UpdateLabel(ctx, userID, updateReq)
	if statusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to perform this update")
	}
	if statusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "failed to update label")
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected response format")
	}

	labelData, ok := data["label"].(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "unexpected label format")
	}

	id, _ := labelData["id"].(int)
	name, _ := labelData["name"].(string)
	color, _ := labelData["color"].(string)

	return &LabelResponse{
		Label: &Label{
			Id:    int32(id),
			Name:  name,
			Color: color,
		},
	}, nil
}

func (s *LabelGRPCServer) DeleteLabel(ctx context.Context, req *DeleteLabelRequest) (*Empty, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user identity")
	}

	statusCode, _ := s.labelService.DeleteLabel(ctx, userID, int(req.GetId()))
	if statusCode != http.StatusNoContent {
		return nil, status.Error(codes.Internal, "failed to delete label")
	}

	return &Empty{}, nil
}
