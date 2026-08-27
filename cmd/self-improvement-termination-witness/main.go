package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	termination "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtermination"
)

func main() {
	inputPath := flag.String("input", "", "self-improvement termination input JSON")
	outputPath := flag.String("output", "", "termination receipt path; stdout when empty")
	check := flag.Bool("check", false, "require a bound receipt")
	flag.Parse()
	if *inputPath == "" {
		fail(fmt.Errorf("input is required"))
	}
	data, err := os.ReadFile(*inputPath)
	if err != nil {
		fail(err)
	}
	var input termination.Input
	if err := json.Unmarshal(data, &input); err != nil {
		fail(err)
	}
	receipt, err := termination.Evaluate(input)
	if err != nil {
		fail(err)
	}
	if *check {
		if err := termination.ValidateReceipt(receipt, input); err != nil {
			fail(err)
		}
	}
	payload, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		fail(err)
	}
	payload = append(payload, '\n')
	if *outputPath == "" {
		_, err = os.Stdout.Write(payload)
	} else {
		err = os.WriteFile(*outputPath, payload, 0o644)
	}
	if err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "termination-witness: decision=%s reason=%s evidence=%d/%d\n",
		receipt.Decision, receipt.Reason, receipt.Summary.Satisfied, receipt.Summary.Total)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "termination-witness:", err)
	os.Exit(2)
}
