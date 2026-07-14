package capture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEnvFile_parsesSupportedEntries(t *testing.T) {
	path := createTestConfigFile(t, "\n# comment\nWEBCAM_URL = http://camera\nWEBCAM_USER=\"camera-user\"\nWEBCAM_PASS='camera-pass'\nnot-an-entry\n")

	values, err := readEnvFile(path)
	if err != nil {
		t.Fatalf("readEnvFile() error = %v", err)
	}

	if values["WEBCAM_URL"] != "http://camera" {
		t.Errorf("WEBCAM_URL = %q", values["WEBCAM_URL"])
	}
	if values["WEBCAM_USER"] != "camera-user" {
		t.Errorf("WEBCAM_USER = %q", values["WEBCAM_USER"])
	}
	if values["WEBCAM_PASS"] != "camera-pass" {
		t.Errorf("WEBCAM_PASS = %q", values["WEBCAM_PASS"])
	}
	if len(values) != 3 {
		t.Errorf("len(values) = %d, want 3", len(values))
	}
}

func createTestConfigFile(t *testing.T, contents string) string {
	t.Helper()
	filepath := filepath.Join(t.TempDir(), ".env")

	if err := os.WriteFile(filepath, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	return filepath
}

func TestReadEnvFile_missingFile_returnsEmptyValues(t *testing.T) {
	values, err := readEnvFile(filepath.Join(t.TempDir(), ".env"))
	if err != nil {
		t.Fatalf("readEnvFile() error = %v", err)
	}
	if len(values) != 0 {
		t.Errorf("len(values) = %d, want 0", len(values))
	}
}

func TestReadEnvFile_preservesEmptyAndEqualsContainingValues(t *testing.T) {
	path := createTestConfigFile(t, "  WEBCAM_URL  =  http://camera?token=one=two  \nEMPTY_VALUE=\n")

	values, err := readEnvFile(path)
	if err != nil {
		t.Fatalf("readEnvFile() error = %v", err)
	}

	if values["WEBCAM_URL"] != "http://camera?token=one=two" {
		t.Errorf("WEBCAM_URL = %q", values["WEBCAM_URL"])
	}
	if values["EMPTY_VALUE"] != "" {
		t.Errorf("EMPTY_VALUE = %q", values["EMPTY_VALUE"])
	}
}
