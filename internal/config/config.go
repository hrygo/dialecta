package config

import (
	"os"

	"github.com/huangzhonghui/dialecta/internal/llm"
)

// RoleConfig holds configuration for a single debate role
type RoleConfig struct {
	Provider    llm.Provider
	Model       string
	Temperature float64
	MaxTokens   int
}

// Config holds the application configuration
type Config struct {
	ProRole   RoleConfig // 正方配置
	ConRole   RoleConfig // 反方配置
	JudgeRole RoleConfig // 裁决方配置
}

// Default role configurations
var (
	DefaultProRole = RoleConfig{
		Provider:    llm.ProviderDeepSeek,
		Model:       "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   4096,
	}
	DefaultConRole = RoleConfig{
		Provider:    llm.ProviderDeepSeek,
		Model:       "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   4096,
	}
	DefaultJudgeRole = RoleConfig{
		Provider:    llm.ProviderGemini,
		Model:       "gemini-2.0-flash",
		Temperature: 0.1,
		MaxTokens:   8192,
	}
)

// New creates a new Config with defaults
func New() *Config {
	return &Config{
		ProRole:   DefaultProRole,
		ConRole:   DefaultConRole,
		JudgeRole: DefaultJudgeRole,
	}
}

// Validate checks if required API keys are set
func (c *Config) Validate() error {
	providers := map[llm.Provider]bool{
		c.ProRole.Provider:   true,
		c.ConRole.Provider:   true,
		c.JudgeRole.Provider: true,
	}

	for p := range providers {
		switch p {
		case llm.ProviderDeepSeek:
			if os.Getenv("DEEPSEEK_API_KEY") == "" {
				return ConfigError("DEEPSEEK_API_KEY environment variable is required")
			}
		case llm.ProviderGemini:
			if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("GOOGLE_API_KEY") == "" {
				return ConfigError("GEMINI_API_KEY or GOOGLE_API_KEY environment variable is required")
			}
		case llm.ProviderDashScope:
			if os.Getenv("DASHSCOPE_API_KEY") == "" {
				return ConfigError("DASHSCOPE_API_KEY environment variable is required")
			}
		}
	}
	return nil
}

// ToLLMConfig converts RoleConfig to llm.Config
func (r *RoleConfig) ToLLMConfig() llm.Config {
	return llm.Config{
		Provider:    r.Provider,
		Model:       r.Model,
		Temperature: r.Temperature,
		MaxTokens:   r.MaxTokens,
	}
}

// Custom errors
type ConfigError string

func (e ConfigError) Error() string { return string(e) }
