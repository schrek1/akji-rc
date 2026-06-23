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

func DownloadStream(config Config) ([]byte, error) {
	parsedURL, err := url.Parse(config.WebcamURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme != "http" {
		return nil, fmt.Errorf("unsupported webcam URL scheme %q", parsedURL.Scheme)
	}

	host := parsedURL.Host
	if !strings.Contains(host, ":") {
		host += ":80"
	}

	conn, err := net.DialTimeout("tcp", host, config.Timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(config.CaptureWindow)); err != nil {
		return nil, err
	}

	path := parsedURL.RequestURI()
	if path == "" {
		path = "/"
	}

	request := fmt.Sprintf("GET %s HTTP/1.0\r\nHost: %s\r\nConnection: close\r\n", path, parsedURL.Host)
	if config.WebcamUser != "" {
		token := base64.StdEncoding.EncodeToString([]byte(config.WebcamUser + ":" + config.WebcamPass))
		request += "Authorization: Basic " + token + "\r\n"
	}
	request += "\r\n"

	if _, err := io.WriteString(conn, request); err != nil {
		return nil, err
	}

	body, err := readUntilDeadline(conn)
	if err != nil && len(body) == 0 {
		return nil, err
	}
	return stripHTTPHeaders(body), nil
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
	return nil
}
