package main

import (
	"flag"
	"fmt"
	"os"

	denominator "github.com/kimjooyoon/meta-ontology-go/internal/meta/denominatorevolution"
)

type options struct {
	head, contract, source, out   string
	repositoryWrites              int
	snapshotBefore, snapshotAfter string
}

func run(args []string) int {
	flags := flag.NewFlagSet("denominator-evolution-witness", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "denominator evolution contract")
	flags.StringVar(&value.source, "source", "", "Gooo denominator governance source")
	flags.StringVar(&value.out, "out", "", "producer report output")
	flags.IntVar(&value.repositoryWrites, "repository-writes", 0, "observed changed repository paths")
	flags.StringVar(&value.snapshotBefore, "snapshot-before", "", "digest of the CI repository snapshot before execution")
	flags.StringVar(&value.snapshotAfter, "snapshot-after", "", "digest of the CI repository snapshot after execution")
	if flags.Parse(args) != nil || value.head == "" || value.contract == "" || value.source == "" || value.out == "" {
		return 2
	}
	contractRaw, err := os.ReadFile(value.contract)
	if err != nil {
		return 2
	}
	source, err := os.ReadFile(value.source)
	if err != nil {
		return 2
	}
	contract, err := denominator.DecodeContract(contractRaw)
	if err != nil {
		return 2
	}
	report := denominator.Evaluate(denominator.Input{Contract: contract, HeadSHA: value.head, Source: source, RepositorySnapshot: denominator.RepositorySnapshot{BeforeDigest: value.snapshotBefore, AfterDigest: value.snapshotAfter, ChangedPaths: value.repositoryWrites}})
	if err := denominator.WriteReport(value.out, report); err != nil {
		return 2
	}
	fmt.Printf("denominator evolution producer: %s %s fixed=%d/%d cases=%d/%d\n", report.Decision, report.Reason, report.Summary.FixedDenominatorNumerator, report.Summary.FixedDenominatorDenominator, report.Summary.CasesSatisfied, report.Summary.CasesTotal)
	if report.Decision != "PASS" {
		return 1
	}
	return 0
}
