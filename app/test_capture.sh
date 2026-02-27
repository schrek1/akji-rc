#!/usr/bin/env bash
#
# Regression tests for akji_webcam_captor/capture.sh
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CAPTURE_SCRIPT="$SCRIPT_DIR/capture.sh"
TEST_WORK_DIR="$SCRIPT_DIR/test_tmp"

# Helper for colorful output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

log_test() {
    echo -e "${GREEN}[TEST]${NC} $*"
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $*"
    test_cleanup
    exit 1
}

# Move real .env files if they exist to avoid interference
if [[ -f "$SCRIPT_DIR/.env" ]]; then mv "$SCRIPT_DIR/.env" "$SCRIPT_DIR/.env.bak"; fi

# Cleanup and setup
test_cleanup() {
    # Restore .env files
    if [[ -f "$SCRIPT_DIR/.env.bak" ]]; then mv "$SCRIPT_DIR/.env.bak" "$SCRIPT_DIR/.env"; fi
    
    # On Windows, deleting the directory from within the script might fail if bash has it open.
    if [[ -d "$TEST_WORK_DIR" ]]; then
        cd "$SCRIPT_DIR" || exit
        rm -rf "$TEST_WORK_DIR"
    fi
}
trap test_cleanup EXIT
trap test_cleanup SIGINT SIGTERM

mkdir -p "$TEST_WORK_DIR"

# 1. Test Help Message
log_test "Testing help message..."
if bash "$CAPTURE_SCRIPT" --help | grep -q "Usage:"; then
    log_test "Help message OK."
else
    log_fail "Help message missing or incorrect."
    exit 1
fi

# 2. Test Argument Parsing (Loop Mode)
log_test "Testing argument parsing (-tl 5)..."
# We source the script but we need to mock CAPTURES_DIR or it might try to create it in the real location
# Actually, sourcing it will run the argument parsing block but not the main execution block now.
# But it will use the current $1 $2... as arguments.
(
    set -- -tl 5
    source "$CAPTURE_SCRIPT"
    if [[ "$LOOP_MODE" == true ]] && [[ "$INTERVAL" == 5 ]]; then
        exit 0
    else
        exit 1
    fi
)
log_test "Argument parsing OK."

# 3. Test extract_jpeg with Mock Data
log_test "Testing JPEG extraction from mock MJPEG..."

MOCK_MJPEG="$TEST_WORK_DIR/mock.mjpg"
MOCK_OUTPUT="$TEST_WORK_DIR/output.jpg"

# Create a mock MJPEG file with:
# [junk] [SOI] [ImageData] [EOI] [junk]
# SOI: ff d8 ff
# EOI: ff d9

printf "some junk\xff\xd8\xffimage data here\xff\xd9more junk" > "$MOCK_MJPEG"

# Source to get extract_jpeg function
# We need to set CAPTURES_DIR etc so functions don't fail on unset vars if they use them
source "$CAPTURE_SCRIPT"

if extract_jpeg "$MOCK_MJPEG" "$MOCK_OUTPUT"; then
    log_test "Extraction reported success."
    # Check if it actually saved a file and if it has the right content
    # In our case, extract_jpeg extracts from SOI to EOI.
    # start is at offset 9 (length of "some junk")
    # end is at offset 29 (start of "more junk" is at 31, ff d9 is at 29,30)
    # length should be 29-9+2 = 22
    if [[ -s "$MOCK_OUTPUT" ]]; then
        log_test "Output file created."
        # Verify content starts with SOI and ends with EOI
        if [[ "$(head -c 3 "$MOCK_OUTPUT" | od -An -t x1 | xargs)" == "ff d8 ff" ]] && \
           [[ "$(tail -c 2 "$MOCK_OUTPUT" | od -An -t x1 | xargs)" == "ff d9" ]]; then
            log_test "Output content valid."
        else
            log_fail "Output content invalid: $(od -An -t x1 "$MOCK_OUTPUT")"
            exit 1
        fi
    else
        log_fail "Output file empty or missing."
        exit 1
    fi
else
    log_fail "Extraction failed on valid mock data."
    exit 1
fi

# 4. Test extract_jpeg with Malformed Data (Missing EOI)
log_test "Testing extraction with missing EOI..."
printf "some junk\xff\xd8\xffimage data without end" > "$MOCK_MJPEG"
rm -f "$MOCK_OUTPUT"

if extract_jpeg "$MOCK_MJPEG" "$MOCK_OUTPUT"; then
    log_fail "Extraction should have failed but reported success."
    exit 1
else
    log_test "Extraction correctly failed on missing EOI."
    if [[ ! -f "$MOCK_OUTPUT" ]]; then
        log_test "Output file correctly cleaned up."
    else
        log_fail "Output file was NOT cleaned up."
        exit 1
    fi
fi

# 5. Test Multiple Frames (Middle Frame Strategy)
log_test "Testing middle frame strategy..."
# Frame 1: Junk1 [SOI1] Data1 [EOI1]
# Frame 2: Junk2 [SOI2] Data2 [EOI2]
# Frame 3: Junk3 [SOI3] Data3 [EOI3]
# Script should pick Frame 2 (mid = 3/2 = 1? wait. num_sois=3, mid=3/2=1. indices=("1" 1).
# In bash $((3/2)) is 1. 
# sois indices: 1, 2, 3
# target_indices: ("1" 1) -> so index 1 (the first one).
# Wait, let's check the code:
# local mid=$((num_sois / 2))
# [[ $mid -lt 1 ]] && mid=1
# target_indices=("$mid" 1)
# if num_sois is 3, mid is 1. target_indices is ("1" 1).
# if num_sois is 4, mid is 2. target_indices is ("2" 1).
# So for 3 frames, it picks the first one. For 4+ it starts picking from middle.
# Let's test with 4 frames.
# Frame 1: AA [SOI] F1 [EOI]
# Frame 2: BB [SOI] F2 [EOI]
# Frame 3: CC [SOI] F3 [EOI]
# Frame 4: DD [SOI] F4 [EOI]
# mid = 4/2 = 2. It should pick index 2 (Frame 2).

printf "AA\xff\xd8\xffF1\xff\xd9BB\xff\xd8\xffF2\xff\xd9CC\xff\xd8\xffF3\xff\xd9DD\xff\xd8\xffF4\xff\xd9" > "$MOCK_MJPEG"
rm -f "$MOCK_OUTPUT"

if extract_jpeg "$MOCK_MJPEG" "$MOCK_OUTPUT"; then
    if grep -q "F2" "$MOCK_OUTPUT"; then
        log_test "Correctly picked middle frame (Frame 2)."
    else
        log_fail "Did not pick expected frame. Content: $(cat "$MOCK_OUTPUT")"
        exit 1
    fi
else
    log_fail "Extraction failed on multiple frames."
    exit 1
fi

# 6. Test --out parameter
log_test "Testing --out parameter..."
OUT_TEST_FILE="$TEST_WORK_DIR/custom_out.jpg"
(
    # We can't easily mock functions when running 'bash script.sh'
    # But we can test the argument parsing block by sourcing.
    OUT_FILE=""
    set -- --out "$OUT_TEST_FILE"
    # shellcheck disable=SC1091
    source "$CAPTURE_SCRIPT"
    if [[ "$OUT_FILE" == "$OUT_TEST_FILE" ]]; then
        log_test "--out parameter parsing OK."
    else
        log_fail "--out parameter parsing failed. OUT_FILE=$OUT_FILE"
        exit 1
    fi
)

# 7. Test ENV variables
log_test "Testing ENV variables..."
(
    # Clear vars to ensure they come from ENV
    unset WEBCAM_URL WEBCAM_USER WEBCAM_PASS USER PASS
    export WEBCAM_URL="http://env-url"
    export WEBCAM_USER="env-user"
    export WEBCAM_PASS="env-pass"
    # Source to check if they are picked up
    # shellcheck disable=SC1091
    source "$CAPTURE_SCRIPT"
    if [[ "$WEBCAM_URL" == "http://env-url" ]] && [[ "$WEBCAM_USER" == "env-user" ]] && [[ "$WEBCAM_PASS" == "env-pass" ]]; then
        log_test "ENV variables OK."
    else
        log_fail "ENV variables not picked up correctly. URL: $WEBCAM_URL, USER: $WEBCAM_USER, PASS: $WEBCAM_PASS"
    fi
)

# 8. Test .env file loading
log_test "Testing .env file loading..."
(
    cat <<EOF > "$TEST_WORK_DIR/.env"
WEBCAM_URL="http://dot-env-url"
WEBCAM_USER="dot-env-user"
WEBCAM_PASS="dot-env-pass"
EOF
    # Go to the directory where .env is
    cd "$TEST_WORK_DIR"
    
    # Source the script from the test work dir (where .env is)
    # The script should pick up $(pwd)/.env
    # shellcheck disable=SC1091
    source "$CAPTURE_SCRIPT"
    
    if [[ "$WEBCAM_URL" == "http://dot-env-url" ]] && [[ "$WEBCAM_USER" == "dot-env-user" ]] && [[ "$WEBCAM_PASS" == "dot-env-pass" ]]; then
        log_test ".env file loading OK."
    else
        log_fail ".env file NOT loaded correctly."
        log_fail "URL: $WEBCAM_URL, USER: $WEBCAM_USER, PASS: $WEBCAM_PASS"
        exit 1
    fi
)

# 9. Test CI usage (no captures/ dir)
log_test "Testing CI usage (no captures/ directory)..."
CI_OUT_FILE="$TEST_WORK_DIR/ci_test.jpg"
(
    # Using a fake TMPDIR to ensure TMP_MJPEG is also elsewhere
    export TMPDIR="$TEST_WORK_DIR"
    
    # Mock curl to avoid network access
    # Since we are on Windows, we need to be careful with paths and how bash executes things.
    # In Git Bash environment, we can create a mock curl script.
    mkdir -p "$TEST_WORK_DIR/bin"
    # shellcheck disable=SC2016
    printf '#!/usr/bin/env bash\nfor arg in "$@"; do if [[ "$arg" == */stream_capture_*.mjpg ]]; then out="$arg"; fi; done\necho -en "\\xff\\xd8\\xffCI_DATA\\xff\\xd9" > "$out"\nexit 0' > "$TEST_WORK_DIR/bin/curl"
    chmod +x "$TEST_WORK_DIR/bin/curl"
    export PATH="$TEST_WORK_DIR/bin:$PATH"
    
    # Run in a clean environment (no 'captures' dir in current work dir)
    mkdir -p "$TEST_WORK_DIR/run"
    cd "$TEST_WORK_DIR/run"
    
    # We must call it with full path or from its dir since we changed directory
    if WEBCAM_URL="mock" WEBCAM_USER="mock" WEBCAM_PASS="mock" bash "$CAPTURE_SCRIPT" --out "$CI_OUT_FILE"; then
        if [[ -s "$CI_OUT_FILE" ]]; then
            log_test "CI usage with --out successful."
        else
            log_fail "CI output file empty."
        fi
    else
        log_fail "CI usage failed."
    fi
    
    # Check that 'captures/' was not created in the current directory
    if [[ ! -d "captures" ]]; then
        log_test "captures/ directory was NOT created (as expected)."
    else
        log_fail "captures/ directory WAS created in $(pwd)."
    fi
)

log_test "All regression tests passed!"
test_cleanup
exit 0
