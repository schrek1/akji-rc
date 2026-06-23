package capture

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig_envOverridesDotEnvFiles(t *testing.T) {
	workDir := t.TempDir()
	appDir := filepath.Join(workDir, "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, ".env"), []byte("WEBCAM_URL=http://app-env\nWEBCAM_USER=app-user\nWEBCAM_PASS=app-pass\nTIMEOUT=7\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, ".env"), []byte("WEBCAM_URL=http://cwd-env\nWEBCAM_USER=cwd-user\nWEBCAM_PASS=cwd-pass\nCAPTURE_WINDOW=4\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	config, err := LoadConfig(map[string]string{
		"WEBCAM_URL": "http://env",
	}, workDir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.WebcamURL != "http://env" {
		t.Fatalf("WebcamURL = %q", config.WebcamURL)
	}
	if config.WebcamUser != "cwd-user" {
		t.Fatalf("WebcamUser = %q", config.WebcamUser)
	}
	if config.WebcamPass != "cwd-pass" {
		t.Fatalf("WebcamPass = %q", config.WebcamPass)
	}
	if config.Timeout != 7*time.Second {
		t.Fatalf("Timeout = %v", config.Timeout)
	}
	if config.CaptureWindow != 4*time.Second {
		t.Fatalf("CaptureWindow = %v", config.CaptureWindow)
	}
}

func TestLoadConfig_missingRequiredValues_returnsError(t *testing.T) {
	_, err := LoadConfig(map[string]string{}, t.TempDir())
	if err == nil {
		t.Fatal("LoadConfig() error = nil")
	}
}
