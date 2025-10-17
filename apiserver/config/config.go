package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database      DatabaseConfig  `mapstructure:"database" yaml:"database"`
	Jwt           JwtConfig       `mapstructure:"jwt" yaml:"jwt"`
	OAuth         OAuthConfig     `mapstructure:"oauth" yaml:"oauth"`
	Server        ServerConfig    `mapstructure:"server" yaml:"server"`
	SchedulerJobs SchedulerConfig `mapstructure:"scheduler_jobs" yaml:"scheduler_jobs"`
	EmailConfig   EmailConfig     `mapstructure:"email" yaml:"email"`
}

type DatabaseConfig struct {
	FilePath  string `mapstructure:"path" yaml:"path" default:"/config/task-wizard.db"`
	Migration bool   `mapstructure:"migration" yaml:"migration"`
}

type JwtConfig struct {
	Secret      string        `mapstructure:"secret" yaml:"secret"`
	SessionTime time.Duration `mapstructure:"session_time" yaml:"session_time"`
	MaxRefresh  time.Duration `mapstructure:"max_refresh" yaml:"max_refresh"`
}

type OAuthConfig struct {
	Enabled      bool   `mapstructure:"enabled" yaml:"enabled"`
	ClientID     string `mapstructure:"client_id" yaml:"client_id"`
	ClientSecret string `mapstructure:"client_secret" yaml:"client_secret"`
	TenantID     string `mapstructure:"tenant_id" yaml:"tenant_id"`
	AuthorizeURL string `mapstructure:"authorize_url" yaml:"authorize_url"`
	TokenURL     string `mapstructure:"token_url" yaml:"token_url"`
	RedirectURL  string `mapstructure:"redirect_url" yaml:"redirect_url"`
	Scope        string `mapstructure:"scope" yaml:"scope"`
	JwksURL      string `mapstructure:"jwks_url" yaml:"jwks_url"`
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
	PasswordResetValidity   time.Duration `mapstructure:"password_reset_validity" yaml:"password_reset_validity" default:"24h"`
	TokenExpirationReminder time.Duration `mapstructure:"token_expiration_reminder" yaml:"token_expiration_reminder" default:"72h"`
	NotificationCleanup     time.Duration `mapstructure:"notification_cleanup" yaml:"notification_cleanup" default:"10m"`
	TokenExpirationCleanup  time.Duration `mapstructure:"token_expiration_cleanup" yaml:"token_expiration_cleanup" default:"24h"`
}

type EmailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
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

	// Allow values with secrets to be set via environment variables
	_ = viper.BindEnv("jwt.secret", "TW_JWT_SECRET")
	_ = viper.BindEnv("oauth.enabled", "TW_OAUTH_ENABLED")
	_ = viper.BindEnv("oauth.client_id", "TW_OAUTH_CLIENT_ID")
	_ = viper.BindEnv("oauth.client_secret", "TW_OAUTH_CLIENT_SECRET")
	_ = viper.BindEnv("oauth.tenant_id", "TW_OAUTH_TENANT_ID")
	_ = viper.BindEnv("oauth.authorize_url", "TW_OAUTH_AUTHORIZE_URL")
	_ = viper.BindEnv("oauth.token_url", "TW_OAUTH_TOKEN_URL")
	_ = viper.BindEnv("oauth.redirect_url", "TW_OAUTH_REDIRECT_URL")
	_ = viper.BindEnv("oauth.scope", "TW_OAUTH_SCOPE")
	_ = viper.BindEnv("oauth.jwks_url", "TW_OAUTH_JWKS_URL")
	_ = viper.BindEnv("email.host", "TW_EMAIL_HOST")
	_ = viper.BindEnv("email.port", "TW_EMAIL_PORT")
	_ = viper.BindEnv("email.email", "TW_EMAIL_SENDER")
	_ = viper.BindEnv("email.password", "TW_EMAIL_PASSWORD")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	if config.Jwt.Secret == "secret" {
		panic("JWT secret must be changed from the default 'secret'. Set TW_JWT_SECRET or update config.yaml")
	}

	return &config
}
