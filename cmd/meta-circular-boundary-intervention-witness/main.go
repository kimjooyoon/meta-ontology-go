package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundary"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundaryconsumer"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

const interventionSchema = "gooo/meta-circular-boundary-causality/v1"

func main() {
	sourcePath := flag.String("source", producer.ExpectedSourcePath, "Gooo source to intervene")
	headSHA := flag.String("head-sha", "", "exact 40-character commit SHA")
	output := flag.String("output", "", "causality report output path")
	flag.Parse()

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatal(err)
	}
	baseInput := contract.Input{Path: *sourcePath, HeadSHA: *headSHA, Source: source}
	baseline := producer.Evaluate(baseInput)
	if err := consumer.Judge(baseline, baseInput); err != nil {
		fatal(fmt.Errorf("baseline consumer judge: %w", err))
	}

	semanticNeedle := []byte("scope=READ_ONLY|handle=fixture|request_execution=true")
	semanticReplacement := []byte("scope=WRITE|handle=fixture|request_execution=true")
	if bytes.Count(source, semanticNeedle) != 1 {
		fatal(fmt.Errorf("semantic intervention needle count is not one"))
	}
	semanticSource := bytes.Replace(source, semanticNeedle, semanticReplacement, 1)
	nonSemanticSource := append(append([]byte(nil), source...), '\n')
	cases := []contract.CausalityCase{
		intervention("semantic-capability-scope", "SEMANTIC", true, baseline, *sourcePath, *headSHA, semanticSource),
		intervention("trailing-newline", "NON_SEMANTIC", false, baseline, *sourcePath, *headSHA, nonSemanticSource),
	}
	report := contract.CausalityReport{Schema: interventionSchema, Cases: cases}
	report.Summary.Total = len(cases)
	for _, item := range cases {
		if item.Passed {
			report.Summary.Passed++
		}
	}
	if report.Summary.Total > 0 {
		report.Summary.CoverageBPS = report.Summary.Passed * 10_000 / report.Summary.Total
	}
	if report.Summary.Passed != report.Summary.Total {
		fatal(fmt.Errorf("semantic-causality contract failed: %d/%d", report.Summary.Passed, report.Summary.Total))
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if *output == "" {
		_, _ = os.Stdout.Write(encoded)
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("semantic-causality: %d/%d (%d BPS)\n", report.Summary.Passed, report.Summary.Total, report.Summary.CoverageBPS)
}

func intervention(id, kind string, expectedSemanticChange bool, baseline contract.Report, path, head string, source []byte) contract.CausalityCase {
	input := contract.Input{Path: path, HeadSHA: head, Source: source}
	intervened := producer.Evaluate(input)
	consumerAccepted := consumer.Judge(intervened, input) == nil
	semanticChanged := baseline.Source.SemanticDigest != intervened.Source.SemanticDigest
	sourceChanged := baseline.Source.SourceDigest != intervened.Source.SourceDigest
	baselineCase := caseByID(baseline, "explicit-read-only-capability")
	intervenedCase := caseByID(intervened, "explicit-read-only-capability")
	baselineReceipt := baselineCase.Receipt
	intervenedReceipt := intervenedCase.Receipt
	semanticOutputsPreserved := semanticChanged == false &&
		baselineCase.Observation.Decision == intervenedCase.Observation.Decision &&
		baselineCase.Observation.Authorization == intervenedCase.Observation.Authorization &&
		baselineCase.Observation.Execution == intervenedCase.Observation.Execution &&
		baselineReceipt.Decision == intervenedReceipt.Decision &&
		claimAfter(baselineReceipt, "authorization") == claimAfter(intervenedReceipt, "authorization") &&
		claimAfter(baselineReceipt, "execution") == claimAfter(intervenedReceipt, "execution")
	semanticOutputsChanged := baselineCase.Observation.Decision != intervenedCase.Observation.Decision &&
		baselineCase.Observation.Authorization != intervenedCase.Observation.Authorization &&
		baselineCase.Observation.Execution != intervenedCase.Observation.Execution &&
		baselineReceipt.Decision != intervenedReceipt.Decision &&
		baselineReceipt.CapabilityDigest != intervenedReceipt.CapabilityDigest &&
		claimAfter(baselineReceipt, "authorization") != claimAfter(intervenedReceipt, "authorization") &&
		claimAfter(baselineReceipt, "execution") != claimAfter(intervenedReceipt, "execution")
	passed := consumerAccepted && sourceChanged && semanticChanged == expectedSemanticChange
	if expectedSemanticChange {
		passed = passed && semanticOutputsChanged && !semanticOutputsPreserved
	} else {
		passed = passed && semanticOutputsPreserved
	}
	return contract.CausalityCase{
		ID: id, Kind: kind,
		BaselineSourceDigest: baseline.Source.SourceDigest, IntervenedSourceDigest: intervened.Source.SourceDigest,
		BaselineSemanticDigest: baseline.Source.SemanticDigest, IntervenedSemanticDigest: intervened.Source.SemanticDigest,
		SourceChanged: sourceChanged, SemanticChanged: semanticChanged, ExpectedSemanticChange: expectedSemanticChange,
		ConsumerAccepted:     consumerAccepted,
		BaselineCaseDecision: baselineCase.Observation.Decision, IntervenedCaseDecision: intervenedCase.Observation.Decision,
		BaselineAuthorization: baselineCase.Observation.Authorization, IntervenedAuthorization: intervenedCase.Observation.Authorization,
		BaselineExecution: baselineCase.Observation.Execution, IntervenedExecution: intervenedCase.Observation.Execution,
		BaselineReceiptDecision: baselineReceipt.Decision, IntervenedReceiptDecision: intervenedReceipt.Decision,
		BaselineCapabilityDigest: baselineReceipt.CapabilityDigest, IntervenedCapabilityDigest: intervenedReceipt.CapabilityDigest,
		BaselineAuthorizationClaim: claimAfter(baselineReceipt, "authorization"), IntervenedAuthorizationClaim: claimAfter(intervenedReceipt, "authorization"),
		BaselineExecutionClaim: claimAfter(baselineReceipt, "execution"), IntervenedExecutionClaim: claimAfter(intervenedReceipt, "execution"),
		SemanticOutputsPreserved: semanticOutputsPreserved, Passed: passed,
	}
}

func caseByID(report contract.Report, id string) contract.CaseResult {
	for _, item := range report.Cases {
		if item.Definition.ID == id {
			return item
		}
	}
	fatal(fmt.Errorf("missing causality case %q", id))
	return contract.CaseResult{}
}

func claimAfter(receipt contract.Receipt, suffix string) string {
	for _, transition := range receipt.ClaimTransitions {
		if transition.ClaimID == receipt.CaseID+"."+suffix {
			return transition.After
		}
	}
	fatal(fmt.Errorf("missing %s claim in %q receipt", suffix, receipt.CaseID))
	return ""
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
