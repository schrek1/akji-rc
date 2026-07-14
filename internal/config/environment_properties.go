package config

import (
	"bufio"
	"os"
	"strings"
)

type EnvironmentProperties map[string]string

// LoadEnvironmentProperties loads file properties and applies process properties as overrides.
func LoadEnvironmentProperties(filePath string, processEnvironment EnvironmentProperties) (EnvironmentProperties, error) {
	fileProperties, err := readEnvironmentPropertiesFile(filePath)
	if err != nil {
		return nil, err
	}

	applyOverrides(fileProperties, processEnvironment)
	return fileProperties, nil
}

func ReadProcessEnvironmentProperties() EnvironmentProperties {
	processProperties := EnvironmentProperties{}
	for _, environmentEntry := range os.Environ() {
		key, value, found := strings.Cut(environmentEntry, "=")
		if found {
			processProperties[key] = value
		}
	}
	return processProperties
}

func readEnvironmentPropertiesFile(filePath string) (EnvironmentProperties, error) {
	environmentFile, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return EnvironmentProperties{}, nil
		}
		return nil, err
	}
	defer environmentFile.Close()

	fileProperties := EnvironmentProperties{}
	scanner := bufio.NewScanner(environmentFile)
	readEnvironmentProperties(scanner, fileProperties)

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return fileProperties, nil
}

func readEnvironmentProperties(scanner *bufio.Scanner, properties EnvironmentProperties) {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if isBlankLineOrComment(line) {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		properties[strings.TrimSpace(key)] = trimEnvironmentPropertyValue(value)
	}
}

func isBlankLineOrComment(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

func trimEnvironmentPropertyValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	return strings.Trim(value, `'`)
}

func applyOverrides(properties EnvironmentProperties, overrides EnvironmentProperties) {
	for key, value := range overrides {
		properties[key] = value
	}
}
