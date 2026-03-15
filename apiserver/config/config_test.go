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
entra:
  enabled: true
  tenant_id: test-tenant
  client_id: test-client
  audience: api://test-client
scheduler_jobs:
  due_frequency: 5m
  overdue_frequency: 24h
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

	assert.Equal(t, true, cfg.Entra.Enabled)
	assert.Equal(t, "test-tenant", cfg.Entra.TenantID)
	assert.Equal(t, "test-client", cfg.Entra.ClientID)
	assert.Equal(t, "api://test-client", cfg.Entra.Audience)

	assert.Equal(t, 5*time.Minute, cfg.SchedulerJobs.DueFrequency)
	assert.Equal(t, 24*time.Hour, cfg.SchedulerJobs.OverdueFrequency)
}

func TestLoadConfig_PanicOnMissingFile(t *testing.T) {
	// Temporarily rename the real config.yaml so viper can't find any config
	renamed := false
	if _, err := os.Stat("./config.yaml"); err == nil {
		err = os.Rename("./config.yaml", "./config.yaml.bak")
		assert.NoError(t, err)
		renamed = true
	}

	_ = os.Remove("./config/config.yaml")
	viper.Reset()

	defer func() {
		if renamed {
			_ = os.Rename("./config.yaml.bak", "./config.yaml")
		}
	}()

	assert.Panics(t, func() {
		_ = LoadConfig("")
	})
}

func TestLoadConfig_EntraEnvOverride(t *testing.T) {
	_ = os.MkdirAll("./config", 0755)
	f, err := os.Create("./config/config.yaml")
	assert.NoError(t, err)
	defer os.Remove("./config/config.yaml")
	defer f.Close()

	_, err = f.WriteString(`entra:
  enabled: false
  tenant_id: file-tenant
  client_id: file-client
  audience: api://file-client
server:
  port: 1234
`)
	assert.NoError(t, err)

	os.Setenv("TW_ENTRA_ENABLED", "true")
	os.Setenv("TW_ENTRA_TENANT_ID", "env-tenant")
	os.Setenv("TW_ENTRA_CLIENT_ID", "env-client")
	os.Setenv("TW_ENTRA_AUDIENCE", "api://env-client")

	viper.Reset()
	cfg := LoadConfig("./config/config.yaml")

	assert.Equal(t, true, cfg.Entra.Enabled)
	assert.Equal(t, "env-tenant", cfg.Entra.TenantID)
	assert.Equal(t, "env-client", cfg.Entra.ClientID)
	assert.Equal(t, "api://env-client", cfg.Entra.Audience)

	os.Unsetenv("TW_ENTRA_ENABLED")
	os.Unsetenv("TW_ENTRA_TENANT_ID")
	os.Unsetenv("TW_ENTRA_CLIENT_ID")
	os.Unsetenv("TW_ENTRA_AUDIENCE")
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
