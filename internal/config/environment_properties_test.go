package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvironmentProperties_processEnvironmentOverridesFileValues(t *testing.T) {
	workDir := t.TempDir()
	filePath := createEnvironmentFile(t, workDir, ".env", "URL=http://file\nUSER=file-user\nPASS=file-pass\n")

	properties, err := LoadEnvironmentProperties(
		filePath,
		EnvironmentProperties{"URL": "http://process", "PASS": ""},
	)
	if err != nil {
		t.Fatalf("LoadEnvironmentProperties() error = %v", err)
	}

	if properties["URL"] != "http://process" {
		t.Errorf("URL = %q, want %q", properties["URL"], "http://process")
	}
	if properties["USER"] != "file-user" {
		t.Errorf("USER = %q, want %q", properties["USER"], "file-user")
	}
	if properties["PASS"] != "file-pass" {
		t.Errorf("PASS = %q, want %q", properties["PASS"], "file-pass")
	}
}

func TestLoadEnvironmentProperties_parsesQuotedValuesAndIgnoresComments(t *testing.T) {
	filePath := createEnvironmentFile(t, t.TempDir(), ".env", "\n# comment\nURL = http://camera?token=one=two\nUSER=\"camera-user\"\nPASS='camera-pass'\nnot-an-entry\n")

	properties, err := LoadEnvironmentProperties(filePath, EnvironmentProperties{})
	if err != nil {
		t.Fatalf("LoadEnvironmentProperties() error = %v", err)
	}

	if properties["URL"] != "http://camera?token=one=two" {
		t.Errorf("URL = %q", properties["URL"])
	}
	if properties["USER"] != "camera-user" {
		t.Errorf("USER = %q", properties["USER"])
	}
	if properties["PASS"] != "camera-pass" {
		t.Errorf("PASS = %q", properties["PASS"])
	}
	if len(properties) != 3 {
		t.Errorf("len(properties) = %d, want 3", len(properties))
	}
}

func TestLoadEnvironmentProperties_missingFile_returnsProcessEnvironment(t *testing.T) {
	properties, err := LoadEnvironmentProperties(
		filepath.Join(t.TempDir(), ".env"),
		EnvironmentProperties{"URL": "http://process"},
	)
	if err != nil {
		t.Fatalf("LoadEnvironmentProperties() error = %v", err)
	}
	if properties["URL"] != "http://process" {
		t.Errorf("URL = %q, want %q", properties["URL"], "http://process")
	}
}

func createEnvironmentFile(t *testing.T, directory string, fileName string, contents string) string {
	t.Helper()

	filePath := filepath.Join(directory, fileName)
	if err := os.WriteFile(filePath, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return filePath
}
