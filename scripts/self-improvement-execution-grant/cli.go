package main

import (
	"errors"
	"fmt"
	"os"

	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
)

func run(settings options) error {
	if settings.outputPath == "" && settings.mode != "canonical-fixture" {
		return errors.New("-output is required")
	}
	program, err := grant.CompilePolicy(os.DirFS("."), settings.contractPath)
	if err != nil {
		return err
	}
	switch settings.mode {
	case "live":
		return runLive(program, settings)
	case "cases":
		return runCases(program, settings)
	case "verify":
		return runVerification(program, settings)
	case "canonical-fixture":
		return runCanonicalFixture(program, settings)
	default:
		return fmt.Errorf("unknown mode %q", settings.mode)
	}
}
