# External Integrations

**Analysis Date:** 2026-02-08

## APIs & External Services
**GitHub API:**
- GitHub Releases API - Self-upgrade feature (`--upgrade` flag)
- SDK/Client: Go `net/http` standard library (no third-party SDK)
- Auth: None (public API, unauthenticated requests)
- Endpoint: `https://api.github.com/repos/jellydn/dotenv-tui/releases/latest`
- Binary downloads: `https://github.com/jellydn/dotenv-tui/releases/download/`

## Data Storage
**Databases:**
- None - No database used

**File Storage:**
- Local filesystem only
- Reads/writes `.env` and `.env.example` files in user-specified directories
- Binary self-replacement during upgrade (writes to executable path)

**Caching:**
- None

## Authentication & Identity
**Auth Provider:**
- None - CLI tool with no authentication required

## Monitoring & Observability
**Error Tracking:**
- None

**Logs:**
- Stderr for error output (`fmt.Fprintf(os.Stderr, ...)`)
- Stdout for informational messages during CLI operations

## CI/CD & Deployment
**Hosting:**
- GitHub Releases - Pre-built binaries for Linux/macOS/Windows
- Install script (`install.sh`) downloads from GitHub Releases

**CI Pipeline:**
- GitHub Actions
  - `ci.yml` - Test (ubuntu + macos matrix), Lint (golangci-lint + gofmt), Build (ubuntu + macos matrix)
  - `release.yml` - Tag-triggered release with cross-compiled binaries (5 targets), SHA256 checksums
- Codecov - Coverage uploads (via `codecov/codecov-action@v5`)

**Dependency Management:**
- Dependabot - Weekly gomod + github-actions updates
- Renovate - Automated dependency updates (`config:recommended`)

## Environment Configuration
**Required env vars:**
- None - No environment variables required

**Optional env vars:**
- `INSTALL_DIR` - Custom install directory for `install.sh` (default: `~/.local/bin`)

**CI secrets:**
- `GITHUB_TOKEN` - Used by release workflow for publishing releases (auto-provided by GitHub Actions)

**Secrets location:**
- GitHub Actions secrets (CI only)
- No application-level secrets

## Webhooks & Callbacks
**Incoming:**
- None

**Outgoing:**
- None

---
*Integration audit: 2026-02-08*
