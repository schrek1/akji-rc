package config

import (
	"bufio"
	"os"
	"strings"
)

type EnvironmentProperties map[string]string

func LoadEnvironmentProperties(filePaths []string, processEnvironment EnvironmentProperties) (EnvironmentProperties, error) {
	loadedProperties := EnvironmentProperties{}
	for _, filePath := range filePaths {
		fileProperties, err := readEnvironmentPropertiesFile(filePath)
		if err != nil {
			return nil, err
		}
		mergeEnvironmentProperties(loadedProperties, fileProperties)
	}

	mergeNonEmptyEnvironmentProperties(loadedProperties, processEnvironment)
	return loadedProperties, nil
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

func mergeEnvironmentProperties(target EnvironmentProperties, source EnvironmentProperties) {
	for key, value := range source {
		target[key] = value
	}
}

func mergeNonEmptyEnvironmentProperties(target EnvironmentProperties, source EnvironmentProperties) {
	for key, value := range source {
		if value != "" {
			target[key] = value
		}
	}
}
