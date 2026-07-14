# AKJI Runway Chronicles

AKJI Runway Chronicles periodically captures a JPEG frame from an MJPEG webcam, validates it, and publishes a temporary public URL.

The runtime pipeline is implemented in Go:

1. `akji-capture` saves a JPEG frame.
2. `akji-validate` verifies the frame before publishing.
3. `akji-publish` uploads the frame to temporary Uguu hosting and writes its URL to stdout.

There is no backend or UI in the current MVP.

## Requirements

- Go version declared in `go.mod`
- Webcam credentials in `.env` or process environment variables

## Local Run

1. Configure credentials:

   ```bash
   cp .env.example .env
   ```

2. Capture a frame:

   ```bash
   go run ./cmd/akji-capture
   ```

   The default output is `captures/webcam_<TIMESTAMP>.jpg`.

3. Capture to a specific file:

   ```bash
   go run ./cmd/akji-capture --out my_frame.jpg
   ```

4. Run time-lapse capture every 30 seconds:

   ```bash
   go run ./cmd/akji-capture --timeLapse 30
   ```

5. Validate and publish a captured file:

   ```bash
   go run ./cmd/akji-validate my_frame.jpg
   go run ./cmd/akji-publish my_frame.jpg
   ```

## Configuration

`.env` provides local defaults for `akji-capture`. Explicit process environment variables override it, which keeps CI and deployment secrets outside the repository.

| Variable | Description | Default |
| --- | --- | --- |
| `WEBCAM_URL` | Full MJPEG stream URL | Required |
| `WEBCAM_USER` | Basic-auth username | Required |
| `WEBCAM_PASS` | Basic-auth password | Required |
| `TIMEOUT` | Connection timeout in seconds | `5` |
| `CAPTURE_WINDOW` | MJPEG capture window in seconds | `3` |

`akji-validate` reads `MIN_SIZE_BYTES` directly from the process environment; its default is `15000`.

## Development

```bash
go test ./...
go vet ./...
go build ./cmd/...
```

GitHub Actions runs the same Go test suite and performs an integration capture against the configured webcam. The scheduled capture workflow builds and runs all three Go commands.

## Direction

The core pipeline is Go. A future lightweight UI, when needed, will use Svelte. S3-compatible storage, including Cloudflare R2, remains the intended replacement for the temporary Uguu publisher.
