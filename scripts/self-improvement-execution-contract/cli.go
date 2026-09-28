package main

import (
	"errors"
	"fmt"
	"os"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
)

func run(settings options) error {
	if settings.outputPath == "" {
		return errors.New("-output is required")
	}
	program, err := contract.CompilePolicy(os.DirFS("."), settings.contractPath)
	if err != nil {
		return err
	}
	switch settings.mode {
	case "live":
		return runLive(program, settings)
	case "cases":
		return runCases(program, settings)
	case "verify":
		return runVerification(settings)
	default:
		return fmt.Errorf("unknown mode %q", settings.mode)
	}
}
