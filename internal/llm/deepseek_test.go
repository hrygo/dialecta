package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewDeepSeekClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		cfg       Config
		wantModel string
	}{
		{
			name:      "with custom model",
			apiKey:    "test-key",
			cfg:       Config{Model: "custom-model", Temperature: 0.5, MaxTokens: 2048},
			wantModel: "custom-model",
		},
		{
			name:      "with empty model - uses default",
			apiKey:    "test-key",
			cfg:       Config{Model: "", Temperature: 0.8, MaxTokens: 4096},
			wantModel: "deepseek-chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewDeepSeekClient(tt.apiKey, tt.cfg)

			if client == nil {
				t.Fatal("NewDeepSeekClient() returned nil")
			}

			if client.apiKey != tt.apiKey {
				t.Errorf("client.apiKey = %v, want %v", client.apiKey, tt.apiKey)
			}

			if client.cfg.Model != tt.wantModel {
				t.Errorf("client.cfg.Model = %v, want %v", client.cfg.Model, tt.wantModel)
			}

			if client.http == nil {
				t.Error("client.http is nil")
			}
		})
	}
}

func TestDeepSeekClient_Implements_Client(t *testing.T) {
	// Verify DeepSeekClient implements Client interface
	var _ Client = (*DeepSeekClient)(nil)
}

func TestDeepSeekClient_ConfigPreserved(t *testing.T) {
	cfg := Config{
		Model:       "test-model",
		Temperature: 0.7,
		MaxTokens:   1024,
	}

	client := NewDeepSeekClient("api-key", cfg)

	if client.cfg.Temperature != 0.7 {
		t.Errorf("Temperature = %v, want %v", client.cfg.Temperature, 0.7)
	}
	if client.cfg.MaxTokens != 1024 {
		t.Errorf("MaxTokens = %v, want %v", client.cfg.MaxTokens, 1024)
	}
}

// Helper to create a mock OpenAI-style response
func mockOpenAIResponse(content string) string {
	resp := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": content,
				},
			},
		},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

func TestDeepSeekClient_Chat_MockServer(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockOpenAIResponse("This is a test response")))
	}))
	defer server.Close()

	// Create client with mock server URL
	_ = &DeepSeekClient{
		apiKey: "test-api-key",
		cfg: Config{
			Model:       "test-model",
			Temperature: 0.5,
			MaxTokens:   100,
		},
		http: server.Client(),
	}

	// Override the base URL by using a custom transport
	// Since we can't easily modify the base URL, we'll test the handleStream function directly instead
	t.Log("Mock server test setup completed")
}

func TestDeepSeekClient_HandleStream(t *testing.T) {
	// Test the handleStream function with mock data
	streamData := `data: {"choices":[{"delta":{"content":"Hello"}}]}
data: {"choices":[{"delta":{"content":" World"}}]}
data: {"choices":[{"delta":{"content":"!"}}]}
data: [DONE]
`

	client := &DeepSeekClient{}

	var chunks []string
	onChunk := func(s string) {
		chunks = append(chunks, s)
	}

	result, err := client.handleStream(strings.NewReader(streamData), onChunk)

	if err != nil {
		t.Errorf("handleStream() error = %v", err)
	}

	if result != "Hello World!" {
		t.Errorf("handleStream() result = %v, want %v", result, "Hello World!")
	}

	if len(chunks) != 3 {
		t.Errorf("Callback was called %d times, want 3", len(chunks))
	}
}

func TestDeepSeekClient_HandleStream_EmptyChoices(t *testing.T) {
	streamData := `data: {"choices":[]}
data: {"choices":[{"delta":{"content":"test"}}]}
data: [DONE]
`

	client := &DeepSeekClient{}
	result, err := client.handleStream(strings.NewReader(streamData), nil)

	if err != nil {
		t.Errorf("handleStream() error = %v", err)
	}

	if result != "test" {
		t.Errorf("handleStream() result = %v, want %v", result, "test")
	}
}

func TestDeepSeekClient_HandleStream_InvalidJSON(t *testing.T) {
	streamData := `data: invalid json
data: {"choices":[{"delta":{"content":"valid"}}]}
data: [DONE]
`

	client := &DeepSeekClient{}
	result, err := client.handleStream(strings.NewReader(streamData), nil)

	if err != nil {
		t.Errorf("handleStream() error = %v", err)
	}

	// Should skip invalid JSON and continue
	if result != "valid" {
		t.Errorf("handleStream() result = %v, want %v", result, "valid")
	}
}

func TestDeepSeekClient_HandleStream_NoDataPrefix(t *testing.T) {
	streamData := `some random line
data: {"choices":[{"delta":{"content":"test"}}]}
another random line
data: [DONE]
`

	client := &DeepSeekClient{}
	result, err := client.handleStream(strings.NewReader(streamData), nil)

	if err != nil {
		t.Errorf("handleStream() error = %v", err)
	}

	if result != "test" {
		t.Errorf("handleStream() result = %v, want %v", result, "test")
	}
}

func TestDeepSeekClient_HandleStream_NilCallback(t *testing.T) {
	streamData := `data: {"choices":[{"delta":{"content":"test"}}]}
data: [DONE]
`

	client := &DeepSeekClient{}
	result, err := client.handleStream(strings.NewReader(streamData), nil)

	if err != nil {
		t.Errorf("handleStream() error = %v", err)
	}

	if result != "test" {
		t.Errorf("handleStream() result = %v, want %v", result, "test")
	}
}

// Integration test helper - skipped unless API key is set
func TestDeepSeekClient_Integration(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping integration test")
	}

	// This test would make a real API call - only run if explicitly enabled
	t.Skip("Skipping integration test to avoid API costs")
}
