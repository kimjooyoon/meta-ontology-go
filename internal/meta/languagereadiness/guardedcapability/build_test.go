package guardedcapability

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"
)

func TestExactCapabilitySeparatesImplementationFromEvent(t *testing.T) {
	receipt := Build(exactSource(t))
	if err := ValidateForHead(receipt, receipt.Source.CurrentHeadSHA); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionPass || receipt.Summary.Satisfied != 8 ||
		receipt.Summary.Total != 8 || receipt.Summary.ReadinessBPS != 10000 {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func TestUnknownAncestryLowersResolution(t *testing.T) {
	source := exactSource(t)
	source.AncestryObserved = false
	receipt := Build(source)
	if receipt.Decision != DecisionFailClosed || receipt.Reason != ReasonUnknown ||
		receipt.Resolution != ResolutionLower || receipt.Summary.Unresolved != 1 {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func TestImplementationDriftRejectsCapability(t *testing.T) {
	source := exactSource(t)
	source.CurrentGuardTree = "tree-drift"
	receipt := Build(source)
	if receipt.Decision != DecisionFailClosed || receipt.Reason != ReasonRejected ||
		receipt.Summary.NotSatisfied != 1 {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func TestEmbeddedFoundationReplaysBySection(t *testing.T) {
	report, err := foundationReport()
	if err != nil {
		t.Fatal(err)
	}
	replay := guardedpromotion.Build(report.Source)
	sections := []struct {
		name       string
		foundation any
		replayed   any
	}{
		{name: "header", foundation: []string{report.Schema, report.Decision, report.Reason, report.Resolution}, replayed: []string{replay.Schema, replay.Decision, replay.Reason, replay.Resolution}},
		{name: "summary", foundation: report.Summary, replayed: replay.Summary},
		{name: "coordinates", foundation: report.Coordinates, replayed: replay.Coordinates},
		{name: "indicators", foundation: report.Indicators, replayed: replay.Indicators},
		{name: "proofs", foundation: report.Proofs, replayed: replay.Proofs},
		{name: "digest", foundation: report.ReportDigest, replayed: replay.ReportDigest},
	}
	for _, section := range sections {
		if !reflect.DeepEqual(section.foundation, section.replayed) {
			t.Fatalf("foundation %s does not replay: foundation=%#v replayed=%#v", section.name, section.foundation, section.replayed)
		}
	}
}
