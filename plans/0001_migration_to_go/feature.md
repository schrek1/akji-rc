# Feature: Migration to Go

## Objective

Migrate the AKJI Runway Chronicles core capture pipeline from Bash scripts to Go in small, mergeable slices. The goal is a codebase that is easier to read, test, maintain, and extend manually while keeping the current pipeline behavior stable during the migration.

The long-term product direction is a lightweight Go + Svelte stack. This feature covers the Go migration of the current script-based core only; Svelte implementation is documented as future direction, not part of this feature.

## Current state

The repository currently contains a minimal Bash-based MVP:

* `app/capture.sh` captures a JPEG frame from an MJPEG webcam stream.
* `app/ci/validate_frame.sh` validates the captured JPEG.
* `app/ci/publish_frame.sh` uploads the frame to a temporary public file host.
* GitHub Actions runs the pipeline on `ubuntu-latest`.

## Target state

The core pipeline is implemented as Go commands built from source in CI:

* `akji-capture`
* `akji-validate`
* `akji-publish`

CI builds Linux binaries during each workflow run and executes them directly. Generated binaries are never committed to git.

## Global decisions

* Migration is incremental, not Big Bang.
* Each slice should be independently reviewable and mergeable.
* Keep current external behavior stable unless a subplan explicitly says otherwise.
* Preserve env-based configuration as the primary runtime contract.
* Keep the current temporary publishing target for now.
* Do not implement S3/R2 upload in this feature; it remains a later task.
* Do not implement Svelte in this feature; only align documentation toward Go + Svelte.

## Subplans

1. `0001_ci_go_binary_proof.md` — prove GitHub Actions can build and execute Go binaries.
2. `0002_migrate_capture_to_go.md` — migrate frame capture from Bash to Go.
3. `0003_migrate_validate_to_go.md` — migrate frame validation from Bash to Go.
4. `0004_migrate_publish_to_go.md` — migrate temporary frame publishing from Bash to Go.
5. `0005_cleanup_docs_and_stack_alignment.md` — remove replaced scripts and align documentation.

## Suggested branch breakdown

Each subplan should be implemented in its own feature branch and merged independently, in order. Later subplans may assume earlier subplans are merged.

## Definition of done

The feature is complete when the scheduled capture workflow runs the Go-based capture, validate, and publish commands end to end, replaced Bash scripts are removed, and README/docs describe the Go-based pipeline and future Go + Svelte direction.
