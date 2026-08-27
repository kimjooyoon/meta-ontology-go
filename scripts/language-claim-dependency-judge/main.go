package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependencyjudge"
)

func main() {
	receiptPath := flag.String("receipt", "", "claim dependency receipt")
	outputPath := flag.String("output", "", "independent judgment output path")
	flag.Parse()
	if *receiptPath == "" || *outputPath == "" {
		fail("-receipt and -output are required")
	}
	data, err := os.ReadFile(*receiptPath)
	if err != nil {
		fail(err.Error())
	}
	var receipt claimdependency.Receipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		fail(err.Error())
	}
	judgment, err := claimdependencyjudge.Judge(receipt)
	if err != nil {
		fail(err.Error())
	}
	encoded, err := json.MarshalIndent(judgment, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	encoded = append(encoded, '\n')
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, encoded, 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("independent claim dependency judgment case=%s decision=%s resolution=%s accepted=%t\n", judgment.Case, judgment.Decision, judgment.Resolution, judgment.Accepted)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
