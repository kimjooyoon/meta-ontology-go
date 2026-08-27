package main

import (
	"encoding/json"
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
	before, semanticAfter, commentAfter, metaAfter, corpusRaw := read(options.before), read(options.semanticAfter), read(options.commentAfter), read(options.metaAfter), read(options.corpus)
	if !ok {
		return 2
	}
	corpus, err := cf.DecodeCorpus(corpusRaw)
	if err != nil {
		return 2
	}
	report, err := counterexamplefirstcompiler.AnalyzeInterventions(before, semanticAfter, commentAfter, metaAfter, corpus)
	if err != nil {
		return 2
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return 2
	}
	if err := os.WriteFile(options.out, append(raw, '\n'), 0o644); err != nil {
		return 2
	}
	return 0
}
