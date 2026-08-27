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
	Receipt            producer.Receipt            `json:"receipt"`
	IndependentVerdict consumer.Verdict            `json:"independent_verdict"`
	Indicators         []producer.OperationBinding `json:"indicators"`
	RepositoryWrites   int                         `json:"repository_writes"`
	ReportDigest       string                      `json:"report_digest"`
}

func evaluate(input producer.Input) Report {
	receipt, err := producer.ProduceFiles(input)
	if err != nil {
		return Report{Schema: producer.ReportSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, RepositoryWrites: 0, ReportDigest: "read-error"}
	}
	raw, _ := json.Marshal(receipt)
	wire := consumer.Receipt{}
	_ = json.Unmarshal(raw, &wire)
	verdict := consumer.AdjudicateFiles(consumer.Input{CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, BeforePath: input.BeforePath, AfterPath: input.AfterPath}, wire)
	summary := summaryFor(receipt, verdict)
	report := Report{Schema: producer.ReportSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, Receipt: receipt, IndependentVerdict: verdict, Indicators: bindings(summary), RepositoryWrites: 0}
	sealReport(&report)
	return report
}

func sealReport(report *Report) {
	copy := *report
	copy.ReportDigest = ""
	report.ReportDigest = digestValue(copy)
}

func summaryFor(receipt producer.Receipt, verdict consumer.Verdict) Summary {
	result := Summary{CasesTotal: 1, CasesPassed: boolInt(verdict.Passed), AdjudicatedCases: boolInt(verdict.Passed), RepositoryWrites: receipt.RepositoryWrites}
	result.TextualChanges = boolInt(receipt.TextualDelta.Changed)
	result.StructuralObservations = boolInt(receipt.StructuralDelta.Status != "")
	result.ClaimTransitionCases = boolInt(len(receipt.ClaimTransitions) > 0)
	switch receipt.Classification {
	case producer.ClassPreserved:
		result.SemanticPreserved = 1
	case producer.ClassChanged:
		result.SemanticChanged = 1
	case producer.ClassIndeterminate:
		result.Indeterminate, result.UnknownPaths = 1, boolInt(receipt.Resolution == producer.ResolutionUnknown)
	}
	return result
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
