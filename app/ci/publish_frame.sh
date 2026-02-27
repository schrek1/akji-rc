#!/usr/bin/env bash
set -euo pipefail

FILE="${1:-/tmp/frame.jpg}"
EXPIRY="${EXPIRY:-1h}" # 1h | 12h | 24h | 72h

log() { printf "[%s] %s\n" "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" "$*"; }
fail() { log "ERROR: $*"; exit 1; }

test -s "$FILE" || fail "Missing or empty file: $FILE"
command -v curl >/dev/null 2>&1 || fail "curl not found"

# Litterbox is Catbox's temporary host. We use anonymous upload (no userhash).
# If this endpoint ever changes, swap provider (uguu / 0x0 / transfer.sh etc).
API_URL="https://litterbox.catbox.moe/resources/internals/api.php"

log "Uploading frame (expiry=${EXPIRY})..."

# Returns a URL on success (plain text)
URL="$(
  curl -fsS \
    -F "reqtype=fileupload" \
    -F "time=${EXPIRY}" \
    -F "fileToUpload=@${FILE}" \
    "${API_URL}"
)"

# Basic sanity
[[ "$URL" == http* ]] || fail "Upload did not return a URL: $URL"

log "Upload OK: $URL"

# Print URL only (useful for CI capturing)
printf "%s" "$URL"