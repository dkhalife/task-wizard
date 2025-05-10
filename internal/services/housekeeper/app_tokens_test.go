package housekeeper

import (
	"context"
	"errors"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	repos "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/utils/email"
	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct {
	repos.IUserRepo
	getTokensFunc    func(context.Context, time.Duration) ([]*models.AppToken, error)
	deleteTokensFunc func(context.Context) error
}

func (m *mockUserRepo) GetAppTokensNearingExpiration(ctx context.Context, d time.Duration) ([]*models.AppToken, error) {
	return m.getTokensFunc(ctx, d)
}
func (m *mockUserRepo) DeleteStaleAppTokens(ctx context.Context) error {
	return m.deleteTokensFunc(ctx)
}

type mockEmailSender struct {
	email.IEmailSender
	sendFunc func(context.Context, string, string) error
}

func (m *mockEmailSender) SendTokenExpirationReminder(ctx context.Context, name, email string) error {
	return m.sendFunc(ctx, name, email)
}

func TestSendTokenExpirationReminder_Success(t *testing.T) {
	ur := &mockUserRepo{
		getTokensFunc: func(ctx context.Context, d time.Duration) ([]*models.AppToken, error) {
			return []*models.AppToken{{Name: "token1", UserID: 1}}, nil
		},
	}
	es := &mockEmailSender{
		sendFunc: func(ctx context.Context, name, email string) error { return nil },
	}
	cfg := &config.Config{SchedulerJobs: config.SchedulerConfig{TokenExpirationReminder: time.Hour}}
	prc := &AppTokenCleaner{uRepo: ur, es: es, cfg: cfg}
	err := prc.SendTokenExpirationReminder(context.Background())
	assert.NoError(t, err)
}

func TestSendTokenExpirationReminder_GetTokensError(t *testing.T) {
	ur := &mockUserRepo{
		getTokensFunc: func(ctx context.Context, d time.Duration) ([]*models.AppToken, error) {
			return nil, errors.New("db error")
		},
	}
	cfg := &config.Config{SchedulerJobs: config.SchedulerConfig{TokenExpirationReminder: time.Hour}}
	prc := &AppTokenCleaner{uRepo: ur, es: &mockEmailSender{}, cfg: cfg}
	err := prc.SendTokenExpirationReminder(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tokens")
}

func TestSendTokenExpirationReminder_EmailError(t *testing.T) {
	ur := &mockUserRepo{
		getTokensFunc: func(ctx context.Context, d time.Duration) ([]*models.AppToken, error) {
			return []*models.AppToken{{Name: "token1", UserID: 1}}, nil
		},
	}
	es := &mockEmailSender{
		sendFunc: func(ctx context.Context, name, email string) error { return errors.New("email error") },
	}
	cfg := &config.Config{SchedulerJobs: config.SchedulerConfig{TokenExpirationReminder: time.Hour}}
	prc := &AppTokenCleaner{uRepo: ur, es: es, cfg: cfg}
	err := prc.SendTokenExpirationReminder(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send token expiration reminder email")
}

func TestCleanupExpiredTokens_Success(t *testing.T) {
	ur := &mockUserRepo{
		deleteTokensFunc: func(ctx context.Context) error { return nil },
	}
	prc := &AppTokenCleaner{uRepo: ur}
	err := prc.CleanupExpiredTokens(context.Background())
	assert.NoError(t, err)
}

func TestCleanupExpiredTokens_Error(t *testing.T) {
	ur := &mockUserRepo{
		deleteTokensFunc: func(ctx context.Context) error { return errors.New("delete error") },
	}
	prc := &AppTokenCleaner{uRepo: ur}
	err := prc.CleanupExpiredTokens(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete stale app tokens")
}
