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
	if report.Improvement.After != coordinate(0, 1) || report.Views[0].Satisfied != 1 {
		t.Fatalf("baseline overclaimed extension: %#v", report)
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
}

func TestUnknownSourceOnlyProgramFailsClosedAtSyntaxResolution(t *testing.T) {
	source := strings.Replace(catalogFixture, extensionDeclaration, extensionDeclaration+` computes "int.magic:2"`, 1)
	report := Evaluate(fstest.MapFS{"main.gooo": {Data: []byte(source)}}, "main.gooo", strings.Repeat("c", 40))
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionSyntaxOnly || report.Reason != ReasonObservationFailed {
		t.Fatalf("unknown program did not fail closed: %#v", report)
	}
}
