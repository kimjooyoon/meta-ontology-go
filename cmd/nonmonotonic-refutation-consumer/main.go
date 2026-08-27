package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/nonmonotonicrefutationoracle"
)

type receipt struct {
	Schema                string                              `json:"schema"`
	ProducerReceiptDigest string                              `json:"producer_receipt_digest"`
	Report                nonmonotonicrefutationoracle.Report `json:"report"`
	ReplayReportDigest    string                              `json:"replay_report_digest"`
	ReplayVerified        bool                                `json:"replay_verified"`
	RepositoryWrites      int                                 `json:"repository_writes"`
	MutationAuthority     bool                                `json:"mutation_authority"`
	ReceiptDigest         string                              `json:"receipt_digest"`
}

func main() {
	producerPath := flag.String("producer", "", "producer receipt")
	sourcePath := flag.String("source", "", "canonical .gooo source")
	outputPath := flag.String("output", "", "consumer receipt output")
	flag.Parse()
	if *producerPath == "" || *sourcePath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "-producer, -source, and -output are required")
		os.Exit(2)
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
	output := receipt{
		Schema: "gooo/meta-nonmonotonic-refutation-receipt/v1", ProducerReceiptDigest: digestBytes(producer),
		Report: report, ReplayReportDigest: replay.ReportDigest, ReplayVerified: report.ReportDigest == replay.ReportDigest,
		RepositoryWrites: 0, MutationAuthority: false,
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
