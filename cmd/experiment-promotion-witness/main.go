package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion"
)

func main() {
	sourcePath := flag.String("source", experimentpromotion.SourcePath, "raw Gooo source")
	observationPath := flag.String("observations", "", "raw observation bundle")
	contractPath := flag.String("contract", "examples/experiment-promotion/contract.json", "validator expectation contract")
	outPath := flag.String("out", "experiment-promotion-report.json", "report output")
	subjectSHA := flag.String("subject-sha", "", "workflow subject SHA")
	beforePath := flag.String("snapshot-before", "", "captured repository snapshot before replay")
	afterPath := flag.String("snapshot-after", "", "captured repository snapshot after replay")
	pathsPath := flag.String("snapshot-paths", "", "captured changed-path list")
	flag.Parse()
	if *observationPath == "" || *beforePath == "" || *afterPath == "" || *pathsPath == "" {
		log.Fatal("-observations, -snapshot-before, -snapshot-after, and -snapshot-paths are required")
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
	before, err := os.ReadFile(*beforePath)
	if err != nil {
		log.Fatal(err)
	}
	after, err := os.ReadFile(*afterPath)
	if err != nil {
		log.Fatal(err)
	}
	pathRaw, err := os.ReadFile(*pathsPath)
	if err != nil {
		log.Fatal(err)
	}
	changed := make([]string, 0)
	for _, path := range strings.Split(strings.TrimSpace(string(pathRaw)), "\n") {
		if path != "" {
			changed = append(changed, path)
		}
	}
	report := experimentpromotion.Evaluate(experimentpromotion.Input{SubjectSHA: *subjectSHA, SourceRaw: source, ObservationRaw: observations, Contract: contract, RepositorySnapshot: experimentpromotion.RepositorySnapshot{BeforeRaw: before, BeforeDigest: experimentpromotion.DigestBytes(before), AfterRaw: after, AfterDigest: experimentpromotion.DigestBytes(after), ChangedPaths: len(changed), ChangedPathList: changed}})
	if err := experimentpromotion.ValidateReport(report); err != nil {
		log.Fatal(err)
	}
	if err := experimentpromotion.WriteReport(*outPath, report); err != nil {
		log.Fatal(err)
	}
	println(experimentpromotion.FormatSummary(report))
}
