package valuecatalog

import (
	"strings"
	"testing"
	"testing/fstest"
)

const catalogFixture = `package operationcatalog
namespace operationcatalog
entity Integer id "gooo://operation-catalog/entity/integer"
activity IncrementOne(Integer) -> Integer computes "int.add:1"
activity IncrementTwo(Integer) -> Integer
`

func TestBaselineReportsGreenConformanceWithoutClaimingExtension(t *testing.T) {
	head := strings.Repeat("a", 40)
	report := Evaluate(fstest.MapFS{"main.gooo": {Data: []byte(catalogFixture)}}, "main.gooo", head)
	if err := Validate(report, head); err != nil {
		t.Fatal(err)
	}
	if report.Improvement.After != coordinate(0, 1) || report.Views[0].Satisfied != 3 {
		t.Fatalf("baseline overclaimed extension: %#v", report)
	}
	if report.OperationSpecMetrics.VerifiedTotal != 9 || report.OperationSpecMetrics.OpenClaims != 0 {
		t.Fatalf("baseline operation spec is not exact: %#v", report.OperationSpecMetrics)
	}
}

func TestSourceOnlyProgramClosesTheFixedExtensionCoordinate(t *testing.T) {
	source := strings.Replace(catalogFixture, extensionDeclaration, extensionProgramLine, 1)
	head := strings.Repeat("b", 40)
	report := Evaluate(fstest.MapFS{"main.gooo": {Data: []byte(source)}}, "main.gooo", head)
	if err := Validate(report, head); err != nil {
		t.Fatal(err)
	}
	if report.Improvement.After != coordinate(1, 1) || report.Extension.Cases[2].Actual != 43 {
		t.Fatalf("extension evidence is not exact: %#v", report)
	}
	if report.OperationSpecMetrics != (OperationSpecMetrics{MetricID: OperationSpecMetricID, FixedAxisTotal: 9, VerifiedTotal: 9, CoverageBasisPoints: 10_000, DischargedClaims: 9}) {
		t.Fatalf("OS9 evidence is not exact: %#v", report.OperationSpecMetrics)
	}
}

func TestUnknownSourceOnlyProgramFailsClosedAtSyntaxResolution(t *testing.T) {
	source := strings.Replace(catalogFixture, extensionDeclaration, extensionDeclaration+` computes "int.magic:2"`, 1)
	report := Evaluate(fstest.MapFS{"main.gooo": {Data: []byte(source)}}, "main.gooo", strings.Repeat("c", 40))
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionSyntaxOnly || report.Reason != ReasonObservationFailed {
		t.Fatalf("unknown program did not fail closed: %#v", report)
	}
	if report.ProcessCoordinate.Stage != "RESOLVE" || report.ProcessCoordinate.Step != "resolve-operation-spec" || report.ProcessCoordinate.Reason != "VALUE_PROGRAM_UNKNOWN" {
		t.Fatalf("unknown process coordinate is not exact: %#v", report.ProcessCoordinate)
	}
	if report.OperationSpecMetrics.UnknownPathCount != 1 || report.OperationSpecMetrics.OpenClaims != 9 || report.OperationSpecMetrics.DischargedClaims != 0 {
		t.Fatalf("unknown claims were hidden: %#v", report.OperationSpecMetrics)
	}
}
