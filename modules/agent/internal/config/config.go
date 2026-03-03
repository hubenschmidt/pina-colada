package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DatabaseURL string
	DBDebug     bool

	// LLM Providers
	AnthropicAPIKey string
	VoyageAPIKey    string

	// AWS S3
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3Bucket           string

	// Observability
	LangfusePublicKey string
	LangfuseSecretKey string
	LangfuseHost      string

	// External APIs
	SerperAPIKey string

	// Pushover
	PushoverUser  string
	PushoverToken string

	// SMTP Email
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromEmail string
}

func Load() *Config {
	port := getEnv("PORT", "8000")
	return &Config{
		Port:        port,
		Env:         getEnv("ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBDebug:           getEnv("DB_DEBUG", "false") == "true",
		AnthropicAPIKey:   getEnv("ANTHROPIC_API_KEY", ""),
		VoyageAPIKey:      getEnv("VOYAGE_API_KEY", ""),
		LangfusePublicKey: getEnv("LANGFUSE_PUBLIC_KEY", ""),
		LangfuseSecretKey: getEnv("LANGFUSE_SECRET_KEY", ""),
		LangfuseHost:      getEnv("LANGFUSE_HOST", "https://cloud.langfuse.com"),
		SerperAPIKey:      getEnv("SERPER_API_KEY", ""),
		PushoverUser:      getEnv("PUSHOVER_USER", ""),
		PushoverToken:     getEnv("PUSHOVER_TOKEN", ""),
		SMTPHost:          getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:          getEnvInt("SMTP_PORT", 587),
		SMTPUsername:      getEnv("SMTP_USERNAME", ""),
		SMTPPassword:      getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:     getEnv("SMTP_FROM_EMAIL", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}
