package main

import (
	"flag"
	"fmt"
	"os"

	verifier "github.com/kimjooyoon/meta-ontology-go/internal/meta/denominatorevolutionverify"
)

type options struct {
	head, contract, report, source, out string
	repositoryWrites                    int
	snapshotBefore, snapshotAfter       string
}

func run(args []string) int {
	flags := flag.NewFlagSet("denominator-evolution-verify", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "denominator evolution contract")
	flags.StringVar(&value.report, "report", "", "producer report")
	flags.StringVar(&value.source, "source", "", "Gooo denominator governance source")
	flags.StringVar(&value.out, "out", "", "independent verification output")
	flags.IntVar(&value.repositoryWrites, "repository-writes", 0, "observed changed repository paths")
	flags.StringVar(&value.snapshotBefore, "snapshot-before", "", "digest of the CI repository snapshot before execution")
	flags.StringVar(&value.snapshotAfter, "snapshot-after", "", "digest of the CI repository snapshot after execution")
	if flags.Parse(args) != nil || value.head == "" || value.contract == "" || value.report == "" || value.source == "" || value.out == "" {
		return 2
	}
	contractRaw, err := os.ReadFile(value.contract)
	if err != nil {
		return 2
	}
	reportRaw, err := os.ReadFile(value.report)
	if err != nil {
		return 2
	}
	source, err := os.ReadFile(value.source)
	if err != nil {
		return 2
	}
	verification := verifier.Verify(verifier.Input{ContractRaw: contractRaw, ReportRaw: reportRaw, HeadSHA: value.head, SourceRaw: source, RepositorySnapshot: verifier.RepositorySnapshot{BeforeDigest: value.snapshotBefore, AfterDigest: value.snapshotAfter, ChangedPaths: value.repositoryWrites}})
	if err := verifier.WriteVerification(value.out, verification); err != nil {
		return 2
	}
	fmt.Printf("denominator evolution consumer: %s %s checks=%d\n", verification.Decision, verification.Reason, len(verification.Checks))
	for _, check := range verification.Checks {
		if check.Status != "PASS" {
			fmt.Printf("denominator evolution consumer check: %s status=%s expected=%s observed=%s\n", check.ID, check.Status, check.Expected, check.Observed)
		}
	}
	if verification.Decision != "PASS" {
		return 1
	}
	return 0
}
