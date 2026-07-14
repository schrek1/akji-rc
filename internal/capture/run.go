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

func CaptureToFile(config CapturingConfiguration, outputPath string, logger Logger) error {
	mjpegStream, err := downloadMJPEGStream(config)
	if err != nil {
		return err
	}

	logDownloadedMJPEGStream(logger, mjpegStream)

	jpegFrame, err := ExtractJPEG(mjpegStream)
	if err != nil {
		return removeInvalidCaptureOutput(outputPath, err)
	}

	return saveJPEGFrame(outputPath, jpegFrame, logger)
}

func downloadMJPEGStream(config CapturingConfiguration) ([]byte, error) {
	mjpegStream, err := DownloadStream(config)
	if err != nil {
		return nil, fmt.Errorf("failed to access MJPEG stream at %s: %w", config.WebcamURL, err)
	}
	return mjpegStream, nil
}

func logDownloadedMJPEGStream(logger Logger, mjpegStream []byte) {
	if logger != nil {
		logger.Printf("Downloaded %d bytes from MJPEG stream.", len(mjpegStream))
	}
}

func removeInvalidCaptureOutput(outputPath string, captureError error) error {
	_ = os.Remove(outputPath)
	return captureError
}

func saveJPEGFrame(outputPath string, jpegFrame []byte, logger Logger) error {
	if err := createCaptureOutputDirectory(outputPath); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, jpegFrame, 0o644); err != nil {
		return err
	}
	logSavedJPEGFrame(logger, outputPath)
	return nil
}

func createCaptureOutputDirectory(outputPath string) error {
	return os.MkdirAll(filepath.Dir(outputPath), 0o755)
}

func logSavedJPEGFrame(logger Logger, outputPath string) {
	if logger != nil {
		logger.Printf("Saved: %s", outputPath)
	}
}

func DefaultOutputPath(workDir string, captureTime time.Time) string {
	return filepath.Join(defaultCapturesDirectory(workDir), captureFileName(captureTime))
}

func defaultCapturesDirectory(workDir string) string {
	capturesDir := filepath.Join(workDir, "app", "captures")
	if _, err := os.Stat(filepath.Join(workDir, "app")); os.IsNotExist(err) {
		capturesDir = filepath.Join(workDir, "captures")
	}
	return capturesDir
}

func captureFileName(captureTime time.Time) string {
	timestamp := captureTime.Format("2006-01-02_15-04-05")
	return "webcam_" + timestamp + ".jpg"
}
