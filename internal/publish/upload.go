package publish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultEndpoint        = "https://uguu.se/upload"
	defaultMaximumAttempts = 3
	defaultRetryDelay      = 5 * time.Second
)

type Configuration struct {
	Endpoint        string
	HTTPClient      *http.Client
	MaximumAttempts int
	RetryDelay      time.Duration
}

type Result struct {
	URL string
}

func DefaultConfiguration() Configuration {
	return Configuration{
		Endpoint:        defaultEndpoint,
		HTTPClient:      http.DefaultClient,
		MaximumAttempts: defaultMaximumAttempts,
		RetryDelay:      defaultRetryDelay,
	}
}

func PublishFile(configuration Configuration, filePath string) (Result, error) {
	if err := requireNonEmptyFile(filePath); err != nil {
		return Result{}, err
	}
	if err := validateConfiguration(configuration); err != nil {
		return Result{}, err
	}

	var lastError error
	for attempt := 1; attempt <= configuration.MaximumAttempts; attempt++ {
		publishedURL, err := uploadFile(configuration, filePath)
		if err == nil {
			return Result{URL: publishedURL}, nil
		}
		lastError = err

		if attempt < configuration.MaximumAttempts {
			time.Sleep(configuration.RetryDelay)
		}
	}

	return Result{}, fmt.Errorf("upload failed after %d attempts: %w", configuration.MaximumAttempts, lastError)
}

func requireNonEmptyFile(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("could not access frame %q: %w", filePath, err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("missing or empty file: %s", filePath)
	}
	return nil
}

func validateConfiguration(configuration Configuration) error {
	if configuration.Endpoint == "" {
		return fmt.Errorf("upload endpoint is required")
	}
	if configuration.HTTPClient == nil {
		return fmt.Errorf("HTTP client is required")
	}
	if configuration.MaximumAttempts <= 0 {
		return fmt.Errorf("maximum upload attempts must be positive")
	}
	return nil
}

func uploadFile(configuration Configuration, filePath string) (string, error) {
	request, err := createUploadRequest(configuration.Endpoint, filePath)
	if err != nil {
		return "", err
	}

	response, err := configuration.HTTPClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("upload returned HTTP %s", response.Status)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return parsePublishedURL(responseBody)
}

func createUploadRequest(endpoint string, filePath string) (*http.Request, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	filePart, err := writer.CreateFormFile("files[]", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(filePart, file); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request, nil
}

func parsePublishedURL(responseBody []byte) (string, error) {
	var response uploadResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("could not parse upload response: %w", err)
	}

	for _, file := range response.Files {
		if publishedURL := firstHTTPURL(file.URL, file.FileURL, file.Link); publishedURL != "" {
			return publishedURL, nil
		}
	}
	if publishedURL := firstHTTPURL(response.URL, response.FileURL, response.Link); publishedURL != "" {
		return publishedURL, nil
	}

	return "", fmt.Errorf("upload response does not contain a URL")
}

func firstHTTPURL(urls ...string) string {
	for _, candidate := range urls {
		if strings.HasPrefix(candidate, "http") {
			return candidate
		}
	}
	return ""
}

type uploadResponse struct {
	Files   []uploadedFile `json:"files"`
	URL     string         `json:"url"`
	FileURL string         `json:"fileUrl"`
	Link    string         `json:"link"`
}

type uploadedFile struct {
	URL     string `json:"url"`
	FileURL string `json:"fileUrl"`
	Link    string `json:"link"`
}
