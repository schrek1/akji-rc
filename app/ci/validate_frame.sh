#!/usr/bin/env bash
set -euo pipefail

FILE="${1:-/tmp/frame.jpg}"

log() { printf "[%s] %s\n" "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" "$*"; }
fail() { log "ERROR: $*"; exit 1; }

test -s "$FILE" || fail "Missing or empty file: $FILE"

SIZE="$(wc -c < "$FILE" | tr -d ' ')"
log "Image size: ${SIZE} bytes"

# Tune this to your camera. Start safe (15 KB).
MIN_SIZE="${MIN_SIZE_BYTES:-15000}"
if [[ "$SIZE" -lt "$MIN_SIZE" ]]; then
  fail "File too small (${SIZE} < ${MIN_SIZE}) - likely not a real frame"
fi

# JPEG markers: SOI (FFD8) and EOI (FFD9)
SOI="$(od -An -t x1 -N 2 "$FILE" | xargs | tr -d ' ')"
EOI="$(tail -c 2 "$FILE" | od -An -t x1 | head -n 1 | xargs | tr -d ' ')"

[[ "$SOI" == "ffd8" ]] || fail "Invalid JPEG SOI marker: $SOI"
[[ "$EOI" == "ffd9" ]] || fail "Invalid JPEG EOI marker: $EOI"

# MIME check (catches HTML/text pages)
if command -v file >/dev/null 2>&1; then
  MIME="$(file --mime-type -b "$FILE" || true)"
  [[ "$MIME" == "image/jpeg" ]] || fail "Invalid mime-type: $MIME"
fi

# Quick HTML guard (some cameras return login pages)
if head -c 512 "$FILE" | strings | grep -qiE '<html|<!doctype'; then
  fail "Looks like HTML page, not a JPEG"
fi

log "Validation OK"