package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Success(t *testing.T) {
	os.MkdirAll("./config", 0755)
	f, err := os.Create("./config/config.yaml")
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}
	defer os.Remove("./config/config.yaml")
	defer f.Close()
	f.WriteString(`server:
  port: 1234
  log_level: debug
  registration: true
  serve_frontend: false
  read_timeout: 10s
  write_timeout: 10s
database:
  path: test.db
  migration: true
jwt:
  secret: testsecret
  session_time: 1h
  max_refresh: 1h
scheduler_jobs:
  due_frequency: 5m
  overdue_frequency: 24h
  password_reset_validity: 24h
  token_expiration_reminder: 72h
email:
  host: smtp.example.com
  port: 587
  email: test@example.com
  password: testpass
`)

	viper.Reset()
	cfg := LoadConfig()
	assert.Equal(t, 1234, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.LogLevel)
	assert.Equal(t, true, cfg.Server.Registration)
	assert.Equal(t, false, cfg.Server.ServeFrontend)
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, cfg.Server.WriteTimeout)

	assert.Equal(t, "test.db", cfg.Database.FilePath)
	assert.Equal(t, true, cfg.Database.Migration)

	assert.Equal(t, "testsecret", cfg.Jwt.Secret)
	assert.Equal(t, 1*time.Hour, cfg.Jwt.SessionTime)
	assert.Equal(t, 1*time.Hour, cfg.Jwt.MaxRefresh)

	assert.Equal(t, 5*time.Minute, cfg.SchedulerJobs.DueFrequency)
	assert.Equal(t, 24*time.Hour, cfg.SchedulerJobs.OverdueFrequency)
	assert.Equal(t, 24*time.Hour, cfg.SchedulerJobs.PasswordResetValidity)
	assert.Equal(t, 72*time.Hour, cfg.SchedulerJobs.TokenExpirationReminder)

	assert.Equal(t, "smtp.example.com", cfg.EmailConfig.Host)
	assert.Equal(t, 587, cfg.EmailConfig.Port)
	assert.Equal(t, "test@example.com", cfg.EmailConfig.Email)
	assert.Equal(t, "testpass", cfg.EmailConfig.Password)
}

func TestLoadConfig_PanicOnMissingFile(t *testing.T) {
	os.Remove("./config/config.yaml")
	viper.Reset()
	defer func() {
		recover() // expect panic
	}()
	LoadConfig()
}
