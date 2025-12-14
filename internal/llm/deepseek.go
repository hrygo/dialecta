package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const deepseekBaseURL = "https://api.deepseek.com/v1"

// DeepSeekClient implements the Client interface for DeepSeek
type DeepSeekClient struct {
	apiKey string
	cfg    Config
	http   *http.Client
}

// NewDeepSeekClient creates a new DeepSeek client
func NewDeepSeekClient(apiKey string, cfg Config) *DeepSeekClient {
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	return &DeepSeekClient{
		apiKey: apiKey,
		cfg:    cfg,
		http:   &http.Client{},
	}
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type openAIStreamDelta struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (c *DeepSeekClient) Chat(ctx context.Context, messages []Message) (string, error) {
	return c.chat(ctx, messages, false, nil)
}

func (c *DeepSeekClient) ChatStream(ctx context.Context, messages []Message, onChunk func(string)) (string, error) {
	return c.chat(ctx, messages, true, onChunk)
}

func (c *DeepSeekClient) chat(ctx context.Context, messages []Message, stream bool, onChunk func(string)) (string, error) {
	// Convert messages
	oaiMessages := make([]openAIMessage, len(messages))
	for i, m := range messages {
		oaiMessages[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}

	req := openAIRequest{
		Model:       c.cfg.Model,
		Messages:    oaiMessages,
		Temperature: c.cfg.Temperature,
		MaxTokens:   c.cfg.MaxTokens,
		Stream:      stream,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", deepseekBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if stream {
		return c.handleStream(resp.Body, onChunk)
	}

	var oaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return oaiResp.Choices[0].Message.Content, nil
}

func (c *DeepSeekClient) handleStream(body io.Reader, onChunk func(string)) (string, error) {
	var fullContent strings.Builder
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var delta openAIStreamDelta
		if err := json.Unmarshal([]byte(data), &delta); err != nil {
			continue
		}

		if len(delta.Choices) > 0 {
			content := delta.Choices[0].Delta.Content
			fullContent.WriteString(content)
			if onChunk != nil {
				onChunk(content)
			}
		}
	}

	return fullContent.String(), nil
}
