package testevents

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDecodeKeepsNamedOutcomeAndOrderWithoutPassingPolicy(t *testing.T) {
	raw := []byte(`{"Action":"start","Package":"example"}
{"Action":"run","Test":"TestZ"}
{"Action":"fail","Test":"TestZ"}
{"Action":"pass","Test":"TestA"}
{"Action":"skip","Test":"TestSkipped"}
{"Action":"fail","Package":"example"}`)
	events, err := Decode(raw)
	want := []Event{{"TestZ", "fail"}, {"TestA", "pass"}, {"TestSkipped", "skip"}}
	if err != nil || !reflect.DeepEqual(events, want) {
		t.Fatal("named-test transport imposed passing policy or changed order", events, err)
	}
	encoded, err := json.Marshal(events[0])
	if err != nil || string(encoded) != `{"name":"TestZ","action":"fail"}` {
		t.Fatal("legacy event wire shape changed", string(encoded), err)
	}
}

func TestDecodeRetainsFailureBeforeInvalidSuffix(t *testing.T) {
	for _, suffix := range []string{
		`{`, `{"Action":"future","Test":"TestOther"}`,
		`{"Action":"pass","Test":"TestFail"}`, `{"Action":"fail","Test":"TestFail"}`,
	} {
		events, err := Decode([]byte(`{"Action":"fail","Test":"TestFail"}` + "\n" + suffix))
		if err == nil || len(events) < 1 || events[0] != (Event{"TestFail", "fail"}) {
			t.Fatal("invalid suffix erased a known failure", events, err)
		}
	}
}

func TestDecodeNoNamedTestsDoesNotInventSuccess(t *testing.T) {
	events, err := Decode([]byte(`{"Action":"pass","Package":"example"}`))
	if err != nil || events == nil || len(events) != 0 {
		t.Fatal("package pass became a named test", events, err)
	}
}
