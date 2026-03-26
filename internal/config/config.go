package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL           string
	WebAuthnRPDisplayName string
	WebAuthnRPID          string
	WebAuthnRPOrigins     []string
	WebAuthnTimeout       time.Duration
}

func Load() *Config {
	return &Config{
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/wechat_task?sslmode=disable"),
		WebAuthnRPDisplayName: getEnv("WEBAUTHN_RP_DISPLAY_NAME", "WeChat Task"),
		WebAuthnRPID:          getEnv("WEBAUTHN_RP_ID", "localhost"),
		WebAuthnRPOrigins:     []string{getEnv("WEBAUTHN_RP_ORIGIN", "http://localhost:8080")},
		WebAuthnTimeout:       5 * time.Minute,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
