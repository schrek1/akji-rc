#!/usr/bin/env bash
#
# akji_webcam_captor - captures a single image from MJPEG stream
# This script accesses a webcam via basic auth, downloads a portion of the MJPEG stream,
# and extracts a valid JPEG frame.
#
# Supported modes:
# - Single capture: Run without parameters.
# - Time-lapse: Run with -tl <seconds> or --timeLapse <seconds>.

set -euo pipefail

# --- LOGGING HELPERS ---
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $*"
}

log_err() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - ERROR: $*" >&2
}

# --- DIRECTORY SETUP ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CAPTURES_DIR="$SCRIPT_DIR/captures"

# --- CONFIGURATION ---
# Load local .env if it exists.
# Variables from files will NOT override existing environment variables.
# Priority: ENV > .env
for env_file in "$SCRIPT_DIR/.env" "$(pwd)/.env"; do
    if [[ -f "$env_file" ]]; then
        # shellcheck disable=SC1091
        source "$env_file"
    fi
done

WEBCAM_URL="${WEBCAM_URL:-}"
WEBCAM_USER="${WEBCAM_USER:-}"
WEBCAM_PASS="${WEBCAM_PASS:-}"
TIMEOUT="${TIMEOUT:-5}"
# Seconds to capture MJPEG data to ensure we have at least one complete frame
CAPTURE_WINDOW="${CAPTURE_WINDOW:-2}"
DEFAULT_INTERVAL="${DEFAULT_INTERVAL:-15}"

# Using PID-unique name for the temp buffer to prevent conflicts
TMP_MJPEG="${TMPDIR:-/tmp}/stream_capture_$$.mjpg"

# --- CLEANUP ON EXIT ---
cleanup() {
    rm -f "$TMP_MJPEG"
}
trap cleanup EXIT

# --- HELP MESSAGE ---
show_help() {
    cat <<EOF
Usage: $0 [OPTIONS]

KISS MJPEG webcam capture script. Extracts one still JPEG frame from a stream.

Options:
  -o, --out <file>       Output file path. (Default: captures/webcam_<TIMESTAMP>.jpg)
  -tl, --timeLapse <N>   Run in loop, capturing every N seconds.
  -h, --help             Show this help message.

If no options are provided, the script captures a single image and exits.
EOF
}

# --- IMAGE CAPTURE FUNCTIONS ---

# Downloads a few seconds of MJPEG stream to a temporary file.
# We download a window of data because MJPEG is a continuous stream and we need 
# to ensure we capture at least one full JPEG frame (from SOI to EOI).
# Returns 0 if data was captured, 1 otherwise.
download_stream() {
    # Using --http0.9 because this camera provides minimal headers.
    # --max-time ensures we don't hang if the camera stops responding.
    # curl will return error code 28 on timeout, so we check if file exists and has size.
    local auth_args=()
    if [[ -n "$WEBCAM_USER" ]]; then
        auth_args=("-u" "$WEBCAM_USER:$WEBCAM_PASS")
    fi

    local curl_err_log
    curl_err_log=$(mktemp)
    
    local exit_code=0
    LC_ALL=C curl --http0.9 -fsSL \
        --connect-timeout "$TIMEOUT" \
        --max-time "$CAPTURE_WINDOW" \
        "${auth_args[@]}" \
        "$WEBCAM_URL" \
        -o "$TMP_MJPEG" 2>"$curl_err_log" || exit_code=$?
    
    if [[ $exit_code -ne 0 ]] && [[ ! -s "$TMP_MJPEG" ]]; then
        if [[ -s "$curl_err_log" ]]; then
            log_err "Curl failed (exit code $exit_code): $(cat "$curl_err_log")"
        fi
        rm -f "$curl_err_log"
        return 1
    fi
    
    rm -f "$curl_err_log"
    return 0
}

# Extracts a valid JPEG frame from MJPEG data.
# It searches for JPEG markers: SOI (FF D8 FF) and EOI (FF D9).
# If multiple frames are present, it tries a "middle frame" strategy to avoid 
# potentially truncated frames at the beginning or end of the capture window.
# Returns 0 if a valid JPEG was saved, 1 otherwise.
extract_jpeg() {
    local source_mjpg="$1"
    local output_jpg="$2"
    
    # MJPEG contains JPEGs concatenated. They start with FF D8 FF and end with FF D9.
    # We use LC_ALL=C for byte-accurate offsets in binary data.
    local sois eois
    sois=$(LC_ALL=C grep -a -b -o $'\xff\xd8\xff' "$source_mjpg" | cut -d: -f1) || true
    eois=$(LC_ALL=C grep -a -b -o $'\xff\xd9' "$source_mjpg" | cut -d: -f1) || true
    
    if [[ -z "$sois" ]] || [[ -z "$eois" ]]; then
        return 1
    fi
    
    # Try the 'middle' frame strategy: frames in the middle are less likely to be 
    # truncated by the network/buffer window.
    local num_sois
    num_sois=$(echo "$sois" | grep -c . || echo 0)
    
    local -a target_indices
    local mid=$((num_sois / 2))
    [[ $mid -lt 1 ]] && mid=1
    
    # We try the middle one first, then the very first one as fallback
    target_indices=("$mid" 1)
    
    for idx in "${target_indices[@]}"; do
        local start
        start=$(echo "$sois" | sed -n "${idx}p")
        [[ -z "$start" ]] && continue
        
        # Find the first EOI that occurs after this SOI
        local end=""
        for eoi in $eois; do
            if [[ "$eoi" -gt "$start" ]]; then
                end="$eoi"
                break
            fi
        done
        
        if [[ -n "$end" ]]; then
            # EOI marker is 2 bytes (ff d9). Length is (end - start + 2).
            local len=$((end - start + 2))
            
            # Extraction using dd
            dd if="$source_mjpg" of="$output_jpg" bs=64K \
               skip="$start" count="$len" \
               iflag=skip_bytes,count_bytes \
               status=none 2>/dev/null
            
            # Final validation: check if last 2 bytes are indeed ff d9
            if [[ -s "$output_jpg" ]] && [[ "$(tail -c 2 "$output_jpg" | LC_ALL=C od -t x1 | head -n 1 | cut -d' ' -f2-)" =~ "ff d9" ]]; then
                return 0
            fi
            # Cleanup bad file and try next index
            rm -f "$output_jpg"
        fi
    done
    
    return 1
}

# Orchestrates one capture cycle
perform_capture() {
    local output_file="$1"
    
    if ! download_stream; then
        log_err "Failed to access MJPEG stream at ${WEBCAM_URL:-[URL NOT SET]}"
        return 1
    fi
    
    if extract_jpeg "$TMP_MJPEG" "$output_file"; then
        log "Saved: $output_file"
        rm -f "$TMP_MJPEG"
        return 0
    else
        log_err "Could not extract a valid JPEG frame from stream."
        rm -f "$TMP_MJPEG" "$output_file"
        return 1
    fi
}

# --- ARGUMENT PARSING ---
LOOP_MODE=false
INTERVAL="$DEFAULT_INTERVAL"
OUT_FILE=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -o|--out)
            if [[ -n "${2:-}" ]]; then
                OUT_FILE="$2"
                shift 2
            else
                log_err "$1 requires a filename argument."
                exit 1
            fi
            ;;
        -tl|--timeLapse)
            LOOP_MODE=true
            if [[ -n "${2:-}" && "$2" =~ ^[0-9]+$ ]]; then
                INTERVAL="$2"
                shift 2
            else
                log_err "$1 requires a numeric argument (seconds)."
                exit 1
            fi
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            # Backward compatibility for old --loop style or unknown params
            LOOP_MODE=true
            shift
            ;;
    esac
done

# Check if required variables are set
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ -z "$WEBCAM_URL" ]]; then
        log_err "WEBCAM_URL is mandatory."
        exit 1
    fi
    if [[ -z "$WEBCAM_USER" ]]; then
        log_err "WEBCAM_USER is mandatory."
        exit 1
    fi
    if [[ -z "$WEBCAM_PASS" ]]; then
        log_err "WEBCAM_PASS is mandatory."
        exit 1
    fi
fi

# --- MAIN EXECUTION ---
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ "$LOOP_MODE" == false ]]; then
        # SINGLE RUN
        if [[ -z "$OUT_FILE" ]]; then
            mkdir -p "$CAPTURES_DIR"
            timestamp=$(date +%Y-%m-%d_%H-%M-%S)
            OUT_FILE="$CAPTURES_DIR/webcam_${timestamp}.jpg"
        fi
        
        if perform_capture "$OUT_FILE"; then
            log "Single capture successful."
            exit 0
        else
            log_err "Single capture failed."
            exit 1
        fi
    else
        # TIME-LAPSE LOOP
        if [[ -n "$OUT_FILE" ]]; then
            log_err "Error: --out is not compatible with --timeLapse."
            exit 1
        fi
        mkdir -p "$CAPTURES_DIR"
        
        log "Time-lapse enabled (Interval: ${INTERVAL}s). Press Ctrl+C to stop."
        
        next_tick=$(date +%s)
        while true; do
            now=$(date +%s)
            
            timestamp=$(date +%Y-%m-%d_%H-%M-%S)
            current_out="$CAPTURES_DIR/webcam_${timestamp}.jpg"
            
            if perform_capture "$current_out"; then
                log "Capture successful."
            else
                log_err "Capture cycle failed."
            fi
            
            # Calculate next tick based on fixed cadence to prevent drift
            next_tick=$(( next_tick + INTERVAL ))
            
            # Determine sleep duration
            after_work=$(date +%s)
            sleep_sec=$(( next_tick - after_work ))
            
            if [[ $sleep_sec -le 0 ]]; then
                # If capture took longer than interval, catch up immediately
                next_tick=$after_work
                sleep_sec=0
            fi
            
            log "Waiting ${sleep_sec}s until next capture..."
            sleep "$sleep_sec"
        done
    fi
fi
