# Subplan 0005: Cleanup Docs And Stack Alignment

## Objective

Finish the Go migration by removing replaced Bash scripts and aligning documentation with the Go-based core pipeline and future Go + Svelte direction.

## Scope

* Remove Bash scripts that are fully replaced by Go commands.
* Remove obsolete Bash-specific test scripts if equivalent Go tests and CI checks exist.
* Update README with the Go build/run/test workflow.
* Update decisions documentation to describe Go as the core pipeline language.
* Document Svelte as the intended lightweight frontend direction for future UI work.
* Keep S3/R2 as future storage direction unless already implemented in a separate feature.

## Implementation notes

* Only remove a shell script after its Go replacement is active in CI.
* Keep local development instructions short and practical.
* Make the current runtime path clear:
  * capture frame
  * validate frame
  * publish temporary URL
* Separate current implementation from future direction.
* Avoid rewriting project vision beyond the migration-related corrections.

## Acceptance criteria

* README no longer instructs users to run replaced Bash scripts as the primary path.
* Docs no longer imply Vue or Kotlin are the current planned stack.
* Docs state Go + Svelte as the intended lightweight direction.
* CI no longer depends on replaced shell scripts.
* No generated binaries are committed.
* Full Go test suite and capture workflow pass.

## Test plan

* Run `go test ./...`.
* Run the GitHub Actions regression workflow.
* Run the scheduled/manual capture workflow.
* Manually read README instructions and verify they match the repository state.

## Explicit non-goals

* Do not implement Svelte.
* Do not implement S3/R2.
* Do not add backend APIs.
* Do not add archive/calendar/timelapse features.
