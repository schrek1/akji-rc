package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type scriptOptions struct {
	outputPath       string
	timeLapseSeconds int
	helpRequested    bool
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
	if errors.Is(flagError, flag.ErrHelp) {
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

func (options scriptOptions) isTimeLapseEnabled() bool {
	return options.timeLapseSeconds > 0
}
