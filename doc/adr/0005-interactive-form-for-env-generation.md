# 5. Interactive form for .env generation from .env.example

Date: 2026-02-07

## Status

Accepted

## Context

When a developer clones a repo and needs to create their `.env` from `.env.example`, the options are:

- **Copy and edit** (`cp .env.example .env && $EDITOR .env`) — works but requires switching to an editor, easy to miss variables.
- **Prompted one-by-one** — ask for each value sequentially. Linear but can't go back to fix earlier entries.
- **Editable form** — show all variables at once in a scrollable form with inline editing. Can review and edit any field before submitting.

## Decision

Use an editable form-style TUI built with Bubbles text input components. All variables are displayed in a scrollable list, pre-filled with example values. Placeholder values (containing `your_` or `_here` or `***`) are shown as empty inputs with hint text. Navigation via Tab/arrow keys. No grouping by comment sections — flat list only.

## Consequences

### Positive
- Developer sees all variables at once and can fill them in any order
- Pre-filled non-secret values (PORT, NODE_ENV) reduce manual work
- Hint text from placeholders guides the expected format
- Confirmation prompt prevents accidental overwrites

### Negative
- Large `.env` files (50+ variables) may require scrolling, which can feel cramped in small terminals
- No grouping means related variables aren't visually clustered (though original comments are preserved)
- Building a multi-field form in Bubble Tea requires careful focus/scroll management
