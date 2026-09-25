package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metabinding"
)

func main() {
	root := flag.String("root", "", "repository root")
	metrics := flag.String("metrics", "", "raw source metrics JSON")
	bound := flag.String("bound-metrics", "", "augmented source metrics JSON")
	reportPath := flag.String("report", "", "meta-binding report JSON")
	check := flag.Bool("check", false, "exit non-zero unless binding is complete")
	flag.Parse()
	if *root == "" || *metrics == "" || *bound == "" || *reportPath == "" {
		fail(fmt.Errorf("root, metrics, bound-metrics, and report are required"))
	}
	in, err := metabinding.Load(*metrics)
	if err != nil {
		fail(err)
	}
	report, indicator, err := metabinding.Build(*root, in)
	if err != nil {
		fail(err)
	}
	boundData, err := metabinding.Augment(in, indicator)
	if err != nil {
		fail(err)
	}
	reportData, err := metabinding.MarshalReport(report)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(*bound, boundData, 0o644); err != nil {
		fail(err)
	}
	if err := os.WriteFile(*reportPath, reportData, 0o644); err != nil {
		fail(err)
	}
	fmt.Printf("meta-binding: decision=%s bound=%d unbound=%d coverage_bps=%d digest=%s\n",
		report.Decision, report.Summary.BoundIndicators, report.Summary.UnboundIndicators,
		report.Summary.CoverageBasisPoints, report.ReportDigest)
	if *check && report.Decision != "PASS" {
		os.Exit(1)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
