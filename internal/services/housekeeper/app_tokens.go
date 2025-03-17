package housekeeper

import (
	"context"
	"fmt"

	"dkhalife.com/tasks/core/config"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/utils/email"
)

type AppTokenCleaner struct {
	cfg   *config.Config
	uRepo *uRepo.UserRepository
	es    *email.EmailSender
}

func NewAppTokenCleaner(cfg *config.Config, ur *uRepo.UserRepository, es *email.EmailSender) *AppTokenCleaner {
	return &AppTokenCleaner{
		cfg:   cfg,
		uRepo: ur,
		es:    es,
	}
}

func (prc *AppTokenCleaner) SendTokenExpirationReminder(c context.Context) error {
	log := logging.FromContext(c)

	tokens, err := prc.uRepo.GetAppTokensNearingExpiration(c, prc.cfg.SchedulerJobs.TokenExpirationReminder)
	log.Debugf("Tokens nearing expiration, count=%d", len(tokens))

	if err != nil {
		return fmt.Errorf("failed to get tokens nearing expiration: %s", err.Error())
	}

	for _, token := range tokens {
		err = prc.es.SendTokenExpirationReminder(c, token.Name, token.User.Email)
		if err != nil {
			return fmt.Errorf("failed to send token expiration reminder email: %s", err.Error())
		}
	}

	return nil
}

func (prc *AppTokenCleaner) CleanupExpiredTokens(c context.Context) error {
	err := prc.uRepo.DeleteStaleAppTokens(c)
	if err != nil {
		return fmt.Errorf("failed to delete stale app tokens: %s", err.Error())
	}

	return nil
}
