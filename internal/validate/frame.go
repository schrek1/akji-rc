package validate

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

const DefaultMinimumSizeBytes int64 = 15_000

var (
	jpegSOI = []byte{0xff, 0xd8}
	jpegEOI = []byte{0xff, 0xd9}
)

type Result struct {
	FilePath string
	Size     int64
}

func ValidateFile(filePath string, minimumSizeBytes int64) (Result, error) {
	if minimumSizeBytes <= 0 {
		return Result{}, fmt.Errorf("minimum frame size must be positive")
	}

	frame, err := os.ReadFile(filePath)
	if err != nil {
		return Result{}, fmt.Errorf("could not read frame %q: %w", filePath, err)
	}
	if len(frame) == 0 {
		return Result{}, fmt.Errorf("missing or empty file: %s", filePath)
	}
	if int64(len(frame)) < minimumSizeBytes {
		return Result{}, fmt.Errorf("file too small (%d < %d) - likely not a real frame", len(frame), minimumSizeBytes)
	}
	if !bytes.HasPrefix(frame, jpegSOI) {
		return Result{}, fmt.Errorf("invalid JPEG SOI marker")
	}
	if !bytes.HasSuffix(frame, jpegEOI) {
		return Result{}, fmt.Errorf("invalid JPEG EOI marker")
	}
	mimeType := http.DetectContentType(frame)
	if mimeType != "image/jpeg" {
		return Result{}, fmt.Errorf("invalid MIME type: %s", mimeType)
	}
	if containsHTMLResponse(frame) {
		return Result{}, fmt.Errorf("looks like HTML page, not a JPEG")
	}

	return Result{FilePath: filePath, Size: int64(len(frame))}, nil
}

func containsHTMLResponse(frame []byte) bool {
	if len(frame) > 512 {
		frame = frame[:512]
	}
	lowercaseFrame := bytes.ToLower(frame)
	return bytes.Contains(lowercaseFrame, []byte("<html")) || bytes.Contains(lowercaseFrame, []byte("<!doctype"))
}
