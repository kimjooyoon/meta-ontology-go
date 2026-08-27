package main

import (
	"encoding/json"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceiptconsumer"
)

type Report struct {
	Schema             string                      `json:"schema"`
	CaseID             string                      `json:"case_id"`
	SubjectSHA         string                      `json:"subject_sha"`
	SourcePaths        []string                    `json:"source_paths"`
	OutputPath         string                      `json:"output_path"`
	Receipt            producer.Receipt            `json:"receipt"`
	IndependentVerdict consumer.Verdict            `json:"independent_verdict"`
	Indicators         []producer.OperationBinding `json:"indicators"`
	RepositoryWrites   int                         `json:"repository_writes"`
	Evidence           Evidence                    `json:"evidence"`
	ReportDigest       string                      `json:"report_digest"`
}

type Evidence struct {
	RawBeforeDigest      string `json:"raw_before_digest"`
	RawAfterDigest       string `json:"raw_after_digest"`
	SemanticBeforeDigest string `json:"semantic_before_digest"`
	SemanticAfterDigest  string `json:"semantic_after_digest"`
	DistinctPropositions int    `json:"distinct_propositions"`
	Added                int    `json:"added"`
	Removed              int    `json:"removed"`
	Changed              int    `json:"changed"`
	Open                 int    `json:"open"`
	Discharged           int    `json:"discharged"`
	Refuted              int    `json:"refuted"`
	TransitionChain      int    `json:"transition_chain"`
	Decision             string `json:"decision"`
	Resolution           string `json:"resolution"`
	Classification       string `json:"classification"`
	Stage                string `json:"stage"`
	Step                 string `json:"step"`
	Reason               string `json:"reason"`
}

func evaluate(input producer.Input, outputPath string) Report {
	receipt, err := producer.ProduceFiles(input)
	if err != nil {
		return Report{Schema: producer.ReportSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, SourcePaths: []string{input.BeforePath, input.AfterPath}, OutputPath: outputPath, RepositoryWrites: 0, ReportDigest: "read-error"}
	}
	raw, _ := json.Marshal(receipt)
	wire := consumer.Receipt{}
	_ = json.Unmarshal(raw, &wire)
	verdict := consumer.AdjudicateFiles(consumer.Input{CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, BeforePath: input.BeforePath, AfterPath: input.AfterPath}, wire)
	summary := summaryFor(receipt, verdict)
	report := Report{Schema: producer.ReportSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, SourcePaths: []string{input.BeforePath, input.AfterPath}, OutputPath: outputPath, Receipt: receipt, IndependentVerdict: verdict, Indicators: bindings(summary), RepositoryWrites: 0, Evidence: evidenceFor(receipt, verdict, summary)}
	sealReport(&report)
	return report
}

func sealReport(report *Report) {
	copy := *report
	copy.ReportDigest = ""
	report.ReportDigest = digestValue(copy)
}

func summaryFor(receipt producer.Receipt, verdict consumer.Verdict) Summary {
	result := Summary{CasesTotal: 1, CasesPassed: boolInt(verdict.Passed), AdjudicatedCases: boolInt(verdict.Passed), RepositoryWrites: receipt.RepositoryWrites, ModeledSemanticComponents: receipt.ModeledSemanticComponents, TotalSemanticComponents: receipt.TotalSemanticComponents,
		TextualChanges:         boolInt(receipt.TextualDelta.Changed),
		StructuralObservations: boolInt(receipt.StructuralDelta.Status != ""),
		ClaimTransitionCases:   boolInt(len(receipt.ClaimTransitions) > 0),
		AddedClaims:            len(receipt.SemanticClaimDelta.Added),
		RemovedClaims:          len(receipt.SemanticClaimDelta.Removed),
		ChangedClaims:          len(receipt.SemanticClaimDelta.Changed),
		TransitionChains:       len(receipt.ClaimTransitions),
		AmbiguousCases:         boolInt(len(receipt.SemanticClaimDelta.Ambiguous) > 0)}
	seen := make(map[string]bool)
	for _, claim := range receipt.ClaimLedger {
		if claim.PropositionDigest != "" {
			seen[claim.PropositionDigest] = true
		}
		switch claim.Status {
		case producer.StatusOpen:
			result.OpenClaims++
		case producer.StatusDischarged:
			result.DischargedClaims++
		case producer.StatusRefuted:
			result.RefutedClaims++
		}
	}
	result.DistinctPropositions = len(seen)
	switch receipt.Classification {
	case producer.ClassPreserved:
		result.SemanticPreserved = 1
	case producer.ClassChanged:
		result.SemanticChanged = 1
	case producer.ClassIndeterminate:
		result.Indeterminate, result.UnknownPaths = 1, boolInt(receipt.Resolution == producer.ResolutionLower)
	}
	return result
}

func evidenceFor(receipt producer.Receipt, verdict consumer.Verdict, summary Summary) Evidence {
	return Evidence{RawBeforeDigest: receipt.Before.SourceDigest, RawAfterDigest: receipt.After.SourceDigest, SemanticBeforeDigest: receipt.Before.SemanticDigest, SemanticAfterDigest: receipt.After.SemanticDigest, DistinctPropositions: summary.DistinctPropositions, Added: summary.AddedClaims, Removed: summary.RemovedClaims, Changed: summary.ChangedClaims, Open: summary.OpenClaims, Discharged: summary.DischargedClaims, Refuted: summary.RefutedClaims, TransitionChain: summary.TransitionChains, Decision: verdict.Decision, Resolution: verdict.Resolution, Classification: verdict.Classification, Stage: verdict.Stage, Step: verdict.Step, Reason: verdict.Reason}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
