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
