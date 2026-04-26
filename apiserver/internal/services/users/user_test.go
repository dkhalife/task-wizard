package users

import (
	"context"
	"net/http"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/utils/test"
	"dkhalife.com/tasks/core/internal/ws"
	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	"github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
	test.DatabaseTestSuite
	repo    uRepo.IUserRepo
	service *UserService
	wsServer *ws.WSServer
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) SetupTest() {
	s.DatabaseTestSuite.SetupTest()

	cfg := &config.Config{
		Server: config.ServerConfig{Registration: true},
	}
	s.repo = uRepo.NewUserRepository(s.DB, cfg)

	authMiddleware, _ := authMW.NewAuthMiddleware(&config.Config{}, s.repo, nil)
	taskRepo := tRepo.NewTaskRepository(s.DB, cfg)
	labelRepo := lRepo.NewLabelRepository(s.DB, cfg)
	s.wsServer = ws.NewWSServer(authMiddleware, taskRepo, labelRepo, s.repo)
	s.service = NewUserService(s.repo, s.wsServer)
}

func (s *UserServiceTestSuite) createUser() *models.User {
	user := &models.User{
		DirectoryID: "svc-dir",
		ObjectID:    "svc-obj",
		CreatedAt:   time.Now(),
	}
	s.Require().NoError(s.repo.CreateUser(context.Background(), user))
	return user
}

func (s *UserServiceTestSuite) TestRequestDeletion_Success() {
	user := s.createUser()

	status, _ := s.service.RequestDeletion(context.Background(), user.ID)
	s.Equal(http.StatusNoContent, status)

	var updated models.User
	s.Require().NoError(s.DB.First(&updated, user.ID).Error)
	s.NotNil(updated.DeletionRequestedAt)
}

func (s *UserServiceTestSuite) TestCancelDeletion_Success() {
	user := s.createUser()
	s.Require().NoError(s.repo.RequestDeletion(context.Background(), user.ID))

	status, _ := s.service.CancelDeletion(context.Background(), user.ID)
	s.Equal(http.StatusNoContent, status)

	var updated models.User
	s.Require().NoError(s.DB.First(&updated, user.ID).Error)
	s.Nil(updated.DeletionRequestedAt)
}

func (s *UserServiceTestSuite) TestProcessDeletions_DeletesExpiredUsers() {
	ctx := context.Background()
	user := s.createUser()

	past := time.Now().Add(-25 * time.Hour)
	s.Require().NoError(s.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("deletion_requested_at", past).Error)

	err := s.service.ProcessDeletions(ctx)
	s.Require().NoError(err)

	var count int64
	s.DB.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	s.Zero(count)
}

func (s *UserServiceTestSuite) TestProcessDeletions_SkipsUsersWithinGracePeriod() {
	ctx := context.Background()
	user := s.createUser()

	recent := time.Now().Add(-1 * time.Hour)
	s.Require().NoError(s.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("deletion_requested_at", recent).Error)

	err := s.service.ProcessDeletions(ctx)
	s.Require().NoError(err)

	var count int64
	s.DB.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	s.Equal(int64(1), count)
}

func (s *UserServiceTestSuite) TestProcessDeletions_SkipsNormalUsers() {
	ctx := context.Background()
	user := s.createUser()

	err := s.service.ProcessDeletions(ctx)
	s.Require().NoError(err)

	var count int64
	s.DB.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	s.Equal(int64(1), count)
}
