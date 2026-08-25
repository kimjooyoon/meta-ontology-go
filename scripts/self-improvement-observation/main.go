package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementobservation"
)

func main() {
	head := flag.String("head-sha", "", "exact language experiment commit")
	runID := flag.Int64("source-run-id", 0, "language experiment workflow run ID")
	report := flag.String("report", "", "language experiment report JSON")
	counterexamples := flag.String("counterexamples", "", "language counterexample summary JSON")
	contract := flag.String("contract", "", "compiled Gooo self-improvement contract JSON")
	output := flag.String("output", "", "observation JSON output path; stdout when empty")
	check := flag.Bool("check", false, "exit non-zero unless the input is exactly observed")
	flag.Parse()

	in, err := selfimprovementobservation.LoadInputs(*report, *counterexamples, *contract)
	if err != nil {
		exitError(err)
	}
	observation := selfimprovementobservation.Build(in, selfimprovementobservation.Options{HeadSHA: *head, SourceRunID: *runID})
	data, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		exitError(err)
	}
	data = append(data, '\n')
	if *output == "" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(*output, data, 0o644)
	}
	if err != nil {
		exitError(err)
	}
	if *check && observation.Decision != "OBSERVED" {
		os.Exit(1)
	}
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
