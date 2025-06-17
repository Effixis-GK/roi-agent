# Chrome Browser Access Tool Makefile

.PHONY: build run clean install deps test

# Build the application
build:
	go build -o chrome-browser-access

# Run the application directly
run:
	go run .

# Run with extended analysis (uncomment runExtendedAnalysis in main.go first)
run-extended:
	@echo "Make sure to enable runExtendedAnalysis() in main.go"
	go run .

# Install dependencies
deps:
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	rm -f chrome-browser-access
	rm -f /tmp/chrome_history_copy.db
	rm -f /tmp/chrome_downloads_copy.db

# Install the tool to GOPATH/bin
install: build
	cp chrome-browser-access $(GOPATH)/bin/

# Run basic tests
test:
	go test -v

# Check for Chrome installation
check-chrome:
	@if [ -d "$(HOME)/Library/Application Support/Google/Chrome" ]; then \
		echo "✅ Chrome is installed"; \
	else \
		echo "❌ Chrome not found. Please install Google Chrome."; \
	fi

# Show help
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make run-extended - Run with extended analysis"
	@echo "  make deps         - Install dependencies"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make install      - Install to GOPATH/bin"
	@echo "  make check-chrome - Check if Chrome is installed"
	@echo "  make help         - Show this help message"
