package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/decider"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	if options.receipt == "" || options.context == "" || options.output == "" {
		return fmt.Errorf("receipt, context, and output are required")
	}
	var source []byte
	if options.source != "" {
		var err error
		source, err = os.ReadFile(options.source)
		if err != nil {
			return err
		}
	}
	receipt, err := os.ReadFile(options.receipt)
	if err != nil {
		return err
	}
	context, err := os.ReadFile(options.context)
	if err != nil {
		return err
	}
	verdict := decider.Decide(source, receipt, context)
	if err := writeVerdict(options.output, verdict); err != nil {
		return err
	}
	fmt.Printf("independent freshness decider: %s %s at %s/%s\n", verdict.State, verdict.Reason, verdict.Coordinate.Stage, verdict.Coordinate.Step)
	return nil
}

func writeVerdict(path string, verdict model.Verdict) error {
	raw, err := model.Marshal(verdict)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
