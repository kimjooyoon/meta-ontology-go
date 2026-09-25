package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/directorykind"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	sourcePath := flag.String("source-metrics", "", "exact source metrics JSON")
	mode := flag.String("mode", "check", "read-only execution mode")
	flag.Parse()
	if *mode != "check" {
		return fmt.Errorf("unsupported mode %q", *mode)
	}
	if *sourcePath == "" {
		return fmt.Errorf("-source-metrics is required")
	}
	payload, err := os.ReadFile(*sourcePath)
	if err != nil {
		return err
	}
	source, err := directorykind.DecodeSource(payload)
	if err != nil {
		return err
	}
	report, err := directorykind.Build(source)
	if err != nil {
		return err
	}
	data, err := report.JSON()
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}
