package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/proofchoicealgebra"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/proofchoicejudge"
)

func main() {
	mode := flag.String("mode", "produce", "produce or judge")
	source := flag.String("source", "", ".gooo source path")
	receipt := flag.String("receipt", "", "receipt path for independent judging")
	output := flag.String("output", "", "output JSON path")
	expect := flag.String("expect", "", "expected decision")
	flag.Parse()
	if err := run(*mode, *source, *receipt, *output, *expect); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(mode, source, receipt, output, expect string) error {
	if output == "" {
		return fmt.Errorf("output is required")
	}
	var value any
	switch mode {
	case "produce":
		if source == "" {
			return fmt.Errorf("source is required")
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		value = proofchoicealgebra.Evaluate(source, data)
	case "judge":
		if receipt == "" {
			return fmt.Errorf("receipt is required")
		}
		data, err := os.ReadFile(receipt)
		if err != nil {
			return err
		}
		value = proofchoicejudge.Judge(data)
	default:
		return fmt.Errorf("unknown mode %q", mode)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, append(data, '\n'), 0o644); err != nil {
		return err
	}
	decision := ""
	switch report := value.(type) {
	case proofchoicealgebra.Receipt:
		decision = report.Decision
	case proofchoicejudge.Verdict:
		decision = report.Decision
	}
	if expect != "" && decision != expect {
		return fmt.Errorf("decision = %s, want %s", decision, expect)
	}
	fmt.Printf("proof-choice algebra: mode=%s decision=%s\n", mode, decision)
	return nil
}
