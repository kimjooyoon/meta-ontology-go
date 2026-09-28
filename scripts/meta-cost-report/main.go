package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	format := flag.String("format", "json", "Output format: json or markdown")
	flag.Parse()
	if *format != "json" && *format != "markdown" {
		fmt.Fprintln(os.Stderr, "unsupported output format")
		os.Exit(1)
	}
	report, err := readCostReport(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *format == "markdown" {
		if err := writeCostMarkdown(os.Stdout, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
