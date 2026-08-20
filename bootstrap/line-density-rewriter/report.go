package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func densityReport(sha string, subjects []rewriteSubject) rewriteReport {
	applied, blocked := 0, 0
	for _, subject := range subjects {
		if subject.Status == "applied" {
			applied++
		} else {
			blocked++
		}
	}
	proof := "axiomatic-foundation"
	return rewriteReport{
		Schema: "gooo.line-density-rewrite.v1", SourceSHA: sha, Subjects: subjects,
		Indicators: []rewriteIndicator{
			{ID: "density.applied", Value: applied, Limit: -1,
				Consumer: "logical-materializer", Operation: "accept-density-rewrite", Proof: proof},
			{ID: "density.blocked", Value: blocked, Limit: 0, Blocking: true,
				Consumer: "line-density-rewriter", Operation: "classify-density-blocker", Proof: proof},
		},
	}
}

func writeDensityReport(name string, report rewriteReport) error {
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, append(encoded, '\n'), 0o644)
}

func requireDensityClosure(report rewriteReport) error {
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}
