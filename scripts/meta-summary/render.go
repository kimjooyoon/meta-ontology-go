package main

import (
	"bytes"
	"fmt"
)

func renderSummary(report summaryReport) []byte {
	var output bytes.Buffer
	fmt.Fprintf(&output, "## Deterministic meta summary\n\n")
	fmt.Fprintf(&output, "- Schema: `%s`\n", report.SchemaVersion)
	fmt.Fprintf(&output, "- Decision: `%s` (`%s`)\n", report.Decision, report.Reason)
	fmt.Fprintf(&output, "- Input commitment: `sha256:%s`\n", report.InputSHA256)
	fmt.Fprintf(&output, "- Summary budget: `%d / %d` bytes\n\n", report.OutputBytes, report.LimitBytes)
	fmt.Fprintf(&output, "### Source metrics inventory (observed)\n\n")
	fmt.Fprintf(&output, "- regular files: `%d`\n", report.SourceMetrics.RegularFiles)
	fmt.Fprintf(&output, "- directories (including root): `%d`\n", report.SourceMetrics.DirectoriesIncludingRoot)
	fmt.Fprintf(&output, "- descendant directories: `%d`\n", report.SourceMetrics.DescendantDirectories)
	fmt.Fprintf(&output, "- Go files / physical lines: `%d / %d`\n", report.SourceMetrics.GoFiles, report.SourceMetrics.GoLines)
	fmt.Fprintf(&output, "- Gooo files / physical lines: `%d / %d`\n", report.SourceMetrics.GoooFiles, report.SourceMetrics.GoooLines)
	fmt.Fprintf(&output, "- root README exclusion: `%s` (`%s`), blocking=`%t`\n", report.SourceMetrics.RootReadme.Applicability, report.SourceMetrics.RootReadme.Reason, report.SourceMetrics.RootReadme.Blocking)
	fmt.Fprintf(&output, "- threshold: Go=`%d`, Gooo=`%d`, function=`%d`\n", report.SourceMetrics.Thresholds.GoFile, report.SourceMetrics.Thresholds.GoooFile, report.SourceMetrics.Thresholds.Function)
	fmt.Fprintf(&output, "- over-threshold candidates: Go=`%d`, Gooo=`%d`, function=`%d`, total=`%d`\n", sourceCandidateCount(report.SourceMetrics.Candidates, "gooo.metric.source.go-file-lines.v1"), sourceCandidateCount(report.SourceMetrics.Candidates, "gooo.metric.source.gooo-file-lines.v1"), sourceCandidateCount(report.SourceMetrics.Candidates, "gooo.metric.source.function-lines.v1"), len(report.SourceMetrics.Candidates))
	fmt.Fprintf(&output, "- selected operations: `%d`\n", report.SourceMetrics.SelectedOperations)
	for _, selected := range report.SourceMetrics.SelectedSubjects {
		fmt.Fprintf(&output, "  - `%s`: `%s`\n", selected.Operation, selected.Subject)
	}
	fmt.Fprintf(&output, "\n")
	fmt.Fprintf(&output, "### Munchhausen indicator routes\n\n")
	fmt.Fprintf(&output, "| Indicator | Route | Verdict | Observation | Bound |\n")
	fmt.Fprintf(&output, "|---|---|---|---:|---:|\n")
	for _, indicator := range report.Indicators {
		fmt.Fprintf(&output, "| `%s` | `%s` | `%s` | `%s` | `%s %s` |\n",
			indicator.ID, indicator.Route, indicator.Verdict,
			indicator.Value, indicator.Relation, indicator.Limit)
	}
	fmt.Fprintf(&output, "\n### Exact artifact commitments\n\n")
	fmt.Fprintf(&output, "| Artifact | Bytes | SHA-256 |\n|---|---:|---|\n")
	for _, artifact := range report.Artifacts {
		fmt.Fprintf(&output, "| `%s` | %d | `%s` |\n", artifact.ID, artifact.Bytes, artifact.SHA256)
	}
	fmt.Fprintf(&output, "\n### Bound provenance\n\n")
	fmt.Fprintf(&output, "- Schema: `%s`\n", report.ProvenanceSchema)
	fmt.Fprintf(&output, "- Decision: `%s` (`%s`)\n", report.Provenance.Decision, report.Provenance.Reason)
	fmt.Fprintf(&output, "- Indicator ledger: `%s` (%d decisions)\n",
		report.Provenance.LedgerDigest, report.Provenance.LedgerCount)
	fmt.Fprintf(&output, "- Bound indicators: `%d`\n", report.Provenance.Pass)
	fmt.Fprintf(&output, "- Envelope: `%s`\n", report.Provenance.Envelope)
	fmt.Fprintf(&output, "- Replay: `%s`\n\n", report.Provenance.Replay)
	fmt.Fprintf(&output, "Full JSON remains available in exact-head GitHub Actions artifacts.\n")
	return output.Bytes()
}

func sourceCandidateCount(candidates []sourceCandidate, metricID string) int {
	count := 0
	for _, candidate := range candidates {
		if candidate.MetricID == metricID {
			count++
		}
	}
	return count
}
