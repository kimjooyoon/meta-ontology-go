package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

type options struct {
	ledger, observationReceipt, effectReceipt, output string
	root                                              string
}

func main() {
	var opts options
	flag.StringVar(&opts.ledger, "ledger", "", "observer-effect ledger to judge")
	flag.StringVar(&opts.observationReceipt, "observation-receipt", "", "observation receipt to judge")
	flag.StringVar(&opts.effectReceipt, "effect-receipt", "", "observer-effect receipt to judge")
	flag.StringVar(&opts.output, "output", "", "independent judgment output")
	flag.StringVar(&opts.root, "root", "", "repository root whose workflow topology is independently checked")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "observer-effect-judge:", err)
		os.Exit(1)
	}
}

func run(opts options) error {
	if opts.ledger == "" || opts.observationReceipt == "" || opts.effectReceipt == "" || opts.output == "" || opts.root == "" {
		return fmt.Errorf("root, ledger, both receipts, and output are required")
	}
	var report Report
	var observationReceipt, effectReceipt Receipt
	if err := readJSON(opts.ledger, &report); err != nil {
		return err
	}
	if err := readJSON(opts.observationReceipt, &observationReceipt); err != nil {
		return err
	}
	if err := readJSON(opts.effectReceipt, &effectReceipt); err != nil {
		return err
	}
	judgment, err := judge(opts.root, report, observationReceipt, effectReceipt)
	if err != nil {
		return err
	}
	if err := writeJSON(opts.output, judgment); err != nil {
		return err
	}
	fmt.Printf("observer-effect-judge: decision=%s subject=%s resolution=%s\n",
		judgment.Decision, judgment.SubjectDecision, judgment.Resolution)
	return nil
}

func readJSON(path string, target any) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	decoder := json.NewDecoder(bytesReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("decode %s: expected one JSON value", path)
	}
	return nil
}

func writeJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode judgment: %w", err)
	}
	payload = append(payload, '\n')
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
