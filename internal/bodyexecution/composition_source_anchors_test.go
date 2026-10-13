package bodyexecution

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"strconv"
	"strings"
	"testing"
)

func TestNativeSourceAnchorUnicode1567(t *testing.T) {
	body := `let 안내 = "줄\n끝"; let 한계 = 37; if input.수량 < 한계 && input.메모 != "a\"b" { return 11 }; if input.수량 < 한계 { return 12 }; return 13`
	quoted := strings.ReplaceAll(strconv.Quote(body), "수량", `\uC218\uB7C9`)
	source := []byte(`package unicode1567
namespace unicode1567
entity Integer id "unicode1567://integer"
entity Request id "unicode1567://request" fields {
 field 수량 id "unicode1567://quantity" type integer required one
 field 메모 id "unicode1567://memo" type string required one
}
activity Decide(Request) -> Integer id "unicode1567://decide" computes ` + quoted + "\n")
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Decide":{"수량":9,"메모":"x"}},"expected":{"Decide":11}},{"inputs":{"Decide":{"수량":10,"메모":"a\"b"}},"expected":{"Decide":12}},{"inputs":{"Decide":{"수량":50,"메모":"x"}},"expected":{"Decide":13}}]}`)
	prior, run := newSourceAnchorRun(t, "unicode1567", source, suite)
	if run.FinitePassed != 3 || len(run.ConditionSites) != 2 {
		t.Fatal("fresh typed execution differs", run.FinitePassed)
	}
	for i, site := range run.ConditionSites {
		anchor := site.Source
		if anchor == nil || anchor.Status != "MAPPED" || anchor.SourceSHA256 != prior.SelectedSourceSHA256 || anchor.SourceOrdinal != i {
			t.Fatal("forward source condition missing", site)
		}
		fragment := strconv.Quote(anchor.Expression)
		fragment = fragment[1 : len(fragment)-1]
		fragment = strings.ReplaceAll(fragment, "수량", `\uC218\uB7C9`)
		if prior.GoooSource[anchor.Start:anchor.End] != fragment {
			t.Fatal("Unicode/escape bytes differ", anchor)
		}
		want := 1
		if i == 0 {
			want = 2
		}
		if len(anchor.Fields) != want || anchor.Fields[0].FieldID != "unicode1567://quantity" || anchor.Fields[0].EntityID != "unicode1567://request" {
			t.Fatal("typed field IDs missing", anchor.Fields)
		}
	}
	if run.ConditionSites[0].Source.Fields[1].FieldID != "unicode1567://memo" || run.ConditionSites[0].Source.Start >= run.ConditionSites[1].Source.Start {
		t.Fatal("source field ownership/order lost")
	}
	events := 0
	for _, trace := range run.Traces {
		events += len(trace.Conditions)
	}
	if events != 5 || !strings.Contains(prior.GoooSource, `\uC218\uB7C9`) || run.ModelCalls != 0 {
		t.Fatal("native path or source changed")
	}
	saveConditionObservation(t, "unicode1567", source, run)
	t.Log("unicode1567 native3/3 conditions5; source sites2/2; typed field references3/3; escaped Unicode bytes exact; model0")
}

func TestNativeSourceAnchorCopiedHelpers1627(t *testing.T) {
	source := []byte(`package helpers1627
namespace helpers1627
entity Integer id "helpers1627://integer"
entity Boolean id "helpers1627://boolean"
activity Gate(Integer) -> Boolean id "helpers1627://gate" computes "if input < 29 { return true }; return false"
activity First(Integer) -> Integer id "helpers1627://first" computes "if Gate(input) { return input + 4 }; return input - 4"
activity Second(Integer) -> Integer id "helpers1627://second" computes "if Gate(input) { return input * 3 }; return input - 3"
bind First.result -> Second.input
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"First":6},"expected":{"First":10,"Second":30}},{"inputs":{"First":39},"expected":{"First":35,"Second":32}}]}`)
	prior, run := newSourceAnchorRun(t, "helpers1627", source, suite, "Second")
	if run.FinitePassed != 4 || len(run.ConditionSites) != 4 {
		t.Fatal("copied helper outputs/sites differ", run.FinitePassed, len(run.ConditionSites))
	}
	var starts []int
	roots := map[string]bool{}
	for _, site := range run.ConditionSites {
		if site.Source == nil || site.Source.Status != "MAPPED" || site.Source.SourceSHA256 != prior.SelectedSourceSHA256 {
			t.Fatal("helper source unmapped", site)
		}
		if site.ActivityID == "helpers1627://gate" {
			starts = append(starts, site.Source.Start)
			roots[site.RootActivityID] = true
		}
	}
	if len(starts) != 2 || starts[0] != starts[1] || len(roots) != 2 {
		t.Fatal("helper identity or source location collapsed", starts, roots)
	}
	changed := prior
	changed.Steps = append([]CompositionStep(nil), prior.Steps...)
	changed.Steps[0].Generation.Report.BodyConditions = append([]bodycodegen.BodyConditionOrigin(nil), prior.Steps[0].Generation.Report.BodyConditions...)
	changed.Steps[0].Generation.Report.BodyConditions[0].Start++
	if VerifyComposition(context.Background(), "helpers1627.gooo", source, changed) == nil {
		t.Fatal("changed forward origin accepted")
	}
	saveConditionObservation(t, "helpers1627", source, run)
	t.Log("helpers1627 native4/4; source sites4/4; one helper source retained across2 roots; source unchanged; model0")
}

func newSourceAnchorRun(t *testing.T, name string, source []byte, suite CompositionCases, entry ...string) (Composition, CompositionRuntime) {
	t.Helper()
	ctx := context.Background()
	options := CompositionOptions{}
	if len(entry) != 0 {
		options.EntryActivity = entry[0]
	}
	prior, err := GenerateCompositionWithOptions(ctx, name+".gooo", source, suite, options)
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(ctx, name+".gooo", source, prior, suite, nativeTool())
	if err != nil || !run.RuntimeReplayed {
		t.Fatal("fresh source execution", err, run.Stage)
	}
	return prior, run
}
