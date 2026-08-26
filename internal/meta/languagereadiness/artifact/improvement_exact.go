package artifact

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"
)

func requireAcceptedTransition(value improvement.Transition) error {
	switch {
	case !value.Comparable || value.Total != improvement.SnapshotTotal:
		return fmt.Errorf("FAIL_CLOSED: readiness transition not comparable")
	case value.Regressions != 0 || value.BeforeUnresolved != 0 ||
		value.AfterUnresolved != 0:
		return fmt.Errorf("FAIL_CLOSED: readiness transition guardrail failed")
	case len(value.Indicators) != 5 || len(value.Proofs) != 4:
		return fmt.Errorf("FAIL_CLOSED: readiness transition evidence incomplete")
	case value.Decision == improvement.Improved &&
		(value.ReasonCode != "IMPROVEMENT_PROVEN" || value.CompletedDelta <= 0 ||
			value.BasisPointsDelta <= 0 || value.Gains != value.CompletedDelta):
		return fmt.Errorf("FAIL_CLOSED: readiness improvement arithmetic mismatch")
	case value.Decision == improvement.NoChange &&
		(value.ReasonCode != "NO_NUMERIC_CHANGE" || value.CompletedDelta != 0 ||
			value.BasisPointsDelta != 0 || value.Gains != 0):
		return fmt.Errorf("FAIL_CLOSED: readiness fixed point arithmetic mismatch")
	case value.Decision != improvement.Improved && value.Decision != improvement.NoChange:
		return fmt.Errorf("FAIL_CLOSED: readiness transition decision rejected")
	}
	for _, proof := range value.Proofs {
		if !proof.Passed {
			return fmt.Errorf("FAIL_CLOSED: readiness transition proof %q failed", proof.ID)
		}
	}
	return nil
}
