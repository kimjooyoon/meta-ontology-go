package main

import (
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func checkSplitPlans(cfg config, report metricevidence.Report, indicators []metricevidence.Indicator) error {
	planned, blocked, matched := 0, 0, 0
	var subjectError error
	for _, indicator := range indicators {
		if cfg.subject != "" && cfg.subject != indicator.Subject {
			continue
		}
		matched++
		plan, err := planSource(cfg.root, indicator.Subject, report.Meta.Policy.MaxFileLines)
		if err == nil {
			err = validateTopology(report, plan)
		}
		if err == nil {
			planned++
			continue
		}
		if errors.Is(err, errSplitBlocked) {
			blocked++
			if cfg.subject != "" {
				subjectError = fmt.Errorf("plan %s: %w", indicator.Subject, err)
			}
			continue
		}
		return fmt.Errorf("plan %s: %w", indicator.Subject, err)
	}
	if cfg.subject != "" && matched == 0 {
		return fmt.Errorf("subject %q is not an actionable split indicator", cfg.subject)
	}
	fmt.Printf("source-splitter: actionable=%d planned=%d blocked=%d write=false\n", matched, planned, blocked)
	return subjectError
}
