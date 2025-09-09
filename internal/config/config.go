package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Upload   UploadConfig   `mapstructure:"upload"`
}

type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	Environment  string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

type RabbitMQConfig struct {
	URL        string `mapstructure:"url"`
	Host       string `mapstructure:"host"`
	Port       string `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	VHost      string `mapstructure:"vhost"`
	Exchange   string `mapstructure:"exchange"`
	Queue      string `mapstructure:"queue"`
	RoutingKey string `mapstructure:"routing_key"`
}

type JWTConfig struct {
	SecretKey       string `mapstructure:"secret_key"`
	AccessTokenTTL  int    `mapstructure:"access_token_ttl"`  // in minutes
	RefreshTokenTTL int    `mapstructure:"refresh_token_ttl"` // in hours
}

type UploadConfig struct {
	MaxFileSize  int64    `mapstructure:"max_file_size"` // in bytes
	AllowedTypes []string `mapstructure:"allowed_types"`
	StoragePath  string   `mapstructure:"storage_path"`
	BaseURL      string   `mapstructure:"base_url"`
	TempTTL      int      `mapstructure:"temp_ttl"` // in hours
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	TimeFormat string `mapstructure:"time_format"`
}

var AppConfig *Config

func LoadConfig(configPath string) (*Config, error) {
	// Determine config file name based on environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	if env == "" {
		env = "development"
	}

	configName := "config"
	if env == "production" {
		configName = "config.prod"
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Environment variable support
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Config file %s.yaml not found, using defaults and environment variables", configName)
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		log.Printf("Loading configuration from: %s", viper.ConfigFileUsed())
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	AppConfig = config
	return config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.environment", "development")

	// Database defaults
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.database", "realtime_db")
	viper.SetDefault("database.ssl_mode", "disable")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.database", 0)

	// RabbitMQ defaults
	viper.SetDefault("rabbitmq.host", "localhost")
	viper.SetDefault("rabbitmq.port", "5672")
	viper.SetDefault("rabbitmq.username", "guest")
	viper.SetDefault("rabbitmq.password", "guest")
	viper.SetDefault("rabbitmq.vhost", "/")
	viper.SetDefault("rabbitmq.exchange", "chat_exchange")
	viper.SetDefault("rabbitmq.queue", "chat_queue")
	viper.SetDefault("rabbitmq.routing_key", "chat")

	// JWT defaults
	viper.SetDefault("jwt.secret_key", "your-secret-key-change-this-in-production")
	viper.SetDefault("jwt.access_token_ttl", 15)   // 15 minutes
	viper.SetDefault("jwt.refresh_token_ttl", 168) // 7 days

	// Upload defaults
	viper.SetDefault("upload.max_file_size", 10485760) // 10MB
	viper.SetDefault("upload.allowed_types", []string{"image/jpeg", "image/png", "image/gif", "video/mp4", "audio/mpeg", "application/pdf"})
	viper.SetDefault("upload.storage_path", "./uploads")
	viper.SetDefault("upload.base_url", "http://localhost:8080/uploads")
	viper.SetDefault("upload.temp_ttl", 24) // 24 hours

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.time_format", "2006-01-02T15:04:05Z07:00")
}

func GetConfig() *Config {
	return AppConfig
}
