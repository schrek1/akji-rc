package main

import (
	"testing"

	"github.com/schrek1/akji-rc/internal/validate"
)

func TestFramePathFromArgs_withoutPath_usesDefaultPath(t *testing.T) {
	framePath, err := framePathFromArgs(nil)

	if err != nil {
		t.Fatalf("framePathFromArgs() error = %v", err)
	}
	if framePath != defaultFramePath {
		t.Errorf("framePath = %q, want %q", framePath, defaultFramePath)
	}
}

func TestFramePathFromArgs_withPath_returnsProvidedPath(t *testing.T) {
	framePath, err := framePathFromArgs([]string{"custom-frame.jpg"})

	if err != nil {
		t.Fatalf("framePathFromArgs() error = %v", err)
	}
	if framePath != "custom-frame.jpg" {
		t.Errorf("framePath = %q, want %q", framePath, "custom-frame.jpg")
	}
}

func TestFramePathFromArgs_withMultiplePaths_returnsError(t *testing.T) {
	_, err := framePathFromArgs([]string{"one.jpg", "two.jpg"})

	if err == nil {
		t.Fatal("framePathFromArgs() error = nil")
	}
}

func TestMinimumSizeBytes_withoutValue_usesDefault(t *testing.T) {
	minimumSize, err := minimumSizeBytes("")

	if err != nil {
		t.Fatalf("minimumSizeBytes() error = %v", err)
	}
	if minimumSize != validate.DefaultMinimumSizeBytes {
		t.Errorf("minimumSize = %d, want %d", minimumSize, validate.DefaultMinimumSizeBytes)
	}
}

func TestMinimumSizeBytes_withInvalidValue_returnsError(t *testing.T) {
	_, err := minimumSizeBytes("invalid")

	if err == nil {
		t.Fatal("minimumSizeBytes() error = nil")
	}
}
