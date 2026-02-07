# Coding Conventions

**Analysis Date:** 2026-02-08

## Naming Patterns
**Files:**
- Snake_case not used; single-word lowercase filenames (e.g., `parser.go`, `detector.go`, `scanner.go`)
- Test files co-located as `*_test.go` in the same package
- TUI components named by function: `menu.go`, `picker.go`, `preview.go`, `form.go`, `logo.go`

**Functions:**
- Exported: PascalCase verbs or verb phrases (`Parse`, `Write`, `Scan`, `ScanExamples`, `IsSecret`, `GeneratePlaceholder`, `GenerateExample`)
- Unexported: camelCase (`parseKeyValue`, `isSecretKey`, `isSecretValue`, `isCommonNonSecret`, `isBase64`, `isHex`, `scanFiles`, `isEnvFile`, `entryToString`, `adjustScroll`, `moveCursor`, `moveCursorByDirection`)
- Constructors: `New` prefix returning model values (`NewMenuModel`, `NewPickerModel`, `NewPreviewModel`, `NewFormModel`)
- Boolean predicates: `is` prefix (`isSecretKey`, `isEnvFile`, `isPlaceholderValue`)

**Variables:**
- Short names in tight scopes (`e`, `kv`, `r`, `w`, `i`)
- Descriptive names in wider scopes (`entries`, `diffLines`, `originalEntries`, `generatedEntries`)
- Package-level maps use `Map` suffix (`secretPatternsMap`, `knownSecretPrefixesMap`, `commonNonSecretsMap`)
- Constants: camelCase for unexported (`menuScreen`, `pickerScreen`), PascalCase for exported (`GenerateExample`, `GenerateEnv`)
- Direction constants use descriptive names (`directionUp`, `directionDown`)

**Types:**
- Interfaces: noun describing behavior (`Entry`)
- Structs: noun describing entity (`KeyValue`, `Comment`, `BlankLine`, `Release`)
- TUI models: `<Component>Model` suffix (`MenuModel`, `PickerModel`, `PreviewModel`, `FormModel`)
- TUI messages: `<Component><Action>Msg` suffix (`PickerFinishedMsg`, `PreviewFinishedMsg`, `FormFinishedMsg`, `pickerInitMsg`, `previewInitMsg`, `formInitMsg`)
- Enums: typed `int` constants with `iota` (`MenuChoice`, `screen`)

## Code Style
**Formatting:**
- `gofmt` with simplify enabled
- `goimports` with local prefix `github.com/jellydn/dotenv-tui`
- Indentation: tabs (Go standard)

**Linting:**
- `golangci-lint` v2 with `.golangci.yml` config
- Enabled linters: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`, `gocritic`, `misspell`, `revive`, `unconvert`, `unparam`
- `revive` exported rule disables stuttering check (`disableStutteringCheck`)
- `gosec` excluded for `_test.go` files and weak-crypto warnings
- Timeout: 5 minutes

## Import Organization
**Order:**
1. Standard library (`bufio`, `fmt`, `io`, `os`, `strings`, `path/filepath`)
2. Third-party (`github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles`)
3. Local packages (`github.com/jellydn/dotenv-tui/internal/...`)

**Path Aliases:**
- `tea` for `github.com/charmbracelet/bubbletea` (`tea "github.com/charmbracelet/bubbletea"`)

**Notes:**
- Local packages sometimes appear between standard library and third-party (e.g., in `picker.go`, `preview.go`, `form.go` local imports precede charmbracelet imports)

## Error Handling
**Patterns:**
- Wrap errors with `fmt.Errorf("context: %w", err)` for all returned errors
- Return early on error (guard clauses)
- Never panic; always return errors
- Deferred close with discarded error: `defer func() { _ = file.Close() }()`
- CLI error handling: `fmt.Fprintf(os.Stderr, ...)` followed by `os.Exit(1)`
- TUI error handling: errors embedded in message structs (`FormFinishedMsg.Error`)
- Graceful fallback: scanner returns empty slice on error instead of propagating

## Logging
**Framework:** None (no logging framework)

**Patterns:**
- User-facing output via `fmt.Printf` / `fmt.Println` to stdout
- Error output via `fmt.Fprintf(os.Stderr, ...)`
- No structured logging; TUI renders errors inline via lipgloss-styled views

## Comments
**When to Comment:**
- Package-level doc comments on every package (`// Package parser provides...`)
- Exported function/type doc comments starting with the name (`// Parse reads a .env file...`, `// Entry represents a line...`)
- Brief inline comments for non-obvious logic (e.g., `// Increase buffer to 1MB to handle large values`)
- No comments on unexported helpers unless logic is complex
- Comments follow Go conventions: complete sentences, start with the identifier name

## Function Design
**Size:** Small, focused functions; largest functions are TUI `Update` and `View` methods (~60 lines max)
**Parameters:** Minimal parameters (1-3); use struct fields for state in TUI models
**Return Values:** `(result, error)` pair for fallible operations; TUI follows `(tea.Model, tea.Cmd)` pattern
**Design:** Higher-order functions used for shared logic (`scanFiles` takes a predicate, `generateFile` takes `entryProcessor`)

## Module Design
**Exports:** Each `internal/` package exports a minimal public API:
- `parser`: `Parse`, `Write`, `Entry`, `KeyValue`, `Comment`, `BlankLine`
- `detector`: `IsSecret`, `GeneratePlaceholder`
- `generator`: `GenerateExample`
- `scanner`: `Scan`, `ScanExamples`
- `upgrade`: `Upgrade`, `Release`
- `tui`: Model types, message types, constructor functions

**Structure:** Flat package structure under `internal/`; no nested sub-packages. `main.go` is the CLI entry point with minimal logic, delegating to internal packages.

**Patterns:**
- Interface-based design for extensibility (`Entry` interface with `KeyValue`, `Comment`, `BlankLine` implementations)
- Bubble Tea Model-Update-View pattern for all TUI components
- Constructor functions return `tea.Cmd` (not model) for async initialization (`NewPickerModel`, `NewPreviewModel`, `NewFormModel`)
- `MenuModel` is the exception: constructor returns the model directly since it has no async init

---
*Convention analysis: 2026-02-08*
