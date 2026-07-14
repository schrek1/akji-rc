package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFile_validJPEGLikeFrame_returnsResult(t *testing.T) {
	filePath := writeFrameFile(t, validJPEGLikeFrame())

	result, err := ValidateFile(filePath, DefaultMinimumSizeBytes)

	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}
	if result.FilePath != filePath {
		t.Errorf("FilePath = %q, want %q", result.FilePath, filePath)
	}
	if result.Size != int64(len(validJPEGLikeFrame())) {
		t.Errorf("Size = %d, want %d", result.Size, len(validJPEGLikeFrame()))
	}
}

func TestValidateFile_missingFile_returnsError(t *testing.T) {
	_, err := ValidateFile(filepath.Join(t.TempDir(), "missing.jpg"), DefaultMinimumSizeBytes)

	if err == nil {
		t.Fatal("ValidateFile() error = nil")
	}
}

func TestValidateFile_emptyFile_returnsError(t *testing.T) {
	filePath := writeFrameFile(t, nil)

	_, err := ValidateFile(filePath, DefaultMinimumSizeBytes)

	if err == nil {
		t.Fatal("ValidateFile() error = nil")
	}
}

func TestValidateFile_tooSmallFile_returnsError(t *testing.T) {
	filePath := writeFrameFile(t, []byte{0xff, 0xd8, 0xff, 0xd9})

	_, err := ValidateFile(filePath, DefaultMinimumSizeBytes)

	if err == nil {
		t.Fatal("ValidateFile() error = nil")
	}
}

func TestValidateFile_invalidSOIMarker_returnsError(t *testing.T) {
	frame := validJPEGLikeFrame()
	frame[0] = 0
	filePath := writeFrameFile(t, frame)

	_, err := ValidateFile(filePath, DefaultMinimumSizeBytes)

	if err == nil {
		t.Fatal("ValidateFile() error = nil")
	}
}

func TestValidateFile_invalidEOIMarker_returnsError(t *testing.T) {
	frame := validJPEGLikeFrame()
	frame[len(frame)-1] = 0
	filePath := writeFrameFile(t, frame)

	_, err := ValidateFile(filePath, DefaultMinimumSizeBytes)

	if err == nil {
		t.Fatal("ValidateFile() error = nil")
	}
}

func TestValidateFile_htmlResponse_returnsError(t *testing.T) {
	frame := validJPEGLikeFrame()
	copy(frame[16:], []byte("<!doctype html><html>camera login</html>"))
	filePath := writeFrameFile(t, frame)

	_, err := ValidateFile(filePath, DefaultMinimumSizeBytes)

	if err == nil {
		t.Fatal("ValidateFile() error = nil")
	}
}

func validJPEGLikeFrame() []byte {
	frame := make([]byte, DefaultMinimumSizeBytes+1)
	copy(frame, []byte{0xff, 0xd8, 0xff})
	copy(frame[len(frame)-len(jpegEOI):], jpegEOI)
	return frame
}

func writeFrameFile(t *testing.T, frame []byte) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), "frame.jpg")
	if err := os.WriteFile(filePath, frame, 0o644); err != nil {
		t.Fatal(err)
	}
	return filePath
}
