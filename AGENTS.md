# AGENTS.md - dotenv-tui

Agent instructions for working in this repository.

## Project Overview

A terminal UI tool for managing `.env` files across projects and monorepos. Built with Go and the Bubble Tea TUI framework.

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

# Run a single test (example)
go test -v -run TestParse ./internal/parser

# Run tests for specific package
go test -v ./internal/parser
```

## Lint/Format Commands

```bash
# Run linter
just lint

# Format code (runs gofmt + goimports)
just fmt
```

## Code Style Guidelines

### Imports

- Group imports: standard library, then third-party
- Use `goimports` for automatic import management
- No blank lines between import groups (gofmt standard)

```go
import (
    "bufio"
    "fmt"
    "io"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
)
```

### Formatting

- Use `gofmt` for all formatting
- Use `goimports` for import organization
- Max line length: follow Go conventions (no strict limit)
- Indentation: tabs (Go standard)

### Naming Conventions

- **Exported types/functions**: PascalCase (e.g., `Parse`, `KeyValue`)
- **Unexported types/functions**: camelCase (e.g., `parseKeyValue`)
- **Interfaces**: Noun describing behavior (e.g., `Entry`)
- **Structs**: Noun describing the entity (e.g., `KeyValue`, `Comment`)
- **Methods**: Verb or verb phrase (e.g., `Type()`, `Write()`)
- **Test functions**: `Test<Name>` with table-driven subtests
- **Variables**: Descriptive, use short names in short scopes

### Error Handling

- Wrap errors with context using `fmt.Errorf("...: %w", err)`
- Return errors early (guard clauses)
- Don't panic; return errors to caller

```go
if err != nil {
    return nil, fmt.Errorf("error reading: %w", err)
}
```

### Types

- Use interfaces to define behavior (see `Entry` interface)
- Use struct tags when needed for serialization
- Prefer concrete types over `interface{}`
- Use type assertions with ok checks

### Comments

- Follow Go conventions: start with function/type name
- Document exported types and functions
- Keep comments concise and informative

```go
// Entry represents a line in a .env file
type Entry interface {
    Type() string
}
```

### Testing

- Use table-driven tests with `t.Run()` for subtests
- Test files: `*_test.go` in same package
- Test both success and error cases
- Use `strings.Builder` for output testing
- Name test cases descriptively

### Architecture Patterns

- Interface-based design for extensibility
- Separate parsing logic into dedicated packages (`internal/parser`)
- Keep `main.go` minimal (CLI entry point only)
- Use Bubble Tea's Model-Update-View pattern for TUI

## Project Structure

```
.
├── main.go              # CLI entry point
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── justfile             # Task definitions
├── internal/            # Internal packages
│   └── parser/          # .env file parser
│       ├── parser.go
│       └── parser_test.go
└── scripts/             # Development scripts
```

## Dependencies

- **Bubble Tea**: TUI framework (github.com/charmbracelet/bubbletea)
- **Lip Gloss**: Styling (github.com/charmbracelet/lipgloss)
- Standard Go library for file I/O

## Tools Required

- Go 1.25.6+
- `just` (command runner)
- `golangci-lint` (linting)
- `goimports` (import formatting)

## Notes

- Preserve comments and blank lines when parsing .env files
- Maintain key ordering from original files
- The parser uses an `Entry` interface for different line types (KeyValue, Comment, BlankLine)
- This is a Bubble Tea TUI application - follow the Model-Update-View architecture
