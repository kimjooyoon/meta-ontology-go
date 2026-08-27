package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependencyjudge"
)

func main() {
	sourcePath := flag.String("source", "", "raw claim dependency Gooo source")
	priorPath := flag.String("prior-receipt", "", "raw prior receipt, required for recovery")
	observationPath := flag.String("observation", "", "raw observation JSON")
	receiptPath := flag.String("receipt", "", "receipt JSON")
	outputPath := flag.String("output", "", "independent judgment output path")
	flag.Parse()
	if *sourcePath == "" || *observationPath == "" || *receiptPath == "" || *outputPath == "" {
		fail("-source, -observation, -receipt, and -output are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	observation, err := os.ReadFile(*observationPath)
	if err != nil {
		fail(err.Error())
	}
	receipt, err := os.ReadFile(*receiptPath)
	if err != nil {
		fail(err.Error())
	}
	var prior []byte
	if *priorPath != "" {
		prior, err = os.ReadFile(*priorPath)
		if err != nil {
			fail(err.Error())
		}
	}
	judgment, err := claimdependencyjudge.Judge(source, *sourcePath, prior, observation, receipt)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*outputPath, judgment)
	fmt.Printf("independent claim dependency predicate=%s decision=%s resolution=%s accepted=%t source_reconstruction=%s producer_import=%d/%d recovery_chain=%d\n", judgment.Predicate, judgment.Decision, judgment.Resolution, judgment.Accepted, judgment.SourceReconstruction, judgment.ProducerPackageImportNumerator, judgment.ProducerPackageImportDenominator, judgment.AppendOnlyRecoveryChainTotal)
}

func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
