package email

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/services/logging"
	"github.com/wneessen/go-mail"
)

type IEmailSender interface {
	SendResetPasswordEmail(ctx context.Context, to string, code string) error
	SendWelcomeEmail(ctx context.Context, name string, to string, activationCode string)
	SendTokenExpirationReminder(ctx context.Context, tokenName string, to string) error
}

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

func (es *EmailSender) sendEmail(to string, subject string, body string) error {
	message := mail.NewMsg()
	if err := message.From(es.Email); err != nil {
		return fmt.Errorf("failed to set From address: %s", err.Error())
	}

	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %s", err.Error())
	}

	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, body)

	client, err := mail.NewClient(es.Host,
		mail.WithPort(es.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(es.Email), mail.WithPassword(es.Password))

	if err != nil {
		return fmt.Errorf("failed to create mail client: %s", err.Error())
	}

	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %s", err.Error())
	}

	return nil
}

func (es *EmailSender) SendResetPasswordEmail(c context.Context, to string, code string) error {
	err := es.validateConfig()
	if err != nil {
		return err
	}

	resetURL := es.AppHost + "/password/update?c=" + url.QueryEscape(code)
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

	err = es.sendEmail(to, "Task Wizard - Password Reset", htmlBody)
	if err != nil {
		return err
	}

	return nil
}

func (es *EmailSender) SendWelcomeEmail(c context.Context, name string, to string, activationCode string) {
	log := logging.FromContext(c)
	err := es.validateConfig()
	if err != nil {
		log.Errorf("failed to validate email config: %s", err.Error())
		return
	}

	activationURL := es.AppHost + "/activate?code=" + url.QueryEscape(activationCode)
	htmlBody := `
		<html>
		<body>
			<p>Dear ` + name + `,</p>
			<p>Welcome to Task Wizard! Please use the link below to activate your account:</p>
			<p><a href="` + activationURL + `">Activate Account</a></p>
			<p>If you did not sign up for this account, please ignore this email.</p>
			<p>Thank you,</p>
			<p><strong>Task Wizard</strong></p>
		</body>
		</html>
	`

	err = es.sendEmail(to, "Task Wizard - Welcome!", htmlBody)
	if err != nil {
		log.Errorf("failed to send welcome email: %s", err.Error())
		return
	}
}

func (es *EmailSender) SendTokenExpirationReminder(c context.Context, tokenName string, to string) error {
	err := es.validateConfig()
	if err != nil {
		return err
	}

	htmlBody := `
		<html>
		<body>
			<p>Dear user,</p>
			<p>Your Task Wizard access token '` + tokenName + `' is about to expire. Please log in to the application to generate a new token.</p>
			<p>If you did not request a new token, please ignore this email.</p>
			<p>Thank you,</p>
			<p><strong>Task Wizard</strong></p>
		</body>
		</html>
	`

	err = es.sendEmail(to, "Task Wizard - Token Expiration Reminder", htmlBody)
	if err != nil {
		return err
	}

	return nil
}
