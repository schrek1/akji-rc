package main

import (
	"os"
	"path/filepath"
	"time"
)

func defaultOutputPath(workDir string, captureTime time.Time) string {
	return filepath.Join(defaultCapturesDirectory(workDir), captureFileName(captureTime))
}

func defaultCapturesDirectory(workDir string) string {
	capturesDirectory := filepath.Join(workDir, "app", "captures")
	if _, err := os.Stat(filepath.Join(workDir, "app")); os.IsNotExist(err) {
		capturesDirectory = filepath.Join(workDir, "captures")
	}
	return capturesDirectory
}

func captureFileName(captureTime time.Time) string {
	return "webcam_" + captureTime.Format("2006-01-02_15-04-05") + ".jpg"
}
