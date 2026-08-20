package main

import (
	"errors"
	"fmt"
)

func checkSplitPlans(cfg config, report metricReport, indicators []metricIndicator) error {
	planned, blocked, matched := 0, 0, 0
	for _, indicator := range indicators {
		if cfg.subject != "" && cfg.subject != indicator.Subject {
			continue
		}
		matched++
		plan, err := planSource(cfg.root, indicator.Subject, report.Meta.Policy.MaxFileLines)
		if err == nil {
			err = report.validateTopology(plan)
		}
		if err == nil {
			planned++
			continue
		}
		if errors.Is(err, errSplitBlocked) {
			blocked++
			continue
		}
		return fmt.Errorf("plan %s: %w", indicator.Subject, err)
	}
	if cfg.subject != "" && matched == 0 {
		return fmt.Errorf("subject %q is not an actionable split indicator", cfg.subject)
	}
	fmt.Printf("source-splitter: actionable=%d planned=%d blocked=%d write=false\n", matched, planned, blocked)
	return nil
}

func (report metricReport) validateTopology(plan splitPlan) error {
	for _, directory := range report.Directories {
		if directory.Path != plan.Directory {
			continue
		}
		added := len(plan.Parts) - 1
		if directory.DirectFolders != 0 {
			return fmt.Errorf("%w: %s contains folders", errSplitBlocked, plan.Directory)
		}
		if directory.DirectFiles+added > report.Meta.Policy.MaxDirectEntries {
			return fmt.Errorf("%w: %s projects %d entries", errSplitBlocked, plan.Directory, directory.DirectFiles+added)
		}
		return nil
	}
	return fmt.Errorf("metric evidence omits directory %q", plan.Directory)
}
