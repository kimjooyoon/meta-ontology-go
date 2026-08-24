package main

import (
	"context"
	"fmt"
	"os"

	execution "github.com/kimjooyoon/meta-ontology-go/internal/meta/externalecosystemexecution"
)

func run(args []string) int {
	o, err := parseOptions(args)
	if err != nil { fmt.Fprintln(os.Stderr, err); return 2 }
	var observation execution.Observation
	if o.mode == "observe" {
		observation, err = execution.Observe(context.Background(), o.sourceRoot, o.externalRoot)
	} else {
		err = readJSON(o.observation, &observation)
	}
	if err != nil { fmt.Fprintln(os.Stderr, err); return 2 }
	suite := execution.RunSuite()
	observation.Regression = execution.RegressionReceipt{Passed:suite.Passed, Total:suite.Total, Unresolved:suite.Unresolved}
	report := execution.Evaluate(&observation)
	if o.mode == "observe" { err = writeJSON(o.observation, observation) }
	if err == nil { err = writeJSON(o.report, report) }
	if err == nil { err = writeJSON(o.suite, suite) }
	if err != nil { fmt.Fprintln(os.Stderr, err); return 2 }
	if report.Decision != execution.DecisionConfirmed || report.Resolution != "EXACT" || suite.Passed != suite.Total { return 1 }
	return 0
}
