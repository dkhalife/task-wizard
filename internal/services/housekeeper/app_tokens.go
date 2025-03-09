package housekeeper

import (
	"context"

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
	log.Debug("Tokens nearing expiration", " count ", len(tokens))

	if err != nil {
		return err
	}

	for _, token := range tokens {
		log.Debug("Sending token expiration reminder", "email", token.User.Email, "token", token.Name)

		err = prc.es.SendTokenExpirationReminder(c, token.Name, token.User.Email)
		if err != nil {
			log.Error("Failed to send token expiration reminder email", "email", token.User.Email, "error", err)
		}
	}

	return nil
}

func (prc *AppTokenCleaner) CleanupExpiredTokens(c context.Context) error {
	err := prc.uRepo.DeleteStaleAppTokens(c)
	if err != nil {
		return err
	}

	return nil
}
