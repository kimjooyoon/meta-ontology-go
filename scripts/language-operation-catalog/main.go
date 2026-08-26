package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/valuecatalog"
)

func main() {
	source := flag.String("source", "examples/language-operation-catalog/main.gooo", "catalog Gooo source")
	head := flag.String("head-sha", "", "exact 40-character commit SHA")
	output := flag.String("output", "", "receipt output path")
	check := flag.Bool("check", false, "validate the exact baseline or extension state")
	flag.Parse()
	report := valuecatalog.Evaluate(os.DirFS("."), *source, *head)
	if *check {
		if err := valuecatalog.Validate(report, *head); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	encoded = append(encoded, '\n')
	if *output == "" {
		_, _ = os.Stdout.Write(encoded)
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
