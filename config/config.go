package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Name          string          `mapstructure:"name" yaml:"name"`
	Database      DatabaseConfig  `mapstructure:"database" yaml:"database"`
	Jwt           JwtConfig       `mapstructure:"jwt" yaml:"jwt"`
	Server        ServerConfig    `mapstructure:"server" yaml:"server"`
	SchedulerJobs SchedulerConfig `mapstructure:"scheduler_jobs" yaml:"scheduler_jobs"`
	EmailConfig   EmailConfig     `mapstructure:"email" yaml:"email"`
}

type DatabaseConfig struct {
	Host      string `mapstructure:"host" yaml:"host"`
	Port      int    `mapstructure:"port" yaml:"port"`
	User      string `mapstructure:"user" yaml:"user"`
	Password  string `mapstructure:"password" yaml:"password"`
	Name      string `mapstructure:"name" yaml:"name"`
	Migration bool   `mapstructure:"migration" yaml:"migration"`
	LogLevel  int    `mapstructure:"logger" yaml:"logger"`
}

type JwtConfig struct {
	Secret      string        `mapstructure:"secret" yaml:"secret"`
	SessionTime time.Duration `mapstructure:"session_time" yaml:"session_time"`
	MaxRefresh  time.Duration `mapstructure:"max_refresh" yaml:"max_refresh"`
}

type ServerConfig struct {
	Port             int           `mapstructure:"port" yaml:"port"`
	RatePeriod       time.Duration `mapstructure:"rate_period" yaml:"rate_period"`
	RateLimit        int           `mapstructure:"rate_limit" yaml:"rate_limit"`
	ReadTimeout      time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout     time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	CorsAllowOrigins []string      `mapstructure:"cors_allow_origins" yaml:"cors_allow_origins"`
	ServeFrontend    bool          `mapstructure:"serve_frontend" yaml:"serve_frontend"`
}

type SchedulerConfig struct {
	DueJob     time.Duration `mapstructure:"due_job" yaml:"due_job"`
	OverdueJob time.Duration `mapstructure:"overdue_job" yaml:"overdue_job"`
	PreDueJob  time.Duration `mapstructure:"pre_due_job" yaml:"pre_due_job"`
}

type EmailConfig struct {
	Email   string `mapstructure:"email"`
	Key     string `mapstructure:"key"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	AppHost string `mapstructure:"appHost"`
}

func LoadConfig() *Config {
	if os.Getenv("DT_ENV") == "debug" {
		viper.SetConfigName("debug")
	} else {
		viper.SetConfigName("prod")
	}

	viper.AddConfigPath("./config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()

	if err != nil {
		panic(err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	return &config
}
