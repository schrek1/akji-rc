package main

import (
	"fmt"
	"os"
	"time"
)

const dateLayout = "2006-01-02 15:04:05"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s - ERROR: %v\n", time.Now().Format(dateLayout), err)
		os.Exit(1)
	}
}
