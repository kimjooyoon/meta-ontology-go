package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
)

func main() {
	root := flag.String("root", ".", "repository root for metric scan")
	jsonMode := flag.Bool("json", false, "emit metrics as JSON")
	summaryMode := flag.Bool("summary", false, "emit failed indicators as bounded text")
	flag.Parse()

	if err := run(*root, *jsonMode, *summaryMode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, jsonMode, summaryMode bool) error {
	if jsonMode && summaryMode {
		return fmt.Errorf("json and summary modes are mutually exclusive")
	}
	report, err := linecaps.AnalyzeLineMetrics(root)
	if err != nil {
		return err
	}
	report.Repository = os.Getenv("GITHUB_REPOSITORY")
	report.CommitSHA = os.Getenv("METRICS_COMMIT_SHA")
	if summaryMode {
		fmt.Print(report.Summary())
		return nil
	}
	if jsonMode {
		payload, err := report.JSON()
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(payload)
		return err
	}
	fmt.Print(report.Text())
	return nil
}
