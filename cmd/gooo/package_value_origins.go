package main

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func formatPackageValueOrigins(out *strings.Builder, flow workspaceexecution.ProgramValueOrigins) {
	fmt.Fprintf(out, "\nSelected program value origins: %s [%s]\n", flow.RootActivityID, flow.Analysis)
	for _, row := range flow.Rows {
		fmt.Fprintf(out, "\n  %s -> %s in %s", row.RootKind, row.Kind, row.ActivityID)
		if row.CandidateID != "" {
			fmt.Fprintf(out, " / candidate %s / hole %s", row.CandidateID, row.HoleID)
		}
		fmt.Fprintln(out)
		for _, step := range row.Path {
			fmt.Fprintf(out, "    %s %s %s: %q\n", step.Via, step.ActivityID, step.Kind, step.Expression)
		}
		for _, source := range row.Locations {
			fmt.Fprintf(out, "    source %s: %s bytes %d..%d (%s)\n", source.View, source.Filename, source.Start, source.End, source.SourceSHA256)
		}
	}
	for _, issue := range flow.Issues {
		fmt.Fprintf(out, "  Partial %s: %s\n", issue.ActivityID, issue.Reason)
	}
	if flow.Truncated {
		fmt.Fprintln(out, "  Value paths truncated; narrow the root or candidate, or raise --limit.")
	}
	fmt.Fprintln(out, "  "+flow.Scope)
}
