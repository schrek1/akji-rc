package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type cliOptions struct {
	outputPath       string
	timeLapseSeconds int
	helpRequested    bool
}

func parseOptions(args []string, output io.Writer) (cliOptions, error) {
	flags := createFlagSet(output)
	options := registerOptions(flags)

	if err := flags.Parse(args); err != nil {
		return handleFlagError(err)
	}

	return validateOptions(*options)
}

func createFlagSet(output io.Writer) *flag.FlagSet {
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

func registerOptions(flags *flag.FlagSet) *cliOptions {
	options := &cliOptions{}
	flags.StringVar(&options.outputPath, "out", "", "Output file path")
	flags.StringVar(&options.outputPath, "o", "", "Output file path")
	flags.IntVar(&options.timeLapseSeconds, "timeLapse", 0, "Run in loop, capturing every N seconds")
	flags.IntVar(&options.timeLapseSeconds, "tl", 0, "Run in loop, capturing every N seconds")
	return options
}

func handleFlagError(flagError error) (cliOptions, error) {
	if errors.Is(flagError, flag.ErrHelp) {
		return cliOptions{helpRequested: true}, nil
	}
	return cliOptions{}, flagError
}

func validateOptions(options cliOptions) (cliOptions, error) {
	if options.timeLapseSeconds > 0 && options.outputPath != "" {
		return cliOptions{}, fmt.Errorf("--out is not compatible with --timeLapse")
	}
	return options, nil
}

func (options cliOptions) isTimeLapseEnabled() bool {
	return options.timeLapseSeconds > 0
}
