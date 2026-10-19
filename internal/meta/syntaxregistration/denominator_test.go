package syntaxregistration

import (
	"encoding/json"
	"testing"
)

func TestGenerateDenominatorPreservesExpandedLinkBaseline(t *testing.T) {
	baseline := denominator{
		Schema:        "gooo/vertical-slice-boundary-denominator/v1",
		DenominatorID: denominatorID(47), Version: 47,
		Boundaries: []boundary{
			{ID: "syntax", Target: 79, LinkTarget: 1},
			{ID: "semantics", LinkTarget: 2},
			{ID: "binding", LinkTarget: 2},
			{ID: "use-cases", LinkTarget: 1},
			{ID: "toolchain", LinkTarget: 3},
			{ID: "release", LinkTarget: 4},
		},
	}
	raw, err := json.Marshal(baseline)
	if err != nil {
		t.Fatal(err)
	}
	record := AppendRecord{BaseVersion: 47, BaseTarget: 79, CorpusTarget: 79,
		CaseID: "expanded-link-fixture", Path: "examples/expanded-link/main.gooo",
		SourceDigest: digest([]byte("fixture source")), BaselineDigest: digest(raw),
		NextVersion: 48, NextTarget: 80, Accepted: true}
	generated, err := generateDenominator(raw, record)
	if err != nil {
		t.Fatalf("expanded baseline was rejected: %v", err)
	}
	var observed denominator
	if err := json.Unmarshal(generated, &observed); err != nil {
		t.Fatal(err)
	}
	links := 0
	for _, item := range observed.Boundaries {
		links += item.LinkTarget
	}
	if observed.Version != 48 || observed.DenominatorID != denominatorID(48) ||
		observed.Boundaries[0].Target != 80 || links != 13 {
		t.Fatalf("generation failed to preserve expanded baseline: %#v", observed)
	}
}
