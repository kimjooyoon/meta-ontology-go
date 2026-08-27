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
	evidencePath := flag.String("evidence", "", "raw CURRENT_EVIDENCE receipt")
	priorPath := flag.String("prior-receipt", "", "raw prior UNKNOWN receipt")
	receiptPath := flag.String("receipt", "", "producer receipt")
	contractPath := flag.String("contract", "", "external validator contract read independently by the judge")
	manifestPath := flag.String("structural-manifest", "", "external structural inventory oracle read independently by the judge")
	outputPath := flag.String("output", "", "independent judgment output")
	flag.Parse()
	if *sourcePath == "" || *evidencePath == "" || *receiptPath == "" || *contractPath == "" || *manifestPath == "" || *outputPath == "" {
		fail("-source, -evidence, -receipt, -contract, -structural-manifest, and -output are required")
	}
	source, evidence, receipt := read(*sourcePath), read(*evidencePath), read(*receiptPath)
	var prior []byte
	if *priorPath != "" {
		prior = read(*priorPath)
	}
	judgment, err := claimdependencyjudge.JudgeWithExternalMaterials(source, *sourcePath, prior, evidence, receipt, *contractPath, *manifestPath)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*outputPath, judgment)
	fmt.Printf("independent judge decision=%s accepted=%t source_reconstruction=%d/%d producer_import=%d/%d causal_edges=%d/%d authority=%s\n", judgment.Decision, judgment.Accepted, judgment.SourceReconstructionNumerator, judgment.SourceReconstructionDenominator, judgment.ProducerPackageImportNumerator, judgment.ProducerPackageImportDenominator, judgment.Metrics.ObservedCausalEdgeTotal, judgment.Metrics.EligibleEdgeTotal, judgment.AuthorityResolution)
}

func read(path string) []byte {
	value, err := os.ReadFile(path)
	if err != nil {
		fail(err.Error())
	}
	return value
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
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(2) }
