# Subplan 0002: Migrate Capture To Go

## Objective

Replace the Bash frame capture implementation with a Go command while preserving the current capture behavior and CI contract.

## Scope

* Implement `akji-capture`.
* Preserve the current single-capture CLI behavior with `--out <file>`.
* Preserve env-based configuration for webcam URL, credentials, timeout, and capture window.
* Extract a valid JPEG frame from MJPEG stream bytes.
* Update CI capture step to run the Go binary.
* Keep the old Bash script temporarily as reference/fallback until parity is proven.

## Implementation notes

* Start from the behavior of `app/capture.sh`, not from a new product design.
* Preserve current config names unless intentionally documented otherwise:
  * `WEBCAM_URL`
  * `WEBCAM_USER`
  * `WEBCAM_PASS`
  * `TIMEOUT`
  * `CAPTURE_WINDOW`
* Keep logs useful but avoid printing secrets.
* The command should write the captured JPEG to the requested output path and exit non-zero on failure.
* If Go HTTP handling cannot exactly reproduce the current camera behavior, document the difference and keep the shell fallback until the real camera integration passes.

## Acceptance criteria

* `akji-capture --out /tmp/frame.jpg` creates a non-empty JPEG when configured with valid webcam credentials.
* Missing required config fails with a clear error.
* Invalid or incomplete MJPEG data fails without leaving a misleading successful output.
* CI capture workflow uses the Go capture command.
* Existing validation and publish steps still work after the capture step changes.

## Test plan

* Unit-test JPEG extraction using byte fixtures:
  * valid single frame
  * missing EOI marker
  * multiple frames
  * junk before/after frame
* Add a test HTTP server for MJPEG-like stream behavior if practical.
* Run `go test ./...`.
* Run the real webcam integration workflow or equivalent manual test with configured secrets.

## Explicit non-goals

* Do not migrate validation.
* Do not migrate publishing.
* Do not add S3/R2 upload.
* Do not redesign time-lapse behavior unless needed for parity.
* Do not remove `app/capture.sh` in this slice.
