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

const dashscopeBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"

// DashScopeClient implements the Client interface for Alibaba DashScope (Qwen)
type DashScopeClient struct {
	apiKey string
	cfg    Config
	http   *http.Client
}

// NewDashScopeClient creates a new DashScope client
func NewDashScopeClient(apiKey string, cfg Config) *DashScopeClient {
	if cfg.Model == "" {
		cfg.Model = "qwen-plus"
	}
	return &DashScopeClient{
		apiKey: apiKey,
		cfg:    cfg,
		http:   &http.Client{},
	}
}

func (c *DashScopeClient) Chat(ctx context.Context, messages []Message) (string, error) {
	return c.chat(ctx, messages, false, nil)
}

func (c *DashScopeClient) ChatStream(ctx context.Context, messages []Message, onChunk func(string)) (string, error) {
	return c.chat(ctx, messages, true, onChunk)
}

func (c *DashScopeClient) chat(ctx context.Context, messages []Message, stream bool, onChunk func(string)) (string, error) {
	// Convert messages (OpenAI compatible format)
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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", dashscopeBaseURL+"/chat/completions", bytes.NewReader(body))
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

func (c *DashScopeClient) handleStream(body io.Reader, onChunk func(string)) (string, error) {
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
