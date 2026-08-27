package metacircularboundary

import (
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

func TestConsumerRejectsSharedWrongCaseInputCounterexample(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	input := Input{Path: ExpectedSourcePath, HeadSHA: "dddddddddddddddddddddddddddddddddddddddd", Source: source}
	report := Evaluate(input)
	tampered := report
	tampered.Cases = append([]CaseResult(nil), report.Cases...)
	// A coupled producer/consumer CaseInput bug could accidentally give the
	// description-only case the explicit capability from case two.
	tampered.Cases[0].Attempt = report.Cases[1].Attempt
	tampered.ReportDigest = sealReport(tampered).ReportDigest
	if err := metacircularboundaryconsumer.Judge(tampered, input); err == nil {
		t.Fatal("consumer accepted a report with a shared wrong CaseInput")
	}
}
