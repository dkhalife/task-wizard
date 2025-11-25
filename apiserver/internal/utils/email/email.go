package email

import (
	"context"
	"errors"
	"fmt"

	"dkhalife.com/tasks/core/config"
	"github.com/wneessen/go-mail"
)

type IEmailSender interface {
	SendTokenExpirationReminder(ctx context.Context, tokenName string, to string) error
}

type EmailSender struct {
	AppHost  string
	Host     string
	Port     int
	Email    string
	Password string
}

var _ IEmailSender = (*EmailSender)(nil)

func NewEmailSender(conf *config.Config) IEmailSender {
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
