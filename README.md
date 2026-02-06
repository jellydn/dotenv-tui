# dotenv-tui

<p align="center">
  <img src="assets/logo.svg" alt="dotenv-tui logo" width="640" />
</p>

A terminal UI tool for managing `.env` files across projects and monorepos.

## Motivation

Every project has `.env` files. Every `.env` file has secrets. And yet:

- **Developers manually create `.env.example`** — tediously copying keys, guessing which values are safe to keep, and inevitably leaking a token or forgetting a variable.
- **New team members stare at a blank `.env.example`** — no idea what format `STRIPE_SECRET_KEY` expects, no hint whether `DATABASE_URL` needs a port number.
- **Monorepos make it worse** — `.env` files scattered across 10+ packages, each needing its own `.env.example`, each drifting out of sync.

dotenv-tui fixes this with two commands:

1. **`.env` → `.env.example`** — Auto-detects secrets (API keys, tokens, passwords) and masks them with format hints (`sk_***`, `ghp_***`, `eyJ***`) so the next developer knows exactly what shape the value should be. Non-secrets like `PORT=3000` stay as-is.

2. **`.env.example` → `.env`** — Interactive form pre-filled with example values. Just tab through, fill in your secrets, and you're set up.

## Features

- Smart secret detection by key name patterns and value shape
- Format-hint placeholders (`sk_***`, `ghp_***`) instead of useless `<REQUIRED>`
- Recursive monorepo scanning with selectable file list
- Preserves comments, blank lines, and key ordering
- Diff preview before writing `.env.example`
- Supports `.env.local`, `.env.production`, and all `.env.*` variants
- CLI flags for non-interactive / CI usage

## Install

```sh
go install github.com/jellydn/env-man@latest
```

Or build from source:

```sh
git clone https://github.com/jellydn/env-man.git && cd env-man
just build
```

## Usage

Launch the TUI:

```sh
dotenv-tui
```

Non-interactive:

```sh
# Generate .env.example from .env
dotenv-tui --generate-example .env

# Generate .env from .env.example
dotenv-tui --generate-env .env.example

# List discovered .env files
dotenv-tui --scan
```

## Development

```sh
just dev      # Run in development
just build    # Build binary
just test     # Run all tests
just lint     # Run linter
just fmt      # Format code
```

## Tech Stack

- [Go](https://go.dev)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — Input components

## License

MIT
