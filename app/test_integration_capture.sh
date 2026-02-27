#!/usr/bin/env bash
#
# Integration test for akji-rc
# Verifies capture against the REAL webcam.
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CAPTURE_SCRIPT="$SCRIPT_DIR/capture.sh"
ENV_FILE="$SCRIPT_DIR/.env"
TEST_OUT="$SCRIPT_DIR/integration_test_result.jpg"

# Helper for colorful output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

log_test() {
    echo -e "${GREEN}[INTEGRATION TEST]${NC} $*"
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $*"
    exit 1
}

# 1. Check if configuration is available
if [[ -z "${WEBCAM_URL:-}" ]] || [[ -z "${WEBCAM_USER:-}" ]] || [[ -z "${WEBCAM_PASS:-}" ]]; then
    if [[ -f "$ENV_FILE" ]]; then
        log_test "Loading configuration from $ENV_FILE"
        # We don't source here to avoid polluting the shell, 
        # but capture.sh will source it anyway.
        # This check is just to ensure we have credentials somewhere.
    else
        log_fail "Webcam credentials not found in environment and $ENV_FILE is missing."
    fi
fi

# 2. Run capture
log_test "Running capture against real webcam..."
if bash "$CAPTURE_SCRIPT" --out "$TEST_OUT"; then
    log_test "Capture command executed successfully."
else
    log_fail "Capture command failed."
fi

# 3. Verify file exists and has content
if [[ -s "$TEST_OUT" ]]; then
    FILE_SIZE=$(wc -c < "$TEST_OUT" | tr -d ' ')
    log_test "Captured image size: $((FILE_SIZE / 1024)) KB ($FILE_SIZE bytes)."
else
    log_fail "Captured image is empty or missing."
fi

# 4. Verify image size is around 22kB
# We'll allow a range from 15kB to 40kB to be safe but specific enough.
MIN_SIZE=$((15 * 1024))
MAX_SIZE=$((40 * 1024))

if [[ $FILE_SIZE -ge $MIN_SIZE ]] && [[ $FILE_SIZE -le $MAX_SIZE ]]; then
    log_test "Image size is within expected range (around 22kB)."
else
    log_fail "Image size $FILE_SIZE is outside expected range ($MIN_SIZE - $MAX_SIZE bytes)."
fi

# 5. Verify JPEG markers
SOI=$(od -An -t x1 -N 2 "$TEST_OUT" | xargs | tr -d ' ')
EOI=$(tail -c 2 "$TEST_OUT" | od -An -t x1 | head -n 1 | xargs | tr -d ' ')

if [[ "$SOI" == "ffd8" ]] && [[ "$EOI" == "ffd9" ]]; then
    log_test "JPEG markers (SOI/EOI) are valid."
else
    log_fail "Invalid JPEG markers. SOI: $SOI (expected ffd8), EOI: $EOI (expected ffd9)."
fi

# Cleanup
rm -f "$TEST_OUT"

log_test "Integration test passed successfully!"
exit 0
