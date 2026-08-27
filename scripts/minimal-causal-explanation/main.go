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
	compilerReceiptPath := flag.String("compiler-receipt", "", "raw Gooo compiler receipt")
	repositoryBeforePath := flag.String("repository-before", "", "raw CI repository before observation")
	repositoryAfterPath := flag.String("repository-after", "", "raw CI repository after observation")
	independencePath := flag.String("independence", "", "raw producer import independence observation")
	repository := flag.String("repository", "github.com/kimjooyoon/meta-ontology-go", "subject repository")
	subjectSHA := flag.String("subject-sha", "", "exact 40-character subject SHA")
	receiptPath := flag.String("receipt", "", "receipt input path for verify")
	outputPath := flag.String("output", "", "output JSON path")
	check := flag.Bool("check", false, "run the independent judge")
	flag.Parse()
	if *outputPath == "" {
		fail("-output is required")
	}

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	compilerReceipt, repositoryBefore, repositoryAfter, independence := readInputs(*compilerReceiptPath, *repositoryBeforePath, *repositoryAfterPath, *independencePath)
	var output any
	switch *mode {
	case "generate":
		receipt, err := explanation.Evaluate(*sourcePath, source, compilerReceipt, repositoryBefore, repositoryAfter, independence, *repository, *subjectSHA)
		if err != nil {
			fail(err.Error())
		}
		if *check {
			encoded, err := json.Marshal(receipt)
			if err != nil {
				fail(err.Error())
			}
			if _, err := judge.Judge(encoded, *sourcePath, source, compilerReceipt, repositoryBefore, repositoryAfter, independence); err != nil {
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
		judgment, err := judge.Judge(data, *sourcePath, source, compilerReceipt, repositoryBefore, repositoryAfter, independence)
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

func readInputs(compilerPath, beforePath, afterPath, independencePath string) ([]byte, []byte, []byte, []byte) {
	read := func(path, name string) []byte {
		if path == "" {
			fail("-" + name + " is required")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fail(err.Error())
		}
		return data
	}
	return read(compilerPath, "compiler-receipt"), read(beforePath, "repository-before"), read(afterPath, "repository-after"), read(independencePath, "independence")
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
