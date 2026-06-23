package capture

import (
	"bytes"
	"fmt"
)

var (
	jpegSOI = []byte{0xff, 0xd8, 0xff}
	jpegEOI = []byte{0xff, 0xd9}
)

func ExtractJPEG(source []byte) ([]byte, error) {
	sois := markerOffsets(source, jpegSOI)
	eois := markerOffsets(source, jpegEOI)
	if len(sois) == 0 || len(eois) == 0 {
		return nil, fmt.Errorf("could not find enough JPEG markers (SOI: %d, EOI: %d)", len(sois), len(eois))
	}

	mid := len(sois) / 2
	if mid < 1 {
		mid = 1
	}

	targetIndices := []int{mid - 1, 0}
	for _, index := range targetIndices {
		if index < 0 || index >= len(sois) {
			continue
		}
		start := sois[index]
		for _, end := range eois {
			if end > start {
				frame := append([]byte(nil), source[start:end+len(jpegEOI)]...)
				if bytes.HasSuffix(frame, jpegEOI) {
					return frame, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("could not extract a valid JPEG frame from stream")
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
		searchFrom = offset + 1
	}
}
