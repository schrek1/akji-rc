package publish

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePublishedURL_supportedResponseFields_returnsURL(t *testing.T) {
	testCases := []struct {
		name         string
		responseBody string
		expectedURL  string
	}{
		{"url", `{"files":[{"url":"https://uguu.se/file-url"}]}`, "https://uguu.se/file-url"},
		{"fileUrl", `{"files":[{"fileUrl":"https://uguu.se/file-url"}]}`, "https://uguu.se/file-url"},
		{"link", `{"files":[{"link":"https://uguu.se/file-url"}]}`, "https://uguu.se/file-url"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			publishedURL, err := parsePublishedURL([]byte(testCase.responseBody))

			if err != nil {
				t.Fatalf("parsePublishedURL() error = %v", err)
			}
			if publishedURL != testCase.expectedURL {
				t.Errorf("publishedURL = %q, want %q", publishedURL, testCase.expectedURL)
			}
		})
	}
}

func TestPublishFile_uploadsFrameAsMultipartFile(t *testing.T) {
	frameContents := []byte("frame data")
	framePath := writeFrameFile(t, frameContents)

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", request.Method, http.MethodPost)
		}
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}

		files := request.MultipartForm.File["files[]"]
		if len(files) != 1 {
			t.Fatalf("files[] count = %d, want 1", len(files))
		}
		if files[0].Filename != filepath.Base(framePath) {
			t.Errorf("filename = %q, want %q", files[0].Filename, filepath.Base(framePath))
		}

		uploadedFile, err := files[0].Open()
		if err != nil {
			t.Fatal(err)
		}
		defer uploadedFile.Close()
		uploadedContents, err := io.ReadAll(uploadedFile)
		if err != nil {
			t.Fatal(err)
		}
		if string(uploadedContents) != string(frameContents) {
			t.Errorf("uploadedContents = %q, want %q", uploadedContents, frameContents)
		}

		_, _ = response.Write([]byte(`{"files":[{"url":"https://uguu.se/published-frame"}]}`))
	}))
	defer server.Close()

	result, err := PublishFile(testConfiguration(server.URL), framePath)

	if err != nil {
		t.Fatalf("PublishFile() error = %v", err)
	}
	if result.URL != "https://uguu.se/published-frame" {
		t.Errorf("URL = %q", result.URL)
	}
}

func TestPublishFile_temporaryFailures_retriesUntilUploadSucceeds(t *testing.T) {
	framePath := writeFrameFile(t, []byte("frame data"))
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		attempts++
		if attempts < 3 {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = response.Write([]byte(`{"files":[{"url":"https://uguu.se/published-frame"}]}`))
	}))
	defer server.Close()

	_, err := PublishFile(testConfiguration(server.URL), framePath)

	if err != nil {
		t.Fatalf("PublishFile() error = %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestPublishFile_persistentFailure_returnsErrorAfterAllAttempts(t *testing.T) {
	framePath := writeFrameFile(t, []byte("frame data"))
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		attempts++
		response.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := PublishFile(testConfiguration(server.URL), framePath)

	if err == nil {
		t.Fatal("PublishFile() error = nil")
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestPublishFile_missingOrEmptyFile_returnsErrorBeforeUpload(t *testing.T) {
	testCases := []struct {
		name       string
		filePath   string
		contents   []byte
		createFile bool
	}{
		{"missing file", filepath.Join(t.TempDir(), "missing.jpg"), nil, false},
		{"empty file", "", nil, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			filePath := testCase.filePath
			if testCase.createFile {
				filePath = writeFrameFile(t, testCase.contents)
			}

			_, err := PublishFile(testConfiguration("http://127.0.0.1:1"), filePath)

			if err == nil {
				t.Fatal("PublishFile() error = nil")
			}
			if !strings.Contains(err.Error(), "file") {
				t.Errorf("error = %q, want file error", err)
			}
		})
	}
}

func testConfiguration(endpoint string) Configuration {
	return Configuration{
		Endpoint:        endpoint,
		HTTPClient:      http.DefaultClient,
		MaximumAttempts: 3,
		RetryDelay:      0,
	}
}

func writeFrameFile(t *testing.T, contents []byte) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), "frame.jpg")
	if err := os.WriteFile(filePath, contents, 0o644); err != nil {
		t.Fatal(err)
	}
	return filePath
}
