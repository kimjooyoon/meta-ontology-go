package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type options struct {
	expectedSHA, density, extraction, split, observed, untracked, expectedOut, output string
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	var value options
	flags := flag.NewFlagSet("authorized-write-set", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&value.expectedSHA, "expected-sha", "", "exact subject SHA")
	flags.StringVar(&value.density, "density-report", "", "line density receipt")
	flags.StringVar(&value.extraction, "extraction-report", "", "function extraction receipt")
	flags.StringVar(&value.split, "split-report", "", "logical source split receipt")
	flags.StringVar(&value.observed, "observed-write-set", "", "observed changed paths")
	flags.StringVar(&value.untracked, "untracked-paths", "", "observed untracked path list")
	flags.StringVar(&value.expectedOut, "expected-output", "", "authorized path output")
	flags.StringVar(&value.output, "output", "", "evidence JSON output")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if value.expectedSHA == "" || value.density == "" || value.extraction == "" || value.split == "" ||
		value.observed == "" || value.untracked == "" || value.expectedOut == "" || value.output == "" {
		fmt.Fprintln(stderr, "authorized-write-set: required input missing")
		return 2
	}
	density, extraction, split, observed, untracked, err := loadInputs(value.density, value.extraction, value.split, value.observed, value.untracked)
	var report evidence
	if err != nil {
		report = evidence{Schema: evidenceSchema, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
			Reason: "WRITE_SET_INPUT_UNKNOWN", Audience: "GOVERNOR", SourceSHA: value.expectedSHA,
			MetaOperation: metaOperation, Coordinates: coordinates{SourceReceiptsTotal: 3, Unknowns: 3}}
		report = seal(report)
	} else {
		report = reduce(value.expectedSHA, density, extraction, split, observed, untracked)
	}
	if err := writeOutputs(value.expectedOut, value.output, report); err != nil {
		fmt.Fprintf(stderr, "authorized-write-set: write: %v\n", err)
		return 2
	}
	c := report.Coordinates
	fmt.Fprintf(stdout, "authorized write set: %s/%s sources=%d/%d density=%d extraction=%d split=%d overlap=%d expected=%d observed=%d untracked=%d unknown=%d\n",
		report.Decision, report.Resolution, c.SourceReceipts, c.SourceReceiptsTotal, c.DensityPaths,
		c.ExtractionPaths, c.SplitPaths, c.OverlapPaths, c.ExpectedPaths, c.ObservedPaths, c.UntrackedPaths, c.Unknowns)
	if report.Decision != "PASS" {
		fmt.Fprintf(stderr, "authorized-write-set: %s\n", report.Reason)
		return 1
	}
	return 0
}
