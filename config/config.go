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
	FilePath  string `mapstructure:"path" yaml:"path" default:"/config/task-wizard.db"`
	Migration bool   `mapstructure:"migration" yaml:"migration"`
}

type JwtConfig struct {
	Secret      string        `mapstructure:"secret" yaml:"secret"`
	SessionTime time.Duration `mapstructure:"session_time" yaml:"session_time"`
	MaxRefresh  time.Duration `mapstructure:"max_refresh" yaml:"max_refresh"`
}

type ServerConfig struct {
	Port          int           `mapstructure:"port" yaml:"port"`
	RatePeriod    time.Duration `mapstructure:"rate_period" yaml:"rate_period"`
	RateLimit     int           `mapstructure:"rate_limit" yaml:"rate_limit"`
	ReadTimeout   time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	ServeFrontend bool          `mapstructure:"serve_frontend" yaml:"serve_frontend"`
	Debug         bool          `mapstructure:"debug" yaml:"debug"`
}

type SchedulerConfig struct {
	DueFrequency     time.Duration `mapstructure:"due_frequency" yaml:"frequency" default:"5m"`
	OverdueFrequency time.Duration `mapstructure:"overdue_frequency" yaml:"overdue_frequency" default:"1d"`
}

type EmailConfig struct {
	Email   string `mapstructure:"email"`
	Key     string `mapstructure:"key"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	AppHost string `mapstructure:"appHost"`
}

func LoadConfig() *Config {
	viper.SetConfigName(os.Getenv("TW_ENV"))
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
