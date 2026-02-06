# 2. Custom .env parser preserving file structure

Date: 2026-02-07

## Status

Accepted

## Context

We need to parse `.env` files and write them back. Existing Go libraries (`joho/godotenv`, `caarlos0/env`) parse key-value pairs but discard comments, blank lines, and ordering. Our tool must produce `.env.example` files that mirror the original structure so developers can recognize and maintain them.

## Decision

Build a custom `.env` parser in `internal/parser/` that returns an ordered list of entries, where each entry is one of: `KeyValue` (key, value, quote style), `Comment`, or `BlankLine`. The writer reproduces the exact original structure.

Supported formats:
- `KEY=VALUE`, `KEY="VALUE"`, `KEY='VALUE'`
- `export KEY=VALUE`
- `# comment` lines
- Blank lines

## Consequences

### Positive
- Full round-trip fidelity: parse → modify → write preserves comments, blank lines, and key order
- No external dependency for a core function
- Can extend to support additional formats (e.g., multiline values) later

### Negative
- Must maintain and test our own parser instead of using a battle-tested library
- Edge cases (multiline values, escaped quotes, inline comments) need careful handling
