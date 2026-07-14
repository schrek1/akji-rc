package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/schrek1/akji-rc/internal/validate"
)

const defaultFramePath = "/tmp/frame.jpg"

func run(args []string) error {
	framePath, err := framePathFromArgs(args)
	if err != nil {
		return err
	}
	minimumSizeBytes, err := minimumSizeBytes(os.Getenv("MIN_SIZE_BYTES"))
	if err != nil {
		return err
	}

	result, err := validate.ValidateFile(framePath, minimumSizeBytes)
	if err != nil {
		return err
	}

	fmt.Printf("Image size: %d bytes\n", result.Size)
	fmt.Println("Validation OK")
	return nil
}

func framePathFromArgs(args []string) (string, error) {
	if len(args) == 0 {
		return defaultFramePath, nil
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("expected at most one frame path")
}

func minimumSizeBytes(value string) (int64, error) {
	if value == "" {
		return validate.DefaultMinimumSizeBytes, nil
	}

	minimumSizeBytes, err := strconv.ParseInt(value, 10, 64)
	if err != nil || minimumSizeBytes <= 0 {
		return 0, fmt.Errorf("MIN_SIZE_BYTES must be a positive integer")
	}
	return minimumSizeBytes, nil
}
