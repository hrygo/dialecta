package llm

import (
	"context"
	"fmt"
	"os"
)

// Provider represents an LLM provider
type Provider string

const (
	ProviderDeepSeek  Provider = "deepseek"
	ProviderGemini    Provider = "gemini"
	ProviderDashScope Provider = "dashscope"
)

// Message represents a chat message
type Message struct {
	Role    string
	Content string
}

// Config holds the configuration for an LLM client
type Config struct {
	Provider    Provider
	Model       string
	Temperature float64
	MaxTokens   int
}

// Client is the interface for LLM clients
type Client interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	ChatStream(ctx context.Context, messages []Message, onChunk func(string)) (string, error)
}

// NewClient creates a new LLM client based on the provider
func NewClient(cfg Config) (Client, error) {
	switch cfg.Provider {
	case ProviderDeepSeek:
		apiKey := os.Getenv("DEEPSEEK_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required")
		}
		return NewDeepSeekClient(apiKey, cfg), nil

	case ProviderGemini:
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY or GOOGLE_API_KEY environment variable is required")
		}
		return NewGeminiClient(apiKey, cfg), nil

	case ProviderDashScope:
		apiKey := os.Getenv("DASHSCOPE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("DASHSCOPE_API_KEY environment variable is required")
		}
		return NewDashScopeClient(apiKey, cfg), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// ParseProvider parses a provider string
func ParseProvider(s string) (Provider, error) {
	switch s {
	case "deepseek":
		return ProviderDeepSeek, nil
	case "gemini", "google":
		return ProviderGemini, nil
	case "dashscope", "qwen", "alibaba":
		return ProviderDashScope, nil
	default:
		return "", fmt.Errorf("unknown provider: %s (supported: deepseek, gemini, dashscope)", s)
	}
}
