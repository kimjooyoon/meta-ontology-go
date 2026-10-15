package verticalsliceclosureshadow

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestEverydayDenominator73RetainsHistorical72Target67471(t *testing.T) {
	current, err := decodeDenominator(activeDenominator())
	if err != nil || !registeredSourceScope69131(current.Version, current.Boundaries[0].Target, languagesyntax.FixedCapabilityTotal) {
		t.Fatal("current registered-source target", current, err)
	}
	historical, err := decodeDenominator(embeddedDenominatorV73)
	if err != nil || historical.Version != 73 || historical.Boundaries[0].Target != 109 ||
		digestBytes(embeddedDenominatorV73) != DenominatorMigrationV73Digest {
		t.Fatal("historical registered-source target changed", historical, err)
	}
	previous, err := decodeDenominator(embeddedDenominatorV72)
	if err != nil || previous.Version != 72 || previous.Boundaries[0].Target != 108 ||
		digestBytes(embeddedDenominatorV72) != DenominatorMigrationV72Digest {
		t.Fatal("historical source target changed", previous, err)
	}
	links := 0
	for _, boundary := range current.Boundaries {
		links += boundary.LinkTarget
	}
	if links != 13 {
		t.Fatal("historical link obligations changed", links)
	}
	if !reflect.DeepEqual(previous.Boundaries[1:], current.Boundaries[1:]) {
		t.Fatal("unrelated boundaries or links changed")
	}
	lowered := bytes.Replace(embeddedDenominatorV73, []byte("\"target\": 109"), []byte("\"target\": 108"), 1)
	if bytes.Equal(lowered, embeddedDenominatorV73) {
		t.Fatal("counterexample did not change its source")
	}
	if _, err := decodeDenominator(lowered); err == nil {
		t.Fatal("unapproved current target accepted")
	}
}

func registeredSourceScope69131(version, target, sourceTarget int) bool {
	return version >= 73 && target == sourceTarget
}
