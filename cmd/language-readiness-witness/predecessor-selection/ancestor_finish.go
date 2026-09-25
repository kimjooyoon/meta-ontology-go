package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func finishAncestry(cfg config, immediate string,
	selected predecessorselection.Result, attempts []predecessorresolution.Attempt,
) (predecessorselection.Result, predecessorresolution.Report, error) {
	input := predecessorresolution.Input{Repository: cfg.repository,
		CurrentHeadSHA: cfg.currentHead, ImmediatePredecessorSHA: immediate,
		SearchLimit: predecessorresolution.SearchLimit, Attempts: attempts}
	report, err := predecessorresolution.Build(input)
	if err != nil {
		return predecessorselection.Result{}, predecessorresolution.Report{}, err
	}
	replay, replayErr := predecessorresolution.Build(input)
	if replayErr != nil || replay.ReportDigest != report.ReportDigest {
		return predecessorselection.Result{}, predecessorresolution.Report{},
			fmt.Errorf("ancestor resolution replay mismatch")
	}
	if report.Decision != predecessorresolution.DecisionResolved {
		return predecessorselection.Result{}, report,
			fmt.Errorf("%s: %s", report.Decision, report.Reason)
	}
	return selected, report, nil
}
