package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvironmentProperties_osOverridesFileValuesIncludingEmptyValues(t *testing.T) {
	fileProperties := EnvironmentProperties{
		"URL":         "http://file",
		"USER":        "file-user",
		"PASS":        "file-pass",
		"STATIC_ONLY": "file-only",
	}
	osProperties := EnvironmentProperties{"URL": "http://process", "USER": "process-user", "PASS": ""}

	properties := LoadEnvironmentProperties(fileProperties, osProperties)

	if properties["URL"] != "http://process" {
		t.Errorf("URL = %q, want %q", properties["URL"], "http://process")
	}
	if properties["USER"] != "process-user" {
		t.Errorf("USER = %q, want %q", properties["USER"], "process-user")
	}
	if properties["PASS"] != "" {
		t.Errorf("PASS = %q, want empty string", properties["PASS"])
	}
	if properties["STATIC_ONLY"] != "file-only" {
		t.Errorf("STATIC_ONLY = %q, want %q", properties["STATIC_ONLY"], "file-only")
	}
}

func TestReadFileEnvProperties_parsesQuotedValuesAndIgnoresComments(t *testing.T) {
	filePath := createEnvironmentFile(t, t.TempDir(), ".env", "\n# comment\nURL = http://camera?token=one=two\nUSER=\"camera-user\"\nPASS='camera-pass'\nnot-an-entry\n")

	properties, err := ReadFileEnvProperties(filePath)
	if err != nil {
		t.Fatalf("ReadFileEnvProperties() error = %v", err)
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

func TestReadFileEnvProperties_missingFile_returnsEmptyProperties(t *testing.T) {
	properties, err := ReadFileEnvProperties(filepath.Join(t.TempDir(), ".env"))
	if err != nil {
		t.Fatalf("ReadFileEnvProperties() error = %v", err)
	}
	if len(properties) != 0 {
		t.Errorf("len(properties) = %d, want 0", len(properties))
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
