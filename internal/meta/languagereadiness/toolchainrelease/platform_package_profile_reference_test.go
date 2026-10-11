package toolchainrelease

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestComputationReferenceRejectsIncompleteQueriesAndProfiles(t *testing.T) {
	for _, p := range computationProfiles {
		for _, raw := range []string{`null`, `{}`, `{"text":""}`, `{"needle":0}`, `{"text":null,"needle":0}`,
			`{"text":"","needle":null}`, `{"text":"","needle":0,"extra":1}`, `{"text":2,"needle":0}`} {
			if _, err := computationPackageValue([]byte(raw), p); err == nil {
				t.Fatalf("query accepted: %s", raw)
			}
		}
		f := newComputationFixture(t, p)
		var suite map[string]any
		if err := json.Unmarshal(f.Cases, &suite); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"inputs", "expected"} {
			row := suite["cases"].([]any)[0].(map[string]any)
			before := row[field]
			row[field] = map[string]any{}
			raw, _ := json.Marshal(suite)
			if _, err := computationPackageReference(raw, p); err == nil {
				t.Fatal("incomplete case accepted")
			}
			row[field] = before
		}
		p.Total--
		if _, err := computationPackageReference(f.Cases, p); err == nil {
			t.Fatal("changed profile accepted")
		}
	}
	for _, needle := range []any{0.5, "a", nil} {
		raw, _ := json.Marshal(map[string]any{"text": "abc", "needle": needle})
		if _, err := computationPackageValue(raw, computationProfiles[1]); err == nil {
			t.Fatal("invalid byte needle accepted")
		}
	}
	raw, _ := json.Marshal(map[string]any{"text": strings.Repeat("a", 1025), "needle": "a"})
	if _, err := computationPackageValue(raw, computationProfiles[0]); err == nil {
		t.Fatal("oversize Text accepted")
	}
	unknown := computationProfiles[0]
	unknown.Name = "unknown"
	if _, err := computationPackageValue([]byte(`{"text":"abc","needle":"a"}`), unknown); err == nil {
		t.Fatal("unknown profile accepted")
	}
}
