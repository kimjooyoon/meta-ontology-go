package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := parseOptions()
	if err != nil {
		return err
	}
	artifact, err := buildArtifact(opts)
	if err != nil {
		return err
	}
	if err := validateArtifact(artifact); err != nil {
		return err
	}
	raw, err := marshalArtifact(artifact)
	if err != nil {
		return err
	}
	if opts.check != "" {
		return checkArtifact(opts.check, raw)
	}
	return writeArtifact(opts.output, raw)
}
