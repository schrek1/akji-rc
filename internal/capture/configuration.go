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

type webcamSettings struct {
	url  string
	user string
	pass string
}

type captureTiming struct {
	timeout       time.Duration
	captureWindow time.Duration
}

func NewConfiguration(properties config.EnvironmentProperties) (Configuration, error) {
	webcam, err := newWebcamSettings(properties)
	if err != nil {
		return Configuration{}, err
	}
	timing, err := newCaptureTiming(properties)
	if err != nil {
		return Configuration{}, err
	}

	return Configuration{
		WebcamURL:     webcam.url,
		WebcamUser:    webcam.user,
		WebcamPass:    webcam.pass,
		Timeout:       timing.timeout,
		CaptureWindow: timing.captureWindow,
	}, nil
}

func newWebcamSettings(properties config.EnvironmentProperties) (webcamSettings, error) {
	url, err := requiredProperty(properties, webcamURLProperty)
	if err != nil {
		return webcamSettings{}, err
	}
	user, err := requiredProperty(properties, webcamUserProperty)
	if err != nil {
		return webcamSettings{}, err
	}
	pass, err := requiredProperty(properties, webcamPassProperty)
	if err != nil {
		return webcamSettings{}, err
	}

	return webcamSettings{url: url, user: user, pass: pass}, nil
}

func newCaptureTiming(properties config.EnvironmentProperties) (captureTiming, error) {
	timeout, err := durationProperty(properties, timeoutProperty, defaultTimeoutSeconds)
	if err != nil {
		return captureTiming{}, err
	}
	captureWindow, err := durationProperty(properties, captureWindowProperty, defaultCaptureWindowSec)
	if err != nil {
		return captureTiming{}, err
	}

	return captureTiming{timeout: timeout, captureWindow: captureWindow}, nil
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
