package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const packageOriginsUsage = "usage: gooo package origins --receipt <execution.json> [--activity <id>] [--candidate <id>] [--limit <1..128>] [--follow] [--json]"
const packageOriginsHelp = `Read where a selected candidate reaches a condition or return

` + packageOriginsUsage + `

Reads a saved package execution record. Paths run from a condition/return back
to a filled hole, with value inputs and controlling conditions distinguished.
Each root/hole pair shows one shortest recorded static path. Other paths may
exist. The query reads only the selected construction in this receipt.

--follow derives a static value path across selected pure calls and binds.
With --follow, --activity selects the root activity (default: workspace entry);
call arguments keep their own invocation context. This includes original
activity declarations and external input boundaries where retained.
Unfiltered --follow also lists field, index and slice source occurrences in
value_flow.accesses (up to 256); inspect accesses_truncated with truncated.
These expressions describe source choices without assigning runtime element identities.

Without --follow, --activity filters the recorded assembly activity ID. --candidate filters a
candidate ID within matching activities. The default limit is 32 rows; inspect
truncated when using JSON. PARTIAL and missing graphs retain their reasons.
Original package file locations appear where they were recorded; otherwise the
assembly input location is shown. Source views and hashes accompany locations.

This command reads records and source snippets. It makes zero body, candidate or
model executions. Current source comparison is available through gooo impact.
See docs/language/recorded-origin-queries.ko.md.
`

type packageOriginsReceipt struct {
	Schema        string                             `json:"schema"`
	ReceiptSHA256 string                             `json:"receipt_sha256"`
	Report        workspaceexecution.OriginUseReport `json:"report"`
}

func runPackageOrigins(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(stdout, packageOriginsHelp)
		return exitOK
	}
	args, jsonMode := parseJSONFlag(args)
	path, query, err := parsePackageOriginArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, packageOriginsUsage)
		return exitUsage
	}
	deadline := time.Now().Add(commandDeadline)
	raw, err := readBodyExecutionFile(path, 32<<20)
	if err != nil {
		fmt.Fprintln(stderr, "gooo package origins:", err)
		return exitFailure
	}
	var saved packageExecutionReceipt
	if err = bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		fmt.Fprintln(stderr, "gooo package origins:", err)
		return exitFailure
	}
	if saved.Schema != "gooo/workspace-body-execution-receipt/v1" || saved.Result == nil {
		fmt.Fprintln(stderr, "gooo package origins: a saved execution result is required")
		return exitFailure
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	report, err := workspaceexecution.InspectRecordedOrigins(ctx, *saved.Result, query)
	if err != nil {
		fmt.Fprintln(stderr, "gooo package origins:", err)
		return exitFailure
	}
	receipt := packageOriginsReceipt{Schema: "gooo/package-origin-query/v1", ReceiptSHA256: workspaceDigest(raw), Report: report}
	var payload []byte
	if jsonMode {
		payload, err = json.Marshal(receipt)
		payload = append(payload, '\n')
	} else {
		payload = []byte(formatPackageOrigins(report))
	}
	if err == nil && int64(len(payload)) > maxOutputBytes {
		err = fmt.Errorf("origin query output exceeds %d bytes", maxOutputBytes)
	}
	if err == nil {
		err = writeInspectOutput(stdout, payload, deadline)
	}
	if err != nil {
		fmt.Fprintln(stderr, "gooo package origins:", err)
		return exitFailure
	}
	return exitOK
}

func parsePackageOriginArgs(args []string) (string, workspaceexecution.OriginQuery, error) {
	query := workspaceexecution.OriginQuery{Limit: 32}
	var filtered []string
	for _, arg := range args {
		if arg == "--follow" {
			if query.Follow {
				return "", query, fmt.Errorf("duplicate follow flag")
			}
			query.Follow = true
		} else {
			filtered = append(filtered, arg)
		}
	}
	args = filtered
	flags := map[string]string{"--receipt": "", "--activity": "", "--candidate": "", "--limit": ""}
	for i := 0; i < len(args); i += 2 {
		old, known := flags[args[i]]
		if !known || old != "" || i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" || strings.HasPrefix(args[i+1], "-") {
			return "", query, fmt.Errorf("invalid origin query argument")
		}
		flags[args[i]] = args[i+1]
	}
	if flags["--receipt"] == "" {
		return "", query, fmt.Errorf("origin query needs a receipt")
	}
	if flags["--limit"] != "" {
		limit, err := strconv.Atoi(flags["--limit"])
		if err != nil || limit < 1 || limit > 128 {
			return "", query, fmt.Errorf("origin query limit must be 1..128")
		}
		query.Limit = limit
	}
	query.ActivityID, query.CandidateID = flags["--activity"], flags["--candidate"]
	return flags["--receipt"], query, nil
}

func formatPackageOrigins(report workspaceexecution.OriginUseReport) string {
	var out strings.Builder
	fmt.Fprintf(&out, "Recorded candidate origins: %d uses; %d matching constructions; %d recorded value graphs\n", len(report.Rows), report.OriginsMatched, report.GraphsAvailable)
	for _, row := range report.Rows {
		fmt.Fprintf(&out, "\n%s / candidate %s / hole %s\n  %s: %q [%s; %s]\n", row.ActivityID, row.CandidateID, row.Use.HoleID, row.Use.RootKind, row.Expressions[0], row.Use.Relation, row.AnalysisStatus)
		for _, location := range row.Locations {
			fmt.Fprintf(&out, "  source %s: %s bytes %d..%d (%s)\n", location.View, location.Filename, location.Start, location.End, location.SourceSHA256)
		}
		fmt.Fprintln(&out, "  path from use to origin:")
		for i, step := range row.Use.Path {
			fmt.Fprintf(&out, "    %s %s %s: %q\n", step.Via, step.Node.Kind, step.Node.Name, row.Expressions[i])
		}
		if row.SnippetsTruncated {
			fmt.Fprintln(&out, "  snippets shortened; full ranges remain in JSON")
		}
	}
	for _, issue := range report.Issues {
		fmt.Fprintf(&out, "\nPartial %s / %s: %s\n", issue.ActivityID, issue.CandidateID, issue.Reason)
	}
	if report.Truncated {
		fmt.Fprintln(&out, "\nOutput truncated. Narrow --activity or --candidate, or raise --limit up to 128.")
	}
	if report.ValueFlow != nil {
		formatPackageValueOrigins(&out, *report.ValueFlow)
	}
	fmt.Fprintln(&out, "\n"+report.Scope)
	return out.String()
}
