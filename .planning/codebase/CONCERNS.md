# Codebase Concerns

**Analysis Date:** 2026-02-08

## Tech Debt

**Regex compiled on every call in detector:**
- Issue: `regexp.MatchString` is called in `isHex()` which compiles the regex on every invocation instead of using a pre-compiled `regexp.Regexp`
- Files: `internal/detector/detector.go:165`
- Impact: Performance degradation when processing many entries with hex-like values; unnecessary allocations per call
- Fix approach: Use a package-level `var hexPattern = regexp.MustCompile(...)` and call `hexPattern.MatchString(s)`

**Hardcoded GitHub API URL — not injectable for testing:**
- Issue: `getLatestVersion()` uses a hardcoded `githubAPIURL` constant making it impossible to inject a test server. The upgrade tests create httptest servers but never actually exercise `getLatestVersion()` against them (the test bodies are empty stubs)
- Files: `internal/upgrade/upgrade.go:19`, `internal/upgrade/upgrade_test.go:93-135`
- Impact: `getLatestVersion()` is effectively untested; upgrade integration path has no real coverage
- Fix approach: Accept an HTTP client or base URL as a parameter, or use a package-level variable that tests can override

**Duplicated `entryToString` logic:**
- Issue: `entryToString()` in `internal/tui/preview.go:108-129` duplicates the serialization logic in `parser.Write()`. If the format changes, both must be updated
- Files: `internal/tui/preview.go`, `internal/parser/parser.go`
- Impact: Risk of inconsistent output between preview and actual file write
- Fix approach: Add an `EntryToString(Entry) string` function in the parser package and reuse it

**`main.go` contains significant business logic:**
- Issue: CLI handler functions (`generateFile`, `generateEnvFile`, `generateAllEnvFiles`, `processExampleFile`, `scanAndList`) are defined in `main.go` (366-506) making them untestable — root package has 0% test coverage
- Files: `main.go:366-506`
- Impact: No test coverage for CLI workflows, file overwrite logic, or yolo mode
- Fix approach: Move CLI command handlers into an `internal/cli` package

## Known Bugs

**Empty upgrade test bodies:**
- Symptoms: `TestGetLatestVersion` subtests for "network error", "non-200 status code", "empty tag name", and "invalid JSON" create httptest servers but have empty test bodies — they never call `getLatestVersion` or assert anything
- Files: `internal/upgrade/upgrade_test.go:104-135`
- Trigger: Run tests — they pass vacuously
- Workaround: None needed (no runtime bug), but false confidence in test coverage

**Picker silently swallows scanner errors:**
- Symptoms: If `scanner.Scan()` or `scanner.ScanExamples()` returns an error, `NewPickerModel` silently replaces the result with an empty list with no user feedback
- Files: `internal/tui/picker.go:88-89`
- Trigger: Permission denied or filesystem error during scan
- Workaround: User sees "No .env files found" message, which is misleading

## Security Considerations

**Upgrade downloads over HTTPS but no TLS certificate pinning:**
- Risk: Man-in-the-middle attack on the upgrade binary download. Checksum verification helps but the checksum itself is also fetched over the same channel
- Files: `internal/upgrade/upgrade.go:154-172`, `install.sh:75-86`
- Current mitigation: SHA256 checksum verification when available
- Recommendations: Consider code signing; warn if checksum file is unavailable (currently only prints to stdout, which is suppressed in non-TTY contexts)

**`http.Get` without timeout:**
- Risk: `http.Get` uses `http.DefaultClient` which has no timeout, allowing requests to hang indefinitely (e.g., during upgrade or version check)
- Files: `internal/upgrade/upgrade.go:132,176`
- Current mitigation: None
- Recommendations: Use an `http.Client` with explicit `Timeout` (e.g., 30s)

**File creation with `os.Create` uses default permissions (0666 minus umask):**
- Risk: Generated `.env` files containing secrets may be world-readable depending on the user's umask
- Files: `main.go:386,489`, `internal/tui/preview.go:210`, `internal/tui/form.go:300`
- Current mitigation: Relies on user's umask
- Recommendations: Use `os.OpenFile` with explicit `0600` permissions for `.env` files containing secrets

**Upgrade binary replacement race condition:**
- Risk: `replaceBinary` does a rename-or-copy. Between removing the old binary and writing the new one (in the copy fallback path), the binary is briefly missing or partially written
- Files: `internal/upgrade/upgrade.go:250-258`
- Current mitigation: Atomic rename attempted first; copy is fallback
- Recommendations: Write to temp file in same directory, then atomic rename (avoids cross-filesystem copy fallback issues)

## Performance Bottlenecks

**`regexp.MatchString` per hex check:**
- Problem: Compiles a new regex on every call to `isHex()`
- Files: `internal/detector/detector.go:165`
- Cause: `regexp.MatchString` compiles the pattern each time
- Improvement path: Pre-compile with `regexp.MustCompile` at package level, or replace with a simple loop checking `unicode.IsDigit`/hex range

**Recursive directory walk for file scanning:**
- Problem: `filepath.WalkDir` traverses the entire directory tree on every scan
- Files: `internal/scanner/scanner.go:27-57`
- Cause: Full recursive walk, no caching, path split + map lookup per directory entry
- Improvement path: Acceptable for typical projects; for massive trees, consider depth limits or parallel walking

## Fragile Areas

**Type assertions on Bubble Tea models:**
- Files: `main.go:96,111,145,162` (e.g., `m.menu.Update(msg)` → `menuModel.(tui.MenuModel)`)
- Why fragile: Unchecked type assertions will panic if the Update method ever returns a different type. All screen update functions use bare type assertions without `,ok` checks
- Safe modification: Add `,ok` checks or use typed model methods
- Test coverage: No tests for `main.go` model transitions at all (0% coverage on root package)

**Parser does not handle multiline values:**
- Files: `internal/parser/parser.go:31-68`
- Why fragile: The parser reads line-by-line and does not support multiline quoted values (e.g., `KEY="line1\nline2"`), heredoc syntax, or values with literal newlines. Real `.env` files from tools like Docker Compose may use these
- Safe modification: Would require switching to a state-machine parser
- Test coverage: No multiline test cases exist

**`visibleDiffLines` is hardcoded to 10:**
- Files: `internal/tui/preview.go:136`
- Why fragile: Does not adapt to terminal height (unlike picker which uses `windowHeight`). Small terminals will still try to show 10 lines; large terminals waste space
- Safe modification: Accept window height and compute dynamically
- Test coverage: No tests for scroll behavior in preview with varying window sizes

## Scaling Limits

**Scanner directory traversal:**
- Current capacity: Works well for typical projects (hundreds of files)
- Limit: Very deep or wide directory trees (e.g., monorepos with 100k+ files) will be slow since it walks everything
- Scaling path: Add max-depth flag, or use `.gitignore`-aware scanning

**Form field rendering:**
- Current capacity: `visibleFields = 7` with scrolling
- Limit: `.env` files with hundreds of keys will work but the UI only shows 7 at a time with no search/filter capability
- Scaling path: Add search/filter within the form, or group by section

## Dependencies at Risk

**Charm libraries (bubbletea, bubbles, lipgloss):**
- Risk: These are actively maintained but the Charm ecosystem has undergone major API changes in past versions (e.g., lipgloss v1.0 breaking changes). Currently on `bubbletea v1.3.10`, `lipgloss v1.1.0`
- Impact: Major version bumps could require significant refactoring of all TUI components
- Migration plan: Pin versions; monitor changelogs. The app's TUI code is well-structured enough to absorb API changes

**No other direct dependencies besides Charm stack:**
- Risk: Low — minimal dependency surface. All core logic (parser, detector, generator, scanner) uses only the standard library
- Impact: N/A
- Migration plan: N/A

## Missing Critical Features

**No `.env` file validation:**
- Problem: The tool generates and writes `.env` files but never validates that the output is syntactically correct or that required keys are present
- Blocks: Users could accidentally write corrupt `.env` files (e.g., empty required values)

**No backup before overwrite:**
- Problem: When using `--force` or confirming overwrite in TUI/yolo mode, the existing file is overwritten without creating a backup
- Blocks: Accidental data loss of existing `.env` files with real secrets

**No `--dry-run` flag for CLI operations:**
- Problem: No way to preview what would happen without side effects when using `--generate-example`, `--generate-env`, or `--yolo`
- Blocks: Safe usage in CI/CD pipelines

## Test Coverage Gaps

**Root package (`main.go`) — 0% coverage:**
- What's not tested: All CLI flag handling, `generateFile`, `generateExampleFile`, `generateEnvFile`, `scanAndList`, `generateAllEnvFiles`, `processExampleFile`, all Bubble Tea model transitions in `updateMenu`/`updatePicker`/`updatePreview`/`updateForm`/`updateDone`
- Files: `main.go`
- Risk: CLI workflow bugs (wrong flag combinations, file overwrite logic, yolo mode edge cases) go undetected
- Priority: High

**TUI package — 47.6% coverage:**
- What's not tested: `View()` methods (visual rendering), `saveForm()` file write, `writeFile()` in preview, `NewPickerModel` (the tea.Cmd factory that calls scanner), `NewFormModel` (the tea.Cmd factory that parses files), `NewPreviewModel` (the tea.Cmd factory)
- Files: `internal/tui/form.go`, `internal/tui/preview.go`, `internal/tui/picker.go`
- Risk: File I/O in TUI components (save/write operations) untested; visual regressions undetected
- Priority: Medium

**Upgrade package — 46.5% coverage:**
- What's not tested: `Upgrade()` integration flow, `getLatestVersion()` (test stubs are empty), `verifyChecksum()` end-to-end with real hash comparison, `replaceBinary` cross-filesystem fallback
- Files: `internal/upgrade/upgrade.go`, `internal/upgrade/upgrade_test.go`
- Risk: Self-update mechanism could silently fail or corrupt the binary
- Priority: High

**Parser edge cases:**
- What's not tested: Multiline values, values containing `=` signs, inline comments (`KEY=value # comment`), empty keys, BOM-prefixed files, Windows line endings (`\r\n`)
- Files: `internal/parser/parser.go`
- Risk: Parser fails silently on real-world `.env` files from other tools
- Priority: Medium

---
*Concerns audit: 2026-02-08*
