package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Logger interface {
	Printf(format string, args ...any)
}

func CaptureToFile(config Config, outputPath string, logger Logger) error {
	stream, err := DownloadStream(config)
	if err != nil {
		return fmt.Errorf("failed to access MJPEG stream at %s: %w", config.WebcamURL, err)
	}
	if logger != nil {
		logger.Printf("Downloaded %d bytes from MJPEG stream.", len(stream))
	}

	frame, err := ExtractJPEG(stream)
	if err != nil {
		_ = os.Remove(outputPath)
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, frame, 0o644); err != nil {
		return err
	}
	if logger != nil {
		logger.Printf("Saved: %s", outputPath)
	}
	return nil
}

func DefaultOutputPath(workDir string, now time.Time) string {
	capturesDir := filepath.Join(workDir, "app", "captures")
	if _, err := os.Stat(filepath.Join(workDir, "app")); os.IsNotExist(err) {
		capturesDir = filepath.Join(workDir, "captures")
	}
	timestamp := now.Format("2006-01-02_15-04-05")
	return filepath.Join(capturesDir, "webcam_"+timestamp+".jpg")
}
