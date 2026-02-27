# Architectural Decisions — AKJI-RC

## MVP1 Scope

Capture → Upload to S3-compatible storage.
No backend. No UI.

## Why Bash?

- No build tooling required
- Works natively in GitHub Actions
- Minimal surface area
- Easy to keep secrets out of logs

## Public Repository Strategy

- Credentials only via environment variables
- GitHub Secrets for CI
- Never print secrets
- Avoid verbose shell debugging

## Storage Strategy

Default target: S3-compatible storage.
Cloudflare R2 recommended for free-tier start.