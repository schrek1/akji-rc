package capture

import (
	"fmt"
	"os"
	"path/filepath"
)

type CaptureResult struct {
	OutputPath      string
	DownloadedBytes int
}

func CaptureToFile(configuration Configuration, outputPath string) (CaptureResult, error) {
	mjpegStream, err := downloadMJPEGStream(configuration)
	if err != nil {
		return CaptureResult{}, err
	}

	jpegFrame, err := ExtractJPEG(mjpegStream)
	if err != nil {
		return CaptureResult{}, removeInvalidCaptureOutput(outputPath, err)
	}
	if err := saveJPEGFrame(outputPath, jpegFrame); err != nil {
		return CaptureResult{}, err
	}

	return CaptureResult{
		OutputPath:      outputPath,
		DownloadedBytes: len(mjpegStream),
	}, nil
}

func downloadMJPEGStream(configuration Configuration) ([]byte, error) {
	mjpegStream, err := DownloadMJPEGStream(configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to access MJPEG stream at %s: %w", configuration.WebcamURL, err)
	}
	return mjpegStream, nil
}

func removeInvalidCaptureOutput(outputPath string, captureError error) error {
	_ = os.Remove(outputPath)
	return captureError
}

func saveJPEGFrame(outputPath string, jpegFrame []byte) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, jpegFrame, 0o644)
}
