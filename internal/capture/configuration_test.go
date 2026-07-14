package capture

import (
	"testing"
	"time"

	"github.com/schrek1/akji-rc/internal/config"
)

func TestNewConfiguration_createsCaptureConfiguration(t *testing.T) {
	configuration, err := NewConfiguration(validEnvironmentProperties())
	if err != nil {
		t.Fatalf("NewConfiguration() error = %v", err)
	}

	if configuration.WebcamURL != "http://camera" {
		t.Errorf("WebcamURL = %q", configuration.WebcamURL)
	}
	if configuration.WebcamUser != "user" {
		t.Errorf("WebcamUser = %q", configuration.WebcamUser)
	}
	if configuration.WebcamPass != "pass" {
		t.Errorf("WebcamPass = %q", configuration.WebcamPass)
	}
	if configuration.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", configuration.Timeout, 5*time.Second)
	}
	if configuration.CaptureWindow != 3*time.Second {
		t.Errorf("CaptureWindow = %v, want %v", configuration.CaptureWindow, 3*time.Second)
	}
}

func TestNewConfiguration_missingWebcamURL_returnsError(t *testing.T) {
	properties := validEnvironmentProperties()
	delete(properties, webcamURLProperty)

	_, err := NewConfiguration(properties)
	if err == nil {
		t.Fatal("NewConfiguration() error = nil")
	}
}

func TestNewConfiguration_invalidTimingProperty_returnsError(t *testing.T) {
	properties := validEnvironmentProperties()
	properties[captureWindowProperty] = "0"

	_, err := NewConfiguration(properties)
	if err == nil {
		t.Fatal("NewConfiguration() error = nil")
	}
}

func TestNewConfiguration_nonNumericTimingProperty_returnsError(t *testing.T) {
	properties := validEnvironmentProperties()
	properties[timeoutProperty] = "abc"

	_, err := NewConfiguration(properties)
	if err == nil {
		t.Fatal("NewConfiguration() error = nil")
	}
}

func validEnvironmentProperties() config.EnvironmentProperties {
	return config.EnvironmentProperties{
		webcamURLProperty:  "http://camera",
		webcamUserProperty: "user",
		webcamPassProperty: "pass",
	}
}
