# 1. Use Go with Bubble Tea for TUI

Date: 2026-02-07

## Status

Accepted

## Context

We need a TUI framework to build an interactive terminal tool for managing `.env` files. Options considered:

- **Rust + ratatui** — Fast, strong type system, but steeper learning curve and slower iteration.
- **Go + Bubble Tea** — Elm-architecture TUI framework from Charm. Rich ecosystem (Lip Gloss for styling, Bubbles for input components). Go compiles fast and produces a single static binary.
- **TypeScript + Ink/Clack** — Familiar for JS developers but requires Node runtime, slower startup, heavier distribution.

## Decision

Use Go with the Charm stack: Bubble Tea for the TUI framework, Lip Gloss for styling, and Bubbles for reusable input components (text inputs, lists, spinners).

## Consequences

### Positive
- Single static binary with no runtime dependencies
- Elm architecture enforces clean separation of state, update, and view
- Rich component library (Bubbles) provides text inputs, lists, and spinners out of the box
- Fast compilation and cross-platform builds

### Negative
- Go's error handling is verbose compared to Rust's `Result` type
- No generics-heavy abstractions (though Go 1.18+ generics help)
- Bubble Tea's Elm architecture has a learning curve for developers unfamiliar with the pattern
