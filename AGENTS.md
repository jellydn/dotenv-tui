# AGENTS.md - dotenv-tui

Agent instructions for working in this repository.

## Project Overview

A terminal UI tool for managing `.env` files across projects and monorepos. Built with Go 1.25.6 and the Bubble Tea TUI framework.

## Build Commands

```bash
# Build binary
just build

# Run in development
just dev

# Build and run
just run

# Clean build artifacts
just clean
```

## Test Commands

```bash
# Run all tests
just test

# Run tests with verbose output
just test-v

# Run a single test function (pattern: go test -v -run TestFunctionName ./package/path)
go test -v -run TestParse ./internal/parser
go test -v -run TestIsSecret ./internal/detector

# Run tests for a specific package
go test -v ./internal/parser
go test -v ./internal/detector
go test -v ./internal/generator
go test -v ./internal/scanner
go test -v ./internal/tui

# Run with race detection and coverage
go test -v -race -coverprofile=coverage.out ./...
```

## Lint/Format Commands

```bash
# Run linter (golangci-lint with .golangci.yml config)
just lint

# Format code (gofmt + goimports)
just fmt
```

## Code Style Guidelines

### Imports

Group imports: standard library, then third-party, then local packages. Use `goimports` for automatic management.

```go
import (
    "bufio"
    "fmt"
    "io"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    "github.com/jellydn/dotenv-tui/internal/parser"
)
```

### Formatting

- Use `gofmt` and `goimports` for all formatting
- Indentation: tabs (Go standard)
- Line length: follow Go conventions

### Naming Conventions

- **Exported types/functions**: PascalCase (e.g., `Parse`, `KeyValue`)
- **Unexported types/functions**: camelCase (e.g., `parseKeyValue`)
- **Interfaces**: Noun describing behavior (e.g., `Entry`)
- **Structs**: Noun describing the entity (e.g., `KeyValue`, `Comment`)
- **Methods**: Verb or verb phrase (e.g., `Type()`, `Write()`)
- **Test functions**: `Test<Name>` with table-driven subtests
- **Variables**: Descriptive; use short names in short scopes

### Error Handling

Wrap errors with context and return early. Don't panic.

```go
if err != nil {
    return nil, fmt.Errorf("error reading: %w", err)
}
```

### Types

- Use interfaces to define behavior (see `Entry` interface)
- Prefer concrete types over `interface{}`
- Use type assertions with ok checks

### Comments

Follow Go conventions: start with function/type name. Document exported items.

```go
// Entry represents a line in a .env file
type Entry interface{}
```

### Testing

- Use table-driven tests with `t.Run()` for subtests
- Test files: `*_test.go` in same package
- Test both success and error cases
- Use `strings.Builder` for output testing

### Architecture Patterns

- Interface-based design for extensibility
- Separate parsing logic into dedicated packages (`internal/parser`)
- Keep `main.go` minimal (CLI entry point only)
- Use Bubble Tea's Model-Update-View pattern for TUI components

## Project Structure

```
├── main.go              # CLI entry point
├── justfile             # Task definitions
├── go.mod/go.sum        # Go module files
├── internal/            # Internal packages
│   ├── parser/          # .env file parser
│   ├── detector/        # Secret detection logic
│   ├── generator/       # .env.example generation
│   ├── scanner/         # Directory scanning
│   └── tui/             # Bubble Tea TUI components
└── testdata/            # Test fixtures
```

## Tools & Dependencies

**Required:**

- Go 1.25.6+
- `just` (command runner)
- `golangci-lint` (linting with revive, staticcheck, gocritic)
- `goimports` (import formatting)

**Key Dependencies:**

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/bubbles` — TUI components
- `github.com/charmbracelet/lipgloss` — Styling

## Notes

- Preserve comments and blank lines when parsing .env files
- Maintain key ordering from original files
- The parser uses an `Entry` interface for different line types (KeyValue, Comment, BlankLine)
- The detector uses pattern matching to identify secrets in key-value pairs
- Placeholders preserve format hints (prefix patterns) for context (e.g., `sk_***`, `ghp_***`)
- Follow Model-Update-View architecture for all TUI components
