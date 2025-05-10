package housekeeper

import (
	"context"
	"fmt"

	uRepo "dkhalife.com/tasks/core/internal/repos/user"
)

type PasswordResetCleaner struct {
	uRepo uRepo.IUserRepo
}

func NewPasswordResetCleaner(ur uRepo.IUserRepo) *PasswordResetCleaner {
	return &PasswordResetCleaner{
		uRepo: ur,
	}
}

func (prc *PasswordResetCleaner) CleanupStalePasswordResets(c context.Context) error {
	err := prc.uRepo.DeleteStalePasswordResets(c)
	if err != nil {
		return fmt.Errorf("error deleting stale password resets: %s", err.Error())
	}

	return nil
}
