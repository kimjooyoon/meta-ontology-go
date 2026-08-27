package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	verifier "github.com/kimjooyoon/meta-ontology-go/internal/meta/denominatorevolutionverify"
)

type options struct{ head, contract, report, source, out string }

func run(args []string) int {
	flags := flag.NewFlagSet("denominator-evolution-verify", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "denominator evolution contract")
	flags.StringVar(&value.report, "report", "", "producer report")
	flags.StringVar(&value.source, "source", "", "Gooo denominator governance source")
	flags.StringVar(&value.out, "out", "", "independent verification output")
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
	sum := sha256.Sum256(source)
	verification := verifier.Verify(verifier.Input{ContractRaw: contractRaw, ReportRaw: reportRaw, HeadSHA: value.head, SourceDigest: "sha256:" + hex.EncodeToString(sum[:])})
	if err := verifier.WriteVerification(value.out, verification); err != nil {
		return 2
	}
	fmt.Printf("denominator evolution consumer: %s %s checks=%d\n", verification.Decision, verification.Reason, len(verification.Checks))
	if verification.Decision != "PASS" {
		return 1
	}
	return 0
}
