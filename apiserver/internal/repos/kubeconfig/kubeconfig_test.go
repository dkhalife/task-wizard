package repos

import (
	"context"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/utils/test"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type KubeconfigTestSuite struct {
	test.DatabaseTestSuite
	repo IKubeconfigRepo
	cfg  *config.Config
}

func TestKubeconfigTestSuite(t *testing.T) {
	suite.Run(t, new(KubeconfigTestSuite))
}

func (s *KubeconfigTestSuite) SetupTest() {
	s.DatabaseTestSuite.SetupTest()

	s.cfg = &config.Config{
		Jwt: config.JwtConfig{
			Secret: "test-secret-for-encryption",
		},
	}
	s.repo = NewKubeconfigRepository(s.DB, s.cfg)

	// Create the kubeconfig table
	err := s.DB.AutoMigrate(&models.KubeContext{})
	s.Require().NoError(err)
}

func (s *KubeconfigTestSuite) createTestUser() *models.User {
	user := &models.User{
		Email:     "test@example.com",
		Password:  "password",
		CreatedAt: time.Now(),
	}
	err := s.DB.Create(user).Error
	s.Require().NoError(err)
	return user
}

func (s *KubeconfigTestSuite) TestCreateContext() {
	ctx := context.Background()
	user := s.createTestUser()

	kubeContext := &models.KubeContext{
		UserID:                   user.ID,
		Name:                     "test-context",
		ClusterName:              "test-cluster",
		Server:                   "https://kubernetes.example.com:6443",
		CertificateAuthorityData: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t",
		ContextUser:              "test-user",
		Token:                    "my-secret-token",
		IsActive:                 false,
	}

	err := s.repo.CreateContext(ctx, kubeContext)
	s.Require().NoError(err)
	s.NotZero(kubeContext.ID)

	// Verify it was created and data is encrypted in the database
	var stored models.KubeContext
	err = s.DB.First(&stored, kubeContext.ID).Error
	s.Require().NoError(err)
	s.NotEqual("my-secret-token", stored.Token) // Should be encrypted
	s.NotEqual("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t", stored.CertificateAuthorityData)
}

func (s *KubeconfigTestSuite) TestGetContextByID() {
	ctx := context.Background()
	user := s.createTestUser()

	kubeContext := &models.KubeContext{
		UserID:                   user.ID,
		Name:                     "test-context",
		ClusterName:              "test-cluster",
		Server:                   "https://kubernetes.example.com:6443",
		CertificateAuthorityData: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t",
		ContextUser:              "test-user",
		Token:                    "my-secret-token",
		ClientCertificate:        "cert-data",
		ClientKey:                "key-data",
	}

	err := s.repo.CreateContext(ctx, kubeContext)
	s.Require().NoError(err)

	// Retrieve and verify decryption
	retrieved, err := s.repo.GetContextByID(ctx, user.ID, kubeContext.ID)
	s.Require().NoError(err)
	s.Equal("test-context", retrieved.Name)
	s.Equal("my-secret-token", retrieved.Token)
	s.Equal("cert-data", retrieved.ClientCertificate)
	s.Equal("key-data", retrieved.ClientKey)
	s.Equal("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t", retrieved.CertificateAuthorityData)
}

func (s *KubeconfigTestSuite) TestGetContextByID_NotFound() {
	ctx := context.Background()
	user := s.createTestUser()

	_, err := s.repo.GetContextByID(ctx, user.ID, 999)
	s.Require().Error(err)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *KubeconfigTestSuite) TestGetContextByID_WrongUser() {
	ctx := context.Background()
	user1 := s.createTestUser()
	
	user2 := &models.User{
		Email:     "user2@example.com",
		Password:  "password",
		CreatedAt: time.Now(),
	}
	err := s.DB.Create(user2).Error
	s.Require().NoError(err)

	kubeContext := &models.KubeContext{
		UserID:      user1.ID,
		Name:        "test-context",
		ClusterName: "test-cluster",
		Server:      "https://kubernetes.example.com:6443",
		ContextUser: "test-user",
		Token:       "token",
	}

	err = s.repo.CreateContext(ctx, kubeContext)
	s.Require().NoError(err)

	// Try to access with different user
	_, err = s.repo.GetContextByID(ctx, user2.ID, kubeContext.ID)
	s.Require().Error(err)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *KubeconfigTestSuite) TestGetAllContexts() {
	ctx := context.Background()
	user := s.createTestUser()

	// Create multiple contexts
	contexts := []*models.KubeContext{
		{
			UserID:      user.ID,
			Name:        "context1",
			ClusterName: "cluster1",
			Server:      "https://k8s1.example.com",
			ContextUser: "user1",
			Token:       "token1",
		},
		{
			UserID:      user.ID,
			Name:        "context2",
			ClusterName: "cluster2",
			Server:      "https://k8s2.example.com",
			ContextUser: "user2",
			Token:       "token2",
		},
	}

	for _, kc := range contexts {
		err := s.repo.CreateContext(ctx, kc)
		s.Require().NoError(err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Retrieve all
	retrieved, err := s.repo.GetAllContexts(ctx, user.ID)
	s.Require().NoError(err)
	s.Len(retrieved, 2)

	// Verify both contexts are present and decrypted
	contextNames := make(map[string]string)
	for _, c := range retrieved {
		contextNames[c.Name] = c.Token
	}
	s.Equal("token1", contextNames["context1"])
	s.Equal("token2", contextNames["context2"])
}

func (s *KubeconfigTestSuite) TestSetActiveContext() {
	ctx := context.Background()
	user := s.createTestUser()

	// Create multiple contexts
	context1 := &models.KubeContext{
		UserID:      user.ID,
		Name:        "context1",
		ClusterName: "cluster1",
		Server:      "https://k8s1.example.com",
		ContextUser: "user1",
		IsActive:    true,
	}
	err := s.repo.CreateContext(ctx, context1)
	s.Require().NoError(err)

	context2 := &models.KubeContext{
		UserID:      user.ID,
		Name:        "context2",
		ClusterName: "cluster2",
		Server:      "https://k8s2.example.com",
		ContextUser: "user2",
		IsActive:    false,
	}
	err = s.repo.CreateContext(ctx, context2)
	s.Require().NoError(err)

	// Set context2 as active
	err = s.repo.SetActiveContext(ctx, user.ID, context2.ID)
	s.Require().NoError(err)

	// Verify context2 is active
	var updated2 models.KubeContext
	err = s.DB.First(&updated2, context2.ID).Error
	s.Require().NoError(err)
	s.True(updated2.IsActive)

	// Verify context1 is no longer active
	var updated1 models.KubeContext
	err = s.DB.First(&updated1, context1.ID).Error
	s.Require().NoError(err)
	s.False(updated1.IsActive)
}

func (s *KubeconfigTestSuite) TestSetActiveContext_NotFound() {
	ctx := context.Background()
	user := s.createTestUser()

	err := s.repo.SetActiveContext(ctx, user.ID, 999)
	s.Require().Error(err)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *KubeconfigTestSuite) TestDeleteContext() {
	ctx := context.Background()
	user := s.createTestUser()

	kubeContext := &models.KubeContext{
		UserID:      user.ID,
		Name:        "test-context",
		ClusterName: "test-cluster",
		Server:      "https://kubernetes.example.com:6443",
		ContextUser: "test-user",
	}

	err := s.repo.CreateContext(ctx, kubeContext)
	s.Require().NoError(err)

	// Delete
	err = s.repo.DeleteContext(ctx, user.ID, kubeContext.ID)
	s.Require().NoError(err)

	// Verify it's deleted
	var count int64
	err = s.DB.Model(&models.KubeContext{}).Where("id = ?", kubeContext.ID).Count(&count).Error
	s.Require().NoError(err)
	s.Equal(int64(0), count)
}

func (s *KubeconfigTestSuite) TestDeleteContext_NotFound() {
	ctx := context.Background()
	user := s.createTestUser()

	err := s.repo.DeleteContext(ctx, user.ID, 999)
	s.Require().Error(err)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *KubeconfigTestSuite) TestGetActiveContext() {
	ctx := context.Background()
	user := s.createTestUser()

	// Create inactive context
	context1 := &models.KubeContext{
		UserID:      user.ID,
		Name:        "context1",
		ClusterName: "cluster1",
		Server:      "https://k8s1.example.com",
		ContextUser: "user1",
		IsActive:    false,
	}
	err := s.repo.CreateContext(ctx, context1)
	s.Require().NoError(err)

	// Create active context
	context2 := &models.KubeContext{
		UserID:      user.ID,
		Name:        "context2",
		ClusterName: "cluster2",
		Server:      "https://k8s2.example.com",
		ContextUser: "user2",
		Token:       "active-token",
		IsActive:    true,
	}
	err = s.repo.CreateContext(ctx, context2)
	s.Require().NoError(err)

	// Get active context
	active, err := s.repo.GetActiveContext(ctx, user.ID)
	s.Require().NoError(err)
	s.Equal("context2", active.Name)
	s.Equal("active-token", active.Token)
	s.True(active.IsActive)
}

func (s *KubeconfigTestSuite) TestGetActiveContext_NoActiveContext() {
	ctx := context.Background()
	user := s.createTestUser()

	// Create only inactive contexts
	context1 := &models.KubeContext{
		UserID:      user.ID,
		Name:        "context1",
		ClusterName: "cluster1",
		Server:      "https://k8s1.example.com",
		ContextUser: "user1",
		IsActive:    false,
	}
	err := s.repo.CreateContext(ctx, context1)
	s.Require().NoError(err)

	// Try to get active context
	_, err = s.repo.GetActiveContext(ctx, user.ID)
	s.Require().Error(err)
	s.Equal(gorm.ErrRecordNotFound, err)
}
