package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceimpact"
)

const sourceImpactUsage = "usage: gooo impact <before.gooo> <after.gooo> [--format text|json]"

func runSourceImpact(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	format := "text"
	if len(args) == 4 && args[2] == "--format" {
		format = args[3]
	} else if len(args) != 2 {
		fmt.Fprintln(stderr, sourceImpactUsage)
		return exitUsage
	}
	if args[0] == "" || args[1] == "" || strings.HasPrefix(args[0], "-") || strings.HasPrefix(args[1], "-") || (format != "text" && format != "json") {
		fmt.Fprintln(stderr, sourceImpactUsage)
		return exitUsage
	}
	deadline := time.Now().Add(commandDeadline)
	before, err := readSourceWithDeadline(reader, args[0], remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintln(stderr, "gooo impact:", err)
		return exitFailure
	}
	after, err := readSourceWithDeadline(reader, args[1], remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintln(stderr, "gooo impact:", err)
		return exitFailure
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	r, err := sourceimpact.Compare(ctx, args[0], before, args[1], after)
	if err != nil {
		fmt.Fprintln(stderr, "gooo impact:", err)
		return exitFailure
	}
	var payload []byte
	if format == "json" {
		payload, err = json.Marshal(r)
		payload = append(payload, '\n')
	} else {
		payload = []byte(formatSourceImpact(r))
	}
	if err == nil && int64(len(payload)) > maxOutputBytes {
		err = fmt.Errorf("impact report exceeds %d bytes", maxOutputBytes)
	}
	if err == nil {
		err = writeInspectOutput(stdout, payload, deadline)
	}
	if err != nil {
		fmt.Fprintln(stderr, "gooo impact:", err)
		return exitFailure
	}
	return exitOK
}

func formatSourceImpact(r sourceimpact.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Source change impact: %s\nDeclarations changed: %d; scope changes: %d; declarations to recheck: %d\n", r.Status, len(r.Changes), len(r.ScopeChanges), len(r.Affected))
	for _, change := range r.Changes {
		fmt.Fprintf(&b, "\nChanged %s (%s -> %s): %s\n", change.ID, change.BeforeName, change.AfterName, strings.Join(change.Facets, ", "))
	}
	for _, change := range r.ScopeChanges {
		fmt.Fprintf(&b, "\nScope changed: %s\n", change.Field)
	}
	for _, row := range r.Affected {
		fmt.Fprintf(&b, "\nRecheck %s [%s], present after: %t\n", row.Name, row.ID, row.PresentAfter)
		for _, cause := range row.Causes {
			if cause.Via == nil {
				fmt.Fprintf(&b, "  %s: declaration changed\n", cause.Revision)
				continue
			}
			fmt.Fprintf(&b, "  %s: %s via %s (%s); origin %s, distance %d\n", cause.Revision, cause.Via.Dependency, cause.Via.Kind, cause.Via.Surface, cause.ChangedID, cause.Depth)
		}
		for _, side := range []struct {
			name string
			o    *sourceimpact.Obligations
		}{{"before", row.BeforeObligations}, {"after", row.AfterObligations}} {
			if o := side.o; o != nil {
				fmt.Fprintf(&b, "  %s obligations: %d selection, %d conditions, %d held evaluation, %d observation inputs; oracle %q\n", side.name, o.Cases, o.Conditions, o.Holdouts, o.ObservationInputs, o.Oracle)
				writeImpactBudgets(&b, side.name, o)
			}
		}
	}
	for _, unknown := range r.Unknowns {
		fmt.Fprintf(&b, "\nUnresolved %s %s %s %q: %s\n", unknown.Revision, unknown.ActivityID, unknown.Surface, unknown.Reference, unknown.Reason)
	}
	fmt.Fprintf(&b, "\nModel calls: %d; candidate executions: %d\n%s\n", r.ModelCalls, r.CandidateExecutions, r.Scope)
	return b.String()
}

func writeImpactBudgets(b *strings.Builder, side string, o *sourceimpact.Obligations) {
	if o.MaxAttempts > 0 {
		fmt.Fprintf(b, "  %s declared attempt limit: %d\n", side, o.MaxAttempts)
	}
	if o.FillCandidates > 0 {
		fmt.Fprintf(b, "  %s explicit fill candidates: %d\n", side, o.FillCandidates)
	}
	if o.GeneratedCandidateLimit > 0 {
		fmt.Fprintf(b, "  %s grammar candidate limit: %d\n", side, o.GeneratedCandidateLimit)
	}
}
