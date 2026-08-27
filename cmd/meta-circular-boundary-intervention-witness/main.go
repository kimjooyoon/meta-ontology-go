package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundary"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundaryconsumer"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

const interventionSchema = "gooo/meta-circular-boundary-causality/v1"

func main() {
	sourcePath := flag.String("source", producer.ExpectedSourcePath, "Gooo source to intervene")
	headSHA := flag.String("head-sha", "", "exact 40-character commit SHA")
	output := flag.String("output", "", "causality report output path")
	flag.Parse()

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatal(err)
	}
	baseInput := contract.Input{Path: *sourcePath, HeadSHA: *headSHA, Source: source}
	baseline := producer.Evaluate(baseInput)
	if err := consumer.Judge(baseline, baseInput); err != nil {
		fatal(fmt.Errorf("baseline consumer judge: %w", err))
	}

	semanticNeedle := []byte("gooo://meta-circular-boundary/entity/self-description")
	semanticReplacement := []byte("gooo://meta-circular-boundary/entity/self-description-v2")
	if bytes.Count(source, semanticNeedle) != 1 {
		fatal(fmt.Errorf("semantic intervention needle count is not one"))
	}
	semanticSource := bytes.Replace(source, semanticNeedle, semanticReplacement, 1)
	nonSemanticSource := append(append([]byte(nil), source...), '\n')
	cases := []contract.CausalityCase{
		intervention("semantic-entity-id", "SEMANTIC", true, baseline, *sourcePath, *headSHA, semanticSource),
		intervention("trailing-newline", "NON_SEMANTIC", false, baseline, *sourcePath, *headSHA, nonSemanticSource),
	}
	report := contract.CausalityReport{Schema: interventionSchema, Cases: cases}
	report.Summary.Total = len(cases)
	for _, item := range cases {
		if item.Passed {
			report.Summary.Passed++
		}
	}
	if report.Summary.Total > 0 {
		report.Summary.CoverageBPS = report.Summary.Passed * 10_000 / report.Summary.Total
	}
	if report.Summary.Passed != report.Summary.Total {
		fatal(fmt.Errorf("semantic-causality contract failed: %d/%d", report.Summary.Passed, report.Summary.Total))
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if *output == "" {
		_, _ = os.Stdout.Write(encoded)
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("semantic-causality: %d/%d (%d BPS)\n", report.Summary.Passed, report.Summary.Total, report.Summary.CoverageBPS)
}

func intervention(id, kind string, expectedSemanticChange bool, baseline contract.Report, path, head string, source []byte) contract.CausalityCase {
	input := contract.Input{Path: path, HeadSHA: head, Source: source}
	intervened := producer.Evaluate(input)
	consumerAccepted := consumer.Judge(intervened, input) == nil
	semanticChanged := baseline.Source.SemanticDigest != intervened.Source.SemanticDigest
	sourceChanged := baseline.Source.SourceDigest != intervened.Source.SourceDigest
	return contract.CausalityCase{
		ID: id, Kind: kind,
		BaselineSourceDigest: baseline.Source.SourceDigest, IntervenedSourceDigest: intervened.Source.SourceDigest,
		BaselineSemanticDigest: baseline.Source.SemanticDigest, IntervenedSemanticDigest: intervened.Source.SemanticDigest,
		SourceChanged: sourceChanged, SemanticChanged: semanticChanged, ExpectedSemanticChange: expectedSemanticChange,
		ConsumerAccepted: consumerAccepted, Passed: consumerAccepted && sourceChanged && semanticChanged == expectedSemanticChange,
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
