package llm

import (
	"os"
	"testing"

	"github.com/google/generative-ai-go/genai"
)

func TestNewGeminiClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		cfg       Config
		wantModel string
	}{
		{
			name:      "with custom model",
			apiKey:    "test-key",
			cfg:       Config{Model: "gemini-pro", Temperature: 0.5, MaxTokens: 2048},
			wantModel: "gemini-pro",
		},
		{
			name:      "with empty model - uses default",
			apiKey:    "test-key",
			cfg:       Config{Model: "", Temperature: 0.8, MaxTokens: 4096},
			wantModel: "gemini-3-pro-preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGeminiClient(tt.apiKey, tt.cfg)

			if client == nil {
				t.Fatal("NewGeminiClient() returned nil")
			}

			if client.apiKey != tt.apiKey {
				t.Errorf("client.apiKey = %v, want %v", client.apiKey, tt.apiKey)
			}

			if client.cfg.Model != tt.wantModel {
				t.Errorf("client.cfg.Model = %v, want %v", client.cfg.Model, tt.wantModel)
			}
		})
	}
}

func TestGeminiClient_Implements_Client(t *testing.T) {
	// Verify GeminiClient implements Client interface
	var _ Client = (*GeminiClient)(nil)
}

func TestGeminiClient_ConfigPreserved(t *testing.T) {
	cfg := Config{
		Model:       "gemini-1.5-pro",
		Temperature: 0.3,
		MaxTokens:   8192,
	}

	client := NewGeminiClient("api-key", cfg)

	if client.cfg.Temperature != 0.3 {
		t.Errorf("Temperature = %v, want %v", client.cfg.Temperature, 0.3)
	}
	if client.cfg.MaxTokens != 8192 {
		t.Errorf("MaxTokens = %v, want %v", client.cfg.MaxTokens, 8192)
	}
}

func TestExtractGeminiText(t *testing.T) {
	tests := []struct {
		name     string
		resp     *genai.GenerateContentResponse
		expected string
	}{
		{
			name: "single text part",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Hello, World!"),
							},
						},
					},
				},
			},
			expected: "Hello, World!",
		},
		{
			name: "multiple text parts",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Hello, "),
								genai.Text("World!"),
							},
						},
					},
				},
			},
			expected: "Hello, World!",
		},
		{
			name: "multiple candidates",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text("First")},
						},
					},
					{
						Content: &genai.Content{
							Parts: []genai.Part{genai.Text("Second")},
						},
					},
				},
			},
			expected: "FirstSecond",
		},
		{
			name:     "empty response",
			resp:     &genai.GenerateContentResponse{},
			expected: "",
		},
		{
			name: "nil content",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{Content: nil},
				},
			},
			expected: "",
		},
		{
			name: "empty parts",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{},
						},
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGeminiText(tt.resp)
			if result != tt.expected {
				t.Errorf("extractGeminiText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractGeminiText_UnicodeContent(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{
						genai.Text("你好，世界！🌍"),
					},
				},
			},
		},
	}

	result := extractGeminiText(resp)
	expected := "你好，世界！🌍"

	if result != expected {
		t.Errorf("extractGeminiText() = %q, want %q", result, expected)
	}
}

// Integration test helper - skipped unless API key is set
func TestGeminiClient_Integration(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY/GOOGLE_API_KEY not set, skipping integration test")
	}

	// This test would make a real API call - only run if explicitly enabled
	t.Skip("Skipping integration test to avoid API costs")
}
