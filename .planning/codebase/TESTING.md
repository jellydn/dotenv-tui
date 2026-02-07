# Testing Patterns

**Analysis Date:** 2026-02-08

## Test Framework
**Runner:**
- Go standard `testing` package
- Go 1.25.6
- Config: No custom test config; uses `go test` defaults

**Assertion Library:**
- Go standard library only (`t.Errorf`, `t.Fatalf`, `t.Logf`)
- No third-party assertion libraries (no testify, no gomega)

**Run Commands:**
```bash
just test               # Run all tests (go test ./...)
just test-v             # Run tests with verbose output (go test -v ./...)
go test -v -run TestFunctionName ./package/path  # Run single test
go test -v -race -coverprofile=coverage.out ./...  # Race detection + coverage
```

## Test File Organization
**Location:**
- Co-located in the same directory as source files
- Same package (white-box testing) for most tests
- Exception: `testdata_test.go` uses `parser_test` package (black-box testing with external imports)

**Naming:**
- Files: `<source>_test.go` (e.g., `parser_test.go`, `detector_test.go`)
- Functions: `Test<FunctionName>` or `Test<TypeMethod>` (e.g., `TestParse`, `TestMenuModelUpdateNavigation`)

**Structure:**
```
internal/
├── parser/
│   ├── parser.go
│   ├── parser_test.go          # White-box tests (package parser)
│   └── testdata_test.go        # Black-box tests (package parser_test)
├── detector/
│   ├── detector.go
│   └── detector_test.go
├── generator/
│   ├── generator.go
│   └── generator_test.go
├── scanner/
│   ├── scanner.go
│   └── scanner_test.go
├── upgrade/
│   ├── upgrade.go
│   └── upgrade_test.go
└── tui/
    ├── menu.go / menu_test.go
    ├── picker.go / picker_test.go
    ├── preview.go / preview_test.go
    ├── form.go / form_test.go
    └── logo.go / logo_test.go
testdata/                        # Shared test fixtures
├── .env.complex
├── .env.example
├── .env.minimal
└── .env.quoted
```

## Test Structure
**Suite Organization:**
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string       // or relevant input fields
        expected string       // or relevant expected fields
    }{
        {
            name:     "descriptive case name",
            input:    "test input",
            expected: "expected output",
        },
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FunctionUnderTest(tt.input)
            if result != tt.expected {
                t.Errorf("FunctionUnderTest(%q) = %v; want %v", tt.input, result, tt.expected)
            }
        })
    }
}
```

**Patterns:**
- **Setup:** Inline in test function or via helper functions with `t.Helper()` (see scanner tests)
- **Teardown:** `t.TempDir()` for automatic cleanup; `defer os.Remove()` for manual cleanup
- **Assertion:** Direct comparison with `if got != want` pattern; `t.Errorf` for non-fatal, `t.Fatalf` for fatal
- **AAA comments:** Some TUI tests use `// Arrange`, `// Act`, `// Assert` comments (menu_test.go, preview_test.go, form_test.go)

## Mocking
**Framework:** No mocking framework

**Patterns:**
```go
// HTTP test servers for network-dependent code (upgrade package)
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"tag_name": "v1.2.3"}`))
}))
defer server.Close()

// Bubble Tea message simulation for TUI tests
msg := tea.KeyMsg{Type: tea.KeyDown}
newModel, cmd := model.Update(msg)
newPickerModel := newModel.(PickerModel)

// Stdout capture for output verification
old := os.Stdout
r, w, _ := os.Pipe()
os.Stdout = w
// ... code that prints ...
_ = w.Close()
os.Stdout = old
output, _ := io.ReadAll(r)
```

## Fixtures and Factories
**Test Data:**
```go
// String readers for parser tests
reader := strings.NewReader("KEY=value\n")
entries, err := Parse(reader)

// Temp directories for scanner tests with helper functions
func writeFile(t *testing.T, base, name, content string) {
    t.Helper()
    path := filepath.Join(base, name)
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatalf("Failed to write file %s: %v", path, err)
    }
}

func mkdir(t *testing.T, base, name string) {
    t.Helper()
    path := filepath.Join(base, name)
    if err := os.MkdirAll(path, 0755); err != nil {
        t.Fatalf("Failed to create directory %s: %v", path, err)
    }
}

// Struct literals for TUI model tests
model := PickerModel{
    items:    []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}},
    selected: map[int]bool{0: true},
    cursor:   0,
    mode:     GenerateExample,
}

// In-memory entries for generator tests
entries := []parser.Entry{
    parser.KeyValue{Key: "PORT", Value: "3000"},
    parser.KeyValue{Key: "API_SECRET", Value: "sk_live_123456789"},
}
```

**Location:**
- `testdata/` directory at project root for real `.env` file fixtures
- Inline test data in `*_test.go` files for most tests
- `strings.NewReader` for parser input, `strings.Builder` for parser output

## Coverage
**Requirements:** None enforced (no minimum coverage threshold)

## Test Types
**Unit Tests:**
- All packages have unit tests
- White-box testing of unexported functions (e.g., `isSecretKey`, `isBase64`, `isHex`, `isCommonNonSecret`, `isEnvFile`, `isPlaceholderValue`, `generateHint`, `entryToString`)
- Table-driven tests are the dominant pattern across all packages
- Tests cover success paths, error/edge cases, and boundary conditions

**Integration Tests:**
- `TestGenerateExampleIntegration` in generator package verifies detector+generator work together
- `TestParseRealEnvFiles` in parser package runs against real fixture files from `testdata/`
- `TestRoundTrip` verifies Parse→Write preserves content exactly

**Benchmark Tests:**
- `BenchmarkGenerateExample` in generator package benchmarks with 1000 entries
- Uses `b.ResetTimer()` after setup

**E2E Tests:**
- Not used (TUI components tested via message simulation, not full program execution)

## Common Patterns
**Error Testing:**
```go
// Table-driven with expectError flag
tests := []struct {
    name        string
    content     string
    expected    string
    expectError bool
}{
    {name: "empty file", content: "", expectError: true},
    {name: "valid hash", content: "abc123", expected: "abc123", expectError: false},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := functionUnderTest(tt.content)
        if tt.expectError && err == nil {
            t.Error("expected error, got nil")
        }
        if !tt.expectError && err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if !tt.expectError && result != tt.expected {
            t.Errorf("got %q, want %q", result, tt.expected)
        }
    })
}
```

**TUI Model Testing:**
```go
// Simulate key press and verify model state
model := NewMenuModel()
msg := tea.KeyMsg{Type: tea.KeyDown}
newModel, cmd := model.Update(msg)
menuModel := newModel.(MenuModel)

if menuModel.choice != expectedChoice {
    t.Errorf("got %v, want %v", menuModel.choice, expectedChoice)
}
if cmd != nil {
    t.Errorf("expected nil command")
}

// Verify command produces correct message
_, cmd := model.Update(enterKey)
if cmd != nil {
    msg := cmd()
    finishedMsg := msg.(PickerFinishedMsg)
    if len(finishedMsg.Selected) != expected {
        t.Errorf("wrong selection count")
    }
}
```

**Boolean Function Testing:**
```go
// Compact table entries for boolean predicates
tests := []struct {
    name     string
    key      string
    expected bool
}{
    {"uppercase secret", "DATABASE_SECRET", true},
    {"lowercase secret", "database_secret", true},
    {"non-secret", "PORT", false},
}
```

---
*Testing analysis: 2026-02-08*
