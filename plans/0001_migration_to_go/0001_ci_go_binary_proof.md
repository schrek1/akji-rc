# Subplan 0001: CI Go Binary Proof

## Objective

Prove that GitHub Actions can build a Go binary from this repository and execute the produced Linux binary in the pipeline. This slice should not migrate any business behavior yet.

## Scope

* Add a minimal Go module.
* Add one tiny command that can be built and executed in CI.
* Update or add a CI check that runs `go test` and `go build`.
* Execute the built binary on `ubuntu-latest`.

## Implementation notes

* Use a conventional Go layout that can grow into the later commands.
* Build binaries into an ignored/generated directory such as `dist/` or a temporary CI path.
* Do not commit generated binaries.
* Keep this slice deliberately boring: the command may only print a version/help/probe message and exit successfully.
* Add or update `.gitignore` if the chosen build output directory is not already ignored.

## Acceptance criteria

* CI runs `go test ./...`.
* CI builds at least one Go command on `ubuntu-latest`.
* CI executes the built binary successfully.
* No generated binary is committed.
* Existing Bash capture pipeline behavior remains unchanged.

## Test plan

* Run `go test ./...`.
* Run `go build` for the proof command locally where possible.
* Verify the GitHub Actions workflow passes on the branch.

## Explicit non-goals

* Do not migrate `capture.sh`.
* Do not migrate validation or publishing.
* Do not introduce S3/R2 support.
* Do not add Svelte or any frontend code.
