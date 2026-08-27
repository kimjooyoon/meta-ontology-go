package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotionverify"
)

func main() {
	sourcePath := flag.String("source", experimentpromotionverify.SourcePath, "raw Gooo source")
	observationPath := flag.String("observations", "", "observation receipt bundle")
	contractPath := flag.String("contract", "examples/experiment-promotion/contract.json", "validator expectation contract")
	reportPath := flag.String("report", "experiment-promotion-report.json", "producer report")
	outPath := flag.String("out", "experiment-promotion-verification.json", "verification output")
	subjectSHA := flag.String("subject-sha", "", "workflow subject SHA")
	beforeDigest := flag.String("before-digest", "", "repository snapshot before digest")
	afterDigest := flag.String("after-digest", "", "repository snapshot after digest")
	changedPaths := flag.Int("changed-paths", 0, "repository snapshot changed path count")
	flag.Parse()

	if *observationPath == "" || *beforeDigest == "" || *afterDigest == "" {
		log.Fatal("-observations, -before-digest, and -after-digest are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		log.Fatal(err)
	}
	observations, err := os.ReadFile(*observationPath)
	if err != nil {
		log.Fatal(err)
	}
	contract, err := loadContract(*contractPath)
	if err != nil {
		log.Fatal(err)
	}
	report, err := loadReport(*reportPath)
	if err != nil {
		log.Fatal(err)
	}
	verification := experimentpromotionverify.Verify(experimentpromotionverify.Input{
		SourceRaw: source, ObservationRaw: observations, Contract: contract, Report: report,
		SubjectSHA:         *subjectSHA,
		RepositorySnapshot: experimentpromotionverify.RepositorySnapshot{BeforeDigest: *beforeDigest, AfterDigest: *afterDigest, ChangedPaths: *changedPaths},
	})
	if err := writeVerification(*outPath, verification); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("decision=%s %s\n", verification.Decision, verification.Reason)
	if verification.Decision != "PASS" {
		os.Exit(1)
	}
}

func loadContract(filename string) (experimentpromotionverify.Contract, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return experimentpromotionverify.Contract{}, err
	}
	var contract experimentpromotionverify.Contract
	if err := decodeStrict(raw, &contract); err != nil {
		return contract, err
	}
	return contract, nil
}

func loadReport(filename string) (experimentpromotionverify.Report, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return experimentpromotionverify.Report{}, err
	}
	var report experimentpromotionverify.Report
	if err := decodeStrict(raw, &report); err != nil {
		return report, err
	}
	return report, nil
}

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return err
	}
	return nil
}

func writeVerification(filename string, verification experimentpromotionverify.Verification) error {
	raw, err := json.MarshalIndent(verification, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(filename, raw, 0o644)
}
