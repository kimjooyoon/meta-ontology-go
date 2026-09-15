package transformationeffect

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSplitGoEvaluationArtifactJSONRoundTrip(t *testing.T) {
	want := SplitGoEvaluationArtifact{Reasons: []string{}}
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got SplitGoEvaluationArtifact
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON roundtrip mismatch: got %#v, want %#v", got, want)
	}
}
