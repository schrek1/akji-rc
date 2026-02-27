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

# --- CONFIGURATION ---
MJPEG_URL="http://01089001.pfw.ji.cz:16170/channel2"
USER="akji"
PASS="akji"
TIMEOUT=5
# Seconds to capture MJPEG data to ensure we have at least one complete frame
CAPTURE_WINDOW=2
DEFAULT_INTERVAL=15

# --- DIRECTORY SETUP ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CAPTURES_DIR="$SCRIPT_DIR/captures"
# Using PID-unique name for the temp buffer to prevent conflicts
TMP_MJPEG="$CAPTURES_DIR/stream_capture_$$.mjpg"

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
  -tl, --timeLapse <N>   Run in loop, capturing every N seconds.
  -h, --help             Show this help message.

If no options are provided, the script captures a single image and exits.
EOF
}

# --- LOGGING HELPERS ---
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $*"
}

log_err() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - ERROR: $*" >&2
}

# --- IMAGE CAPTURE FUNCTIONS ---

# Downloads a few seconds of MJPEG stream
# Returns 0 if we got data (even if curl timed out as expected)
download_stream() {
    # Using --http0.9 because this camera provides minimal headers.
    # --max-time ensures we don't hang if the camera stops responding.
    # curl will return error code 28 on timeout, so we check if file exists and has size.
    LC_ALL=C curl --http0.9 -fsSL \
        --connect-timeout "$TIMEOUT" \
        --max-time "$CAPTURE_WINDOW" \
        -u "$USER:$PASS" \
        "$MJPEG_URL" \
        -o "$TMP_MJPEG" 2>/dev/null || [[ -s "$TMP_MJPEG" ]]
}

# Extracts a valid JPEG frame from MJPEG data
# Returns 0 if saved successfully, 1 otherwise
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
    local timestamp
    timestamp=$(date +%Y-%m-%d_%H-%M-%S)
    local output_file="$CAPTURES_DIR/webcam_${timestamp}.jpg"
    
    if ! download_stream; then
        log_err "Failed to access MJPEG stream at $MJPEG_URL"
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

while [[ $# -gt 0 ]]; do
    case "$1" in
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

# --- MAIN EXECUTION ---
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    mkdir -p "$CAPTURES_DIR"

    if [[ "$LOOP_MODE" == false ]]; then
        # SINGLE RUN
        if perform_capture; then
            log "Single capture successful."
            exit 0
        else
            log_err "Single capture failed."
            exit 1
        fi
    else
        # TIME-LAPSE LOOP
        log "Time-lapse enabled (Interval: ${INTERVAL}s). Press Ctrl+C to stop."
        
        next_tick=$(date +%s)
        while true; do
            now=$(date +%s)
            
            if perform_capture; then
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
