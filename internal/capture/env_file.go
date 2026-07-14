package capture

import (
	"bufio"
	"os"
	"strings"
)

type EnvConfigValues map[string]string

func readEnvFile(path string) (EnvConfigValues, error) {
	configFile, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	defer configFile.Close()

	resolvedConfigValues := EnvConfigValues{}
	scanner := bufio.NewScanner(configFile)

	readConfigFile(scanner, resolvedConfigValues)

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return resolvedConfigValues, nil
}

func readConfigFile(scanner *bufio.Scanner, values EnvConfigValues) {
	for scanner.Scan() {
		line := getScannedLine(scanner)
		if isBlankLineOrComment(line) {
			continue
		}
		key, value, ok := parseConfigEntry(line)
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = trimEnvValue(value)
	}
}

func parseConfigEntry(line string) (string, string, bool) {
	return strings.Cut(line, "=")
}

func getScannedLine(scanner *bufio.Scanner) string {
	return strings.TrimSpace(scanner.Text())
}

func isBlankLineOrComment(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

func trimEnvValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)
	return value
}
