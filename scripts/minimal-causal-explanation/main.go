package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	explanation "github.com/kimjooyoon/meta-ontology-go/internal/meta/minimalcausalexplanation"
	judge "github.com/kimjooyoon/meta-ontology-go/internal/meta/minimalcausalexplanation/verify"
)

func main() {
	mode := flag.String("mode", "generate", "generate or verify")
	sourcePath := flag.String("source", "examples/minimal-causal-explanation/main.gooo", "Gooo source")
	repository := flag.String("repository", "github.com/kimjooyoon/meta-ontology-go", "subject repository")
	subjectSHA := flag.String("subject-sha", "", "exact 40-character subject SHA")
	receiptPath := flag.String("receipt", "", "receipt input path for verify")
	outputPath := flag.String("output", "", "output JSON path")
	check := flag.Bool("check", false, "run the independent judge")
	flag.Parse()
	if *outputPath == "" {
		fail("-output is required")
	}

	var output any
	switch *mode {
	case "generate":
		source, err := os.ReadFile(*sourcePath)
		if err != nil {
			fail(err.Error())
		}
		receipt, err := explanation.Evaluate(*sourcePath, source, *repository, *subjectSHA)
		if err != nil {
			fail(err.Error())
		}
		if *check {
			if _, err := judge.Judge(receipt); err != nil {
				fail(err.Error())
			}
		}
		output = receipt
	case "verify":
		if *receiptPath == "" {
			fail("-receipt is required for verify")
		}
		data, err := os.ReadFile(*receiptPath)
		if err != nil {
			fail(err.Error())
		}
		var receipt explanation.Receipt
		if err := json.Unmarshal(data, &receipt); err != nil {
			fail(err.Error())
		}
		judgment, err := judge.Judge(receipt)
		if err != nil {
			fail(err.Error())
		}
		output = judgment
	default:
		fail("-mode must be generate or verify")
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, data, 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("minimal causal explanation mode=%s output=%s\n", *mode, *outputPath)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
