package config

import (
	"os"
)

// ModelConfig holds configuration for a single LLM model
type ModelConfig struct {
	Model       string
	Temperature float64
	MaxTokens   int
}

// Config holds the application configuration
type Config struct {
	APIKey     string
	BaseURL    string
	ProModel   ModelConfig // 正方模型
	ConModel   ModelConfig // 反方模型
	JudgeModel ModelConfig // 裁决方模型
}

// Default model configurations
var (
	DefaultProModel = ModelConfig{
		Model:       "deepseek/deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   4096,
	}
	DefaultConModel = ModelConfig{
		Model:       "deepseek/deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   4096,
	}
	DefaultJudgeModel = ModelConfig{
		Model:       "anthropic/claude-sonnet-4-20250514",
		Temperature: 0.1,
		MaxTokens:   8192,
	}
)

// New creates a new Config with defaults
func New() *Config {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	return &Config{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		ProModel:   DefaultProModel,
		ConModel:   DefaultConModel,
		JudgeModel: DefaultJudgeModel,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return ErrMissingAPIKey
	}
	return nil
}

// Custom errors
type ConfigError string

func (e ConfigError) Error() string { return string(e) }

const ErrMissingAPIKey = ConfigError("API key is required. Set OPENROUTER_API_KEY or OPENAI_API_KEY environment variable")
