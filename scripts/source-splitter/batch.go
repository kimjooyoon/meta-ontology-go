package main

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func runBatch(cfg config, metrics metricevidence.Report, indicators []metricevidence.Indicator) error {
	logical, err := readLogicalSplitPlan(cfg.plan)
	if err != nil {
		return err
	}
	if logical.Schema != logicalSplitPlanSchema || logical.SourceSHA != cfg.sha {
		return fmt.Errorf("logical split plan identity is unknown")
	}
	selected := make([]logicalSplitSubject, 0)
	seen := map[string]bool{}
	for _, subject := range logical.Subjects {
		if seen[subject.Logical] || !metricevidence.Contains(indicators, subject.Logical) {
			return fmt.Errorf("logical subject %q has unknown metric routing", subject.Logical)
		}
		seen[subject.Logical] = true
		if subject.Reason != "projectable" {
			continue
		}
		if subject.Consumer != "logical-source-splitter" || subject.Operation != splitBatchOperation ||
			subject.Logical == "" {
			return fmt.Errorf("projectable subject %q has unknown routing", subject.Logical)
		}
		selected = append(selected, subject)
	}
	if len(seen) != len(indicators) {
		return fmt.Errorf("logical subjects=%d actionable indicators=%d", len(seen), len(indicators))
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Logical < selected[j].Logical })
	plans := make([]plannedSplit, 0, len(selected))
	for _, subject := range selected {
		plan, planErr := planSource(cfg.root, subject.Logical, metrics.Meta.Policy.MaxFileLines)
		if planErr == nil {
			planErr = validateTopology(metrics, plan)
		}
		if planErr != nil {
			return fmt.Errorf("plan %s: %w", subject.Logical, planErr)
		}
		plans = append(plans, plannedSplit{logical: subject.Logical, plan: plan})
	}
	if err := validateBatchTopology(metrics, plans); err != nil {
		return err
	}
	applied := make([]splitBatchSubject, 0, len(plans))
	for _, item := range plans {
		writeCfg := cfg
		writeCfg.subject = item.logical
		evidence, applyErr := applySplitEvidence(writeCfg, item.plan)
		if applyErr != nil {
			return fmt.Errorf("apply %s: %w", item.logical, applyErr)
		}
		subject, evidenceErr := summarizeSplitEvidence(cfg.sha, item.logical, evidence)
		if evidenceErr != nil {
			return evidenceErr
		}
		applied = append(applied, subject)
	}
	return writeSplitBatchReport(cfg.output, buildSplitBatchReport(cfg.sha, plans, applied))
}
