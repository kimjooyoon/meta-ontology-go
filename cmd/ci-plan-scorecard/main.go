package main

import (
	"fmt"
	"os"
)

type options struct {
	contract   string
	source     string
	generatedA string
	generatedB string
	reports    string
	replays    string
	golden     string
	profile    string
	output     string
	check      bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
