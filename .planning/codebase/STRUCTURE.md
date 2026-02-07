# Codebase Structure

**Analysis Date:** 2026-02-08

## Directory Layout
```
dotenv-tui/
├── internal/              # Internal packages (domain logic)
│   ├── detector/          # Secret detection engine
│   ├── generator/         # .env.example generation
│   ├── parser/            # .env file parser/writer
│   ├── scanner/           # Directory tree file discovery
│   ├── tui/               # Bubble Tea UI components
│   └── upgrade/           # Self-update from GitHub releases
├── testdata/              # Test fixture .env files
├── assets/                # Static assets (screenshots, etc.)
├── doc/                   # Documentation
│   └── adr/               # Architecture Decision Records
├── scripts/               # Build/automation scripts
│   └── ralph/             # Ralph agent scripts
├── tasks/                 # Task definitions (empty)
├── .github/               # GitHub Actions & Dependabot
│   ├── workflows/         # CI workflows
│   └── dependabot.yml     # Dependency updates
├── .planning/             # Planning documents
│   └── codebase/          # Architecture analysis (this file)
├── main.go                # CLI entry point + top-level TUI model
├── justfile               # Task runner (build, test, lint, fmt)
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── install.sh             # Installation script
├── renovate.json          # Renovate bot config
├── .golangci.yml          # Linter configuration
├── AGENTS.md              # Agent instructions
├── CLAUDE.md              # Claude AI instructions
├── README.md              # Project documentation
└── LICENSE                # License file
```

## Directory Purposes
**internal/parser/:**
- Purpose: Parse .env files into structured Entry types, serialize back to files
- Contains: `parser.go` (Entry interface, KeyValue/Comment/BlankLine types, Parse/Write), `parser_test.go`, `testdata_test.go`
- Key files: `parser.go`

**internal/detector/:**
- Purpose: Detect secrets via key name patterns, value patterns (JWT, base64, hex), and known prefixes
- Contains: `detector.go` (IsSecret, GeneratePlaceholder), `detector_test.go`
- Key files: `detector.go`

**internal/generator/:**
- Purpose: Transform .env entries by masking secrets with format-aware placeholders
- Contains: `generator.go` (GenerateExample), `generator_test.go`
- Key files: `generator.go`

**internal/scanner/:**
- Purpose: Walk directory trees to find .env and .env.example files, skip dependency dirs
- Contains: `scanner.go` (Scan, ScanExamples, skip list), `scanner_test.go`
- Key files: `scanner.go`

**internal/tui/:**
- Purpose: All Bubble Tea interactive UI components
- Contains: `menu.go` (main menu), `picker.go` (file selector), `preview.go` (diff preview), `form.go` (env editor), `logo.go` (ASCII art branding), plus corresponding `*_test.go` files
- Key files: `menu.go`, `picker.go`, `preview.go`, `form.go`

**internal/upgrade/:**
- Purpose: Self-update binary by downloading from GitHub releases with checksum verification
- Contains: `upgrade.go` (Upgrade, download, verify), `upgrade_test.go`
- Key files: `upgrade.go`

**testdata/:**
- Purpose: Fixture .env files for parser integration tests
- Contains: `.env.complex`, `.env.example`, `.env.minimal`, `.env.quoted`
- Key files: All fixtures used by `internal/parser/testdata_test.go`

## Key File Locations
**Entry Points:**
- `main.go`: CLI flag parsing, headless commands, TUI launch, top-level screen routing

**Configuration:**
- `go.mod`: Module path `github.com/jellydn/dotenv-tui`, Go 1.25.6, Charm dependencies
- `.golangci.yml`: Linter rules (errcheck, govet, staticcheck, revive, gocritic, misspell)
- `justfile`: Task runner commands (build, dev, test, lint, fmt, clean, run)
- `renovate.json`: Dependency update bot config

**Core Logic:**
- `internal/parser/parser.go`: Entry interface, Parse(), Write()
- `internal/detector/detector.go`: IsSecret(), GeneratePlaceholder()
- `internal/generator/generator.go`: GenerateExample()
- `internal/scanner/scanner.go`: Scan(), ScanExamples()

**Testing:**
- `internal/parser/parser_test.go`: Parser unit tests
- `internal/parser/testdata_test.go`: Integration tests using testdata/ fixtures
- `internal/detector/detector_test.go`: Secret detection tests
- `internal/generator/generator_test.go`: Generator tests
- `internal/scanner/scanner_test.go`: Scanner tests
- `internal/tui/*_test.go`: TUI component tests (menu, picker, preview, form, logo)
- `internal/upgrade/upgrade_test.go`: Upgrade logic tests

## Naming Conventions
**Files:**
- `<package>.go`: Primary implementation (e.g., `parser.go`, `detector.go`)
- `<package>_test.go`: Tests for the package
- `<component>.go`: TUI components named by screen (e.g., `menu.go`, `picker.go`, `form.go`)

**Directories:**
- `internal/<domain>/`: One package per domain concern
- Flat structure within each package (no sub-packages)

**Types:**
- `<Component>Model`: Bubble Tea models (e.g., `MenuModel`, `PickerModel`)
- `<Component>FinishedMsg`: Completion messages (e.g., `PickerFinishedMsg`)
- `<component>InitMsg`: Internal initialization messages (unexported)

## Where to Add New Code
**New Feature:**
- Primary code: `internal/<feature>/` (new package)
- Tests: `internal/<feature>/<feature>_test.go`

**New TUI Screen:**
- Implementation: `internal/tui/<screen>.go`
- Tests: `internal/tui/<screen>_test.go`
- Wire up: Add screen constant and update/view cases in `main.go`

**New CLI Command:**
- Implementation: Add flag in `main.go`, add handler function in `main.go` or new internal package
- Tests: Package-level tests or manual CLI testing

**Utilities:**
- Shared helpers: Within the relevant `internal/` package (no shared utils package exists)

## Special Directories
**testdata/:**
- Purpose: .env fixture files for parser integration tests
- Generated: No
- Committed: Yes

**.planning/:**
- Purpose: Architecture analysis and planning documents
- Generated: Partially (by analysis tools)
- Committed: Yes

**doc/adr/:**
- Purpose: Architecture Decision Records
- Generated: No
- Committed: Yes

**scripts/ralph/:**
- Purpose: Ralph autonomous agent automation scripts
- Generated: No
- Committed: Yes

---
*Structure analysis: 2026-02-08*
