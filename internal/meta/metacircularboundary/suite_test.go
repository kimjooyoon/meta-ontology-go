package metacircularboundary

import (
	"bytes"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundaryconsumer"
)

func TestBoundarySuite(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	head := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	report := Evaluate(Input{Path: ExpectedSourcePath, HeadSHA: head, Source: source})
	if err := metacircularboundaryconsumer.Judge(report, Input{Path: ExpectedSourcePath, HeadSHA: head, Source: source}); err != nil {
		t.Fatal(err)
	}
	if report.Summary.CasesTotal != 4 || report.Summary.CasesPassed != 4 || report.Summary.ExplicitAuthorizations != 1 || report.Summary.AllowedExecutions != 1 || report.Summary.DescriptionEscalationPaths != 0 || report.Summary.RepositoryWrites != 0 || report.Summary.MutationAuthority != 0 {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
}

func TestDescriptionCannotMintAuthority(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	report := Evaluate(Input{Path: ExpectedSourcePath, HeadSHA: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Source: source})
	caseResult := report.Cases[0]
	if caseResult.Observation.Authorization != AuthorizationDenied || caseResult.Observation.Execution != ExecutionBlocked || caseResult.Observation.Reason != ReasonDescriptionOnly {
		t.Fatalf("description gained authority: %+v", caseResult.Observation)
	}
}

func TestForgedAndOutOfScopeCapabilitiesAreBlocked(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	report := Evaluate(Input{Path: ExpectedSourcePath, HeadSHA: "cccccccccccccccccccccccccccccccccccccccc", Source: source})
	for _, index := range []int{2, 3} {
		if report.Cases[index].Observation.Authorization != AuthorizationDenied || report.Cases[index].Observation.Execution != ExecutionBlocked {
			t.Fatalf("case %d crossed boundary: %+v", index, report.Cases[index].Observation)
		}
	}
}

func TestConsumerRejectsSharedWrongCaseFactCounterexample(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	input := Input{Path: ExpectedSourcePath, HeadSHA: "dddddddddddddddddddddddddddddddddddddddd", Source: source}
	report := Evaluate(input)
	tampered := report
	tampered.Cases = append([]CaseResult(nil), report.Cases...)
	// A coupled producer/consumer case-fact bug could silently change this
	// request from true to false: the description-only observation remains
	// blocked, so a shared wrong input would still pass every outcome check.
	tampered.Cases[0].Attempt.RequestExecution = false
	tampered.ReportDigest = sealReport(tampered).ReportDigest
	if err := metacircularboundaryconsumer.Judge(tampered, input); err == nil {
		t.Fatal("consumer accepted a report with a shared wrong case fact")
	}
}

func TestExpectedDecisionCannotMintAuthorization(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	input := Input{Path: ExpectedSourcePath, HeadSHA: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", Source: source}
	report := Evaluate(input)
	definition := report.Cases[1].Definition
	definition.ExpectedDecision = DecisionPass
	observation := report.Cases[1].Observation
	observation.Decision = DecisionFailClosed
	observation.Authorization = AuthorizationDenied
	observation.Execution = ExecutionBlocked
	observation.Reason = ReasonForgedCapability
	receipt := buildReceipt(report.Source, definition, report.Cases[1].Attempt, observation)
	if receipt.Decision == definition.ExpectedDecision || receipt.Decision != DecisionFailClosed || receipt.Authorization != AuthorizationDenied {
		t.Fatalf("ExpectedDecision minted authorization: definition=%+v receipt=%+v", definition, receipt)
	}
}

func TestUnknownComputationDataRemainsOpen(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	mutated := bytes.Replace(source,
		[]byte("id=description-only|description=source|capability=none|request_execution=true"),
		[]byte("id=description-only|description=source|capability=none|request_execution=unknown"), 1)
	if bytes.Equal(mutated, source) {
		t.Fatal("unknown computation intervention did not apply")
	}
	input := Input{Path: ExpectedSourcePath, HeadSHA: "ffffffffffffffffffffffffffffffffffffffff", Source: mutated}
	report := Evaluate(input)
	if report.Decision != DecisionOpen || report.Resolution != ResolutionLower || report.Reason != ReasonCaseDataUnknown || report.Coordinate != (Coordinate{Stage: "PARSE_COMPUTES", Step: "read-case-facts", Reason: ReasonCaseDataUnknown}) {
		t.Fatalf("unknown computation was not lower-resolution open: %+v", report)
	}
	if report.Cases[0].Observation.Decision != DecisionOpen || report.Cases[0].Observation.Reason != ReasonCaseDataUnknown {
		t.Fatalf("unknown case fact was not open: %+v", report.Cases[0])
	}
	if err := metacircularboundaryconsumer.Judge(report, input); err != nil {
		t.Fatalf("independent consumer rejected valid open report: %v", err)
	}
}

func TestContradictoryCapabilityEvidenceIsRefuted(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	mutated := bytes.Replace(source,
		[]byte("scope=READ_ONLY|handle=fixture|request_execution=true"),
		[]byte("scope=READ_ONLY|scope=WRITE|handle=fixture|request_execution=true"), 1)
	if bytes.Equal(mutated, source) {
		t.Fatal("contradictory capability intervention did not apply")
	}
	input := Input{Path: ExpectedSourcePath, HeadSHA: "1212121212121212121212121212121212121212", Source: mutated}
	report := Evaluate(input)
	caseResult := report.Cases[1]
	if report.Decision != DecisionRefuted || report.Resolution != ResolutionExact || report.Reason != ReasonContradictory || caseResult.Observation.Decision != DecisionRefuted || caseResult.Receipt.Decision != DecisionRefuted {
		t.Fatalf("contradictory evidence was not refuted: report=%+v case=%+v", report, caseResult)
	}
	if caseResult.Receipt.ClaimTransitions[1].After != "REFUTED" || caseResult.Receipt.ClaimTransitions[2].After != "REFUTED" {
		t.Fatalf("contradictory claim state did not persist: %+v", caseResult.Receipt.ClaimTransitions)
	}
	if err := metacircularboundaryconsumer.Judge(report, input); err != nil {
		t.Fatalf("independent consumer rejected valid refutation: %v", err)
	}
}
