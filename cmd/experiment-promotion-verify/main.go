package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotionverify"
)

func main() {
	sourcePath := flag.String("source", experimentpromotionverify.SourcePath, "raw Gooo source")
	observationPath := flag.String("observations", "", "raw observation bundle")
	contractPath := flag.String("contract", "examples/experiment-promotion/contract.json", "validator expectation contract")
	reportPath := flag.String("report", "experiment-promotion-report.json", "producer report")
	outPath := flag.String("out", "experiment-promotion-verification.json", "verification output")
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
	contract, err := loadJSON[experimentpromotionverify.Contract](*contractPath)
	if err != nil {
		log.Fatal(err)
	}
	report, err := loadJSON[experimentpromotionverify.Report](*reportPath)
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
	verification := experimentpromotionverify.Verify(experimentpromotionverify.Input{SubjectSHA: *subjectSHA, SourceRaw: source, ObservationRaw: observations, Contract: contract, Report: report, RepositorySnapshot: experimentpromotionverify.RepositorySnapshot{BeforeRaw: before, BeforeDigest: digestBytes(before), AfterRaw: after, AfterDigest: digestBytes(after), ChangedPaths: len(changed), ChangedPathList: changed}})
	if err := writeJSON(*outPath, verification); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("decision=%s resolution=%s reason=%s\n", verification.Decision, verification.Resolution, verification.Reason)
	if verification.Decision != "PASS" {
		os.Exit(1)
	}
}

func loadJSON[T any](filename string) (T, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		var zero T
		return zero, err
	}
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		var zero T
		return zero, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			var zero T
			return zero, fmt.Errorf("trailing JSON")
		}
		var zero T
		return zero, err
	}
	return value, nil
}
func writeJSON(filename string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(filename, raw, 0o644)
}
func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("sha256:%x", sum[:])
}
