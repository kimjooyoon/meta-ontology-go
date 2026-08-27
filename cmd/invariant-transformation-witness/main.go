package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/judge"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/producer"
)

func main() {
	sourcePath := flag.String("source", "", "Gooo source value contract")
	contractPath := flag.String("contract", "", "fixed invariant transformation contract")
	headSHA := flag.String("head-sha", "", "exact subject commit")
	outputPath := flag.String("output", "", "suite report output")
	receiptDir := flag.String("receipt-dir", "", "individual receipt output directory")
	check := flag.Bool("check", false, "run the independent judge and validate the report")
	flag.Parse()
	if *sourcePath == "" || *contractPath == "" || *headSHA == "" || *outputPath == "" || *receiptDir == "" {
		fail("-source, -contract, -head-sha, -output, and -receipt-dir are required")
	}

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	contractRaw, err := os.ReadFile(*contractPath)
	if err != nil {
		fail(err.Error())
	}
	var contract model.Contract
	if err := json.Unmarshal(contractRaw, &contract); err != nil {
		fail(fmt.Sprintf("decode contract: %v", err))
	}
	if err := model.ValidateContract(contract); err != nil {
		fail(err.Error())
	}

	report := buildReport(source, *headSHA, contract)
	if *check {
		if err := judge.ValidateReport(report, source); err != nil {
			fail(err.Error())
		}
	}
	if err := writeReceipts(*receiptDir, report.Cases); err != nil {
		fail(err.Error())
	}
	if err := writeJSON(*outputPath, report); err != nil {
		fail(err.Error())
	}
	fmt.Printf("invariant transformation: %s %d/%d authorized=%d refuted=%d open=%d claims=%d/%d writes=%d approved-effects=%d\n",
		report.Decision, report.Summary.CasesSatisfied, report.Summary.CasesTotal, report.Summary.AuthorizedCases,
		report.Summary.RefutedCases, report.Summary.OpenCases, report.Summary.DischargedClaims,
		report.Summary.ClaimsTotal, report.Summary.RepositoryWrites, report.Summary.ApprovedArtifactEffects)
}

func buildReport(source []byte, headSHA string, contract model.Contract) model.Report {
	report := model.Report{Schema: model.ReportSchema, HeadSHA: headSHA, SourcePath: model.SourcePath,
		SourceDigest: model.DigestBytes(source), ContractDigest: model.Digest(contract), DenominatorID: contract.Denominator,
		DenominatorTotal: len(contract.Cases), Decision: model.DecisionPass, Resolution: model.ResolutionExact,
		Reason: "INVARIANT_TRANSFORMATION_SUITE_SATISFIED", Cases: []model.CaseResult{}, NotClaimed: []string{
			"full semantic equivalence for arbitrary programs", "complete transformation verification",
			"toolchain correctness", "repository mutation or promotion authority",
		}}
	for _, spec := range contract.Cases {
		receipt, err := producer.Build(source, headSHA, spec.ID)
		if err != nil {
			fail(err.Error())
		}
		judgment := judge.Judge(receipt, source)
		satisfied := judgment.Independent && judgment.Decision == spec.ExpectedDecision && judgment.Resolution == spec.ExpectedResolution &&
			judgment.Reason == spec.ExpectedReason && judgment.Status == spec.ExpectedStatus && len(receipt.Effects) == spec.ExpectedEffects
		report.Cases = append(report.Cases, model.CaseResult{Spec: spec, Receipt: receipt, Judgment: judgment, Satisfied: satisfied})
		accumulate(&report.Summary, receipt, judgment, satisfied)
	}
	if report.Summary.CasesTotal != 0 {
		report.Summary.CoverageBPS = report.Summary.CasesSatisfied * 10_000 / report.Summary.CasesTotal
	}
	report.Indicators = judge.Indicators(report.Summary)
	return model.SealReport(report)
}

func accumulate(summary *model.Summary, receipt model.Receipt, judgment model.Judgment, satisfied bool) {
	summary.CasesTotal++
	if satisfied {
		summary.CasesSatisfied++
	}
	summary.ClaimsTotal += judgment.CheckedClaims
	summary.DischargedClaims += judgment.DischargedClaims
	summary.RefutedClaims += judgment.RefutedClaims
	summary.OpenClaims += judgment.OpenClaims
	summary.TransitionEvents += len(receipt.Claims)
	summary.ApprovedArtifactEffects += len(receipt.Effects)
	summary.RepositoryWrites += receipt.RepositoryWrites
	if receipt.MutationAuthority {
		summary.MutationAuthority++
	}
	switch judgment.Decision {
	case model.DecisionAllowed:
		summary.AuthorizedCases++
	case model.DecisionRefuted:
		summary.RefutedCases++
	case model.DecisionBlocked:
		summary.OpenCases++
	}
}

func writeReceipts(directory string, cases []model.CaseResult) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	for _, item := range cases {
		if err := writeJSON(filepath.Join(directory, item.Spec.ID+".json"), item.Receipt); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
