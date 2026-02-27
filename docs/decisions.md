# Architectural Decisions — AKJI-RC

## MVP1 Scope

Capture → Upload to S3-compatible storage.
No backend. No UI.

## Why Bash?

- No build tooling required
- Works natively in GitHub Actions
- Minimal surface area
- Easy to keep secrets out of logs
- High compatibility with coreutils for binary data processing

## Public Repository Strategy

- Credentials must be provided only via environment variables or local `.env` files (which are excluded from git).
- No hardcoded secret values (URL, user, password) should ever appear in committed script files or templates.
- `.env.template` contains placeholders for mandatory variables.
- GitHub Secrets for CI
- Never print secrets
- Avoid verbose shell debugging

## Storage Strategy

Default target: S3-compatible storage.
Cloudflare R2 recommended for free-tier start.
## Configuration Strategy

- Support loading from `.env` for local development convenience.
- Prioritize direct environment variables over `.env` file.
