# Architecture

**Analysis Date:** 2026-02-08

## Pattern Overview
**Overall:** Model-Update-View (Elm Architecture via Bubble Tea) with layered internal packages

**Key Characteristics:**
- Unidirectional data flow through Bubble Tea's message-passing loop
- Interface-based abstractions for .env file line types (`Entry` interface)
- Clean separation between CLI commands (headless) and TUI screens (interactive)
- Internal packages organized by domain responsibility with no circular dependencies

## Layers
**CLI Entry (`main.go`):**
- Purpose: Parses flags, dispatches to headless commands or launches TUI
- Location: `main.go`
- Contains: Flag parsing, headless generate/scan/yolo/upgrade commands, top-level Bubble Tea model and screen routing
- Depends on: `internal/parser`, `internal/generator`, `internal/scanner`, `internal/upgrade`, `internal/tui`
- Used by: End user via `dotenv-tui` binary

**TUI Presentation (`internal/tui/`):**
- Purpose: Interactive terminal UI screens using Bubble Tea Model-Update-View
- Location: `internal/tui/`
- Contains: MenuModel, PickerModel, PreviewModel, FormModel, Logo/Wordmark
- Depends on: `internal/parser`, `internal/generator`, `internal/scanner`, `bubbletea`, `bubbles`, `lipgloss`
- Used by: `main.go` top-level model

**Parser (`internal/parser/`):**
- Purpose: Parse and serialize .env files while preserving structure (comments, blank lines, ordering)
- Location: `internal/parser/`
- Contains: `Entry` interface, `KeyValue`/`Comment`/`BlankLine` types, `Parse()`, `Write()`
- Depends on: Standard library only
- Used by: `internal/generator`, `internal/tui`, `main.go`

**Detector (`internal/detector/`):**
- Purpose: Identify secrets in key-value pairs using key patterns, value patterns, and encoding detection
- Location: `internal/detector/`
- Contains: `IsSecret()`, `GeneratePlaceholder()`, pattern matchers for known prefixes (Stripe, GitHub, Slack, etc.)
- Depends on: Standard library only
- Used by: `internal/generator`

**Generator (`internal/generator/`):**
- Purpose: Transform .env entries into .env.example by masking detected secrets
- Location: `internal/generator/`
- Contains: `GenerateExample()` — iterates entries, replaces secret values with placeholders
- Depends on: `internal/parser`, `internal/detector`
- Used by: `internal/tui/preview.go`, `main.go`

**Scanner (`internal/scanner/`):**
- Purpose: Recursively discover .env and .env.example files, skipping dependency directories
- Location: `internal/scanner/`
- Contains: `Scan()`, `ScanExamples()`, directory skip list (node_modules, .git, vendor, etc.)
- Depends on: Standard library only
- Used by: `internal/tui/picker.go`, `main.go`

**Upgrade (`internal/upgrade/`):**
- Purpose: Self-update binary from GitHub releases with SHA256 checksum verification
- Location: `internal/upgrade/`
- Contains: `Upgrade()`, GitHub API client, binary download/replace logic
- Depends on: Standard library only (net/http, crypto/sha256)
- Used by: `main.go`

## Data Flow
**Generate .env.example (TUI):**
1. Menu → user selects "Generate .env.example"
2. Picker → scanner finds .env files, user selects files
3. Preview → parser reads file, generator masks secrets, diff shown
4. User confirms → parser.Write() outputs .env.example
5. Done screen → Tab/Shift+Tab to navigate between files

**Generate .env from .env.example (TUI):**
1. Menu → user selects "Generate .env from .env.example"
2. Picker → scanner finds .env.example files, user selects
3. Form → parser reads .env.example, text inputs for placeholder values
4. User submits (Enter/Ctrl+S) → parser.Write() outputs .env
5. Done screen

**Headless CLI:**
1. Flag parsed → file opened → parser.Parse() → processor applied → parser.Write() → output file

**State Management:**
- Top-level `model` struct in main.go holds current screen enum and all sub-models
- Screen transitions via `currentScreen` field with message-driven navigation
- Each TUI component manages its own internal state (cursor, scroll, selected items)
- Sub-models initialized lazily via Bubble Tea commands (returning init messages)

## Key Abstractions
**Entry Interface:**
- Purpose: Polymorphic representation of .env file lines
- Examples: `internal/parser/parser.go` — `KeyValue`, `Comment`, `BlankLine`
- Pattern: Type switch dispatch in Parse/Write/generator

**Screen Routing:**
- Purpose: Top-level state machine for TUI navigation
- Examples: `main.go` — `screen` enum (menuScreen, pickerScreen, previewScreen, formScreen, doneScreen)
- Pattern: Switch-based dispatch in Update/View, message-driven transitions

**Bubble Tea Messages:**
- Purpose: Inter-component communication and async initialization
- Examples: `PickerFinishedMsg`, `PreviewFinishedMsg`, `FormFinishedMsg`, `pickerInitMsg`, `previewInitMsg`, `formInitMsg`
- Pattern: Commands return closures that emit typed messages

## Entry Points
**Interactive TUI:**
- Location: `main.go:331`
- Triggers: `dotenv-tui` with no flags
- Responsibilities: Launch alt-screen Bubble Tea program with menu → picker → preview/form flow

**CLI Generate Example:**
- Location: `main.go:286`
- Triggers: `--generate-example <path>`
- Responsibilities: Parse .env, mask secrets via generator, write .env.example

**CLI Generate Env:**
- Location: `main.go:294`
- Triggers: `--generate-env <path>`
- Responsibilities: Parse .env.example, write .env (pass-through)

**CLI Scan:**
- Location: `main.go:302`
- Triggers: `--scan [directory]`
- Responsibilities: Discover and list .env files

**CLI Yolo:**
- Location: `main.go:315`
- Triggers: `--yolo`
- Responsibilities: Find all .env.example files, generate .env for each with optional overwrite prompt

**CLI Upgrade:**
- Location: `main.go:323`
- Triggers: `--upgrade`
- Responsibilities: Self-update binary from GitHub releases

## Error Handling
**Strategy:** Return errors with context wrapping (`fmt.Errorf` with `%w`), exit with code 1 on CLI errors

**Patterns:**
- CLI commands: `fmt.Fprintf(os.Stderr, ...)` + `os.Exit(1)`
- Parser: Wrapped errors with line context
- TUI components: Graceful degradation (show error in UI, empty state on parse failure)
- Upgrade: Checksum verification, cleanup temp files on failure
- File operations: Deferred close with `_ = file.Close()` to suppress close errors

## Cross-Cutting Concerns
**Logging:** None — CLI uses stdout/stderr directly, TUI uses Bubble Tea rendering
**Validation:** Inline in parser (key-value format), detector (pattern matching), form (placeholder detection)
**Authentication:** N/A — no auth; upgrade uses unauthenticated GitHub API

---
*Architecture analysis: 2026-02-08*
