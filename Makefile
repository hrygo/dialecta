.PHONY: build test clean install lint fmt cover demo run help all debate-ui debate-demo debate-mixed

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
	@echo "  🎭 Debate Capabilities:"
	@echo "    debate-ui     Run in interactive mode"
	@echo "    debate-demo   Run a quick demo usage"
	@echo "    debate-file   Run debate on a file (usage: make debate-file FILE=doc.md)"
	@echo "    debate-mixed  Run with mixed providers (DeepSeek/Qwen/Gemini)"
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
# 🎭 Debate Capabilities
# =============================================================================

debate-ui: build
	@echo "🚀 Starting Interactive Debate..."
	@$(BUILD_DIR)/$(BINARY) --interactive

debate-demo: build
	@echo "📢 Running Demo Debate..."
	@echo "我们应该在明年启动一个 AI 创业项目" | $(BUILD_DIR)/$(BINARY) -

debate-file: build
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Error: Please specify FILE argument (e.g., make debate-file FILE=proposal.md)"; \
		exit 1; \
	fi
	@echo "📄 Analyzing $(FILE)..."
	@$(BUILD_DIR)/$(BINARY) $(FILE)

debate-mixed: build
	@echo "🔀 Running Mixed Provider Debate..."
	@echo "Using: Pro=DeepSeek, Con=DashScope, Judge=Gemini"
	@echo "话题: 远程办公是否应该成为主流？" | $(BUILD_DIR)/$(BINARY) \
		--pro-provider deepseek \
		--con-provider dashscope \
		--judge-provider gemini \
		-

