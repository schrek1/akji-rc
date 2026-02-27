# akji-rc (AKJI Runway Chronicles)

Minimal MVP: periodically capture one JPEG snapshot and upload it to S3-compatible storage.

No backend. No UI.

---

## MVP1 Scope

- `capture.sh` generates a JPEG into `./out/YYYY/MM/DD/HH/mm.jpg`
- It also updates `./out/latest.jpg`
- `upload.sh` uploads the image to S3-compatible storage
- GitHub Actions runs it on a CRON schedule (every 15 minutes) and manually

---

## Local Run

### Requirements

- bash
- ffmpeg
- AWS CLI (only if upload is needed)

---

### Run (mock source, default)

```bash
./app/run.sh