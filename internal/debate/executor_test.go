package debate

import (
	"strings"
	"testing"

	"github.com/hrygo/dialecta/internal/config"
	"github.com/hrygo/dialecta/internal/llm"
)

func TestNewExecutor(t *testing.T) {
	cfg := config.New()
	executor := NewExecutor(cfg)

	if executor == nil {
		t.Fatal("NewExecutor() returned nil")
	}

	if executor.cfg != cfg {
		t.Error("NewExecutor() did not set cfg correctly")
	}

	if executor.stream != false {
		t.Error("NewExecutor() should set stream to false by default")
	}
}

func TestExecutor_SetStream(t *testing.T) {
	cfg := config.New()
	executor := NewExecutor(cfg)

	proCalled := false
	conCalled := false
	judgeCalled := false

	onPro := func(s string, done bool) { proCalled = true }
	onCon := func(s string, done bool) { conCalled = true }
	onJudge := func(s string, done bool) { judgeCalled = true }

	executor.SetStream(onPro, onCon, onJudge)

	if !executor.stream {
		t.Error("SetStream() did not set stream to true")
	}

	// Test callbacks are set by calling them
	if executor.onPro == nil {
		t.Error("SetStream() did not set onPro callback")
	}
	if executor.onCon == nil {
		t.Error("SetStream() did not set onCon callback")
	}
	if executor.onJudge == nil {
		t.Error("SetStream() did not set onJudge callback")
	}

	// Verify callbacks work
	executor.onPro("test", false)
	executor.onCon("test", false)
	executor.onJudge("test", false)

	if !proCalled {
		t.Error("onPro callback was not called")
	}
	if !conCalled {
		t.Error("onCon callback was not called")
	}
	if !judgeCalled {
		t.Error("onJudge callback was not called")
	}
}

func TestResult(t *testing.T) {
	result := &Result{
		Material:        "test material",
		ProOneLiner:     "pro short",
		ProFullBody:     "pro full",
		ConOneLiner:     "con short",
		ConFullBody:     "con full",
		VerdictOneLiner: "verdict short",
		VerdictFullBody: "verdict full",
	}

	if result.Material != "test material" {
		t.Errorf("Result.Material = %v, want %v", result.Material, "test material")
	}
	if result.ProOneLiner != "pro short" {
		t.Errorf("Result.ProOneLiner = %v, want %v", result.ProOneLiner, "pro short")
	}
	if result.ProFullBody != "pro full" {
		t.Errorf("Result.ProFullBody = %v, want %v", result.ProFullBody, "pro full")
	}
	if result.ConOneLiner != "con short" {
		t.Errorf("Result.ConOneLiner = %v, want %v", result.ConOneLiner, "con short")
	}
	if result.ConFullBody != "con full" {
		t.Errorf("Result.ConFullBody = %v, want %v", result.ConFullBody, "con full")
	}
}

func TestNewExecutor_WithCustomConfig(t *testing.T) {
	cfg := &config.Config{
		ProRole: config.RoleConfig{
			Provider:    llm.ProviderDeepSeek,
			Model:       "custom-model",
			Temperature: 0.5,
			MaxTokens:   2048,
		},
		ConRole: config.RoleConfig{
			Provider:    llm.ProviderGemini,
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		JudgeRole: config.RoleConfig{
			Provider:    llm.ProviderDashScope,
			Model:       "qwen-max",
			Temperature: 0.1,
			MaxTokens:   8192,
		},
	}

	executor := NewExecutor(cfg)

	if executor.cfg.ProRole.Model != "custom-model" {
		t.Errorf("executor.cfg.ProRole.Model = %v, want %v", executor.cfg.ProRole.Model, "custom-model")
	}
	if executor.cfg.ConRole.Provider != llm.ProviderGemini {
		t.Errorf("executor.cfg.ConRole.Provider = %v, want %v", executor.cfg.ConRole.Provider, llm.ProviderGemini)
	}
	if executor.cfg.JudgeRole.MaxTokens != 8192 {
		t.Errorf("executor.cfg.JudgeRole.MaxTokens = %v, want %v", executor.cfg.JudgeRole.MaxTokens, 8192)
	}
}

func TestExecutor_SetStream_NilCallbacks(t *testing.T) {
	cfg := config.New()
	executor := NewExecutor(cfg)

	// Should not panic with nil callbacks
	executor.SetStream(nil, nil, nil)

	if !executor.stream {
		t.Error("SetStream() should still set stream to true even with nil callbacks")
	}
}

func TestResult_EmptyFields(t *testing.T) {
	result := &Result{}

	if result.Material != "" {
		t.Error("Empty Result.Material should be empty string")
	}
	if result.ProFullBody != "" {
		t.Error("Empty Result.ProFullBody should be empty string")
	}
}

// Helper to mock LLM behavior or just check simple things
// Real execution tests require mocking LLM client which is complex here
// So we focus on Executor structure logic

func TestParserParsing(t *testing.T) {
	parser := NewStreamParser("## 📝 Full Argument")

	// Test streaming input
	inputPart1 := "## 💡 One-Liner\nThis is a short "
	inputPart2 := "point.\n\n## 📝 Full Argument\nThis is the body."

	_, found := parser.Feed(inputPart1)
	if found {
		t.Error("Should not have found one-liner yet")
	}

	ol, found := parser.Feed(inputPart2)
	if !found {
		t.Error("Should have found one-liner now")
	}

	// Remove newlines and compare content
	cleanOL := strings.TrimSpace(ol)
	if cleanOL != "This is a short point." {
		t.Errorf("Got '%s', want 'This is a short point.'", cleanOL)
	}

	parser.Finalize()
	if parser.fullBody != "This is the body." {
		t.Errorf("Got full body '%s', want 'This is the body.'", parser.fullBody)
	}
}
