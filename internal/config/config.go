package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	WebAuthn WebAuthnConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type CORSConfig struct {
	AllowedOrigins []string
}

type JWTConfig struct {
	Secret string
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	URL string
}

type WebAuthnConfig struct {
	RPDisplayName string
	RPID          string
	RPOrigins     []string
	Timeout       time.Duration
}

var cfg *Config

func Load() *Config {
	if cfg != nil {
		return cfg
	}

	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/wechat-task/")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("failed to read config file: %w", err))
		}
	}

	v.SetEnvPrefix("WECHAT_TASK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("webauthn.rp_display_name", "WEBAUTHN_RP_DISPLAY_NAME")
	v.BindEnv("webauthn.rp_id", "WEBAUTHN_RP_ID")
	v.BindEnv("webauthn.rp_origins", "WEBAUTHN_RP_ORIGINS")
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.mode", "GIN_MODE")
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")

	cfg = &Config{}

	if err := v.Unmarshal(cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %w", err))
	}

	// Viper Unmarshal doesn't populate values from environment variables.
	// Explicitly read bound env vars to override struct fields.
	if val := v.GetString("database.url"); val != "" {
		cfg.Database.URL = val
	}
	if val := v.GetString("webauthn.rp_display_name"); val != "" {
		cfg.WebAuthn.RPDisplayName = val
	}
	if val := v.GetString("webauthn.rp_id"); val != "" {
		cfg.WebAuthn.RPID = val
	}
	if val := v.GetStringSlice("webauthn.rp_origins"); len(val) > 0 {
		cfg.WebAuthn.RPOrigins = val
	}
	if val := v.GetInt("server.port"); val != 0 {
		cfg.Server.Port = val
	}
	if val := v.GetString("server.mode"); val != "" {
		cfg.Server.Mode = val
	}
	if val := v.GetString("jwt.secret"); val != "" {
		cfg.JWT.Secret = val
	}
	if val := v.GetStringSlice("cors.allowed_origins"); len(val) > 0 {
		cfg.CORS.AllowedOrigins = val
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}

	if cfg.Database.URL == "" {
		cfg.Database.URL = "postgres://postgres:postgres@localhost:5432/wechat_task?sslmode=disable"
	}

	if cfg.WebAuthn.RPDisplayName == "" {
		cfg.WebAuthn.RPDisplayName = "WeChat Task"
	}

	if cfg.WebAuthn.RPID == "" {
		cfg.WebAuthn.RPID = "localhost"
	}

	if len(cfg.WebAuthn.RPOrigins) == 0 {
		cfg.WebAuthn.RPOrigins = []string{"http://localhost:8080"}
	}

	if cfg.WebAuthn.Timeout == 0 {
		cfg.WebAuthn.Timeout = 5 * time.Minute
	}

	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "change-this-secret-in-production"
	}

	return cfg
}

func Get() *Config {
	if cfg == nil {
		return Load()
	}
	return cfg
}
