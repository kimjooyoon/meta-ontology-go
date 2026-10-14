package bodycodegen

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRetainedLocalUseContract11017(t *testing.T) {
	source := []byte(`package use11017
namespace use11017
entity Integer id "urn:gooo:type:integer"
activity Adjust(Integer) -> Integer id "urn:use11017:adjust" computes "let offset=__GOOO_BODY_HOLE_offset__;let copy=offset;offset=offset+input;return copy+offset" assembling {
 source_fill intent "대입 기원을 보존하고 저장되지 않은 사용 좌표는 비워 둔다" {
  hole "offset"
  candidate "offset11017" {fill "offset" "347"}
  candidate "other11017" {fill "offset" "349"}
 }
 case "353" -> "1047"
 case "-359" -> "335"
}
`)
	prior := generateOriginFill3371(t, "use11017.gooo", source, "Adjust")
	if len(prior.Report.BodyLocals) != 3 || len(prior.Report.BodyLocals[0].Uses) != 3 {
		t.Fatal("fresh local write/use population", prior.Report.BodyLocals)
	}
	encoded, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	var absent Result
	if err := json.Unmarshal(encoded, &absent); err != nil {
		t.Fatal(err)
	}
	for i := range absent.Report.BodyLocals {
		absent.Report.BodyLocals[i].Uses = nil
	}
	realized, err := RealizeSourceAssembly(context.Background(), "use11017.gooo", source, absent)
	if err != nil || realized.ModelCalls != 0 {
		t.Fatal("omitted optional coordinates stopped realization", err)
	}
	for _, origin := range absent.Report.BodyLocals {
		if origin.Uses != nil {
			t.Fatal("historical use claim fabricated")
		}
	}
	checkRetainedUseContradictions11017(t, encoded, prior)
	after, err := json.Marshal(prior)
	if err != nil || string(encoded) != string(after) {
		t.Fatal("comparison mutated current metadata", err)
	}
	saveLocalOrigin9811(t, "use11017.json", map[string]any{"source": string(source), "generation": prior,
		"optional_use_absence": "preserved", "contradictions_rejected": 4, "native_executions": 0, "model_calls": 0})
	t.Log("fresh source_fill two cases; three writes; four retained-coordinate contradictions rejected; omitted uses stay absent; native0/model0")
}

func checkRetainedUseContradictions11017(t *testing.T, encoded []byte, expected Result) {
	t.Helper()
	for _, field := range []string{"use-position", "use-ordinal", "empty-claim", "write-position"} {
		var changed Result
		if err := json.Unmarshal(encoded, &changed); err != nil {
			t.Fatal(err)
		}
		local := &changed.Report.BodyLocals[0]
		switch field {
		case "use-position":
			local.Uses[0].Start++
		case "use-ordinal":
			local.Uses[0].NativeOrdinal++
		case "empty-claim":
			local.Uses = []BodyLocalUseOrigin{}
		case "write-position":
			local.Uses = nil
			local.Start++
		}
		if EqualRetainedLocalOrigins(changed.Report.BodyLocals, expected.Report.BodyLocals) {
			t.Fatal("retained contradiction accepted", field)
		}
		if err := compareSourceFillReplay(changed, expected); err == nil {
			t.Fatal("source_fill ignored retained contradiction", field)
		}
	}
}
