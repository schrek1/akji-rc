package capture

import (
	"fmt"
	"strconv"
	"time"

	"github.com/schrek1/akji-rc/internal/config"
)

const (
	webcamURLProperty       = "WEBCAM_URL"
	webcamUserProperty      = "WEBCAM_USER"
	webcamPassProperty      = "WEBCAM_PASS"
	timeoutProperty         = "TIMEOUT"
	captureWindowProperty   = "CAPTURE_WINDOW"
	defaultTimeoutSeconds   = 5
	defaultCaptureWindowSec = 3
)

type Configuration struct {
	WebcamURL     string
	WebcamUser    string
	WebcamPass    string
	Timeout       time.Duration
	CaptureWindow time.Duration
}

func NewConfiguration(properties config.EnvironmentProperties) (Configuration, error) {
	webcamURL, err := requiredProperty(properties, webcamURLProperty)
	if err != nil {
		return Configuration{}, err
	}
	webcamUser, err := requiredProperty(properties, webcamUserProperty)
	if err != nil {
		return Configuration{}, err
	}
	webcamPass, err := requiredProperty(properties, webcamPassProperty)
	if err != nil {
		return Configuration{}, err
	}
	timeout, err := durationProperty(properties, timeoutProperty, defaultTimeoutSeconds)
	if err != nil {
		return Configuration{}, err
	}
	captureWindow, err := durationProperty(properties, captureWindowProperty, defaultCaptureWindowSec)
	if err != nil {
		return Configuration{}, err
	}

	return Configuration{
		WebcamURL:     webcamURL,
		WebcamUser:    webcamUser,
		WebcamPass:    webcamPass,
		Timeout:       timeout,
		CaptureWindow: captureWindow,
	}, nil
}

func requiredProperty(properties config.EnvironmentProperties, propertyName string) (string, error) {
	value := properties[propertyName]
	if value == "" {
		return "", fmt.Errorf("%s is mandatory", propertyName)
	}
	return value, nil
}

func durationProperty(properties config.EnvironmentProperties, propertyName string, defaultSeconds int) (time.Duration, error) {
	seconds, err := positiveSeconds(properties[propertyName], defaultSeconds)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", propertyName, err)
	}
	return time.Duration(seconds) * time.Second, nil
}

func positiveSeconds(value string, defaultSeconds int) (int, error) {
	if value == "" {
		return defaultSeconds, nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("must be positive")
	}
	return seconds, nil
}
