package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/schrek1/akji-rc/internal/capture"
	"github.com/schrek1/akji-rc/internal/config"
)

const dateLayout = "2006-01-02 15:04:05"

type scriptOptions struct {
	outputPath       string
	timeLapseSeconds int
	helpRequested    bool
}

var logger = log.New(os.Stdout, time.Now().Format(dateLayout)+" - ", 0)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s - ERROR: %v\n", time.Now().Format(dateLayout), err)
		os.Exit(1)
	}
}

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

	config, err := loadCaptureConfiguration(workDir)
	if err != nil {
		return err
	}

	if options.isTimeLapseEnabled() {
		return runLoop(config, time.Duration(options.timeLapseSeconds)*time.Second, workDir)
	}

	return runSingleCapture(config, options.outputPath, workDir)
}

func loadCaptureConfiguration(workDir string) (capture.Configuration, error) {
	properties, err := config.LoadEnvironmentProperties(
		filepath.Join(workDir, "app", ".env"),
		config.ReadProcessEnvironmentProperties(),
	)
	if err != nil {
		return capture.Configuration{}, err
	}
	return capture.NewConfiguration(properties)
}

func runSingleCapture(config capture.Configuration, outputPath string, workDir string) error {
	if outputPath == "" {
		outputPath = capture.DefaultOutputPath(workDir, time.Now())
	}

	captureResult, err := capture.CaptureToFile(config, outputPath)
	if err != nil {
		return err
	}

	logger.Printf("Downloaded %d bytes from MJPEG stream.", captureResult.DownloadedBytes)
	logger.Printf("Saved: %s", captureResult.OutputPath)
	logger.Println("Single capture successful.")
	return nil
}

func parseScriptOptions(args []string, output io.Writer) (scriptOptions, error) {
	flags := createScriptFlagSet(output)
	options := registerScriptOptions(flags)

	if err := flags.Parse(args); err != nil {
		return handleScriptFlagError(err)
	}

	return validateScriptOptions(*options)
}

func createScriptFlagSet(output io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet("akji-capture", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: akji-capture [OPTIONS]")
		fmt.Fprintln(flags.Output())
		fmt.Fprintln(flags.Output(), "KISS MJPEG webcam capture command. Extracts one still JPEG frame from a stream.")
		fmt.Fprintln(flags.Output())
		fmt.Fprintln(flags.Output(), "Options:")
		flags.PrintDefaults()
	}
	return flags
}

func registerScriptOptions(flags *flag.FlagSet) *scriptOptions {
	options := &scriptOptions{}
	flags.StringVar(&options.outputPath, "out", "", "Output file path")
	flags.StringVar(&options.outputPath, "o", "", "Output file path")
	flags.IntVar(&options.timeLapseSeconds, "timeLapse", 0, "Run in loop, capturing every N seconds")
	flags.IntVar(&options.timeLapseSeconds, "tl", 0, "Run in loop, capturing every N seconds")
	return options
}

func handleScriptFlagError(flagError error) (scriptOptions, error) {
	if flagError == flag.ErrHelp {
		return scriptOptions{helpRequested: true}, nil
	}
	return scriptOptions{}, flagError
}

func validateScriptOptions(options scriptOptions) (scriptOptions, error) {
	if options.timeLapseSeconds > 0 && options.outputPath != "" {
		return scriptOptions{}, fmt.Errorf("--out is not compatible with --timeLapse")
	}
	return options, nil
}

func runLoop(config capture.Configuration, interval time.Duration, workDir string) error {
	logger.Printf("Time-lapse enabled (Interval: %.0fs). Press Ctrl+C to stop.", interval.Seconds())
	nextTick := time.Now()
	for {
		outputPath := capture.DefaultOutputPath(workDir, time.Now())
		captureResult, err := capture.CaptureToFile(config, outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s - ERROR: Capture cycle failed: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		} else {
			logger.Printf("Downloaded %d bytes from MJPEG stream.", captureResult.DownloadedBytes)
			logger.Printf("Saved: %s", captureResult.OutputPath)
			logger.Println("Capture successful.")
		}

		nextTick = nextTick.Add(interval)
		sleepDuration := time.Until(nextTick)
		if sleepDuration < 0 {
			nextTick = time.Now()
			sleepDuration = 0
		}
		logger.Printf("Waiting %.0fs until next capture...", sleepDuration.Seconds())
		time.Sleep(sleepDuration)
	}
}

func (options scriptOptions) isTimeLapseEnabled() bool {
	return options.timeLapseSeconds > 0
}
