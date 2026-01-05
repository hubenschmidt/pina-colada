package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port      string
	Env       string
	PublicURL string // Base URL for generating absolute URLs (e.g., short URLs)

	// Database
	DatabaseURL string
	DBDebug     bool

	// LLM Providers
	OpenAIAPIKey    string
	OpenAIModel     string
	AnthropicAPIKey string

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
		Port:              port,
		Env:               getEnv("ENV", "development"),
		PublicURL:         getEnv("PUBLIC_URL", "http://localhost:"+port),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		DBDebug:           getEnv("DB_DEBUG", "false") == "true",
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:       getEnv("OPENAI_MODEL", "gpt-5.2"),
		AnthropicAPIKey:   getEnv("ANTHROPIC_API_KEY", ""),
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
