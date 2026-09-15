package selectiveci

import (
	"reflect"
	"testing"
)

func TestCorpusContract(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if corpus.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", corpus.SchemaVersion)
	}
	if len(corpus.Cases) < 20 {
		t.Fatalf("corpus has %d cases, want at least 20 partitions", len(corpus.Cases))
	}

	seen := make(map[string]struct{}, len(corpus.Cases))
	for _, fixture := range corpus.Cases {
		if _, exists := seen[fixture.Name]; exists {
			t.Fatalf("duplicate fixture name %q", fixture.Name)
		}
		seen[fixture.Name] = struct{}{}

		digest := CanonicalDigest(fixture)
		if fixture.Expected.CanonicalDigest == "" {
			t.Logf("%s canonicalDigest=%q", fixture.Name, digest)
			t.Errorf("%s has no canonical digest", fixture.Name)
		}
		if digest != fixture.Expected.CanonicalDigest {
			t.Errorf("%s digest = %s, want %s", fixture.Name, digest, fixture.Expected.CanonicalDigest)
		}

		got := Evaluate(fixture)
		want := Result{
			Decision:        fixture.Expected.Decision,
			Reason:          fixture.Expected.Reason,
			CommandIDs:      fixture.Expected.CommandIDs,
			Argv:            fixture.Expected.Argv,
			CPUUnits:        fixture.Expected.CPUUnits,
			MemoryCeiling:   fixture.Expected.MemoryCeiling,
			PathCount:       fixture.Expected.PathCount,
			CanonicalDigest: fixture.Expected.CanonicalDigest,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s result = %#v, want %#v", fixture.Name, got, want)
		}
		if got.Decision == FullSuiteFallback && (len(got.CommandIDs) != 0 || len(got.Argv) != 0 || got.CPUUnits != 0 || got.MemoryCeiling != 0 || got.PathCount != 0) {
			t.Errorf("%s fallback retained partial selection: %#v", fixture.Name, got)
		}
	}
}
