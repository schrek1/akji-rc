package capture

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testLogger struct {
	t *testing.T
}

func (logger testLogger) Printf(format string, args ...any) {
	logger.t.Logf(format, args...)
}

func TestCaptureToFile_downloadsAndExtractsMockFrame(t *testing.T) {
	const user = "mock-user"
	const pass = "mock-pass"
	expectedJPEGFrame := []byte("\xff\xd8\xffCI_DATA\xff\xd9")
	mjpegStream := append([]byte("some junk"), expectedJPEGFrame...)
	mjpegStream = append(mjpegStream, []byte("more junk")...)

	webcamServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestUser, requestPass, ok := request.BasicAuth()
		if !ok || requestUser != user || requestPass != pass {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.Header().Set("Content-Type", "multipart/x-mixed-replace")
		_, _ = response.Write(mjpegStream)
	}))
	defer webcamServer.Close()

	outputPath := filepath.Join(t.TempDir(), "captures", "ci_test.jpg")
	capturingConfiguration := CapturingConfiguration{
		WebcamURL:     webcamServer.URL,
		WebcamUser:    user,
		WebcamPass:    pass,
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	if err := CaptureToFile(capturingConfiguration, outputPath, testLogger{t: t}); err != nil {
		t.Fatalf("CaptureToFile() error = %v", err)
	}

	savedJPEGFrame, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(savedJPEGFrame, expectedJPEGFrame) {
		t.Fatalf("savedJPEGFrame = %q, want %q", savedJPEGFrame, expectedJPEGFrame)
	}
}

func TestCaptureToFile_invalidStreamRemovesOutput(t *testing.T) {
	webcamServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte("some junk\xff\xd8\xffimage data without end"))
	}))
	defer webcamServer.Close()

	outputPath := filepath.Join(t.TempDir(), "bad.jpg")
	if err := os.WriteFile(outputPath, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	capturingConfiguration := CapturingConfiguration{
		WebcamURL:     webcamServer.URL,
		WebcamUser:    "user",
		WebcamPass:    "pass",
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	if err := CaptureToFile(capturingConfiguration, outputPath, testLogger{t: t}); err == nil {
		t.Fatal("CaptureToFile() error = nil")
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output file should be removed, stat err = %v", err)
	}
}

func TestDefaultOutputPath_existingAppDirectory_usesAppCapturesDirectory(t *testing.T) {
	workDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(workDir, "app"), 0o755); err != nil {
		t.Fatal(err)
	}

	outputPath := DefaultOutputPath(workDir, time.Date(2026, time.July, 14, 17, 30, 0, 0, time.UTC))

	expectedPath := filepath.Join(workDir, "app", "captures", "webcam_2026-07-14_17-30-00.jpg")
	if outputPath != expectedPath {
		t.Fatalf("DefaultOutputPath() = %q, want %q", outputPath, expectedPath)
	}
}

func TestDefaultOutputPath_missingAppDirectory_usesWorkDirectoryCapturesDirectory(t *testing.T) {
	workDir := t.TempDir()

	outputPath := DefaultOutputPath(workDir, time.Date(2026, time.July, 14, 17, 30, 0, 0, time.UTC))

	expectedPath := filepath.Join(workDir, "captures", "webcam_2026-07-14_17-30-00.jpg")
	if outputPath != expectedPath {
		t.Fatalf("DefaultOutputPath() = %q, want %q", outputPath, expectedPath)
	}
}
