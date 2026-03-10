package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database      DatabaseConfig  `mapstructure:"database" yaml:"database"`
	Entra         EntraConfig     `mapstructure:"entra" yaml:"entra"`
	AppTokens     AppTokensConfig `mapstructure:"app_tokens" yaml:"app_tokens"`
	Server        ServerConfig    `mapstructure:"server" yaml:"server"`
	SchedulerJobs SchedulerConfig `mapstructure:"scheduler_jobs" yaml:"scheduler_jobs"`
}

type DatabaseConfig struct {
	Type      string `mapstructure:"type" yaml:"type" default:"sqlite"`
	FilePath  string `mapstructure:"path" yaml:"path" default:"/config/task-wizard.db"`
	Host      string `mapstructure:"host" yaml:"host"`
	Port      int    `mapstructure:"port" yaml:"port" default:"3306"`
	Database  string `mapstructure:"database" yaml:"database"`
	Username  string `mapstructure:"username" yaml:"username"`
	Password  string `mapstructure:"password" yaml:"password"`
	Migration bool   `mapstructure:"migration" yaml:"migration"`
}

type EntraConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	TenantID string `mapstructure:"tenant_id" yaml:"tenant_id"`
	ClientID string `mapstructure:"client_id" yaml:"client_id"`
	Audience string `mapstructure:"audience" yaml:"audience"`
}

type AppTokensConfig struct {
	Secret string `mapstructure:"secret" yaml:"secret"`
}

type ServerConfig struct {
	HostName             string        `mapstructure:"host_name" yaml:"host_name"`
	Port                 int           `mapstructure:"port" yaml:"port"`
	RatePeriod           time.Duration `mapstructure:"rate_period" yaml:"rate_period"`
	RateLimit            int           `mapstructure:"rate_limit" yaml:"rate_limit"`
	ReadTimeout          time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout         time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	ServeFrontend        bool          `mapstructure:"serve_frontend" yaml:"serve_frontend"`
	Registration         bool          `mapstructure:"registration" yaml:"registration"`
	LogLevel             string        `mapstructure:"log_level" yaml:"log_level"`
	AllowedOrigins       []string      `mapstructure:"allowed_origins" yaml:"allowed_origins"`
	AllowCorsCredentials bool          `mapstructure:"allow_cors_credentials" yaml:"allow_cors_credentials"`
}

type SchedulerConfig struct {
	DueFrequency            time.Duration `mapstructure:"due_frequency" yaml:"due_frequency" default:"5m"`
	OverdueFrequency        time.Duration `mapstructure:"overdue_frequency" yaml:"overdue_frequency" default:"1d"`
	TokenExpirationReminder time.Duration `mapstructure:"token_expiration_reminder" yaml:"token_expiration_reminder" default:"72h"`
	NotificationCleanup     time.Duration `mapstructure:"notification_cleanup" yaml:"notification_cleanup" default:"10m"`
	TokenExpirationCleanup  time.Duration `mapstructure:"token_expiration_cleanup" yaml:"token_expiration_cleanup" default:"24h"`
}

func LoadConfig(configFile string) *Config {
	viper.SetConfigType("yaml")

	if configFile == "" {
		if envFile := os.Getenv("TW_CONFIG_FILE"); envFile != "" {
			configFile = envFile
		}
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	_ = viper.BindEnv("entra.enabled", "TW_ENTRA_ENABLED")
	_ = viper.BindEnv("entra.tenant_id", "TW_ENTRA_TENANT_ID")
	_ = viper.BindEnv("entra.client_id", "TW_ENTRA_CLIENT_ID")
	_ = viper.BindEnv("entra.audience", "TW_ENTRA_AUDIENCE")
	_ = viper.BindEnv("app_tokens.secret", "TW_APP_TOKEN_SECRET")
	_ = viper.BindEnv("database.type", "TW_DATABASE_TYPE")
	_ = viper.BindEnv("database.host", "TW_DATABASE_HOST")
	_ = viper.BindEnv("database.port", "TW_DATABASE_PORT")
	_ = viper.BindEnv("database.database", "TW_DATABASE_NAME")
	_ = viper.BindEnv("database.username", "TW_DATABASE_USERNAME")
	_ = viper.BindEnv("database.password", "TW_DATABASE_PASSWORD")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	if config.AppTokens.Secret == "secret" {
		panic("App token secret must be changed from the default 'secret'. Set TW_APP_TOKEN_SECRET or update config.yaml")
	}

	return &config
}
