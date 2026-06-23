package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/schrek1/akji-rc/internal/capture"
)

func main() {
	logger := log.New(os.Stdout, time.Now().Format("2006-01-02 15:04:05")+" - ", 0)
	if err := run(os.Args[1:], logger); err != nil {
		fmt.Fprintf(os.Stderr, "%s - ERROR: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		os.Exit(1)
	}
}

func run(args []string, logger *log.Logger) error {
	flags := flag.NewFlagSet("akji-capture", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)

	var outputPath string
	var timeLapse int
	flags.StringVar(&outputPath, "out", "", "Output file path")
	flags.StringVar(&outputPath, "o", "", "Output file path")
	flags.IntVar(&timeLapse, "timeLapse", 0, "Run in loop, capturing every N seconds")
	flags.IntVar(&timeLapse, "tl", 0, "Run in loop, capturing every N seconds")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: akji-capture [OPTIONS]")
		fmt.Fprintln(flags.Output())
		fmt.Fprintln(flags.Output(), "KISS MJPEG webcam capture command. Extracts one still JPEG frame from a stream.")
		fmt.Fprintln(flags.Output())
		fmt.Fprintln(flags.Output(), "Options:")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if timeLapse > 0 && outputPath != "" {
		return fmt.Errorf("--out is not compatible with --timeLapse")
	}

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	config, err := capture.LoadConfig(capture.Environment(), workDir)
	if err != nil {
		return err
	}

	if timeLapse > 0 {
		return runLoop(config, time.Duration(timeLapse)*time.Second, workDir, logger)
	}

	if outputPath == "" {
		outputPath = capture.DefaultOutputPath(workDir, time.Now())
	}
	if err := capture.CaptureToFile(config, outputPath, logger); err != nil {
		return err
	}
	logger.Println("Single capture successful.")
	return nil
}

func runLoop(config capture.Config, interval time.Duration, workDir string, logger *log.Logger) error {
	logger.Printf("Time-lapse enabled (Interval: %.0fs). Press Ctrl+C to stop.", interval.Seconds())
	nextTick := time.Now()
	for {
		outputPath := capture.DefaultOutputPath(workDir, time.Now())
		if err := capture.CaptureToFile(config, outputPath, logger); err != nil {
			fmt.Fprintf(os.Stderr, "%s - ERROR: Capture cycle failed: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		} else {
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
