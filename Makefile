.PHONY: build test clean install lint

# Binary name
BINARY := ./bin/dialecta

# Build the binary
build:
	go build -o $(BINARY) ./cmd/dialecta

# Run tests
test:
	go test -v ./...

# Run tests with coverage
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install to GOPATH/bin
install:
	go install ./cmd/dialecta

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run the binary with example input
demo:
	@echo "我们应该在明年启动一个 AI 创业项目" | ./$(BINARY) -
