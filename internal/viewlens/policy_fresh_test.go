package viewlens

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/originlenspolicy"
)

func TestSourceLensPolicyFresh1814(t *testing.T) {
	if err := originlenspolicy.VerifyProjection(); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name, requested, preferred, want, reason string
		full, named                              int64
		person                                   bool
	}{
		{"explicit", Full, NamedInputs, Full, "EXPLICIT_VIEW_REQUEST", 0, 9, true},
		{"pinned", "", Full, Full, "EXPLICIT_PREFERENCE", 0, 9, true},
		{"named-count", "", "", NamedInputs, "MOST_SELECTED_VIEW", 3, 9, false},
		{"full-count", "", "", Full, "MOST_SELECTED_VIEW", 9, 3, true},
		{"person-tie", "", "", NamedInputs, "TIED_SELECTION_COUNTS", 9, 9, true},
		{"agent-default", "", "", Full, "AUDIENCE_DEFAULT", 0, 0, false},
	} {
		t.Run(item.name, func(t *testing.T) {
			if got := originlenspolicy.SelectLens(item.requested, item.preferred, item.full, item.named, item.person); got != item.want {
				t.Fatalf("source-owned view: %s", got)
			}
			if got := originlenspolicy.SelectionReason(item.requested, item.preferred, item.full, item.named, item.person); got != item.reason {
				t.Fatalf("source-owned reason: %s", got)
			}
		})
	}
}
