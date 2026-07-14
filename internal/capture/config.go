package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CapturingConfiguration struct {
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

func LoadConfig(environmentValues EnvConfigValues, workDir string) (CapturingConfiguration, error) {
	configurationValues, err := loadConfigurationValues(environmentValues, workDir)
	if err != nil {
		return CapturingConfiguration{}, err
	}

	return createCapturingConfiguration(configurationValues)
}

func loadConfigurationValues(environmentValues EnvConfigValues, workDir string) (EnvConfigValues, error) {
	configurationValues, err := loadConfigurationValuesFromFiles(workDir)
	if err != nil {
		return nil, err
	}

	mergeNonEmptyEnvironmentValues(configurationValues, environmentValues)
	return configurationValues, nil
}

func loadConfigurationValuesFromFiles(workDir string) (EnvConfigValues, error) {
	configurationValues := EnvConfigValues{}
	for _, configurationFilePath := range configurationFilePaths(workDir) {
		fileValues, err := readEnvFile(configurationFilePath)
		if err != nil {
			return nil, err
		}
		mergeConfigurationValues(configurationValues, fileValues)
	}
	return configurationValues, nil
}

func configurationFilePaths(workDir string) []string {
	return []string{
		filepath.Join(workDir, "app", ".env"),
		filepath.Join(workDir, ".env"),
	}
}

func mergeConfigurationValues(target EnvConfigValues, source EnvConfigValues) {
	for key, value := range source {
		target[key] = value
	}
}

func mergeNonEmptyEnvironmentValues(target EnvConfigValues, source EnvConfigValues) {
	for key, value := range source {
		if value != "" {
			target[key] = value
		}
	}
}

func createCapturingConfiguration(configurationValues EnvConfigValues) (CapturingConfiguration, error) {
	webcam, err := loadWebcamSettings(configurationValues)
	if err != nil {
		return CapturingConfiguration{}, err
	}
	timing, err := loadCaptureTiming(configurationValues)
	if err != nil {
		return CapturingConfiguration{}, err
	}

	return CapturingConfiguration{
		WebcamURL:     webcam.url,
		WebcamUser:    webcam.user,
		WebcamPass:    webcam.pass,
		Timeout:       timing.timeout,
		CaptureWindow: timing.captureWindow,
	}, nil
}

func loadWebcamSettings(configurationValues EnvConfigValues) (webcamSettings, error) {
	url, err := loadRequiredConfigurationValue(configurationValues, "WEBCAM_URL")
	if err != nil {
		return webcamSettings{}, err
	}
	user, err := loadRequiredConfigurationValue(configurationValues, "WEBCAM_USER")
	if err != nil {
		return webcamSettings{}, err
	}
	pass, err := loadRequiredConfigurationValue(configurationValues, "WEBCAM_PASS")
	if err != nil {
		return webcamSettings{}, err
	}

	return webcamSettings{
		url:  url,
		user: user,
		pass: pass,
	}, nil
}

func loadCaptureTiming(configurationValues EnvConfigValues) (captureTiming, error) {
	timeout, err := loadCaptureDuration(configurationValues, "TIMEOUT", 5)
	if err != nil {
		return captureTiming{}, err
	}
	captureWindow, err := loadCaptureDuration(configurationValues, "CAPTURE_WINDOW", 3)
	if err != nil {
		return captureTiming{}, err
	}

	return captureTiming{
		timeout:       timeout,
		captureWindow: captureWindow,
	}, nil
}

func loadCaptureDuration(configurationValues EnvConfigValues, key string, fallbackSeconds int) (time.Duration, error) {
	duration, err := parsePositiveDurationSeconds(configurationValues[key], fallbackSeconds)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return duration, nil
}

func loadRequiredConfigurationValue(configurationValues EnvConfigValues, key string) (string, error) {
	value := configurationValues[key]
	if value == "" {
		return "", fmt.Errorf("%s is mandatory", key)
	}
	return value, nil
}

func Environment() EnvConfigValues {
	environmentValues := EnvConfigValues{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			environmentValues[key] = value
		}
	}
	return environmentValues
}

func parsePositiveDurationSeconds(raw string, fallbackSeconds int) (time.Duration, error) {
	if raw == "" {
		return time.Duration(fallbackSeconds) * time.Second, nil
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("must be positive")
	}
	return time.Duration(seconds) * time.Second, nil
}
