# akji-rc (AKJI Runway Chronicles)

Minimal MVP: periodically capture one JPEG snapshot from an MJPEG webcam and prepare it for upload.

No backend. No UI.

---

## MVP1 Scope

- `app/capture.sh` extracts a valid JPEG frame from an MJPEG stream.
- Supports single capture and time-lapse mode.
- Configurable via environment variables or `.env` files.
- GitHub Actions ready (works without local `captures/` directory if `--out` is used).

---

## Local Run

### Requirements

- `bash` (version 4+)
- `curl` (with HTTP 0.9 support)
- `grep`, `dd`, `od`, `tail` (standard coreutils)

### Quick Start

1. **Configure credentials:**
   Copy the template and edit it with your webcam details:
   ```bash
   cp app/.env.template app/.env
   # Edit app/.env with your WEBCAM_URL, WEBCAM_USER, and WEBCAM_PASS
   ```

2. **Run a single capture:**
   ```bash
   bash app/capture.sh
   # Image will be saved to app/captures/webcam_<TIMESTAMP>.jpg
   ```

3. **Capture to a specific file:**
   ```bash
   bash app/capture.sh --out my_frame.jpg
   ```

4. **Run time-lapse (every 30 seconds):**
   ```bash
   bash app/capture.sh --timeLapse 30
   ```

### Configuration (Environment Variables)

The script loads variables from `.env` files in the script directory or current directory. Direct environment variables have priority.

| Variable      | Description                  | Recommended value                        |
|---------------|------------------------------|------------------------------------------|
| `WEBCAM_URL`  | Full URL to the MJPEG stream | http://01089001.pfw.ji.cz:16170/channel2 |
| `WEBCAM_USER` | Username for basic auth      |                                          |
| `WEBCAM_PASS` | Password for basic auth      |                                          |

---

## Development & Testing

Run regression tests to ensure everything is working correctly:

```bash
bash app/test_capture.sh
```

