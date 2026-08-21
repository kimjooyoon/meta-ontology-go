package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	contract := flag.String(
		"contract", "examples/self-improvement/main.gooo", "Gooo contract path",
	)
	commit := flag.String("commit", "", "exact checked-out commit SHA")
	output := flag.String("output", "", "JSON output path; stdout when empty")
	check := flag.Bool("check", false, "exit non-zero for a failing report")
	flag.Parse()

	source, err := os.ReadFile(*contract)
	if err != nil {
		exitError(err)
	}
	report := buildReport(*contract, source, *commit)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		exitError(err)
	}
	data = append(data, '\n')
	if *output == "" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(*output, data, 0o644)
	}
	if err != nil {
		exitError(err)
	}
	if *check && report.Status != "PASS" {
		os.Exit(1)
	}
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
