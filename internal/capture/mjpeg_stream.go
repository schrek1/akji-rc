package capture

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"
)

type webcamEndpoint struct {
	connectionHost string
	requestHost    string
	requestPath    string
}

// DownloadMJPEGStream speaks minimal HTTP/1.0 over a raw TCP socket instead of using net/http,
// because the target cameras serve an HTTP/0.9-style MJPEG stream with minimal/no response headers
// that net/http rejects. Mirrors the legacy `curl --http0.9` behavior.
func DownloadMJPEGStream(config Configuration) ([]byte, error) {
	endpoint, err := parseWebcamEndpoint(config.WebcamURL)
	if err != nil {
		return nil, err
	}

	connection, err := openWebcamConnection(endpoint, config.Timeout)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	if err := setCaptureDeadline(connection, config.CaptureWindow); err != nil {
		return nil, err
	}

	request := createMJPEGRequest(endpoint, config)
	return downloadMJPEGResponse(connection, request)
}

func parseWebcamEndpoint(webcamURL string) (webcamEndpoint, error) {
	parsedURL, err := url.Parse(webcamURL)
	if err != nil {
		return webcamEndpoint{}, err
	}
	if parsedURL.Scheme != "http" {
		return webcamEndpoint{}, fmt.Errorf("unsupported webcam URL scheme %q", parsedURL.Scheme)
	}

	return webcamEndpoint{
		connectionHost: connectionHost(parsedURL),
		requestHost:    parsedURL.Host,
		requestPath:    parsedURL.RequestURI(),
	}, nil
}

func connectionHost(parsedURL *url.URL) string {
	if strings.Contains(parsedURL.Host, ":") {
		return parsedURL.Host
	}
	return parsedURL.Host + ":80"
}

func openWebcamConnection(endpoint webcamEndpoint, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", endpoint.connectionHost, timeout)
}

func setCaptureDeadline(connection net.Conn, captureWindow time.Duration) error {
	return connection.SetDeadline(time.Now().Add(captureWindow))
}

func createMJPEGRequest(endpoint webcamEndpoint, config Configuration) string {
	request := fmt.Sprintf("GET %s HTTP/1.0\r\nHost: %s\r\nConnection: close\r\n", endpoint.requestPath, endpoint.requestHost)
	if config.WebcamUser != "" {
		request += basicAuthorizationHeader(config.WebcamUser, config.WebcamPass)
	}
	return request + "\r\n"
}

func basicAuthorizationHeader(user string, pass string) string {
	credentials := user + ":" + pass
	token := base64.StdEncoding.EncodeToString([]byte(credentials))
	return "Authorization: Basic " + token + "\r\n"
}

func downloadMJPEGResponse(connection net.Conn, request string) ([]byte, error) {
	if _, err := io.WriteString(connection, request); err != nil {
		return nil, err
	}

	response, err := readUntilDeadline(connection)
	if err != nil && len(response) == 0 {
		return nil, err
	}
	return stripHTTPHeaders(response), nil
}

func readUntilDeadline(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(reader)
	if err == nil {
		return data, nil
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return data, nil
	}
	return data, err
}

func stripHTTPHeaders(data []byte) []byte {
	buffer := bufio.NewReader(bytes.NewReader(data))
	line, err := buffer.ReadString('\n')
	if err != nil {
		return data
	}
	if !strings.HasPrefix(line, "HTTP/") {
		return data
	}

	headerEnd := []byte("\r\n\r\n")
	if index := bytes.Index(data, headerEnd); index >= 0 {
		return data[index+len(headerEnd):]
	}
	headerEnd = []byte("\n\n")
	if index := bytes.Index(data, headerEnd); index >= 0 {
		return data[index+len(headerEnd):]
	}
	// No header terminator (e.g. a response truncated by the capture window): keep the data so the
	// JPEG scan can still find a frame rather than discarding everything.
	return data
}
