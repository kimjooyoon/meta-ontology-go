package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion"
)

func main() {
	sourcePath := flag.String("source", experimentpromotion.SourcePath, "raw Gooo source")
	observationPath := flag.String("observations", "", "observation receipt bundle")
	contractPath := flag.String("contract", "examples/experiment-promotion/contract.json", "validator expectation contract")
	outPath := flag.String("out", "experiment-promotion-report.json", "report output")
	subjectSHA := flag.String("subject-sha", "", "workflow subject SHA")
	beforeDigest := flag.String("before-digest", experimentpromotion.DigestBytes(nil), "repository snapshot before digest")
	afterDigest := flag.String("after-digest", experimentpromotion.DigestBytes(nil), "repository snapshot after digest")
	changedPaths := flag.Int("changed-paths", 0, "repository snapshot changed path count")
	flag.Parse()

	if *observationPath == "" {
		log.Fatal("-observations is required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		log.Fatal(err)
	}
	observations, err := os.ReadFile(*observationPath)
	if err != nil {
		log.Fatal(err)
	}
	contract, err := experimentpromotion.LoadContract(*contractPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := validDigestFlag(*beforeDigest, *afterDigest); err != nil {
		log.Fatal(err)
	}
	report := experimentpromotion.Evaluate(experimentpromotion.Input{
		SubjectSHA: *subjectSHA, SourceRaw: source, ObservationRaw: observations, Contract: contract,
		RepositorySnapshot: experimentpromotion.RepositorySnapshot{BeforeDigest: *beforeDigest, AfterDigest: *afterDigest, ChangedPaths: *changedPaths},
	})
	if err := experimentpromotion.ValidateReport(report); err != nil {
		log.Fatal(err)
	}
	if err := experimentpromotion.WriteReport(*outPath, report); err != nil {
		log.Fatal(err)
	}
	fmt.Println(experimentpromotion.FormatSummary(report))
}

func validDigestFlag(before, after string) error {
	if before == "" || after == "" {
		return fmt.Errorf("repository snapshot digests are required")
	}
	if len(before) != len(experimentpromotion.DigestBytes(nil)) || len(after) != len(experimentpromotion.DigestBytes(nil)) {
		return fmt.Errorf("repository snapshot digest is malformed")
	}
	return nil
}
