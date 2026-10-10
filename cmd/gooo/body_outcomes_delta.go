package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/outcomedelta"
)

const bodyOutcomesDeltaUsage = "usage: gooo body-outcomes-delta --before <saved.json> --after <saved.json> [--json]"

func runBodyOutcomesDelta(args []string, stdout, stderr io.Writer) int {
	flags := map[string]string{"--before": "", "--after": ""}
	asJSON := false
	for i := 0; i < len(args); i++ {
		if args[i] == "--json" && !asJSON {
			asJSON = true
			continue
		}
		value, ok := flags[args[i]]
		if !ok || value != "" || i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			fmt.Fprintln(stderr, bodyOutcomesDeltaUsage)
			return exitUsage
		}
		flags[args[i]] = args[i+1]
		i++
	}
	if flags["--before"] == "" || flags["--after"] == "" {
		fmt.Fprintln(stderr, bodyOutcomesDeltaUsage)
		return exitUsage
	}
	fail := func(err error) int { fmt.Fprintln(stderr, "gooo body-outcomes-delta:", err); return exitFailure }
	before, err := readBodyExecutionFile(flags["--before"], outcomedelta.MaxInputBytes)
	if err != nil {
		return fail(fmt.Errorf("before: %w", err))
	}
	after, err := readBodyExecutionFile(flags["--after"], outcomedelta.MaxInputBytes)
	if err != nil {
		return fail(fmt.Errorf("after: %w", err))
	}
	report, err := outcomedelta.Compare(before, after)
	if err != nil {
		return fail(err)
	}
	if asJSON {
		err = json.NewEncoder(stdout).Encode(report)
	} else {
		err = writeOutcomeDelta(stdout, report)
	}
	if err != nil {
		return fail(err)
	}
	return exitOK
}

func writeOutcomeDelta(out io.Writer, r *outcomedelta.OutcomeDelta) error {
	var text strings.Builder
	fmt.Fprintf(&text, "Compared %d activity/input groups from saved observations.\n", len(r.Rows))
	fmt.Fprintf(&text, "Outcomes changed: %d; unchanged: %d; ambiguous: %d; unobserved: %d\n",
		r.Counts["outcome:CHANGED"], r.Counts["outcome:UNCHANGED"], r.Counts["outcome:AMBIGUOUS"], r.Counts["outcome:UNOBSERVED"])
	fmt.Fprintf(&text, "Requirements changed: %d; regressions under the same requirement: %d; improvements: %d\n",
		r.Counts["requirement:CHANGED"], r.Counts["assessment:REGRESSION"], r.Counts["assessment:IMPROVEMENT"])
	fmt.Fprintf(&text, "Only before: %d; only after: %d\n", r.Counts["presence:BEFORE_ONLY"], r.Counts["presence:AFTER_ONLY"])
	fmt.Fprintf(&text, "Not assessed against a shared observed requirement: %d; conflicting duplicate groups: %d\n",
		r.Counts["assessment:UNOBSERVED"], r.Counts["assessment:AMBIGUOUS"])
	for _, row := range r.Rows {
		if row.OutcomeChange == "UNCHANGED" && row.Assessment == "STILL_MATCHING" {
			continue
		}
		inputs, _ := json.Marshal(row.RootInputs)
		fmt.Fprintf(&text, "  %s %s: %s; outcome=%s; requirement=%s; %s\n", row.ActivityID, inputs,
			row.Presence, row.OutcomeChange, row.RequirementChange, row.Assessment)
		fmt.Fprintf(&text, "    actual: %s -> %s; expected: %s -> %s\n",
			outcomeValue(row.Before, "actual"), outcomeValue(row.After, "actual"),
			outcomeValue(row.Before, "expected"), outcomeValue(row.After, "expected"))
	}
	fmt.Fprintln(&text, "Comparison made no model calls or program executions. Missing observations remain unobserved.")
	fmt.Fprintln(&text, "Use --json for original values, case indices, source identities and every unchanged group.")
	_, err := io.WriteString(out, text.String())
	return err
}

func outcomeValue(records []map[string]any, key string) string {
	if len(records) == 0 {
		return "(absent)"
	}
	if len(records) > 1 {
		return fmt.Sprintf("(%d observations; see --json)", len(records))
	}
	value, present := records[0][key]
	if !present {
		return "(unobserved or unavailable)"
	}
	b, _ := json.Marshal(value)
	runes := []rune(string(b))
	if len(runes) > 120 {
		return string(runes[:120]) + "… (full value in --json)"
	}
	return string(b)
}
