package capture

import (
	"bytes"
	"testing"
)

func TestExtractJPEG_validMockData_extractsFrameWithMarkers(t *testing.T) {
	source := []byte("some junk\xff\xd8\xffimage data here\xff\xd9more junk")

	frame, err := ExtractJPEG(source)
	if err != nil {
		t.Fatalf("ExtractJPEG() error = %v", err)
	}

	if !bytes.HasPrefix(frame, jpegSOI) {
		t.Fatalf("frame does not start with SOI marker: % x", frame)
	}
	if !bytes.HasSuffix(frame, jpegEOI) {
		t.Fatalf("frame does not end with EOI marker: % x", frame)
	}
	expected := []byte("\xff\xd8\xffimage data here\xff\xd9")
	if !bytes.Equal(frame, expected) {
		t.Fatalf("frame = %q, want %q", frame, expected)
	}
}

func TestExtractJPEG_missingEOI_returnsError(t *testing.T) {
	source := []byte("some junk\xff\xd8\xffimage data without end")

	frame, err := ExtractJPEG(source)
	if err == nil {
		t.Fatalf("ExtractJPEG() error = nil, frame = %q", frame)
	}
}

func TestExtractJPEG_multipleFrames_picksSameMiddleFrameAsShellScript(t *testing.T) {
	source := []byte("AA\xff\xd8\xffF1\xff\xd9BB\xff\xd8\xffF2\xff\xd9CC\xff\xd8\xffF3\xff\xd9DD\xff\xd8\xffF4\xff\xd9")

	frame, err := ExtractJPEG(source)
	if err != nil {
		t.Fatalf("ExtractJPEG() error = %v", err)
	}

	if !bytes.Contains(frame, []byte("F2")) {
		t.Fatalf("frame = %q, want frame containing F2", frame)
	}
}

func TestExtractJPEG_ciMockData_extractsFrame(t *testing.T) {
	source := []byte("\xff\xd8\xffCI_DATA\xff\xd9")

	frame, err := ExtractJPEG(source)
	if err != nil {
		t.Fatalf("ExtractJPEG() error = %v", err)
	}

	if !bytes.Equal(frame, source) {
		t.Fatalf("frame = %q, want %q", frame, source)
	}
}

func TestExtractJPEG_ignoresEndMarkerBeforeFrameStart(t *testing.T) {
	source := []byte("\xff\xd9before\xff\xd8\xffFRAME\xff\xd9")

	frame, err := ExtractJPEG(source)
	if err != nil {
		t.Fatalf("ExtractJPEG() error = %v", err)
	}

	expectedFrame := []byte("\xff\xd8\xffFRAME\xff\xd9")
	if !bytes.Equal(frame, expectedFrame) {
		t.Fatalf("frame = %q, want %q", frame, expectedFrame)
	}
}
