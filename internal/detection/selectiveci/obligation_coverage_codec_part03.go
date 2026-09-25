package selectiveci

import (
	"fmt"
)

func validateCoverageResult(result ObligationCoverageResult) (ObligationCoverageResult, error) {
	if result.SchemaVersion != ObligationCoverageSchemaVersion {
		return ObligationCoverageResult{}, fmt.Errorf("unsupported obligation coverage schema")
	}
	if result.Decision != CoverageDecisionExact && result.Decision != CoverageDecisionUnknown {
		return ObligationCoverageResult{}, fmt.Errorf("invalid obligation coverage decision")
	}
	if !validCoverageReason(result.Decision, result.Reason) {
		return ObligationCoverageResult{}, fmt.Errorf("invalid obligation coverage reason")
	}
	if result.FullSuiteRequired != (result.Decision == CoverageDecisionUnknown) {
		return ObligationCoverageResult{}, fmt.Errorf("inconsistent full-suite flag")
	}
	result = normalizeCoverageResult(result)
	if result.Decision == CoverageDecisionUnknown && len(result.RequiredObligationIDs) != 0 {
		return ObligationCoverageResult{}, fmt.Errorf("unknown coverage exposes required obligations")
	}
	return result, nil
}
func validCoverageReason(decision CoverageDecision, reason CoverageReason) bool {
	if decision == CoverageDecisionExact {
		return reason == CoverageReasonComplete || reason == CoverageReasonNoChange
	}
	switch reason {
	case CoverageReasonMissingInput, CoverageReasonInvalidInput, CoverageReasonUnsupportedSchema, CoverageReasonInvalidGraph,
		CoverageReasonInvalidRegistry, CoverageReasonInvalidSnapshot, CoverageReasonStaleGraph,
		CoverageReasonStaleRegistry, CoverageReasonStaleSnapshot, CoverageReasonUnknownRoot,
		CoverageReasonDuplicateRoot, CoverageReasonMissingObligation, CoverageReasonMissingCommand,
		CoverageReasonDanglingCommand, CoverageReasonWorkOverflow:
		return true
	default:
		return false
	}
}
