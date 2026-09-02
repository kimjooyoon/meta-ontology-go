package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/integrationprogress"
)

func run(args []string) error {
	settings, err := parse(args)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(settings.input)
	if err != nil {
		return err
	}
	observation, err := integrationprogress.DecodeObservation(raw)
	if err != nil {
		return err
	}
	programA, programB := integrationprogress.RenderProgram(), integrationprogress.RenderProgram()
	if !bytes.Equal(programA, programB) {
		return fmt.Errorf("generated Gooo replay differs")
	}
	reportA := integrationprogress.Evaluate(observation, true)
	reportB := integrationprogress.Evaluate(observation, true)
	if err := integrationprogress.Validate(reportA); err != nil {
		return err
	}
	if err := integrationprogress.Validate(reportB); err != nil {
		return err
	}
	reportRawA, err := json.MarshalIndent(reportA, "", "  ")
	if err != nil {
		return err
	}
	reportRawB, err := json.MarshalIndent(reportB, "", "  ")
	if err != nil {
		return err
	}
	reportRawA, reportRawB = append(reportRawA, '\n'), append(reportRawB, '\n')
	if !bytes.Equal(reportRawA, reportRawB) {
		return fmt.Errorf("report replay differs")
	}
	if settings.check {
		return compareOutputs(settings, reportRawA, programA)
	}
	if err := os.WriteFile(settings.report, reportRawA, 0o644); err != nil {
		return err
	}
	return os.WriteFile(settings.program, programA, 0o644)
}

func compareOutputs(settings config, report, program []byte) error {
	existingReport, err := os.ReadFile(settings.report)
	if err != nil {
		return err
	}
	existingProgram, err := os.ReadFile(settings.program)
	if err != nil {
		return err
	}
	if !bytes.Equal(existingReport, report) || !bytes.Equal(existingProgram, program) {
		return fmt.Errorf("existing integration progress outputs differ from replay")
	}
	return nil
}
