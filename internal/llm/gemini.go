package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GeminiClient implements the Client interface for Google Gemini
type GeminiClient struct {
	apiKey string
	cfg    Config
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(apiKey string, cfg Config) *GeminiClient {
	if cfg.Model == "" {
		cfg.Model = "gemini-3-pro-preview"
	}
	return &GeminiClient{
		apiKey: apiKey,
		cfg:    cfg,
	}
}

func (c *GeminiClient) Chat(ctx context.Context, messages []Message) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return "", fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(c.cfg.Model)
	model.SetTemperature(float32(c.cfg.Temperature))
	if c.cfg.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(c.cfg.MaxTokens))
	}

	// Build chat history
	cs := model.StartChat()
	for i := 0; i < len(messages)-1; i++ {
		msg := messages[i]
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		cs.History = append(cs.History, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	// Send the last message
	lastMsg := messages[len(messages)-1]
	resp, err := cs.SendMessage(ctx, genai.Text(lastMsg.Content))
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	return extractGeminiText(resp), nil
}

func (c *GeminiClient) ChatStream(ctx context.Context, messages []Message, onChunk func(string)) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return "", fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(c.cfg.Model)
	model.SetTemperature(float32(c.cfg.Temperature))
	if c.cfg.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(c.cfg.MaxTokens))
	}

	// Build chat history
	cs := model.StartChat()
	for i := 0; i < len(messages)-1; i++ {
		msg := messages[i]
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		cs.History = append(cs.History, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	// Send the last message with streaming
	lastMsg := messages[len(messages)-1]
	iter := cs.SendMessageStream(ctx, genai.Text(lastMsg.Content))

	var fullContent strings.Builder
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fullContent.String(), fmt.Errorf("stream error: %w", err)
		}

		text := extractGeminiText(resp)
		fullContent.WriteString(text)
		if onChunk != nil {
			onChunk(text)
		}
	}

	return fullContent.String(), nil
}

func extractGeminiText(resp *genai.GenerateContentResponse) string {
	var result strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					result.WriteString(string(text))
				}
			}
		}
	}
	return result.String()
}
