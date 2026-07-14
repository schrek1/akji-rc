package capture

import (
	"fmt"
	"os"
	"path/filepath"
)

type Result struct {
	OutputPath      string
	DownloadedBytes int
}

func ToFile(configuration Configuration, outputPath string) (Result, error) {
	mjpegStream, err := fetchMJPEGStream(configuration)
	if err != nil {
		return Result{}, err
	}

	jpegFrame, err := ExtractJPEGFrame(mjpegStream)
	if err != nil {
		return Result{}, removeInvalidCaptureOutput(outputPath, err)
	}
	if err := saveJPEGFrame(outputPath, jpegFrame); err != nil {
		return Result{}, err
	}

	return Result{
		OutputPath:      outputPath,
		DownloadedBytes: len(mjpegStream),
	}, nil
}

func fetchMJPEGStream(configuration Configuration) ([]byte, error) {
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
