package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	Env                 string
	ExaAPIKey           string
	TinyfishAPIKey      string
	OpenAIAPIKey        string
	ExaBaseURL          string
	TinyfishBaseURL     string
	OpenAIBaseURL       string
	OpenAIModel         string
	HTTPTimeout         time.Duration
	CallbackTimeout     time.Duration
	MaxCandidates       int
	MaxRecommendations  int
	MaxSearchAttempts   int
	MaxDetailFetches    int
	OpenAIRetryAttempts int
	CallbackRetryCount  int
	CallbackRetryDelay  time.Duration
	FallbackEnabled     bool
	AllowOpenAIFill     bool
	DefaultReportFormat string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	return &Config{
		Port:                getEnv("PORT", "8895"),
		Env:                 getEnv("ENV", "development"),
		ExaAPIKey:           os.Getenv("EXA_API_KEY"),
		TinyfishAPIKey:      os.Getenv("TINYFISH_API_KEY"),
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
		ExaBaseURL:          getEnv("EXA_BASE_URL", "https://api.exa.ai"),
		TinyfishBaseURL:     getEnv("TINYFISH_BASE_URL", "https://agent.tinyfish.ai"),
		OpenAIBaseURL:       getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:         getEnv("OPENAI_MODEL", "gpt-4.1-mini"),
		HTTPTimeout:         getDurationSeconds("HTTP_TIMEOUT_SECONDS", 30),
		CallbackTimeout:     getDurationSeconds("CALLBACK_TIMEOUT_SECONDS", 15),
		MaxCandidates:       getInt("MAX_CANDIDATES", 24),
		MaxRecommendations:  getInt("MAX_RECOMMENDATIONS", 6),
		MaxSearchAttempts:   getInt("MAX_SEARCH_ATTEMPTS", 5),
		MaxDetailFetches:    getInt("MAX_DETAIL_FETCHES", 3),
		OpenAIRetryAttempts: getInt("OPENAI_RETRY_ATTEMPTS", 5),
		CallbackRetryCount:  getInt("CALLBACK_RETRY_COUNT", 3),
		CallbackRetryDelay:  getDurationMillis("CALLBACK_RETRY_DELAY_MS", 300),
		FallbackEnabled:     getBool("FALLBACK_ENABLED", true),
		AllowOpenAIFill:     getBool("ALLOW_OPENAI_FILL", true),
		DefaultReportFormat: getEnv("DEFAULT_REPORT_FORMAT", "markdown"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func getDurationSeconds(key string, fallback int) time.Duration {
	return time.Duration(getInt(key, fallback)) * time.Second
}

func getDurationMillis(key string, fallback int) time.Duration {
	return time.Duration(getInt(key, fallback)) * time.Millisecond
}
