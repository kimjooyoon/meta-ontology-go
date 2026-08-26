package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/symbolicinvocationusecase/claimledger"
)

func main() {
	contractPath := flag.String("contract", "", "claim ledger contract")
	observationPath := flag.String("observation", "", "reader observation JSON")
	subject := flag.String("subject", "", "exact subject commit")
	outputPath := flag.String("out", "", "claim ledger output")
	flag.Parse()
	if *contractPath == "" || *observationPath == "" || *subject == "" || *outputPath == "" {
		fail("contract, observation, subject, and out are required")
	}
	contractData, err := os.ReadFile(*contractPath)
	if err != nil {
		fail("read contract: %v", err)
	}
	observationData, err := os.ReadFile(*observationPath)
	if err != nil {
		fail("read observation: %v", err)
	}
	report, err := claimledger.Project(contractData, observationData, *subject)
	if err != nil {
		fail("project claim ledger: %v", err)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fail("encode claim ledger: %v", err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(*outputPath, encoded, 0o644); err != nil {
		fail("write claim ledger: %v", err)
	}
	fmt.Printf(
		"claim-ledger: conformance=%s claim-set=%s resolution=%s discharged=%d/%d unknown=%d excluded=%d false-promotions=%d\n",
		report.Conformance.Decision, report.ClaimSet.Decision, report.ClaimSet.Resolution,
		report.Metrics.DischargedTotal, report.Metrics.InScopeClaimTotal,
		report.Metrics.UnknownTotal, report.Metrics.ExcludedTotal, report.Metrics.FalsePromotionCount,
	)
	if report.Conformance.Decision != "PASS" {
		fail("fixed metric contract did not match")
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "claim-ledger: "+format+"\n", args...)
	os.Exit(1)
}
