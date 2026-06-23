# Subplan 0003: Migrate Validate To Go

## Objective

Replace the Bash frame validation script with a Go command that validates the captured JPEG before publishing.

## Scope

* Implement `akji-validate`.
* Accept a frame path as the first positional argument.
* Default to `/tmp/frame.jpg` when no path is provided.
* Preserve the meaningful checks from `app/ci/validate_frame.sh`.
* Update CI validation step to run the Go binary.

## Implementation notes

* Preserve the current validation intent:
  * file exists and is non-empty
  * file size is at least `MIN_SIZE_BYTES`, defaulting to `15000`
  * SOI marker is `FFD8`
  * EOI marker is `FFD9`
  * obvious HTML/text responses are rejected
* MIME detection may use Go standard library behavior rather than shelling out to `file`, as long as JPEG false positives are still reasonably guarded.
* Log validation details to stdout/stderr in a CI-readable format.

## Acceptance criteria

* Valid captured JPEG passes validation.
* Missing file fails.
* Empty file fails.
* Too-small file fails.
* Invalid JPEG markers fail.
* Obvious HTML response fails.
* CI validation step uses the Go validation command.

## Test plan

* Unit-test validation with generated temporary files:
  * valid JPEG-like bytes over minimum size
  * missing file
  * empty file
  * too-small file
  * invalid SOI
  * invalid EOI
  * HTML content
* Run `go test ./...`.
* Run the capture workflow after `akji-capture` produces `/tmp/frame.jpg`.

## Explicit non-goals

* Do not migrate capture.
* Do not migrate publishing.
* Do not introduce image decoding requirements unless needed for validation parity.
* Do not remove `app/ci/validate_frame.sh` until the whole Go path is stable.
