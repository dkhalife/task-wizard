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
  read_timeout: 10s
  write_timeout: 10s
database:
  type: sqlite
  path: test.db
  migration: true
jwt:
  secret: testsecret
entra:
  tenant_id: your-tenant-id
  audience: your-audience
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
	cfg := LoadConfig("./config/config.yaml")
	assert.Equal(t, 1234, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.LogLevel)
	assert.Equal(t, true, cfg.Server.Registration)
	assert.Equal(t, false, cfg.Server.ServeFrontend)
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, cfg.Server.WriteTimeout)

	assert.Equal(t, "sqlite", cfg.Database.Type)
	assert.Equal(t, "test.db", cfg.Database.FilePath)
	assert.Equal(t, true, cfg.Database.Migration)

	assert.Equal(t, "testsecret", cfg.Jwt.Secret)

	assert.Equal(t, "your-tenant-id", cfg.Entra.TenantID)
	assert.Equal(t, "your-audience", cfg.Entra.Audience)

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
		_ = LoadConfig("")
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
`)
	assert.NoError(t, err)

	os.Setenv("TW_EMAIL_HOST", "smtp.override.com")
	os.Setenv("TW_EMAIL_PORT", "2525")
	os.Setenv("TW_EMAIL_SENDER", "override@example.com")
	os.Setenv("TW_EMAIL_PASSWORD", "overridepass")

	viper.Reset()
	cfg := LoadConfig("./config/config.yaml")

	assert.Equal(t, "smtp.override.com", cfg.EmailConfig.Host)
	assert.Equal(t, 2525, cfg.EmailConfig.Port)
	assert.Equal(t, "override@example.com", cfg.EmailConfig.Email)
	assert.Equal(t, "overridepass", cfg.EmailConfig.Password)

	os.Unsetenv("TW_EMAIL_HOST")
	os.Unsetenv("TW_EMAIL_PORT")
	os.Unsetenv("TW_EMAIL_SENDER")
	os.Unsetenv("TW_EMAIL_PASSWORD")
}

func TestLoadConfig_EnvFile(t *testing.T) {
	f, err := os.Create("envconfig.yaml")
	assert.NoError(t, err)
	defer os.Remove("envconfig.yaml")
	defer f.Close()

	_, err = f.WriteString("server:\n  port: 4444\n")
	assert.NoError(t, err)

	os.Setenv("TW_CONFIG_FILE", "envconfig.yaml")
	viper.Reset()
	cfg := LoadConfig("")
	assert.Equal(t, 4444, cfg.Server.Port)
	os.Unsetenv("TW_CONFIG_FILE")
}

func TestLoadConfig_CLIOverridesEnv(t *testing.T) {
	f1, err := os.Create("env.yaml")
	assert.NoError(t, err)
	defer os.Remove("env.yaml")
	defer f1.Close()
	_, err = f1.WriteString("server:\n  port: 3333\n")
	assert.NoError(t, err)

	f2, err := os.Create("cli.yaml")
	assert.NoError(t, err)
	defer os.Remove("cli.yaml")
	defer f2.Close()
	_, err = f2.WriteString("server:\n  port: 2222\n")
	assert.NoError(t, err)

	os.Setenv("TW_CONFIG_FILE", "env.yaml")
	viper.Reset()
	cfg := LoadConfig("cli.yaml")
	assert.Equal(t, 2222, cfg.Server.Port)
	os.Unsetenv("TW_CONFIG_FILE")
}

func TestLoadConfig_DatabaseEnvOverride(t *testing.T) {
	_ = os.MkdirAll("./config", 0755)
	f, err := os.Create("./config/config.yaml")
	assert.NoError(t, err)
	defer os.Remove("./config/config.yaml")
	defer f.Close()

	_, err = f.WriteString(`database:
  type: sqlite
  path: /config/task-wizard.db
server:
  port: 1234
`)
	assert.NoError(t, err)

	os.Setenv("TW_DATABASE_TYPE", "mysql")
	os.Setenv("TW_DATABASE_HOST", "localhost")
	os.Setenv("TW_DATABASE_PORT", "3307")
	os.Setenv("TW_DATABASE_NAME", "taskwizard")
	os.Setenv("TW_DATABASE_USERNAME", "dbuser")
	os.Setenv("TW_DATABASE_PASSWORD", "dbpass")

	viper.Reset()
	cfg := LoadConfig("./config/config.yaml")

	assert.Equal(t, "mysql", cfg.Database.Type)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 3307, cfg.Database.Port)
	assert.Equal(t, "taskwizard", cfg.Database.Database)
	assert.Equal(t, "dbuser", cfg.Database.Username)
	assert.Equal(t, "dbpass", cfg.Database.Password)

	os.Unsetenv("TW_DATABASE_TYPE")
	os.Unsetenv("TW_DATABASE_HOST")
	os.Unsetenv("TW_DATABASE_PORT")
	os.Unsetenv("TW_DATABASE_NAME")
	os.Unsetenv("TW_DATABASE_USERNAME")
	os.Unsetenv("TW_DATABASE_PASSWORD")
}

func TestLoadConfig_MySQLConfig(t *testing.T) {
	_ = os.MkdirAll("./config", 0755)
	f, err := os.Create("./config/config.yaml")
	assert.NoError(t, err)
	defer os.Remove("./config/config.yaml")
	defer f.Close()

	_, err = f.WriteString(`database:
  type: mysql
  host: mysql.example.com
  port: 3306
  database: taskwizard
  username: testuser
  password: testpass
  migration: true
server:
  port: 1234
`)
	assert.NoError(t, err)

	viper.Reset()
	cfg := LoadConfig("./config/config.yaml")

	assert.Equal(t, "mysql", cfg.Database.Type)
	assert.Equal(t, "mysql.example.com", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.Equal(t, "taskwizard", cfg.Database.Database)
	assert.Equal(t, "testuser", cfg.Database.Username)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, true, cfg.Database.Migration)
}
