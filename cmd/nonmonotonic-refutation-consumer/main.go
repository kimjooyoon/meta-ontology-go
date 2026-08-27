package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/nonmonotonicrefutationoracle"
)

type receipt struct {
	Schema                       string                              `json:"schema"`
	ProducerReceiptDigest        string                              `json:"producer_receipt_digest"`
	Report                       nonmonotonicrefutationoracle.Report `json:"report"`
	ReplayReportDigest           string                              `json:"replay_report_digest"`
	ReplayVerified               bool                                `json:"replay_verified"`
	NetRepositoryStatusUnchanged bool                                `json:"net_repository_status_unchanged"`
	RepositoryWriteObservation   string                              `json:"repository_write_observation"`
	MutationAuthorityResolution  string                              `json:"mutation_authority_resolution"`
	PromotionOperationsObserved  int                                 `json:"promotion_operations_observed"`
	ReceiptDigest                string                              `json:"receipt_digest"`
}

func main() {
	producerPath := flag.String("producer", "", "producer receipt")
	sourcePath := flag.String("source", "", "canonical .gooo source")
	outputPath := flag.String("output", "", "consumer receipt output")
	repositoryRoot := flag.String("repository-root", ".", "repository root for before/after write observation")
	flag.Parse()
	if *producerPath == "" || *sourcePath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "-producer, -source, and -output are required")
		os.Exit(2)
	}
	before, err := repositorySnapshot(*repositoryRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	producer, err := os.ReadFile(*producerPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	report, err := nonmonotonicrefutationoracle.Judge(producer, source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	replay, err := nonmonotonicrefutationoracle.Judge(producer, source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	after, err := repositorySnapshot(*repositoryRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	netRepositoryStatusUnchanged := bytes.Equal(before, after)
	output := receipt{
		Schema: "gooo/meta-nonmonotonic-refutation-receipt/v4", ProducerReceiptDigest: digestBytes(producer),
		Report: report, ReplayReportDigest: replay.ReportDigest, ReplayVerified: report.ReportDigest == replay.ReportDigest,
		NetRepositoryStatusUnchanged: netRepositoryStatusUnchanged,
		RepositoryWriteObservation:   repositoryStatusObservation(netRepositoryStatusUnchanged),
		MutationAuthorityResolution:  "UNKNOWN", PromotionOperationsObserved: 0,
	}
	output.ReceiptDigest = digestJSON(output)
	file, err := os.Create(*outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		_ = file.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := file.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("consumer receipt: decision=%s transitions=%d\n", report.Decision, report.Metrics.TransitionTotal)
	if report.Decision != "PASS" || !output.ReplayVerified {
		os.Exit(1)
	}
}

func repositorySnapshot(root string) ([]byte, error) {
	return exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all").Output()
}

func repositoryStatusObservation(unchanged bool) string {
	if unchanged {
		return "NONE_OBSERVED_IN_NET_STATUS"
	}
	return "NET_STATUS_CHANGED"
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value receipt) string {
	value.ReceiptDigest = ""
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(data)
}
