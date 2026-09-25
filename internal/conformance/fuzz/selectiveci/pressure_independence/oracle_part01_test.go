package pressureindependence

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCorpus(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if corpus.Schema != CorpusSchemaV1 || len(corpus.Cases) < 11 {
		t.Fatalf("corpus schema/count = %q/%d", corpus.Schema, len(corpus.Cases))
	}
	if corpus.CanonicalDigest != CorpusDigest(corpus) {
		t.Logf("corpus canonical_digest=%q", CorpusDigest(corpus))
		t.Errorf("corpus digest mismatch")
	}
	seen := make(map[string]struct{}, len(corpus.Cases))
	for _, row := range corpus.Cases {
		if row.Name == "" {
			t.Errorf("empty corpus case name")
		}
		if _, exists := seen[row.Name]; exists {
			t.Errorf("duplicate corpus case %q", row.Name)
		}
		seen[row.Name] = struct{}{}
		got := Evaluate(row.Input)
		if row.Expected.CanonicalOutputDigest == "" {
			encoded, _ := json.Marshal(got)
			t.Logf("%s expected=%s", row.Name, encoded)
			continue
		}
		if !reflect.DeepEqual(got, row.Expected) {
			t.Errorf("%s output = %#v, want %#v", row.Name, got, row.Expected)
		}
	}
}
func TestStrictInputJSON(t *testing.T) {
	base := mustCorpusInput(t, "two-independent-groups-pass")
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeInput(append(data, []byte(` {"schema":"extra"}`)...)); err == nil {
		t.Fatal("trailing JSON value accepted")
	}
	duplicate := strings.Replace(string(data), `"schema":"gooo/pressure-independence/v1"`,
		`"schema":"gooo/pressure-independence/v1","schema":"gooo/pressure-independence/v1"`, 1)
	if _, err := DecodeInput([]byte(duplicate)); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
	unknown := strings.TrimSuffix(string(data), "}") + `,"unknown":true}`
	if _, err := DecodeInput([]byte(unknown)); err == nil {
		t.Fatal("unknown JSON field accepted")
	}
}
