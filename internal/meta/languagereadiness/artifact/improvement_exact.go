package artifact

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"
)

func requireFirstImprovement(value improvement.Transition) error {
	switch {
	case value.Decision != improvement.Improved ||
		value.ReasonCode != "IMPROVEMENT_PROVEN" || !value.Comparable:
		return fmt.Errorf("FAIL_CLOSED: first improvement decision not proven")
	case value.BeforeCompleted != 7 || value.AfterCompleted != 8 ||
		value.Total != improvement.SnapshotTotal:
		return fmt.Errorf("FAIL_CLOSED: first improvement obligation counts mismatch")
	case value.CompletedDelta != 1 || value.Gains != 1:
		return fmt.Errorf("FAIL_CLOSED: first improvement gain mismatch")
	case value.BeforeBasisPoints != 2916 || value.AfterBasisPoints != 3333 ||
		value.BasisPointsDelta != 417:
		return fmt.Errorf("FAIL_CLOSED: first improvement basis points mismatch")
	case value.Regressions != 0 || value.BeforeUnresolved != 0 ||
		value.AfterUnresolved != 0:
		return fmt.Errorf("FAIL_CLOSED: first improvement guardrail failed")
	case len(value.Indicators) != 5 || len(value.Proofs) != 4:
		return fmt.Errorf("FAIL_CLOSED: first improvement evidence incomplete")
	}
	for _, proof := range value.Proofs {
		if !proof.Passed {
			return fmt.Errorf("FAIL_CLOSED: first improvement proof %q failed", proof.ID)
		}
	}
	return nil
}
