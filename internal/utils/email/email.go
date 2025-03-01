package email

import (
	"context"
	"errors"
	"fmt"

	"dkhalife.com/tasks/core/config"
	"gopkg.in/gomail.v2"
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
			<p>Dear User,</p>
			<p>Please use the link below to reset your password:</p>
			<p><a href="` + resetURL + `">Reset Password</a></p>
			<p>If you did not request a password reset, please ignore this email.</p>
			<p>Thank you,</p>
			<p><strong>Task Wizard</strong></p>
		</body>
		</html>
	`

	msg := gomail.NewMessage()
	msg.SetHeader("From", es.Email)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Task Wizard - Password Reset")
	msg.SetBody("text/html", htmlBody)

	fmt.Printf("Sending password reset email to %s\n", to)

	dialer := gomail.NewDialer(es.Host, es.Port, es.Email, es.Password)
	if err := dialer.DialAndSend(); err != nil {
		fmt.Println("Failed to send email:", err)
		return err
	}
	return nil
}
