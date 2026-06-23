package capture

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	WebcamURL     string
	WebcamUser    string
	WebcamPass    string
	Timeout       time.Duration
	CaptureWindow time.Duration
}

func LoadConfig(env map[string]string, workDir string) (Config, error) {
	values := map[string]string{}
	for _, envFile := range []string{
		filepath.Join(workDir, "app", ".env"),
		filepath.Join(workDir, ".env"),
	} {
		fileValues, err := readEnvFile(envFile)
		if err != nil {
			return Config{}, err
		}
		for key, value := range fileValues {
			values[key] = value
		}
	}
	for key, value := range env {
		if value != "" {
			values[key] = value
		}
	}

	timeout, err := secondsValue(values["TIMEOUT"], 5)
	if err != nil {
		return Config{}, fmt.Errorf("invalid TIMEOUT: %w", err)
	}
	captureWindow, err := secondsValue(values["CAPTURE_WINDOW"], 3)
	if err != nil {
		return Config{}, fmt.Errorf("invalid CAPTURE_WINDOW: %w", err)
	}

	config := Config{
		WebcamURL:     values["WEBCAM_URL"],
		WebcamUser:    values["WEBCAM_USER"],
		WebcamPass:    values["WEBCAM_PASS"],
		Timeout:       timeout,
		CaptureWindow: captureWindow,
	}
	if config.WebcamURL == "" {
		return Config{}, fmt.Errorf("WEBCAM_URL is mandatory")
	}
	if config.WebcamUser == "" {
		return Config{}, fmt.Errorf("WEBCAM_USER is mandatory")
	}
	if config.WebcamPass == "" {
		return Config{}, fmt.Errorf("WEBCAM_PASS is mandatory")
	}
	return config, nil
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

func readEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = trimEnvValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func trimEnvValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)
	return value
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
