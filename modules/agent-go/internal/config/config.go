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

	// LLM Providers
	OpenAIAPIKey    string
	AnthropicAPIKey string
	GeminiAPIKey    string
	GeminiModel     string

	// AWS S3
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3Bucket           string

	// Observability
	LangfusePublicKey string
	LangfuseSecretKey string
	LangfuseHost      string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		Env:                getEnv("ENV", "development"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
		AnthropicAPIKey:    getEnv("ANTHROPIC_API_KEY", ""),
		GeminiAPIKey:       getEnv("GEMINI_API_KEY", ""),
		GeminiModel:        getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3Bucket:           getEnv("S3_BUCKET", ""),
		LangfusePublicKey:  getEnv("LANGFUSE_PUBLIC_KEY", ""),
		LangfuseSecretKey:  getEnv("LANGFUSE_SECRET_KEY", ""),
		LangfuseHost:       getEnv("LANGFUSE_HOST", "https://cloud.langfuse.com"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
