package bodycodegen

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSourceFillOriginSurvivesRelocation4177(t *testing.T) {
	source := []byte(`package movement4177
namespace movement4177
entity Integer id "urn:gooo:type:integer"
activity Offset(Integer) -> Integer id "urn:movement4177:offset" computes "return __GOOO_BODY_HOLE_delta__" assembling {
 source_fill intent "원래 파일의 이름을 조립 기원으로 남긴다" {
  hole "delta"
  candidate "distance4177" { fill "delta" "input + 29" }
  candidate "identity4177" { fill "delta" "input" }
 }
 case "17" -> "46"
}
`)
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "최초/거리.gooo", source, "Offset")
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateWithSourceIRBodyFill(ctx, "최초/거리.gooo", source, "Offset", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(prior)
	var saved Result
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if _, err := RealizeSourceAssembly(ctx, "<moved-source>", source, saved); err != nil {
		t.Fatal("relocated source", err)
	}
	if saved.Report.BodyFill.SourceOrigin.Filename != "최초/거리.gooo" {
		t.Fatal("original filename overwritten")
	}
	saved.Report.BodyFill.SourceOrigin.Fragments[0].ExpressionSource.End++
	if _, err := RealizeSourceAssembly(ctx, "<changed-claim>", source, saved); err == nil {
		t.Fatal("changed source range accepted")
	}
	t.Log("relocation4177: source label retained across a different read path; actual declaration ranges still compared; model0")
}
