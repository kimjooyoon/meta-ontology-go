package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	termination "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtermination"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtermination/verify"
)

func main() {
	inputPath := flag.String("input", "", "self-improvement termination input JSON")
	receiptPath := flag.String("receipt", "", "termination receipt JSON")
	outputPath := flag.String("output", "", "judge report path; stdout when empty")
	flag.Parse()
	if *inputPath == "" || *receiptPath == "" {
		fail(fmt.Errorf("input and receipt are required"))
	}
	var input termination.Input
	if err := decode(*inputPath, &input); err != nil {
		fail(err)
	}
	var receipt termination.Receipt
	if err := decode(*receiptPath, &receipt); err != nil {
		fail(err)
	}
	report, err := verify.Verify(input, receipt)
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

func decode(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "termination-judge:", err)
	os.Exit(1)
}
