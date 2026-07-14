package main

import (
	"fmt"
	"os"
	"time"

	"github.com/schrek1/akji-rc/internal/capture"
)

func run(args []string) error {
	options, err := parseScriptOptions(args, os.Stdout)
	if err != nil {
		return err
	}
	if options.helpRequested {
		return nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	configuration, err := loadCaptureConfiguration(workDir)
	if err != nil {
		return err
	}

	if options.isTimeLapseEnabled() {
		return runLoop(configuration, time.Duration(options.timeLapseSeconds)*time.Second, workDir)
	}

	return runSingleCapture(configuration, options.outputPath, workDir)
}

func runSingleCapture(configuration capture.Configuration, outputPath string, workDir string) error {
	if outputPath == "" {
		outputPath = defaultOutputPath(workDir, time.Now())
	}

	captureResult, err := capture.CaptureToFile(configuration, outputPath)
	if err != nil {
		return err
	}

	printCaptureResult(captureResult, "Single capture successful.")
	return nil
}

func runLoop(configuration capture.Configuration, interval time.Duration, workDir string) error {
	fmt.Printf("Time-lapse enabled (Interval: %.0fs). Press Ctrl+C to stop.\n", interval.Seconds())
	nextTick := time.Now()
	for {
		outputPath := defaultOutputPath(workDir, time.Now())
		captureResult, err := capture.CaptureToFile(configuration, outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s - ERROR: Capture cycle failed: %v\n", time.Now().Format(dateLayout), err)
		} else {
			printCaptureResult(captureResult, "Capture successful.")
		}

		nextTick = nextTick.Add(interval)
		sleepDuration := time.Until(nextTick)
		if sleepDuration < 0 {
			nextTick = time.Now()
			sleepDuration = 0
		}
		fmt.Printf("Waiting %.0fs until next capture...\n", sleepDuration.Seconds())
		time.Sleep(sleepDuration)
	}
}

func printCaptureResult(captureResult capture.CaptureResult, successMessage string) {
	fmt.Printf("Downloaded %d bytes from MJPEG stream.\n", captureResult.DownloadedBytes)
	fmt.Printf("Saved: %s\n", captureResult.OutputPath)
	fmt.Println(successMessage)
}
