# Subplan 0004: Migrate Publish To Go

## Objective

Replace the Bash temporary publish script with a Go command while preserving the current Uguu upload behavior and CI output contract.

## Scope

* Implement `akji-publish`.
* Accept a frame path as the first positional argument.
* Default to `/tmp/frame.jpg` when no path is provided.
* Upload to the current temporary Uguu endpoint.
* Preserve retry behavior.
* Print only the resulting URL to stdout on success.
* Log operational messages to stderr.
* Update CI publish step to run the Go binary.

## Implementation notes

* Preserve the current endpoint behavior from `app/ci/publish_frame.sh`.
* Keep the stdout contract strict because GitHub Actions captures it into `GITHUB_OUTPUT`.
* Parse known response shapes robustly enough to preserve current behavior:
  * `url`
  * `fileUrl`
  * `link`
* Keep retry count and delay close to current behavior unless a test requires adjustment.
* Keep the note about temporary/public hosting in logs or docs, not in stdout.

## Acceptance criteria

* Existing capture and validation outputs can be published by `akji-publish`.
* Successful publish writes only the URL to stdout.
* Failed publish exits non-zero.
* Missing or empty file fails before upload.
* CI publish step uses the Go publish command.
* S3/R2 remains untouched.

## Test plan

* Unit-test response parsing for supported URL fields.
* Use a local HTTP test server to verify multipart upload request shape.
* Test retry behavior with temporary failures.
* Run `go test ./...`.
* Run the full workflow manually or in CI with the temporary publisher.

## Explicit non-goals

* Do not implement S3/R2 upload.
* Do not change the public hosting provider in this slice.
* Do not change the GitHub summary rendering except for calling the Go command.
* Do not remove `app/ci/publish_frame.sh` until cleanup.
