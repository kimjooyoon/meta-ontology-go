package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func buildReport(sha string, subjects []planSubject, counterexamples []planCounterexample) planReport {
	counts := indicatorCounts(subjects)
	proof := "axiomatic-foundation"
	metrics := []planIndicator{
		{ID: "logical-split.static-density", Value: counts["static-density-rewrite"], Limit: -1,
			Consumer: "line-density-rewriter", Operation: "compact-static-literal", Proof: proof},
		{ID: "logical-split.large-density", Value: counts["large-density-rewrite"], Limit: -1,
			Consumer: "line-density-rewriter", Operation: "compact-large-expression", Proof: proof},
		{ID: "logical-split.density-rewrite", Value: counts["density-rewrite"], Limit: -1,
			Consumer: "line-density-rewriter", Operation: "compact-obvious-lines", Proof: proof},
		{ID: "logical-split.projectable", Value: counts["projectable"], Limit: -1,
			Consumer: "logical-source-splitter", Operation: "split-logical-declarations", Proof: proof},
		{ID: "logical-split.no-movable", Value: counts["no-movable-declaration"], Limit: -1,
			Consumer: "source-extractor", Operation: "extract-indivisible-source", Proof: proof},
		{ID: "logical-split.fixed-capacity", Value: counts["fixed-declaration-capacity"], Limit: -1,
			Consumer: "source-extractor", Operation: "extract-fixed-declaration", Proof: proof},
		{ID: "logical-split.movable-capacity", Value: counts["movable-declaration-capacity"], Limit: -1,
			Consumer: "source-extractor", Operation: "extract-movable-declaration", Proof: proof},
		{ID: "logical-split.declaration-capacity-contradiction", Value: counts["declaration-capacity-contradiction"], Limit: -1,
			Consumer: "source-extractor", Operation: "report-contradiction", Proof: proof},
		{ID: "logical-split.unclassified", Value: counts["unclassified"], Limit: 0,
			Blocking: true, Consumer: "logical-split-planner",
			Operation: "inspect-parse-domain", Proof: proof},
	}
	sortedCounterexamples := append([]planCounterexample(nil), counterexamples...)
	sort.SliceStable(sortedCounterexamples, func(i, j int) bool {
		return sortedCounterexamples[i].BlockerID < sortedCounterexamples[j].BlockerID
	})
	return planReport{Schema: "gooo.logical-split-plan.v1", SourceSHA: sha,
		Subjects: subjects, Counterexamples: sortedCounterexamples, Indicators: metrics}
}

func writeReport(name string, report planReport) error {
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, append(encoded, '\n'), 0o644)
}

func requireClassified(report planReport) error {
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}
