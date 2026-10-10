package main

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/outcomedelta"
)

func writeOutcomeMarkdown(out io.Writer, r *outcomedelta.OutcomeDelta) error {
	var text strings.Builder
	fmt.Fprintf(&text, "# Workflow outcome comparison\n\nObserved activity/input groups: **%d**.\n\n", len(r.Rows))
	writeOutcomeMarkdownMetrics(&text, r)
	fmt.Fprintln(&text, "\nShares use the eligible populations named in each row. They measure these saved observations;")
	fmt.Fprintln(&text, "model accuracy, whole-program coverage and customer savings remain unmeasured.")
	fmt.Fprint(&text, "\n## Observed inputs\n\n")
	if len(r.Rows) == 0 {
		fmt.Fprintln(&text, "No activity/input groups were observed.")
	} else {
		fmt.Fprintln(&text, "| Activity | Caller inputs | Assessment | Actual before → after | Expected before → after |")
		fmt.Fprintln(&text, "| --- | --- | --- | --- | --- |")
		for _, row := range r.Rows {
			fmt.Fprintf(&text, "| %s | %s | %s | %s → %s | %s → %s |\n",
				outcomeMarkdownCell(row.ActivityID), outcomeMarkdownCell(outcomeJSONSummary(row.RootInputs)),
				outcomeMarkdownCell(row.Assessment), outcomeMarkdownCell(outcomeValue(row.Before, "actual")),
				outcomeMarkdownCell(outcomeValue(row.After, "actual")), outcomeMarkdownCell(outcomeValue(row.Before, "expected")),
				outcomeMarkdownCell(outcomeValue(row.After, "expected")))
		}
	}
	fmt.Fprint(&text, "\n## Evidence scope\n\n")
	fmt.Fprintf(&text, "- Recorded runtime stage: %s → %s\n",
		outcomeMarkdownCell(outcomeJSONSummary(r.BeforeRuntime["stage"])), outcomeMarkdownCell(outcomeJSONSummary(r.AfterRuntime["stage"])))
	fmt.Fprintf(&text, "- Before input SHA256: %s\n- After input SHA256: %s\n",
		outcomeMarkdownCell(r.BeforeInputSha256), outcomeMarkdownCell(r.AfterInputSha256))
	fmt.Fprintln(&text, "- No model calls or program executions were made by this comparison.")
	fmt.Fprintln(&text, "- Exit 0 means the report was produced. Missing and conflicting observations remain explicit.")
	fmt.Fprintln(&text, "- Long values are abbreviated at 120 characters. Use --json for full values, faults, case indices and source identities.")
	_, err := io.WriteString(out, text.String())
	return err
}

func writeOutcomeMarkdownMetrics(out io.Writer, r *outcomedelta.OutcomeDelta) {
	c := r.Counts
	previousMatch := c["assessment:REGRESSION"] + c["assessment:STILL_MATCHING"]
	previousFail := c["assessment:IMPROVEMENT"] + c["assessment:STILL_FAILING"]
	fmt.Fprintln(out, "| Observation | Observed / eligible groups |")
	fmt.Fprintln(out, "| --- | ---: |")
	for _, metric := range []struct {
		label           string
		observed, total int
	}{
		{"Unchanged expectations with observed outcomes / all groups", previousMatch + previousFail, len(r.Rows)},
		{"Regressions / previously matching groups under unchanged expectations", c["assessment:REGRESSION"], previousMatch},
		{"Improvements / previously failing groups under unchanged expectations", c["assessment:IMPROVEMENT"], previousFail},
		{"Changed requirements / all groups", c["assessment:REQUIREMENT_CHANGED"], len(r.Rows)},
		{"Unobserved assessment / all groups", c["assessment:UNOBSERVED"], len(r.Rows)},
		{"Conflicting duplicate observations / all groups", c["assessment:AMBIGUOUS"], len(r.Rows)},
	} {
		share := "n/a (no eligible groups)"
		if metric.total > 0 {
			share = fmt.Sprintf("%d/%d (%.2f%%)", metric.observed, metric.total, 100*float64(metric.observed)/float64(metric.total))
		}
		fmt.Fprintf(out, "| %s | %s |\n", metric.label, share)
	}
}

func outcomeMarkdownCell(value string) string {
	escaped := html.EscapeString(value)
	escaped = strings.NewReplacer("|", "&#124;", "\r", "&#13;", "\n", "&#10;", "`", "&#96;",
		"*", "&#42;", "_", "&#95;", "[", "&#91;", "]", "&#93;", "!", "&#33;", "\\", "&#92;", ":", "&#58;").Replace(escaped)
	return "<code>" + escaped + "</code>"
}
