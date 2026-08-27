package main

import (
	"fmt"
	"os"

	oracle "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageartifactoracle"
)

func run(args []string) int {
	options, ok := parseOptions(args)
	if !ok { return 2 }
	read := func(path string) []byte {
		raw, err := os.ReadFile(path)
		if err != nil { ok = false }
		return raw
	}
	contractRaw, source, unsupported := read(options.contract), read(options.source), read(options.unsupported)
	genuine, forged := read(options.genuine), read(options.forged)
	unknown, legacy, independenceRaw := read(options.unknown), read(options.legacy), read(options.independence)
	if !ok { return 2 }
	contract, err := oracle.DecodeContract(contractRaw)
	if err != nil { return 2 }
	independence, err := oracle.DecodeIndependence(independenceRaw)
	if err != nil { return 2 }
	report := oracle.Evaluate(oracle.Input{Contract: contract, HeadSHA: options.head,
		Filename: options.source, UnsupportedFilename: options.unsupported, Entry: options.entry,
		Source: source, UnsupportedSource: unsupported, Genuine: genuine, Forged: forged,
		UnknownDecision: unknown, LegacyAcceptance: legacy, Independence: independence})
	if err := oracle.WriteReport(options.out, report); err != nil { return 2 }
	fmt.Printf("artifact oracle: %s %d/%d counterexample=%d\n", report.Decision,
		report.Summary.CasesSatisfied, report.Summary.CasesTotal, report.Summary.LegacyValidatorCounterexamples)
	if report.Decision != "PASS" { return 1 }
	return 0
}
