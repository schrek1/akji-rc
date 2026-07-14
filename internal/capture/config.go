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

func LoadConfig(env map[string]string, workDir string) (CapturingConfiguration, error) {
	values, err := loadConfigurationValues(env, workDir)
	if err != nil {
		return CapturingConfiguration{}, err
	}

	return configurationFromValues(values)
}

func loadConfigurationValues(env map[string]string, workDir string) (map[string]string, error) {
	values := map[string]string{}
	for _, envFile := range []string{
		filepath.Join(workDir, "app", ".env"),
		filepath.Join(workDir, ".env"),
	} {
		fileValues, err := readEnvFile(envFile)
		if err != nil {
			return nil, err
		}
		mergeValues(values, fileValues)
	}
	mergeNonEmptyValues(values, env)
	return values, nil
}

func mergeValues(target map[string]string, source map[string]string) {
	for key, value := range source {
		target[key] = value
	}
}

func mergeNonEmptyValues(target map[string]string, source map[string]string) {
	for key, value := range source {
		if value != "" {
			target[key] = value
		}
	}
}

func configurationFromValues(values map[string]string) (CapturingConfiguration, error) {
	timeout, err := durationValue(values, "TIMEOUT", 5)
	if err != nil {
		return CapturingConfiguration{}, fmt.Errorf("invalid TIMEOUT: %w", err)
	}
	captureWindow, err := durationValue(values, "CAPTURE_WINDOW", 3)
	if err != nil {
		return CapturingConfiguration{}, fmt.Errorf("invalid CAPTURE_WINDOW: %w", err)
	}

	webcamURL, err := requiredValue(values, "WEBCAM_URL")
	if err != nil {
		return CapturingConfiguration{}, err
	}
	webcamUser, err := requiredValue(values, "WEBCAM_USER")
	if err != nil {
		return CapturingConfiguration{}, err
	}
	webcamPass, err := requiredValue(values, "WEBCAM_PASS")
	if err != nil {
		return CapturingConfiguration{}, err
	}

	return CapturingConfiguration{
		WebcamURL:     webcamURL,
		WebcamUser:    webcamUser,
		WebcamPass:    webcamPass,
		Timeout:       timeout,
		CaptureWindow: captureWindow,
	}, nil
}

func durationValue(values map[string]string, key string, fallback int) (time.Duration, error) {
	return secondsValue(values[key], fallback)
}

func requiredValue(values map[string]string, key string) (string, error) {
	value := values[key]
	if value == "" {
		return "", fmt.Errorf("%s is mandatory", key)
	}
	return value, nil
}

func Environment() map[string]string {
	values := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}

func secondsValue(raw string, fallback int) (time.Duration, error) {
	if raw == "" {
		return time.Duration(fallback) * time.Second, nil
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
