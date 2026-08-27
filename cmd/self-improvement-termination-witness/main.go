package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	termination "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtermination"
)

func main() {
	root := flag.String("root", ".", "repository root")
	sourcePath := flag.String("source", termination.SourcePath, "executable Gooo source")
	repository := flag.String("repository", "kimjooyoon/meta-ontology-go", "repository identity")
	caseID := flag.String("case", "", "source-defined termination case")
	outputPath := flag.String("output", "", "termination receipt path; stdout when empty")
	check := flag.Bool("check", false, "require a bound receipt")
	flag.Parse()
	if *caseID == "" {
		fail(fmt.Errorf("case is required"))
	}
	input, err := termination.BuildInput(*root, *repository, *sourcePath, *caseID)
	if err != nil {
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
	fmt.Fprintf(os.Stderr, "termination-witness: case=%s decision=%s resolution=%s claim=%s conformance=%d/%d\n",
		receipt.Source.CaseID, receipt.Decision, receipt.Resolution, receipt.Outcome.ClaimState,
		receipt.Conformance.Satisfied, receipt.Conformance.Total)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "termination-witness:", err)
	os.Exit(2)
}
