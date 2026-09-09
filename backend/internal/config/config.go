package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Probe    ProbeConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	MaxConns int
	MinConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type ProbeConfig struct {
	ID       string
	Name     string
	Location string
	APIKey   string
	ServerURL string
	HeartbeatInterval time.Duration
}

type LoggingConfig struct {
	Level  string
	Format string
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Override with environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("NETMON")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.readtimeout", 30*time.Second)
	v.SetDefault("server.writetimeout", 30*time.Second)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "netmon")
	v.SetDefault("database.password", "")
	v.SetDefault("database.database", "netmon")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.maxconns", 25)
	v.SetDefault("database.minconns", 5)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Probe defaults
	v.SetDefault("probe.heartbeatinterval", 30*time.Second)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}
