package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultOutputPath_usesWorkDirectoryCapturesDirectory(t *testing.T) {
	workDir := t.TempDir()

	outputPath := defaultOutputPath(workDir, time.Date(2026, time.July, 14, 17, 30, 0, 0, time.UTC))

	expectedPath := filepath.Join(workDir, "captures", "webcam_2026-07-14_17-30-00.jpg")
	if outputPath != expectedPath {
		t.Fatalf("defaultOutputPath() = %q, want %q", outputPath, expectedPath)
	}
}
