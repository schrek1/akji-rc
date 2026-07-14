package capture

import (
	"bytes"
	"fmt"
)

var (
	jpegSOI = []byte{0xff, 0xd8, 0xff}
	jpegEOI = []byte{0xff, 0xd9}
)

func ExtractJPEGFrame(source []byte) ([]byte, error) {
	startOffsets := markerOffsets(source, jpegSOI)
	endOffsets := markerOffsets(source, jpegEOI)
	if !hasCompleteJPEGMarkers(startOffsets, endOffsets) {
		return nil, fmt.Errorf("could not find enough JPEG markers (SOI: %d, EOI: %d)", len(startOffsets), len(endOffsets))
	}

	return extractPreferredJPEGFrame(source, startOffsets, endOffsets)
}

func hasCompleteJPEGMarkers(startOffsets []int, endOffsets []int) bool {
	return len(startOffsets) > 0 && len(endOffsets) > 0
}

func extractPreferredJPEGFrame(source []byte, startOffsets []int, endOffsets []int) ([]byte, error) {
	for _, startOffset := range preferredJPEGStartOffsets(startOffsets) {
		endOffset, found := firstJPEGEndOffsetAfter(startOffset, endOffsets)
		if found {
			return copyJPEGFrame(source, startOffset, endOffset), nil
		}
	}

	return nil, fmt.Errorf("could not extract a valid JPEG frame from stream")
}

func preferredJPEGStartOffsets(startOffsets []int) []int {
	// Preserve the legacy shell selection order: favor an interior frame, then retry from the first frame.
	// This reduces the chance of extracting a partial frame from either edge of a captured stream chunk.
	preferredIndex := len(startOffsets)/2 - 1
	if preferredIndex < 0 {
		preferredIndex = 0
	}

	return uniqueStartOffsets(startOffsets, preferredIndex)
}

func uniqueStartOffsets(startOffsets []int, preferredIndex int) []int {
	preferredStartOffset := startOffsets[preferredIndex]
	firstStartOffset := startOffsets[0]
	if preferredStartOffset == firstStartOffset {
		return []int{preferredStartOffset}
	}
	return []int{preferredStartOffset, firstStartOffset}
}

func firstJPEGEndOffsetAfter(startOffset int, endOffsets []int) (int, bool) {
	// EOI markers before this SOI belong to an earlier partial frame and cannot close the selected frame.
	for _, endOffset := range endOffsets {
		if endOffset > startOffset {
			return endOffset, true
		}
	}
	return 0, false
}

func copyJPEGFrame(source []byte, startOffset int, endOffset int) []byte {
	return append([]byte(nil), source[startOffset:endOffset+len(jpegEOI)]...)
}

func markerOffsets(source []byte, marker []byte) []int {
	offsets := []int{}
	searchFrom := 0
	for {
		index := bytes.Index(source[searchFrom:], marker)
		if index < 0 {
			return offsets
		}
		offset := searchFrom + index
		offsets = append(offsets, offset)
		// Advance past the whole marker so counting stays non-overlapping, matching the legacy `grep -o` scan.
		searchFrom = offset + len(marker)
	}
}
