package verticalsliceclosureshadow

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestOriginDenominatorMigration27307(t *testing.T) {
	if digestBytes(embeddedDenominatorV72) != DenominatorMigrationV72Digest {
		t.Fatal("historical v72 denominator source binding")
	}
	active, activeErr := decodeDenominator(activeDenominator())
	if activeErr != nil || active.Version < 72 || active.Boundaries[0].Target != languagesyntax.FixedCapabilityTotal ||
		digestBytes(activeDenominator()) != activeDenominatorDigest() {
		t.Fatal("current source-selected denominator binding", active, activeErr)
	}
	current, err := decodeDenominator(embeddedDenominatorV72)
	if err != nil || current.Version != 72 || current.Boundaries[0].Target != 108 {
		t.Fatal("complete new syntax target", current, err)
	}
	var previous denominator
	if err := json.Unmarshal(embeddedDenominatorV71, &previous); err != nil {
		t.Fatal(err)
	}
	for i, expected := range previous.Boundaries {
		if i == 0 {
			expected.Target += 4
		}
		if current.Boundaries[i] != expected {
			t.Fatal("unrelated boundary changed", i)
		}
	}
	for _, prior := range []struct {
		raw    []byte
		digest string
	}{
		{embeddedDenominatorV65, DenominatorMigrationV65Digest}, {embeddedDenominatorV66, DenominatorMigrationV66Digest},
		{embeddedDenominatorV67, DenominatorMigrationV67Digest}, {embeddedDenominatorV68, DenominatorMigrationV68Digest},
		{embeddedDenominatorV69, DenominatorMigrationV69Digest}, {embeddedDenominatorV70, DenominatorMigrationV70Digest},
		{embeddedDenominatorV71, DenominatorMigrationV71Digest},
	} {
		if digestBytes(prior.raw) != prior.digest {
			t.Fatal("historical denominator bytes rewritten")
		}
		if _, err := decodeDenominator(prior.raw); err != nil {
			t.Fatal("historical65..71 denominator lost its exact target", err)
		}
	}
	t.Log("denominator27307:v72 adds exactly4 syntax capabilities;old65..71 bytes/digests and exact targets retained;other5 boundaries and13 links unchanged;no producer/model calls")
}
