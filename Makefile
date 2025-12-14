.PHONY: build test clean install lint fmt cover demo run help all

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
# Default target
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
	@echo "  Targets:"
	@echo "    build       Build the binary"
	@echo "    install     Install to GOPATH/bin"
	@echo "    test        Run tests"
	@echo "    cover       Run tests with coverage report"
	@echo "    lint        Run linter (requires golangci-lint)"
	@echo "    fmt         Format code"
	@echo "    clean       Remove build artifacts"
	@echo "    deps        Download dependencies"
	@echo "    demo        Run with example input"
	@echo "    run         Run interactive mode"
	@echo "    all         Format, lint, test, and build"
	@echo ""

# =============================================================================
# Build
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

# =============================================================================
# Install & Run
# =============================================================================
install:
	@echo "📦 Installing $(BINARY)..."
	$(GOCMD) install $(LDFLAGS) ./cmd/dialecta
	@echo "✅ Installed to $(shell go env GOPATH)/bin/$(BINARY)"

run:
	@$(BUILD_DIR)/$(BINARY) --interactive

# =============================================================================
# Test
# =============================================================================
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...

test-short:
	@echo "🧪 Running short tests..."
	$(GOTEST) -short ./...

cover:
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./internal/...
	$(GOCMD) tool cover -func=coverage.out | tail -1
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# =============================================================================
# Code Quality
# =============================================================================
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt:
	@echo "📝 Formatting code..."
	$(GOFMT) ./...

vet:
	@echo "🔬 Running go vet..."
	$(GOCMD) vet ./...

# =============================================================================
# Dependencies
# =============================================================================
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies ready"

# =============================================================================
# Clean
# =============================================================================
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

# =============================================================================
# Demo
# =============================================================================
demo:
	@echo ""
	@echo "📢 Running demo..."
	@echo ""
	@echo "我们应该在明年启动一个 AI 创业项目" | $(BUILD_DIR)/$(BINARY) -
