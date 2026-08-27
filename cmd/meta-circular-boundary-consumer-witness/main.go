package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundaryconsumer"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

func main() {
	sourcePath := flag.String("source", "examples/meta-circular-boundary/main.gooo", "Gooo source to re-read")
	reportPath := flag.String("report", "", "producer report to judge")
	headSHA := flag.String("head-sha", "", "exact 40-character commit SHA")
	flag.Parse()
	if *reportPath == "" {
		fatal(fmt.Errorf("--report is required"))
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatal(err)
	}
	reportBytes, err := os.ReadFile(*reportPath)
	if err != nil {
		fatal(err)
	}
	var report contract.Report
	decoder := json.NewDecoder(bytes.NewReader(reportBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		fatal(err)
	}
	input := contract.Input{Path: *sourcePath, HeadSHA: *headSHA, Source: source}
	if err := metacircularboundaryconsumer.Judge(report, input); err != nil {
		fatal(err)
	}
	fmt.Printf("consumer judge: PASS cases=%d indicators=%d\n", report.Summary.CasesPassed, len(report.Indicators))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
