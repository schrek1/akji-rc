package main

import (
	"path/filepath"

	"github.com/schrek1/akji-rc/internal/capture"
	"github.com/schrek1/akji-rc/internal/config"
)

func loadCaptureConfiguration(workDir string) (capture.Configuration, error) {
	properties, err := config.LoadEnvironmentProperties(filepath.Join(workDir, "app", ".env"))
	if err != nil {
		return capture.Configuration{}, err
	}

	return capture.NewConfiguration(properties)
}
