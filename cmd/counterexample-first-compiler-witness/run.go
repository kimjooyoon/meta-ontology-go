package main

import (
	"fmt"
	"os"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirstcompiler"
)

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
	receipts, err := counterexamplefirstcompiler.Compile(contract, options.head, options.source, source, corpus)
	if err != nil {
		return 2
	}
	if err := cf.WriteReceipts(options.out, receipts); err != nil {
		return 2
	}
	fmt.Printf("counterexample-first compiler: receipts=%d promoted=%d\n", len(receipts), promoted(receipts))
	return 0
}

func promoted(receipts []cf.DecisionReceipt) int {
	count := 0
	for _, receipt := range receipts {
		if receipt.Decision == "PASS" && receipt.Resolution == "EXACT" {
			count++
		}
	}
	return count
}
