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
	flag.Parse()

	if err := run(*root, *jsonMode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, jsonMode bool) error {
	report, err := linecaps.AnalyzeLineMetrics(root)
	if err != nil {
		return err
	}
	report.Repository = os.Getenv("GITHUB_REPOSITORY")
	report.CommitSHA = os.Getenv("METRICS_COMMIT_SHA")
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
