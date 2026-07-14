package main

import (
	"bytes"
	"testing"
)

func TestParseScriptOptions_shortAliases_resolveOutputAndTimeLapse(t *testing.T) {
	outputOptions, err := parseOptions([]string{"-o", "frame.jpg"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}
	if outputOptions.outputPath != "frame.jpg" {
		t.Errorf("outputPath = %q, want %q", outputOptions.outputPath, "frame.jpg")
	}

	timeLapseOptions, err := parseOptions([]string{"-tl", "5"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}
	if timeLapseOptions.timeLapseSeconds != 5 {
		t.Errorf("timeLapseSeconds = %d, want %d", timeLapseOptions.timeLapseSeconds, 5)
	}
}

func TestParseScriptOptions_outputAndTimeLapse_returnsError(t *testing.T) {
	_, err := parseOptions([]string{"--out", "frame.jpg", "--timeLapse", "5"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("parseOptions() error = nil")
	}
}

func TestParseScriptOptions_help_marksOptionsAsHelpRequested(t *testing.T) {
	output := &bytes.Buffer{}

	options, err := parseOptions([]string{"--help"}, output)
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}
	if !options.helpRequested {
		t.Fatal("helpRequested = false")
	}
	if !bytes.Contains(output.Bytes(), []byte("Usage: akji-capture [OPTIONS]")) {
		t.Fatalf("help output = %q", output.String())
	}
}
