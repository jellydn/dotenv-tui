default:
    @just --list

# Build the binary
build:
    go build -o dotenv-tui .

# Run in development
dev:
    go run .

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-v:
    go test -v ./...

# Run linter
lint:
    golangci-lint run ./...

# Format code
fmt:
    gofmt -w .
    goimports -w .

# Clean build artifacts
clean:
    rm -f dotenv-tui

# Build and run
run: build
    ./dotenv-tui

# Create symlink for local development
dev-symlink:
    ln -sf $(pwd)/dotenv-tui ~/.local/bin/dotenv-tui-dev
