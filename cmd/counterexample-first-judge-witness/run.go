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

type effectsEvidence struct {
	Schema             string `json:"schema"`
	BeforeStatusDigest string `json:"before_status_digest"`
	AfterStatusDigest  string `json:"after_status_digest"`
	RepositoryWrites   int    `json:"repository_writes"`
	MutationAuthority  string `json:"mutation_authority"`
	CapabilityEvidence string `json:"capability_evidence"`
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
	receiptsRaw, independenceRaw, effectsRaw := read(options.receipts), read(options.independence), read(options.effects)
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
	var effects effectsEvidence
	if err := json.Unmarshal(effectsRaw, &effects); err != nil || effects.Schema != "gooo/counterexample-first-effects/v2" ||
		effects.BeforeStatusDigest == "" || effects.AfterStatusDigest == "" || effects.RepositoryWrites < 0 ||
		effects.BeforeStatusDigest != effects.AfterStatusDigest || effects.MutationAuthority == "" {
		return 2
	}
	report := counterexamplefirstjudge.Evaluate(cf.JudgeInput{Contract: contract, HeadSHA: options.head,
		SourcePath: options.source, Source: source, Corpus: corpus, Receipts: receipts,
		ProducerDependencies: independence.ProducerDependencies,
		WorkspaceEffects:     cf.Effects{RepositoryWrites: effects.RepositoryWrites, MutationAuthority: effects.MutationAuthority, CapabilityEvidence: effects.CapabilityEvidence}})
	if err := cf.WriteReport(options.out, report); err != nil {
		return 2
	}
	fmt.Printf("counterexample-first judge: %s receipts=%d/%d indicators=%d/%d\n", report.Decision,
		report.Summary.ReceiptsReconstructed, report.Summary.CasesTotal, satisfied(report.Indicators), len(report.Indicators))
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
