# 4. Recursive scan with user confirmation via TUI

Date: 2026-02-07

## Status

Accepted

## Context

In monorepos, `.env` files are scattered across many packages. We need a strategy to discover them:

- **Manual path specification** — user provides explicit paths. Safe but tedious in large repos.
- **Automatic recursive scan** — finds everything but may surface unwanted files (e.g., in test fixtures or vendored code).
- **Workspace config parsing** — reads `pnpm-workspace.yaml` or `turbo.json` to find packages. Accurate but couples to specific tools.
- **Scan + confirm** — recursively scan, then let the user review and select which files to process.

## Decision

Recursively scan from the current directory for all `.env.*` variants (`.env`, `.env.local`, `.env.production`, etc.), excluding `.env.example` and `.env.*.example`. Skip common dependency directories (`node_modules`, `.git`, `vendor`, `dist`, `build`, `.next`, `.nuxt`, `__pycache__`). Present results in a TUI checklist where the user can select/deselect files before proceeding.

## Consequences

### Positive
- Works with any project structure — not coupled to specific monorepo tools
- User has full control over which files get processed
- Scanning is fast (just a directory walk with filters)
- Supports all `.env.*` variants out of the box

### Negative
- May surface unexpected `.env` files in deeply nested directories
- Skip-list of directories is hardcoded and may need extending for uncommon setups
- No awareness of workspace boundaries (treats all found files equally)
