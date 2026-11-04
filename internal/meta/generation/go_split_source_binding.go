package generation

import (
	"crypto/sha256"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// ValidatePreparedGoSplitSource binds the selected inspection to the actual
// disposable executor input. Logical plans without inspections retain their
// existing executor checks; an attached inspection cannot be silently ignored.
func ValidatePreparedGoSplitSource(plan Plan, action Action, source []byte) error {
	if err := validateGoSplitObservations(plan); err != nil {
		return err
	}
	if len(plan.GoSplitPreparations) == 0 {
		return nil
	}
	selected := false
	for _, expected := range plan.Selected {
		selected = selected || (expected.IndicatorID == action.IndicatorID && expected.Subject == action.Subject &&
			expected.Operation == sourcepolicy.OperationSplitGo && action.Operation == expected.Operation)
	}
	if !selected {
		return fmt.Errorf("source prerequisite does not belong to this selected source occurrence")
	}
	for _, observed := range plan.GoSplitPreparations {
		if observed.IndicatorID != action.IndicatorID {
			continue
		}
		if !observed.readyForProjection() || observed.SourceSHA256 != fmt.Sprintf("%x", sha256.Sum256(source)) {
			return fmt.Errorf("selected split source differs from its current-file prerequisite observation")
		}
		return nil
	}
	return fmt.Errorf("selected split has no source prerequisite observation")
}
