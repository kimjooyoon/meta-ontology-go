package main

import (
	"fmt"
	"io"
	"strings"
)

func writeCostMarkdown(output io.Writer, report costReport) error {
	var text strings.Builder
	text.WriteString("# Meta execution cost\n\n")
	text.WriteString("Diagnostic intervals, not semantic verification or execution authority.\n\n")
	fmt.Fprintf(&text, "Source authenticity: **%s**. Improvement: **%s**.\n\n",
		costMarkdownCell(report.Authenticity), costMarkdownCell(report.Improvement))
	text.WriteString("| Observation | Count |\n| --- | ---: |\n")
	fmt.Fprintf(&text, "| Input events | %d |\n| Measured intervals | %d |\n",
		report.Events, len(report.Rows))
	fmt.Fprintf(&text, "| Unmeasured events | %d |\n| Unpaired starts | %d |\n| Unknown returns | %d |\n\n",
		report.UnmeasuredEvents, report.UnpairedStarts, report.UnknownReturns)
	text.WriteString("Action intervals contain process intervals. Do not add these rows into a grand total.\n\n")
	text.WriteString("| Invocation | Activity | Pass | Kind | Start event | Return event | Elapsed ns |\n")
	text.WriteString("| --- | --- | --- | --- | ---: | ---: | ---: |\n")
	for _, row := range report.Rows {
		fmt.Fprintf(&text, "| %s | %s | %s | %s | %d | %d | %d |\n",
			costMarkdownCell(row.Invocation), costMarkdownCell(row.Activity),
			costMarkdownCell(row.Pass), costMarkdownCell(row.Kind), row.Start, row.Return, row.Elapsed)
	}
	text.WriteString("\nUse the JSON report for full source, contract, plan and manifest bindings.\n")
	text.WriteString("Zero measured intervals is not zero execution cost. Missing observations do not establish success.\n")
	_, err := io.WriteString(output, text.String())
	return err
}

func costMarkdownCell(value string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;",
		"\\", "&#92;", "|", "&#124;", "`", "&#96;", "[", "&#91;", "]", "&#93;",
		"*", "&#42;", "_", "&#95;", "\n", " ", "\r", " ").Replace(value)
}
