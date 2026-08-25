package languagepackageexecution

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"
)

func TestEvaluateFixedPackageCases(t *testing.T) {
	report := Evaluate(testInput(t))
	if report.Decision != "PASS" || report.Summary.CasesSatisfied != 5 || report.Summary.CasesTotal != 5 {
		t.Fatalf("decision=%s cases=%d/%d", report.Decision, report.Summary.CasesSatisfied, report.Summary.CasesTotal)
	}
	if report.Views[0].FactsDigest != report.Views[2].FactsDigest || len(report.Views[0].VisibleFacts) >= len(report.Views[2].VisibleFacts) {
		t.Fatal("reader views must change resolution without changing facts")
	}
}

func TestEvaluateUnknownDecisionLowersResolution(t *testing.T) {
	input := testInput(t)
	input.Cases[0].Receipt.Decision = "MYSTERY"
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Reason != "PACKAGE_EXECUTION_DECISION_UNKNOWN" || report.Resolution != "LOWER_RESOLUTION" {
		t.Fatalf("decision=%s reason=%s resolution=%s", report.Decision, report.Reason, report.Resolution)
	}
}

func testInput(t *testing.T) Input {
	t.Helper()
	sources, err := packageexecution.LoadDirectory(filepath.Join("..", "..", "..", "examples", "billing-package"))
	if err != nil {
		t.Fatal(err)
	}
	request := packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources}
	positive := packageexecution.Execute(request)
	mismatch := append([]packageexecution.Source(nil), sources...)
	mismatch[0].Content = strings.Replace(mismatch[0].Content, "package billing", "package other", 1)
	duplicate := append(append([]packageexecution.Source(nil), sources...), packageexecution.Source{Filename: "duplicate.gooo", Content: sources[1].Content})
	return Input{HeadSHA: strings.Repeat("a", 40), Contract: CanonicalContract(), Cases: []CaseEvidence{
		{ID: "positive-package-execution", Receipt: positive},
		{ID: "deterministic-replay", Receipt: packageexecution.Execute(request)},
		{ID: "header-mismatch-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: mismatch})},
		{ID: "duplicate-declaration-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: duplicate})},
		{ID: "source-count-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources[:1]})},
	}}
}
