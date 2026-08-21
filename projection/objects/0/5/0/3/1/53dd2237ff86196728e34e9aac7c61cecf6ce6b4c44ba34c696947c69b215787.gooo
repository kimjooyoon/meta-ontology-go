package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type options struct {
	root, metrics, sha, subject string
	check                       bool
}

type candidate struct {
	raw    string
	source sourcepolicy.SourceSubject
}

func run(opts options) error {
	if opts.root == "" || opts.metrics == "" || opts.sha == "" {
		return fmt.Errorf("root, metrics, and sha are required")
	}
	if !opts.check && opts.subject == "" {
		return fmt.Errorf("apply mode requires one exact subject")
	}
	payload, err := os.ReadFile(opts.metrics)
	if err != nil {
		return err
	}
	var report linecaps.LineMetricsReport
	if err := json.Unmarshal(payload, &report); err != nil {
		return err
	}
	if report.CommitSHA == "" || report.CommitSHA != opts.sha {
		return fmt.Errorf("metrics SHA %q does not match expected %q", report.CommitSHA, opts.sha)
	}
	candidates := make([]candidate, 0)
	for _, indicator := range report.Meta.Actionable() {
		if indicator.Operation != sourcepolicy.OperationCollapseAssign || opts.subject != "" && indicator.Subject != opts.subject {
			continue
		}
		source, err := sourcepolicy.ParseSourceSubject(indicator.Subject)
		if err != nil {
			return err
		}
		candidates = append(candidates, candidate{raw: indicator.Subject, source: source})
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].raw < candidates[j].raw })
	if opts.subject != "" && len(candidates) != 1 {
		return fmt.Errorf("subject %q resolved to %d candidates", opts.subject, len(candidates))
	}
	for _, candidate := range candidates {
		if err := transformCandidate(opts.root, candidate, !opts.check); err != nil {
			return err
		}
	}
	fmt.Printf("refactor-metrics: checked=%d operation=%s write=%t\n", len(candidates), sourcepolicy.OperationCollapseAssign, !opts.check)
	return nil
}
