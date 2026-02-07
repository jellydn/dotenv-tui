# Technology Stack

**Analysis Date:** 2026-02-08

## Languages
**Primary:**
- Go 1.25.6 - Entire application (CLI + TUI)

**Secondary:**
- Shell (POSIX sh) - Install script (`install.sh`), CI workflows
- YAML - CI/CD configuration, linter config, Dependabot config
- JSON - Renovate config

## Runtime
**Environment:**
- Go 1.25.6 (compiled binary, no runtime required)

**Package Manager:**
- Go Modules (`go mod`)
- Lockfile: `go.sum` (present)

## Frameworks
**Core:**
- Bubble Tea v1.3.10 (`charmbracelet/bubbletea`) - Terminal UI framework (Model-Update-View)
- Bubbles v0.21.1 (`charmbracelet/bubbles`) - Pre-built TUI components (text input, list, viewport)
- Lip Gloss v1.1.0 (`charmbracelet/lipgloss`) - Terminal styling/layout

**Testing:**
- Go standard `testing` package - Unit tests with table-driven subtests

**Build/Dev:**
- `just` (justfile) - Task runner (build, test, lint, fmt, dev, run, clean)
- `golangci-lint` - Linting (errcheck, govet, staticcheck, gocritic, revive, misspell, etc.)
- `gofmt` + `goimports` - Code formatting

## Key Dependencies
**Critical:**
- `charmbracelet/bubbletea` v1.3.10 - Core TUI framework, entire interactive UI depends on it
- `charmbracelet/bubbles` v0.21.1 - Text input, list selection, viewport components
- `charmbracelet/lipgloss` v1.1.0 - All terminal styling and layout rendering

**Infrastructure:**
- `atotto/clipboard` v0.1.4 - Clipboard integration (indirect, via bubbles)
- `charmbracelet/x/ansi` v0.11.5 - ANSI escape sequence handling (indirect)
- `charmbracelet/x/term` v0.2.2 - Terminal capability detection (indirect)
- `muesli/termenv` v0.16.0 - Terminal environment detection (indirect)
- `golang.org/x/sys` v0.38.0 - System calls (indirect)

## Configuration
**Environment:**
- No environment variables required at runtime
- `INSTALL_DIR` optional for install script (defaults to `~/.local/bin`)

**Build:**
- `.golangci.yml` - Linter configuration (v2 format, 10+ enabled linters)
- `justfile` - Build/dev task definitions
- `renovate.json` - Automated dependency updates (Renovate)
- `.github/dependabot.yml` - Automated dependency updates (Dependabot, gomod + github-actions)

## Platform Requirements
**Development:**
- Go 1.25.6+
- `just` command runner
- `golangci-lint` for linting
- `goimports` for import formatting

**Production:**
- Standalone compiled binary (no runtime dependencies)
- Supported: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- Terminal with ANSI support for TUI mode

---
*Stack analysis: 2026-02-08*
