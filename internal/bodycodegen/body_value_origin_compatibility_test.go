package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestValueOriginAbsentMetadataAndChangedClaim5179(t *testing.T) {
	const source = `package retained5179
namespace retained5179
entity Integer id "urn:gooo:type:integer"
activity Adjust(Integer) -> Integer id "urn:retained5179:adjust" computes "let adjusted = __GOOO_BODY_HOLE_delta__; return adjusted" assembling {
 source_fill intent "반환값으로 이어지는 대입의 기원을 보존한다" {
  hole "delta"
  candidate "delta5179" { fill "delta" "input + 41" }
  candidate "same5179" { fill "delta" "input" }
 }
 case "19" -> "60"
}
`
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "반환기원.gooo", []byte(source), "Adjust")
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateWithSourceIRBodyFill(ctx, "반환기원.gooo", []byte(source), "Adjust", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	flow := prior.Report.BodyFill.SourceOrigin.DataFlow
	if flow == nil || flow.Status != "MAPPED" || len(flow.Roots) != 1 {
		t.Fatal("return graph absent", flow)
	}
	// This fresh receipt exercises optional-field compatibility; no historical
	// source, saved model, or previously observed test cohort is executed.
	raw, _ := json.Marshal(prior)
	var absent Result
	if err := json.Unmarshal(raw, &absent); err != nil {
		t.Fatal(err)
	}
	absent.Report.BodyFill.SourceOrigin.DataFlow = nil
	if err := compareSourceFillReplay(absent, prior); err != nil {
		t.Fatal("absent optional metadata", err)
	}
	if prior.Report.BodyFill.SourceOrigin.DataFlow == nil {
		t.Fatal("comparison mutated original")
	}
	var changed Result
	if err := json.Unmarshal(raw, &changed); err != nil {
		t.Fatal(err)
	}
	changed.Report.BodyFill.SourceOrigin.DataFlow.Nodes[0].Decoded.End++
	if err := compareSourceFillReplay(changed, prior); err == nil {
		t.Fatal("changed origin claim accepted")
	}
	t.Log("5179: fresh return lineage; absent optional metadata accepted; changed recorded range differs; original graph preserved")
}

func TestSimultaneousAssignmentAndDiscardedCalls5231(t *testing.T) {
	const body = "let left = input + 47; let right = input - 31; left, right = right, left; if left > 163 { return right }; return 0"
	flow := InspectBodyValueOrigins(body, 1)
	start := strings.Index(body, "left > 163")
	link := flow.conditionSlice(FillOriginRange{start, start + len("left > 163")}, flow.BodySHA256)
	found := false
	for _, node := range link.Nodes {
		text := body[node.Decoded.Start:node.Decoded.End]
		if text == "input - 31" {
			found = true
		}
		if text == "input + 47" {
			t.Fatal("simultaneous assignment read updated environment")
		}
	}
	if !found {
		t.Fatal("previous right definition absent")
	}
	discard := InspectBodyValueOrigins("let n = input; _ = alter(&n); return n", 1)
	if discard.Status != "PARTIAL" {
		t.Fatal("discarded call effects hidden")
	}
	large := InspectBodyValueOrigins(strings.Repeat(" ", 128<<10+1), 1)
	if large.Status != "UNMAPPED" || len(large.Reasons) != 1 || large.Reasons[0] != "BODY_SIZE_LIMIT" {
		t.Fatal("unbounded input", large)
	}
	t.Log("5231: parallel assignment uses preceding definitions; discarded call effects remain partial; source size bounded")
}
