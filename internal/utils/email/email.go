package email

import (
	"context"
	"errors"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/logging"
	"github.com/wneessen/go-mail"
)

type EmailSender struct {
	AppHost  string
	Host     string
	Port     int
	Email    string
	Password string
}

func NewEmailSender(conf *config.Config) *EmailSender {
	return &EmailSender{
		AppHost:  conf.Server.HostName,
		Host:     conf.EmailConfig.Host,
		Port:     conf.EmailConfig.Port,
		Email:    conf.EmailConfig.Email,
		Password: conf.EmailConfig.Password,
	}
}

func (es *EmailSender) validateConfig() error {
	if es.AppHost == "" {
		return errors.New("appHost is required to send emails")
	}

	if es.Host == "" {
		return errors.New("SMTP Host is required to send emails")
	}

	if es.Port == 0 {
		return errors.New("SMTP Port is required to send emails")
	}

	if es.Email == "" {
		return errors.New("email is required to send emails")
	}

	if es.Password == "" {
		return errors.New("password is required to send emails")
	}

	return nil
}

func (es *EmailSender) SendResetPasswordEmail(c context.Context, to string, code string) error {
	err := es.validateConfig()
	if err != nil {
		return err
	}

	resetURL := es.AppHost + "/password/update?c=" + code
	htmlBody := `
		<html>
		<body>
			<p>Dear user,</p>
			<p>Please use the link below to reset your password:</p>
			<p><a href="` + resetURL + `">Reset Password</a></p>
			<p>If you did not request a password reset, please ignore this email.</p>
			<p>Thank you,</p>
			<p><strong>Task Wizard</strong></p>
		</body>
		</html>
	`

	log := logging.FromContext(c)
	message := mail.NewMsg()
	if err := message.From(es.Email); err != nil {
		log.Fatalf("failed to set From address: %s", err)
	}
	if err := message.To(to); err != nil {
		log.Fatalf("failed to set To address: %s", err)
	}

	message.Subject("Task Wizard - Password Reset")
	message.SetBodyString(mail.TypeTextHTML, htmlBody)

	client, err := mail.NewClient(es.Host,
		mail.WithPort(es.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(es.Email), mail.WithPassword(es.Password))

	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}

	if err := client.DialAndSend(message); err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}

	return err
}
