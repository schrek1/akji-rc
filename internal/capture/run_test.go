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
	frame := []byte("\xff\xd8\xffCI_DATA\xff\xd9")
	stream := append([]byte("some junk"), frame...)
	stream = append(stream, []byte("more junk")...)

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestUser, requestPass, ok := request.BasicAuth()
		if !ok || requestUser != user || requestPass != pass {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.Header().Set("Content-Type", "multipart/x-mixed-replace")
		_, _ = response.Write(stream)
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "ci_test.jpg")
	config := CapturingConfiguration{
		WebcamURL:     server.URL,
		WebcamUser:    user,
		WebcamPass:    pass,
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	if err := CaptureToFile(config, outputPath, testLogger{t: t}); err != nil {
		t.Fatalf("CaptureToFile() error = %v", err)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, frame) {
		t.Fatalf("output = %q, want %q", output, frame)
	}
}

func TestCaptureToFile_invalidStreamRemovesOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte("some junk\xff\xd8\xffimage data without end"))
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "bad.jpg")
	if err := os.WriteFile(outputPath, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	config := CapturingConfiguration{
		WebcamURL:     server.URL,
		WebcamUser:    "user",
		WebcamPass:    "pass",
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	if err := CaptureToFile(config, outputPath, testLogger{t: t}); err == nil {
		t.Fatal("CaptureToFile() error = nil")
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output file should be removed, stat err = %v", err)
	}
}
