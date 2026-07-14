package capture

import (
	"bytes"
	"testing"
)

func TestExtractJPEGFrame_validMockData_extractsFrameWithMarkers(t *testing.T) {
	source := []byte("some junk\xff\xd8\xffimage data here\xff\xd9more junk")

	frame, err := ExtractJPEGFrame(source)
	if err != nil {
		t.Fatalf("ExtractJPEGFrame() error = %v", err)
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

func TestExtractJPEGFrame_missingEOI_returnsError(t *testing.T) {
	source := []byte("some junk\xff\xd8\xffimage data without end")

	frame, err := ExtractJPEGFrame(source)
	if err == nil {
		t.Fatalf("ExtractJPEGFrame() error = nil, frame = %q", frame)
	}
}

func TestExtractJPEGFrame_multipleFrames_picksSameMiddleFrameAsShellScript(t *testing.T) {
	source := []byte("AA\xff\xd8\xffF1\xff\xd9BB\xff\xd8\xffF2\xff\xd9CC\xff\xd8\xffF3\xff\xd9DD\xff\xd8\xffF4\xff\xd9")

	frame, err := ExtractJPEGFrame(source)
	if err != nil {
		t.Fatalf("ExtractJPEGFrame() error = %v", err)
	}

	if !bytes.Contains(frame, []byte("F2")) {
		t.Fatalf("frame = %q, want frame containing F2", frame)
	}
}

func TestExtractJPEGFrame_ciMockData_extractsFrame(t *testing.T) {
	source := []byte("\xff\xd8\xffCI_DATA\xff\xd9")

	frame, err := ExtractJPEGFrame(source)
	if err != nil {
		t.Fatalf("ExtractJPEGFrame() error = %v", err)
	}

	if !bytes.Equal(frame, source) {
		t.Fatalf("frame = %q, want %q", frame, source)
	}
}

func TestMarkerOffsets_adjacentMarkers_countedNonOverlappingLikeShell(t *testing.T) {
	// A naive one-byte advance would report an overlapping SOI at offset 2;
	// the shell `grep -o` scan counts non-overlapping markers only.
	source := []byte("\xff\xd8\xff\xd8\xff")

	offsets := markerOffsets(source, jpegSOI)

	if len(offsets) != 1 || offsets[0] != 0 {
		t.Fatalf("markerOffsets() = %v, want [0]", offsets)
	}
}

func TestExtractJPEGFrame_ignoresEndMarkerBeforeFrameStart(t *testing.T) {
	source := []byte("\xff\xd9before\xff\xd8\xffFRAME\xff\xd9")

	frame, err := ExtractJPEGFrame(source)
	if err != nil {
		t.Fatalf("ExtractJPEGFrame() error = %v", err)
	}

	expectedFrame := []byte("\xff\xd8\xffFRAME\xff\xd9")
	if !bytes.Equal(frame, expectedFrame) {
		t.Fatalf("frame = %q, want %q", frame, expectedFrame)
	}
}
