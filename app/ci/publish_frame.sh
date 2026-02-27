#!/usr/bin/env bash
set -euo pipefail

FILE="${1:-/tmp/frame.jpg}"

log() { printf "[%s] %s\n" "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" "$*"; }
fail() { log "ERROR: $*"; exit 1; }

test -s "$FILE" || fail "Missing or empty file: $FILE"
command -v curl >/dev/null 2>&1 || fail "curl not found"
command -v python3 >/dev/null 2>&1 || fail "python3 not found"

API_URL="https://uguu.se/upload"

log "Uploading frame to Uguu..."

MAX_RETRIES=3
RETRY_DELAY=5
URL=""

for (( i=1; i<=MAX_RETRIES; i++ )); do
  RESP="$(
    curl -fsS \
      -F "files[]=@${FILE}" \
      "${API_URL}" || true
  )"

  # Try to parse JSON safely
  RESP_FILE="upload_resp_$$.json"
  echo "$RESP" > "$RESP_FILE"
  URL="$(
    python3 - "$RESP_FILE" <<'PY' 2>/dev/null || true
import json, sys
try:
    with open(sys.argv[1], 'r') as f:
        data = json.load(f)
    files = data.get("files") or []
    if files and isinstance(files, list):
        url = files[0].get("url") or files[0].get("fileUrl") or files[0].get("link")
        if url:
            print(url)
except Exception:
    pass
PY
  )"
  rm -f "$RESP_FILE"

  if [[ "$URL" == http* ]]; then
    break
  fi

  log "Upload attempt $i failed. Retrying in ${RETRY_DELAY}s..."
  sleep "${RETRY_DELAY}"
done

[[ "$URL" == http* ]] || fail "Upload failed after ${MAX_RETRIES} attempts"

log "Upload OK: $URL"

# Print URL only (for CI output capture)
printf "%s" "$URL"