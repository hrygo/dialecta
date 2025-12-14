.PHONY: build test clean install lint fmt cover demo run help all ui demo gemini gemini-deepseek gemini-qwen deepseek-qwen

# =============================================================================
# Variables
# =============================================================================
BINARY := dialecta
BUILD_DIR := ./bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOFMT := $(GOCMD) fmt
GOMOD := $(GOCMD) mod

# =============================================================================
# Default Goal
# =============================================================================
.DEFAULT_GOAL := help

# =============================================================================
# Main Build Task
# =============================================================================
all: fmt lint test build

# =============================================================================
# Help
# =============================================================================
help:
	@echo ""
	@echo "  ╭──────────────────────────────────────────────────────────────╮"
	@echo "  │             DIALECTA - Build System                          │"
	@echo "  ╰──────────────────────────────────────────────────────────────╯"
	@echo ""
	@echo "  Usage: make [target]"
	@echo ""
	@echo "  🛠️  Engineering Capabilities:"
	@echo "    build         Build the binary"
	@echo "    build-all     Build for Linux, macOS, and Windows"
	@echo "    clean         Remove build artifacts"
	@echo "    deps          Download dependencies"
	@echo "    fmt           Format code"
	@echo "    lint          Run linter"
	@echo "    test          Run tests"
	@echo "    cover         Run tests with coverage"
	@echo "    install       Install to GOPATH/bin"
	@echo ""
	@echo "  🎭 Debate (Default: Pro=DeepSeek, Con=Qwen, Judge=Gemini):"
	@echo "    ui                 Interactive mode"
	@echo "    demo               Quick demo"
	@echo ""
	@echo "  🔀 Model Combinations (Judge=Gemini, pipe input):"
	@echo "    gemini             Pro=Gemini,   Con=Gemini"
	@echo "    gemini-deepseek    Pro=Gemini,   Con=DeepSeek"
	@echo "    gemini-qwen        Pro=Gemini,   Con=Qwen"
	@echo "    deepseek-qwen      Pro=DeepSeek, Con=Qwen"
	@echo ""
	@echo "  Example: echo 'AI是否会取代人类？' | make gemini"
	@echo ""

# =============================================================================
# 🛠️ Engineering Capabilities
# =============================================================================

build:
	@echo "🔨 Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/dialecta
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY)"

build-all: build-linux build-darwin build-windows

build-linux:
	@echo "🔨 Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/dialecta

build-darwin:
	@echo "🔨 Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/dialecta
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/dialecta

build-windows:
	@echo "🔨 Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/dialecta

clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies ready"

fmt:
	@echo "📝 Formatting code..."
	$(GOFMT) ./...

lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

vet:
	@echo "🔬 Running go vet..."
	$(GOCMD) vet ./...

test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./internal/...

cover:
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./internal/...
	$(GOCMD) tool cover -func=coverage.out | tail -1
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

install:
	@echo "📦 Installing $(BINARY)..."
	$(GOCMD) install $(LDFLAGS) ./cmd/dialecta
	@echo "✅ Installed to $(shell go env GOPATH)/bin/$(BINARY)"

# =============================================================================
# 🎭 Debate
# =============================================================================
# Basic debate commands without model specification (using defaults)
# Default: Pro=DeepSeek, Con=Qwen, Judge=Gemini

ui: build
	@echo "🚀 Interactive Mode"
	@$(BUILD_DIR)/$(BINARY) -i

demo: build
	@echo "📢 Quick Demo"
	@echo "我们应该在明年启动一个 AI 创业项目" | $(BUILD_DIR)/$(BINARY) -

# =============================================================================
# 🔀 Model Combinations (Judge=Gemini)
# =============================================================================
# All combinations use Gemini as Judge, with different Pro/Con combinations:
# - gemini:          Pro=Gemini,   Con=Gemini
# - gemini-deepseek: Pro=Gemini,   Con=DeepSeek
# - gemini-qwen:     Pro=Gemini,   Con=Qwen
# - deepseek-qwen:   Pro=DeepSeek, Con=Qwen
#
# Usage: echo 'your topic' | make <command>
# Example: echo 'AI是否会取代人类？' | make gemini

gemini: build
	@echo "🌟 Gemini vs Gemini"
	@$(BUILD_DIR)/$(BINARY) --pro-provider gemini --con-provider gemini --judge-provider gemini -

gemini-deepseek: build
	@echo "⚔️  Gemini vs DeepSeek"
	@$(BUILD_DIR)/$(BINARY) --pro-provider gemini --con-provider deepseek --judge-provider gemini -

gemini-qwen: build
	@echo "⚔️  Gemini vs Qwen"
	@$(BUILD_DIR)/$(BINARY) --pro-provider gemini --con-provider dashscope --judge-provider gemini -

deepseek-qwen: build
	@echo "⚔️  DeepSeek vs Qwen"
	@$(BUILD_DIR)/$(BINARY) --pro-provider deepseek --con-provider dashscope --judge-provider gemini -

