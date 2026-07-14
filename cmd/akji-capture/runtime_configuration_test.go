package main

import (
	"path/filepath"
	"testing"
)

func TestEnvironmentFilePath_usesWorkDirectoryRoot(t *testing.T) {
	workDir := t.TempDir()

	filePath := environmentFilePath(workDir)

	expectedPath := filepath.Join(workDir, ".env")
	if filePath != expectedPath {
		t.Errorf("filePath = %q, want %q", filePath, expectedPath)
	}
}
