package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Database DBConfig
	CORS     CORSConfig
	Log      LogConfig
}

type AppConfig struct {
	Name  string
	Env   string
	Port  string
	Debug bool
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
}

type CORSConfig struct {
	AllowOrigins     []string
	AllowCredentials bool
}

type LogConfig struct {
	Level    string
	Format   string
	FilePath string
}

// Load reads configuration from environment variables and files
func Load() *Config {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("APP_NAME", "TodoAI")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_MAX_LIFETIME", "5m")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")

	// Try to load .env file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	cfg := &Config{
		App: AppConfig{
			Name:  viper.GetString("APP_NAME"),
			Env:   viper.GetString("APP_ENV"),
			Port:  viper.GetString("APP_PORT"),
			Debug: viper.GetBool("APP_DEBUG"),
		},
		Database: DBConfig{
			Host:         viper.GetString("DB_HOST"),
			Port:         viper.GetString("DB_PORT"),
			User:         viper.GetString("DB_USER"),
			Password:     viper.GetString("DB_PASSWORD"),
			Name:         viper.GetString("DB_NAME"),
			SSLMode:      viper.GetString("DB_SSLMODE"),
			MaxOpenConns: viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: viper.GetInt("DB_MAX_IDLE_CONNS"),
			MaxConnLifetime: func() time.Duration {
				if lifetime := viper.GetString("DB_MAX_LIFETIME"); lifetime != "" {
					if duration, err := time.ParseDuration(lifetime); err == nil {
						return duration
					}
				}
				return 5 * time.Minute
			}(),
		},
		CORS: CORSConfig{
			AllowOrigins:     viper.GetStringSlice("CORS_ALLOW_ORIGINS"),
			AllowCredentials: viper.GetBool("CORS_ALLOW_CREDENTIALS"),
		},
		Log: LogConfig{
			Level:    viper.GetString("LOG_LEVEL"),
			Format:   viper.GetString("LOG_FORMAT"),
			FilePath: viper.GetString("LOG_FILE_PATH"),
		},
	}

	return cfg
}