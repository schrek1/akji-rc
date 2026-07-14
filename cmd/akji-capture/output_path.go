package main

import (
	"path/filepath"
	"time"
)

func defaultOutputPath(workDir string, captureTime time.Time) string {
	return filepath.Join(workDir, "captures", captureFileName(captureTime))
}

func captureFileName(captureTime time.Time) string {
	return "webcam_" + captureTime.Format("2006-01-02_15-04-05") + ".jpg"
}
