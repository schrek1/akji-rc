# Architectural Decisions - AKJI-RC

## Current Scope

The MVP is a Go pipeline that captures one JPEG frame from an MJPEG webcam, validates it, and publishes a temporary URL. There is no backend or UI yet.

## Why Go?

- One portable binary per command without runtime shell-tool dependencies.
- Explicit configuration, networking, validation, and error handling.
- Fast deterministic tests for MJPEG parsing, JPEG validation, and multipart publishing.
- A small codebase that stays easy to inspect and maintain.

## Public Repository Strategy

- Credentials are provided only through process environment variables or a local `.env` file ignored by Git.
- `.env.example` contains placeholders only.
- CI reads webcam credentials from GitHub variables and secrets.
- Commands never print credentials.

## Configuration Strategy

- `.env` provides local development defaults for capture.
- Explicit process environment variables override capture values from `.env`.
- The validator reads `MIN_SIZE_BYTES` directly from the process environment.
- This keeps deployment-specific changes and CI secrets out of version control.

## Storage Strategy

Temporary Uguu hosting is the current publisher. S3-compatible storage remains the future target, with Cloudflare R2 a suitable free-tier starting point.

## Future UI Direction

The intended lightweight stack is Go + Svelte. Svelte is not part of the current MVP.
