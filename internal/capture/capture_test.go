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

func TestCaptureToFile_downloadsAndSavesJPEGFrame(t *testing.T) {
	const user = "mock-user"
	const pass = "mock-pass"
	expectedJPEGFrame := []byte("\xff\xd8\xffCI_DATA\xff\xd9")
	mjpegStream := append([]byte("some junk"), expectedJPEGFrame...)
	mjpegStream = append(mjpegStream, []byte("more junk")...)

	webcamServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestUser, requestPass, authenticated := request.BasicAuth()
		if !authenticated || requestUser != user || requestPass != pass {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = response.Write(mjpegStream)
	}))
	defer webcamServer.Close()

	outputPath := filepath.Join(t.TempDir(), "captures", "ci_test.jpg")
	configuration := Configuration{
		WebcamURL:     webcamServer.URL,
		WebcamUser:    user,
		WebcamPass:    pass,
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	captureResult, err := CaptureToFile(configuration, outputPath)
	if err != nil {
		t.Fatalf("CaptureToFile() error = %v", err)
	}

	if captureResult.OutputPath != outputPath {
		t.Errorf("OutputPath = %q, want %q", captureResult.OutputPath, outputPath)
	}
	if captureResult.DownloadedBytes != len(mjpegStream) {
		t.Errorf("DownloadedBytes = %d, want %d", captureResult.DownloadedBytes, len(mjpegStream))
	}

	savedJPEGFrame, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(savedJPEGFrame, expectedJPEGFrame) {
		t.Fatalf("savedJPEGFrame = %q, want %q", savedJPEGFrame, expectedJPEGFrame)
	}
}

func TestCaptureToFile_invalidStreamRemovesExistingOutput(t *testing.T) {
	webcamServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte("some junk\xff\xd8\xffimage data without end"))
	}))
	defer webcamServer.Close()

	outputPath := filepath.Join(t.TempDir(), "bad.jpg")
	if err := os.WriteFile(outputPath, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	configuration := Configuration{
		WebcamURL:     webcamServer.URL,
		WebcamUser:    "user",
		WebcamPass:    "pass",
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	if _, err := CaptureToFile(configuration, outputPath); err == nil {
		t.Fatal("CaptureToFile() error = nil")
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output file should be removed, stat err = %v", err)
	}
}
