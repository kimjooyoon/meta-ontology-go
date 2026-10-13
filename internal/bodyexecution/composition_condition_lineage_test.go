package bodyexecution

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestNativeConditionLineageHelpers1699(t *testing.T) {
	source := []byte(`package helpers1699
namespace helpers1699
entity Integer id "helpers1699://integer"
entity Boolean id "helpers1699://boolean"
activity Gate(Integer) -> Boolean id "helpers1699://gate" computes "if input < 33 { return true }; return false"
activity First(Integer) -> Integer id "helpers1699://first" computes "if input < 0 { return 0 }; if Gate(input + 1) { return input + 6 }; return input - 6"
activity Second(Integer) -> Integer id "helpers1699://second" computes "if Gate(input - 2) { return input * 4 }; return input - 5"
bind First.result -> Second.input
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"First":9},"expected":{"First":15,"Second":60}},{"inputs":{"First":45},"expected":{"First":39,"Second":34}},{"inputs":{"First":-4},"expected":{"First":0,"Second":0}}]}`)
	prior, run := newSourceAnchorRun(t, "helpers1699", source, suite, "Second")
	saveConditionObservation(t, "helpers1699", source, run)
	if run.FinitePassed != 6 || len(run.ConditionSites) != 5 {
		t.Fatal("fresh helper behavior differs", run.FinitePassed, len(run.ConditionSites))
	}
	var starts []int
	roots := map[string]bool{}
	events := 0
	for _, trace := range run.Traces {
		events += len(trace.Conditions)
	}
	for _, site := range run.ConditionSites {
		anchor := site.Source
		if anchor == nil || anchor.Status != "MAPPED" || anchor.Declaration == nil || anchor.Declaration.Status != "MAPPED" || anchor.Declaration.StableID != site.ActivityID {
			t.Fatal("authored declaration lineage missing", site)
		}
		if site.ExpressionID != conditionSiteIdentity(site) || !strings.Contains(prior.GoooSource[anchor.Declaration.Start:anchor.Declaration.End], "activity "+anchor.Declaration.Name) {
			t.Fatal("projection identity or declaration location differs", site)
		}
		if site.ActivityID == "helpers1699://gate" {
			starts = append(starts, anchor.Start)
			roots[site.RootActivityID] = true
		}
	}
	if len(starts) != 2 || starts[0] != starts[1] || len(roots) != 2 || events != 13 {
		t.Fatal("helper origin collapsed or native path differs", starts, roots, events)
	}
	changed := prior
	changed.Steps = append([]CompositionStep(nil), prior.Steps...)
	changed.Steps[0].Generation.Report.BodyConditions = append([]bodycodegen.BodyConditionOrigin(nil), prior.Steps[0].Generation.Report.BodyConditions...)
	changed.Steps[0].Generation.Report.BodyConditions[0].Start++
	if VerifyComposition(context.Background(), "helpers1699.gooo", source, changed) == nil {
		t.Fatal("source origin alteration went unnoticed")
	}
	t.Log("helpers1699: shared helper origin across2 roots; activity declarations5/5; native6/6; events13; model0")
}

func TestNativeConditionLineageSelectedFill1753(t *testing.T) {
	source := []byte(`package fill1753
namespace fill1753
entity Integer id "fill1753://integer"
activity Calculate(Integer) -> Integer id "fill1753://calculate" computes ` + "`" + `let offset = __GOOO_BODY_HOLE_offset__
if input < 8 { return input + offset }; return input + 1` + "`" + ` assembling {
 source_fill intent "작은 입력에 지정된 큰 상수를 더한다." {
  hole "offset"
  candidate "large" { fill "offset" "300000000000" }
  candidate "small" { fill "offset" "7" }
 }
 case "1" -> "300000000001"
 holdout_case "5" -> "300000000005"
}
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Calculate":2},"expected":{"Calculate":300000000002}},{"inputs":{"Calculate":11},"expected":{"Calculate":12}}]}`)
	prior, run := newSourceAnchorRun(t, "fill1753", source, suite)
	saveConditionObservation(t, "fill1753", source, run)
	if run.FinitePassed != 2 || len(run.ConditionSites) != 1 || prior.OriginalSourceSHA256 == prior.SelectedSourceSHA256 {
		t.Fatal("selected source construction differs", run.FinitePassed)
	}
	anchor := run.ConditionSites[0].Source
	if anchor == nil || anchor.Status != "MAPPED" || anchor.SourceSHA256 != prior.SelectedSourceSHA256 || prior.GoooSource[anchor.Start:anchor.End] != "input < 8" {
		t.Fatal("condition uses original template offsets", anchor)
	}
	if prior.Steps[0].Generation.Report.BodyFill == nil || anchor.Declaration == nil || anchor.Declaration.StableID != "fill1753://calculate" || !strings.Contains(prior.GoooSource, "300000000000") {
		t.Fatal("selection or declaration origin lost")
	}
	t.Log("fill1753: selected-carrier origin after earlier hole replacement; native2/2; original/selected source hashes distinct; model0")
}
