# OpenCode Makefile
# Build and development commands for the OpenCode project

.PHONY: help build build-go build-tui build-all clean install test run serve tui dev

# Default target
help:
	@echo "OpenCode Build Commands"
	@echo ""
	@echo "  make build        - Build the opencode CLI binary with Go"
	@echo "  make build-go     - Same as build"
	@echo "  make build-tui    - Build the TUI binary with Go"
	@echo "  make build-all    - Build both CLI and TUI binaries"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make install      - Build and install to /usr/local/bin"
	@echo "  make test         - Run Go tests"
	@echo "  make test-cli     - Test CLI commands (version, models, etc.)"
	@echo "  make run          - Build and run TUI (WIP: server missing /path endpoint)"
	@echo "  make serve        - Run server only"
	@echo "  make tui          - Run TUI only (requires server to be running)"
	@echo "  make dev          - Run with bun (TypeScript version - recommended)"
	@echo ""
	@echo "Note: The Go server is still in development. Use 'make dev' for full functionality."
	@echo ""

# Build the opencode binary with Go
build: build-go

build-go:
	@echo "Building opencode binary with Go..."
	@cd packages/opencode-go && go build -o ../../bin/opencode ./cmd/opencode
	@echo "Build complete: bin/opencode"

# Build the TUI binary with Go
build-tui:
	@echo "Building TUI binary with Go..."
	@mkdir -p packages/tui/cmd/opencode/dist
	@cd packages/tui && go build -o cmd/opencode/dist/tui ./cmd/opencode
	@echo "Build complete: packages/tui/cmd/opencode/dist/tui"

# Build both CLI and TUI
build-all: build-go build-tui

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/opencode
	@rm -rf packages/tui/cmd/opencode/dist
	@cd packages/opencode-go && go clean
	@cd packages/tui && go clean
	@echo "Clean complete"

# Build and install to /usr/local/bin
install: build
	@echo "Installing opencode to /usr/local/bin..."
	@sudo cp bin/opencode /usr/local/bin/opencode
	@sudo chmod +x /usr/local/bin/opencode
	@echo "Installation complete"

# Run Go tests
test:
	@echo "Running Go tests..."
	@cd packages/opencode-go && go test -v ./...

# Build and run TUI (starts server and TUI together)
# Note: Open two terminals - one for 'make serve', one for 'make tui'
# Or use the run script: ./bin/run-opencode.sh
run: build-all
	@echo "Starting opencode server and TUI with random port..."
	@./script/run-opencode.sh --random-port

# Run server only
serve: build
	@echo "Starting opencode server..."
	@./bin/opencode serve

# Run TUI only (requires server to be running)
tui: build
	@echo "Starting opencode TUI..."
	@./bin/opencode tui

# Run with bun (TypeScript version)
dev:
	@echo "Running opencode with bun (TypeScript version)..."
	@cd packages/opencode && bun dev
