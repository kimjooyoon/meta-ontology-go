package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/promotioncontinuity"
)

func run(arguments []string, output io.Writer) error {
	flags := flag.NewFlagSet("promotion-authorized-continuity", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	head := flags.String("head", "", "expected merged head SHA")
	guard := flags.String("guard", "", "guarded promotion report")
	recovery := flags.String("recovery", "", "rollback fixed-point report")
	out := flags.String("out", "", "output receipt")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *head == "" || *guard == "" || *recovery == "" || *out == "" {
		return errors.New("head, guard, recovery, and out are required")
	}
	report, err := promotioncontinuity.Build(promotioncontinuity.Input{
		ExpectedHeadSHA: *head, GuardPath: *guard, RecoveryPath: *recovery,
	})
	if err != nil {
		return err
	}
	if err := promotioncontinuity.Validate(report); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode receipt: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		return fmt.Errorf("write receipt: %w", err)
	}
	fmt.Fprintf(output, "decision=%s resolution=%s satisfied=%d/%d\n",
		report.Decision, report.Resolution, report.Summary.Satisfied, report.Summary.Total)
	if report.Decision != "PASS" {
		return fmt.Errorf("%s: %s", report.Decision, report.Reason)
	}
	return nil
}
