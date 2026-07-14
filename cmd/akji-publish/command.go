package main

import (
	"fmt"
	"os"

	"github.com/schrek1/akji-rc/internal/publish"
)

const defaultFramePath = "/tmp/frame.jpg"

func run(args []string) error {
	framePath, err := framePathFromArgs(args)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Uploading frame to Uguu...")
	fmt.Fprintln(os.Stderr, "Note: temporary public hosting may be blocked by a corporate proxy or ISP.")
	result, err := publish.PublishFile(publish.DefaultConfiguration(), framePath)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Upload OK: %s\n", result.URL)
	fmt.Print(result.URL)
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
