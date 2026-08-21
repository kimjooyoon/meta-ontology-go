package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	opts := options{}
	flag.StringVar(&opts.MetricsPath, "metrics", "", "source metrics JSON")
	flag.StringVar(&opts.PlanPath, "plan", "", "self-improvement plan JSON")
	flag.StringVar(&opts.ExecutionPath, "execution", "", "execution manifest JSON")
	flag.StringVar(&opts.ReceiptsPath, "receipts", "", "receipt report JSON")
	flag.StringVar(&opts.ProvenancePath, "provenance", "", "artifact provenance JSON")
	flag.StringVar(&opts.OutputPath, "output", "", "bounded Markdown output")
	flag.StringVar(&opts.ReportPath, "report", "", "machine-readable report output")
	flag.IntVar(&opts.LimitBytes, "limit", defaultLimitBytes, "maximum Markdown bytes")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(opts options) error {
	if opts.OutputPath == "" || opts.ReportPath == "" {
		return fmt.Errorf("output and report paths are required")
	}
	summary, report, err := build(opts)
	if err != nil {
		return err
	}
	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode summary report: %w", err)
	}
	if err := os.WriteFile(opts.OutputPath, summary, 0o644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	if err := os.WriteFile(opts.ReportPath, append(reportData, '\n'), 0o644); err != nil {
		return fmt.Errorf("write summary report: %w", err)
	}
	fmt.Printf("meta summary: decision=%s bytes=%d limit=%d digest=%s\n",
		report.Decision, report.OutputBytes, report.LimitBytes, report.OutputSHA256)
	return nil
}
