package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtermination/verify"
)

func main() {
	root := flag.String("root", ".", "repository root")
	sourcePath := flag.String("source", "examples/self-improvement-termination/main.gooo", "executable Gooo source")
	repository := flag.String("repository", "kimjooyoon/meta-ontology-go", "repository identity")
	caseID := flag.String("case", "", "source-defined termination case")
	receiptPath := flag.String("receipt", "", "termination receipt JSON")
	outputPath := flag.String("output", "", "judge report path; stdout when empty")
	flag.Parse()
	if *caseID == "" || *receiptPath == "" {
		fail(fmt.Errorf("case and receipt are required"))
	}
	receipt, err := os.ReadFile(*receiptPath)
	if err != nil {
		fail(err)
	}
	report, err := verify.Verify(*root, *sourcePath, *repository, *caseID, receipt)
	if err != nil {
		fail(err)
	}
	payload, err := json.MarshalIndent(report, "", "  ")
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
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "termination-judge:", err)
	os.Exit(1)
}
