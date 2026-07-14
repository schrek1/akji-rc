package capture

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"net"
	"strings"
	"testing"
	"time"
)

func TestStripHTTPHeaders_httpResponse_returnsBody(t *testing.T) {
	data := []byte("HTTP/1.0 200 OK\r\nContent-Type: multipart/x-mixed-replace\r\n\r\n\xff\xd8\xffDATA\xff\xd9")

	body := stripHTTPHeaders(data)

	if !bytes.Equal(body, []byte("\xff\xd8\xffDATA\xff\xd9")) {
		t.Fatalf("body = %q", body)
	}
}

func TestStripHTTPHeaders_headerlessStream_keepsData(t *testing.T) {
	data := []byte("\xff\xd8\xffDATA\xff\xd9")

	body := stripHTTPHeaders(data)

	if !bytes.Equal(body, data) {
		t.Fatalf("body = %q, want %q", body, data)
	}
}

func TestStripHTTPHeaders_httpPrefixWithoutTerminator_keepsData(t *testing.T) {
	data := []byte("HTTP/1.0 200 OK\r\nContent-Type: image/jpeg\r\n\xff\xd8\xffDATA\xff\xd9")

	body := stripHTTPHeaders(data)

	if !bytes.Equal(body, data) {
		t.Fatalf("body = %q, want %q", body, data)
	}
}

func TestDownloadMJPEGStream_httpEndpoint_sendsRequestAndReturnsBody(t *testing.T) {
	webcamListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer webcamListener.Close()

	expectedJPEGFrame := []byte("\xff\xd8\xffDATA\xff\xd9")
	receivedRequest := make(chan string, 1)
	serverErrors := make(chan error, 1)
	go serveMJPEGResponse(webcamListener, expectedJPEGFrame, receivedRequest, serverErrors)

	capturingConfiguration := Configuration{
		WebcamURL:     "http://" + webcamListener.Addr().String() + "/camera?quality=high",
		WebcamUser:    "camera-user",
		WebcamPass:    "camera-pass",
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	jpegFrame, err := DownloadMJPEGStream(capturingConfiguration)
	if err != nil {
		t.Fatalf("DownloadMJPEGStream() error = %v", err)
	}
	if err := <-serverErrors; err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(jpegFrame, expectedJPEGFrame) {
		t.Fatalf("jpegFrame = %q, want %q", jpegFrame, expectedJPEGFrame)
	}
	assertMJPEGRequest(t, <-receivedRequest, webcamListener.Addr().String())
}

func TestParseWebcamEndpoint_resolvesConnectionHostAndRequestPath(t *testing.T) {
	testCases := []struct {
		name               string
		webcamURL          string
		expectedConnection string
		expectedPath       string
	}{
		{"host without port defaults to 80", "http://camera/stream", "camera:80", "/stream"},
		{"explicit port is preserved", "http://camera:8080/stream", "camera:8080", "/stream"},
		{"query string is kept in request path", "http://camera/cam?quality=high", "camera:80", "/cam?quality=high"},
		{"empty path resolves to root", "http://camera", "camera:80", "/"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			endpoint, err := parseWebcamEndpoint(testCase.webcamURL)
			if err != nil {
				t.Fatalf("parseWebcamEndpoint() error = %v", err)
			}
			if endpoint.connectionHost != testCase.expectedConnection {
				t.Errorf("connectionHost = %q, want %q", endpoint.connectionHost, testCase.expectedConnection)
			}
			if endpoint.requestPath != testCase.expectedPath {
				t.Errorf("requestPath = %q, want %q", endpoint.requestPath, testCase.expectedPath)
			}
		})
	}
}

func TestDownloadMJPEGStream_httpsEndpoint_returnsError(t *testing.T) {
	capturingConfiguration := Configuration{
		WebcamURL:     "https://camera",
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	_, err := DownloadMJPEGStream(capturingConfiguration)
	if err == nil {
		t.Fatal("DownloadMJPEGStream() error = nil")
	}
}

func serveMJPEGResponse(listener net.Listener, jpegFrame []byte, receivedRequest chan<- string, serverErrors chan<- error) {
	connection, err := listener.Accept()
	if err != nil {
		serverErrors <- err
		return
	}
	defer connection.Close()

	request, err := readHTTPRequest(connection)
	if err != nil {
		serverErrors <- err
		return
	}
	receivedRequest <- request

	response := "HTTP/1.0 200 OK\r\nContent-Type: multipart/x-mixed-replace\r\n\r\n"
	_, err = connection.Write(append([]byte(response), jpegFrame...))
	serverErrors <- err
}

func readHTTPRequest(connection net.Conn) (string, error) {
	reader := bufio.NewReader(connection)
	var request strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		request.WriteString(line)
		if line == "\r\n" {
			return request.String(), nil
		}
	}
}

func assertMJPEGRequest(t *testing.T, request string, webcamHost string) {
	t.Helper()

	if !strings.Contains(request, "GET /camera?quality=high HTTP/1.0\r\n") {
		t.Errorf("request does not contain expected path: %q", request)
	}
	if !strings.Contains(request, "Host: "+webcamHost+"\r\n") {
		t.Errorf("request does not contain expected host: %q", request)
	}
	expectedToken := base64.StdEncoding.EncodeToString([]byte("camera-user:camera-pass"))
	if !strings.Contains(request, "Authorization: Basic "+expectedToken+"\r\n") {
		t.Errorf("request does not contain expected authorization: %q", request)
	}
}
