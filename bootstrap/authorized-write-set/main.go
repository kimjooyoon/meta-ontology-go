package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type options struct {
	expectedSHA, density, extraction, observed, expectedOut, output string
	untracked                                                       int
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	var value options
	flags := flag.NewFlagSet("authorized-write-set", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&value.expectedSHA, "expected-sha", "", "exact subject SHA")
	flags.StringVar(&value.density, "density-report", "", "line density receipt")
	flags.StringVar(&value.extraction, "extraction-report", "", "function extraction receipt")
	flags.StringVar(&value.observed, "observed-write-set", "", "observed changed paths")
	flags.IntVar(&value.untracked, "untracked-count", -1, "untracked workspace paths")
	flags.StringVar(&value.expectedOut, "expected-output", "", "authorized path output")
	flags.StringVar(&value.output, "output", "", "evidence JSON output")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if value.expectedSHA == "" || value.density == "" || value.extraction == "" ||
		value.observed == "" || value.expectedOut == "" || value.output == "" {
		fmt.Fprintln(stderr, "authorized-write-set: required input missing")
		return 2
	}
	density, extraction, observed, err := loadInputs(value.density, value.extraction, value.observed)
	var report evidence
	if err != nil {
		report = evidence{Schema: evidenceSchema, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
			Reason: "WRITE_SET_INPUT_UNKNOWN", Audience: "GOVERNOR", SourceSHA: value.expectedSHA,
			MetaOperation: metaOperation, Coordinates: coordinates{SourceReceiptsTotal: 2, Unknowns: 2}}
		report = seal(report)
	} else {
		report = reduce(value.expectedSHA, density, extraction, observed, value.untracked)
	}
	if err := writeOutputs(value.expectedOut, value.output, report); err != nil {
		fmt.Fprintf(stderr, "authorized-write-set: write: %v\n", err)
		return 2
	}
	c := report.Coordinates
	fmt.Fprintf(stdout, "authorized write set: %s/%s sources=%d/%d density=%d extraction=%d overlap=%d expected=%d observed=%d untracked=%d unknown=%d\n",
		report.Decision, report.Resolution, c.SourceReceipts, c.SourceReceiptsTotal, c.DensityPaths,
		c.ExtractionPaths, c.OverlapPaths, c.ExpectedPaths, c.ObservedPaths, c.UntrackedPaths, c.Unknowns)
	if report.Decision != "PASS" {
		fmt.Fprintf(stderr, "authorized-write-set: %s\n", report.Reason)
		return 1
	}
	return 0
}
