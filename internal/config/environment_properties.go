package config

import (
	"bufio"
	"os"
	"strings"
)

type EnvironmentProperties map[string]string

// LoadEnvironmentProperties merges file and OS properties, with OS properties overriding file properties.
func LoadEnvironmentProperties(envPropertiesFilePath string) (EnvironmentProperties, error) {
	envProperties, err := readFileEnvProperties(envPropertiesFilePath)
	if err != nil {
		return EnvironmentProperties{}, err
	}

	applyOverrides(envProperties, readOsEnvProperties())

	return envProperties, nil
}

func readFileEnvProperties(filePath string) (EnvironmentProperties, error) {
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

func readOsEnvProperties() EnvironmentProperties {
	processProperties := EnvironmentProperties{}
	for _, environmentEntry := range os.Environ() {
		key, value, found := strings.Cut(environmentEntry, "=")
		if found {
			processProperties[key] = value
		}
	}
	return processProperties
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
