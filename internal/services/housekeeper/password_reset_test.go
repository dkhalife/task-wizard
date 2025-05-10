package housekeeper

import (
	"context"
	"errors"
	"testing"

	repos "dkhalife.com/tasks/core/internal/repos/user"
	"github.com/stretchr/testify/assert"
)

type mockPasswordResetUserRepo struct {
	repos.IUserRepo
	deleteFunc func(context.Context) error
}

func (m *mockPasswordResetUserRepo) DeleteStalePasswordResets(ctx context.Context) error {
	return m.deleteFunc(ctx)
}

func TestCleanupStalePasswordResets_Success(t *testing.T) {
	repo := &mockPasswordResetUserRepo{
		deleteFunc: func(ctx context.Context) error { return nil },
	}
	prc := NewPasswordResetCleaner(repo)
	err := prc.CleanupStalePasswordResets(context.Background())
	assert.NoError(t, err)
}

func TestCleanupStalePasswordResets_Error(t *testing.T) {
	repo := &mockPasswordResetUserRepo{
		deleteFunc: func(ctx context.Context) error { return errors.New("db error") },
	}
	prc := NewPasswordResetCleaner(repo)
	err := prc.CleanupStalePasswordResets(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error deleting stale password resets")
}
