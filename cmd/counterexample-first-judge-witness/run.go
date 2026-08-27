package main

import (
	"encoding/json"
	"fmt"
	"os"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirstjudge"
)

const independenceSchema = "gooo/counterexample-first-independence/v1"

type independenceEvidence struct {
	Schema               string `json:"schema"`
	ProducerDependencies int    `json:"producer_dependencies"`
}

func run(args []string) int {
	options, ok := parseOptions(args)
	if !ok {
		return 2
	}
	read := func(path string) []byte {
		raw, err := os.ReadFile(path)
		if err != nil {
			ok = false
		}
		return raw
	}
	contractRaw, source, corpusRaw := read(options.contract), read(options.source), read(options.corpus)
	receiptsRaw, independenceRaw := read(options.receipts), read(options.independence)
	if !ok {
		return 2
	}
	contract, err := cf.DecodeContract(contractRaw)
	if err != nil {
		return 2
	}
	corpus, err := cf.DecodeCorpus(corpusRaw)
	if err != nil {
		return 2
	}
	receipts, err := cf.DecodeReceipts(receiptsRaw)
	if err != nil {
		return 2
	}
	var independence independenceEvidence
	if err := json.Unmarshal(independenceRaw, &independence); err != nil ||
		independence.Schema != independenceSchema || independence.ProducerDependencies < 0 {
		return 2
	}
	report := counterexamplefirstjudge.Evaluate(cf.JudgeInput{Contract: contract, HeadSHA: options.head,
		SourcePath: options.source, Source: source, Corpus: corpus, Receipts: receipts,
		ProducerDependencies: independence.ProducerDependencies})
	if err := cf.WriteReport(options.out, report); err != nil {
		return 2
	}
	fmt.Printf("counterexample-first judge: %s cases=%d/%d indicators=%d/%d\n", report.Decision,
		report.Summary.CasesSatisfied, report.Summary.CasesTotal, satisfied(report.Indicators), len(report.Indicators))
	if report.Decision != "PASS" {
		return 1
	}
	return 0
}

func satisfied(values []cf.Indicator) int {
	count := 0
	for _, value := range values {
		if value.Satisfied {
			count++
		}
	}
	return count
}
