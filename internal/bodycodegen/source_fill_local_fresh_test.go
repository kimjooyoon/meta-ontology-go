package bodycodegen

import "testing"

func TestCompoundLocalFillOrigins11247(t *testing.T) {
	source := []byte(`package compound11247
namespace compound11247
entity Integer id "urn:gooo:type:integer"
activity Triple(Integer) -> Integer id "urn:compound11247:triple" computes "let base=__GOOO_BODY_HOLE_base__+input;let unchanged=base;return unchanged*3" assembling {
 source_fill intent "하나의 대입 표현 안에서 후보 조각과 템플릿 조각을 구분한다" {
  hole "base"
  candidate "base11247" {fill "base" "421"}
  candidate "other11247" {fill "base" "433"}
 }
 case "439" -> "2580"
 case "-443" -> "-66"
}
`)
	generated := generateOriginFill3371(t, "compound11247.gooo", source, "Triple")
	origin := generated.Report.BodyFill.SourceOrigin
	local := generated.Report.BodyLocals[0]
	link := origin.LocalOrigin(local)
	if link == nil || len(link.Fragments) != 2 || link.CandidateID != "base11247" || link.SourceSHA256 != digest(source) {
		t.Fatal("compound local construction origin", link)
	}
	if link.Fragments[0].Kind != "declared_candidate" || link.Fragments[1].Kind != "retained_template" {
		t.Fatal("candidate/template fragments mixed", link.Fragments)
	}
	location := link.Fragments[0].ExpressionSource
	if location == nil || string(source[location.Start:location.End]) != "421" {
		t.Fatal("candidate literal location", location)
	}
	location.Start++
	link.CandidateSource.Start++
	again := origin.LocalOrigin(local)
	if again.Fragments[0].ExpressionSource.Start == location.Start || again.CandidateSource.Start == link.CandidateSource.Start {
		t.Fatal("returned origin aliases retained source")
	}
	for _, field := range []string{"activity", "body", "negative", "past-end", "schema", "gap"} {
		changed, write := *origin, local
		switch field {
		case "activity":
			write.ActivityID = "urn:another:activity"
		case "body":
			write.BodySHA256 = "sha256:another"
		case "negative":
			write.Start = -1
		case "past-end":
			write.End += 65536
		case "schema":
			changed.Schema = ""
		case "gap":
			changed.Fragments = nil
			for _, fragment := range origin.Fragments {
				if fragment.HoleID != "base" {
					changed.Fragments = append(changed.Fragments, fragment)
				}
			}
		}
		if changed.LocalOrigin(write) != nil {
			t.Fatal("contradictory local construction origin accepted", field)
		}
	}
	saveLocalOrigin9811(t, "local-fill11247.json", map[string]any{"source": string(source), "generation": generated,
		"local_construction": again, "contradictions_rejected": 6, "native_executions": 0, "model_calls": 0})
	t.Log("compound RHS has candidate plus template fragments; original candidate literal exact; detached copies; six contradictions rejected; native0/model0")
}

func TestLocalFillIntervalContract11271(t *testing.T) {
	// Synthetic metadata intervals isolate coverage rules; no Gooo program or
	// historical source cohort is executed by this contract check.
	origin := SourceFillOrigin{Schema: "gooo/source-fill-origin/v1", ActivityID: "urn:interval11271:activity",
		SourceSHA256: digest([]byte("assembly11271")), SelectedBodySHA256: digest([]byte("selected11271")),
		Fragments: []SourceFillFragment{
			{SelectedDecoded: FillOriginRange{11, 16}, Kind: "declared_candidate"},
			{SelectedDecoded: FillOriginRange{16, 19}, Kind: "retained_template"},
			{SelectedDecoded: FillOriginRange{19, 23}, Kind: "declared_candidate"},
		}}
	local := BodyLocalOrigin{ActivityID: origin.ActivityID, BodySHA256: origin.SelectedBodySHA256, Start: 12, End: 22}
	if link := origin.LocalOrigin(local); link == nil || len(link.Fragments) != 3 {
		t.Fatal("covered partial-edge interval rejected", link)
	}
	for _, fragments := range [][]SourceFillFragment{
		{origin.Fragments[0], origin.Fragments[2]},
		{origin.Fragments[0], origin.Fragments[1]},
		{origin.Fragments[1], origin.Fragments[2]},
	} {
		changed := origin
		changed.Fragments = fragments
		if changed.LocalOrigin(local) != nil {
			t.Fatal("interior or edge gap acquired a construction claim", fragments)
		}
	}
	t.Log("synthetic interval coverage1/1; interior/start/end gaps3/3 rejected; program executions0/models0")
}
