package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/actionability"
)

func main() {
	root := flag.String("root", "", "repository root")
	metrics := flag.String("metrics", "", "meta-bound source metrics JSON")
	binding := flag.String("binding", "", "meta-binding report JSON")
	reportPath := flag.String("report", "", "actionability report JSON")
	check := flag.Bool("check", false, "reject malformed or unbound input")
	flag.Parse()
	if *root == "" || *metrics == "" || *binding == "" || *reportPath == "" {
		fail(fmt.Errorf("root, metrics, binding, and report are required"))
	}
	in, err := actionability.Load(*metrics, *binding)
	if err != nil {
		fail(err)
	}
	report, err := actionability.Build(*root, in)
	if err != nil {
		fail(err)
	}
	data, err := actionability.Marshal(report)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(*reportPath, data, 0o644); err != nil {
		fail(err)
	}
	fmt.Printf("meta-actionability: decision=%s indicators=%d/%d operations=%d/%d selected=%s digest=%s\n",
		report.Decision, report.Summary.ActionableIndicators,
		report.Summary.ApplicableBlockingIndicators, report.Summary.ExecutableOperations,
		report.Summary.RequiredOperations, report.SelectedOperation, report.ReportDigest)
	if *check && report.Decision == "FAIL_CLOSED" {
		os.Exit(1)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
