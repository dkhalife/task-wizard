package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Success(t *testing.T) {
	_ = os.MkdirAll("./config", 0755)
	f, err := os.Create("./config/config.yaml")
	assert.NoError(t, err, "failed to create config.yaml")

	defer os.Remove("./config/config.yaml")
	defer f.Close()
	_, err = f.WriteString(`server:
  port: 1234
  log_level: debug
  registration: true
  serve_frontend: false
  allowed_origins:
    - "http://example.com"
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

	assert.NoError(t, err)

	viper.Reset()
	cfg := LoadConfig()
	assert.Equal(t, 1234, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.LogLevel)
	assert.Equal(t, true, cfg.Server.Registration)
	assert.Equal(t, false, cfg.Server.ServeFrontend)
	assert.Equal(t, []string{"http://example.com"}, cfg.Server.AllowedOrigins)
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
	_ = os.Remove("./config/config.yaml")
	viper.Reset()

	assert.Panics(t, func() {
		_ = LoadConfig()
	})
}

func TestLoadConfig_EmailEnvOverride(t *testing.T) {
	_ = os.MkdirAll("./config", 0755)
	f, err := os.Create("./config/config.yaml")
	assert.NoError(t, err)
	defer os.Remove("./config/config.yaml")
	defer f.Close()

	_, err = f.WriteString(`email:
  host: smtp.example.com
  port: 25
  email: test@example.com
  password: testpass
server:
  port: 1234
jwt:
  secret: secret
`)
	assert.NoError(t, err)

	os.Setenv("TW_EMAIL_HOST", "smtp.override.com")
	os.Setenv("TW_EMAIL_PORT", "2525")
	os.Setenv("TW_EMAIL_SENDER", "override@example.com")
	os.Setenv("TW_EMAIL_PASSWORD", "overridepass")
	os.Setenv("TW_JWT_SECRET", "s3cret")

	viper.Reset()
	cfg := LoadConfig()

	assert.Equal(t, "smtp.override.com", cfg.EmailConfig.Host)
	assert.Equal(t, 2525, cfg.EmailConfig.Port)
	assert.Equal(t, "override@example.com", cfg.EmailConfig.Email)
	assert.Equal(t, "overridepass", cfg.EmailConfig.Password)

	os.Unsetenv("TW_EMAIL_HOST")
	os.Unsetenv("TW_EMAIL_PORT")
	os.Unsetenv("TW_EMAIL_SENDER")
	os.Unsetenv("TW_EMAIL_PASSWORD")
	os.Unsetenv("TW_JWT_SECRET")
}
