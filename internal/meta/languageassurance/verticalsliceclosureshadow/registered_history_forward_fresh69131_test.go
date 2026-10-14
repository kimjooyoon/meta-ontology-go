package verticalsliceclosureshadow

import "testing"

func TestRegisteredHistoryForwardScope69131(t *testing.T) {
	for _, view := range []struct{ version, target, source int }{
		{73, 109, 109}, {74, 110, 110}, {75, 111, 111},
	} {
		if !registeredSourceScope69131(view.version, view.target, view.source) {
			t.Fatal("source-bound continued scope rejected", view)
		}
	}
	for _, view := range []struct{ version, target, source int }{
		{72, 108, 108}, {74, 109, 110}, {74, 111, 110}, {75, 110, 111},
	} {
		if registeredSourceScope69131(view.version, view.target, view.source) {
			t.Fatal("stale or mismatched source scope accepted", view)
		}
	}
	for _, frozen := range []struct {
		raw     []byte
		version int
		target  int
	}{
		{embeddedDenominatorV72, 72, 108}, {embeddedDenominatorV73, 73, 109},
	} {
		value, err := decodeDenominator(frozen.raw)
		if err != nil || value.Version != frozen.version || value.Boundaries[0].Target != frozen.target {
			t.Fatal("historical source target changed", frozen.version, value, err)
		}
	}
}
