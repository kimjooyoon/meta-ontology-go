package verticalsliceclosureshadow

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestOriginSourceRegistrationV48KeepsOtherBoundaries6599(t *testing.T) {
	if languagesyntax.FixedTotal != 82 || languagesyntax.FixedCapabilityTotal != 80 || languagesyntax.FixedGovernanceTotal != 2 {
		t.Fatal("new source slot not reflected in exact counters")
	}
	if digestBytes(embeddedDenominatorV48) != DenominatorMigrationV48Digest {
		t.Fatal("new version digest")
	}
	next, err := decodeDenominator(embeddedDenominatorV48)
	if err != nil || next.Version != 48 {
		t.Fatal("new version shape", err)
	}
	var prior denominator
	if err := json.Unmarshal(embeddedDenominatorV47, &prior); err != nil {
		t.Fatal(err)
	}
	for i, want := range prior.Boundaries {
		if i == 0 {
			want.Target++
		}
		if next.Boundaries[i] != want {
			t.Fatal("unrelated boundary changed", i)
		}
	}
	t.Log("6599: new v48 syntax target80; other five boundaries and all link targets unchanged; no historical source cases executed")
}
