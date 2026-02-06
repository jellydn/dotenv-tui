# 3. Format-hint placeholders for masked secrets

Date: 2026-02-07

## Status

Accepted

## Context

When generating `.env.example`, secret values must be replaced with placeholders. Common approaches:

- **Empty values** (`API_KEY=`) — gives no hint about expected format.
- **Generic placeholders** (`API_KEY=your_api_key_here` or `API_KEY=<REQUIRED>`) — tells you it's needed but not what it looks like.
- **Format-hint placeholders** (`API_KEY=sk_***`) — preserves the prefix/shape so the developer knows the expected format at a glance.

## Decision

Use format-hint placeholders that preserve recognizable prefixes from the original value:

| Original value | Placeholder |
|---|---|
| `sk_live_abc123def456` | `sk_***` |
| `ghp_xxxxxxxxxxxx` | `ghp_***` |
| `eyJhbGciOiJIUzI1NiJ9...` | `eyJ***` |
| `https://user:pass@host` | `https://***` |
| Unknown pattern | `***` |

Detection is based on both key name patterns (SECRET, TOKEN, PASSWORD, etc.) and value shape (base64, JWT, URL with credentials, hex strings).

## Consequences

### Positive
- Developers immediately know what format/provider a secret belongs to (Stripe, GitHub, JWT, etc.)
- Reduces onboarding friction — no need to ask "what does this key look like?"
- More useful than generic placeholders while still safe to commit

### Negative
- Prefix leaks minimal information about the service provider (generally acceptable since `.env.example` keys already reveal this)
- Heuristic-based detection may not cover all prefix patterns; needs a fallback (`***`)
- Must be kept up-to-date as new providers/formats emerge
