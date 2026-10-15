package verticalsliceclosureshadow

import (
	"bytes"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestFrozenOriginSnapshotWithGrowingRegistry6863(t *testing.T) {
	historic, err := decodeDenominator(embeddedDenominatorV48)
	if err != nil || historic.Version != 48 || historic.Boundaries[0].Target != 80 {
		t.Fatal("frozen origin snapshot", historic, err)
	}
	live := expectedBoundarySpecs()
	if live[0].Target != languagesyntax.FixedCapabilityTotal {
		t.Fatal("live scope disconnected from actual registry", live[0])
	}
	changed := bytes.Replace(embeddedDenominatorV48, []byte(`"target": 80`), []byte(`"target": 81`), 1)
	if bytes.Equal(changed, embeddedDenominatorV48) {
		t.Fatal("missing frozen syntax target")
	}
	if _, err := decodeDenominator(changed); err == nil {
		t.Fatal("changed historical bytes were accepted")
	}
	t.Logf("6863: frozen v48 target=%d; live syntax target=%d; scope separated without source or model execution", historic.Boundaries[0].Target, live[0].Target)
}
