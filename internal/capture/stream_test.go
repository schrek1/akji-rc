package capture

import (
	"bytes"
	"testing"
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
