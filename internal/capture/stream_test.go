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

func TestDownloadStream_httpEndpoint_sendsRequestAndReturnsBody(t *testing.T) {
	webcamListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer webcamListener.Close()

	expectedJPEGFrame := []byte("\xff\xd8\xffDATA\xff\xd9")
	receivedRequest := make(chan string, 1)
	serverErrors := make(chan error, 1)
	go serveMJPEGResponse(webcamListener, expectedJPEGFrame, receivedRequest, serverErrors)

	capturingConfiguration := CapturingConfiguration{
		WebcamURL:     "http://" + webcamListener.Addr().String() + "/camera?quality=high",
		WebcamUser:    "camera-user",
		WebcamPass:    "camera-pass",
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	jpegFrame, err := DownloadStream(capturingConfiguration)
	if err != nil {
		t.Fatalf("DownloadStream() error = %v", err)
	}
	if err := <-serverErrors; err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(jpegFrame, expectedJPEGFrame) {
		t.Fatalf("jpegFrame = %q, want %q", jpegFrame, expectedJPEGFrame)
	}
	assertMJPEGRequest(t, <-receivedRequest, webcamListener.Addr().String())
}

func TestDownloadStream_httpsEndpoint_returnsError(t *testing.T) {
	capturingConfiguration := CapturingConfiguration{
		WebcamURL:     "https://camera",
		Timeout:       time.Second,
		CaptureWindow: time.Second,
	}

	_, err := DownloadStream(capturingConfiguration)
	if err == nil {
		t.Fatal("DownloadStream() error = nil")
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
