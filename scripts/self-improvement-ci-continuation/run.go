package main

import (
	"errors"
	"fmt"
	"os"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcontinuation"
)

func run(settings options) error {
	if settings.outputPath == "" {
		return errors.New("-output is required")
	}
	program, err := continuation.CompilePolicy(os.DirFS("."), settings.contractPath)
	if err != nil {
		return err
	}
	switch settings.mode {
	case "live":
		return runLive(program, settings)
	case "cases":
		return runCases(program, settings)
	case "verify":
		return runVerify(program, settings)
	default:
		return fmt.Errorf("unknown mode %q", settings.mode)
	}
}
