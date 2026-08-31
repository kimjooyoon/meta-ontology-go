package improvement

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func classify(result *Transition) {
	resolved := result.BeforeUnresolved == 0 && result.AfterUnresolved == 0
	switch {
	case !resolved:
		result.ReasonCode = "UNRESOLVED_EVIDENCE"
	case result.Regressions > 0 || result.CompletedDelta < 0 || result.BasisPointsDelta < 0:
		result.Decision = Regressed
		result.ReasonCode = "NUMERIC_REGRESSION"
	case result.CompletedDelta > 0 && result.BasisPointsDelta > 0 && result.Gains > 0:
		result.Decision = Improved
		result.ReasonCode = "IMPROVEMENT_PROVEN"
	case result.CompletedDelta == 0 && result.BasisPointsDelta == 0 && result.Gains == 0:
		result.Decision = NoChange
		result.ReasonCode = "NO_NUMERIC_CHANGE"
	default:
		result.ReasonCode = "POSITIVE_DELTA_NOT_PROVEN"
	}
}

func indicators(result Transition) []Indicator {
	return []Indicator{
		{ID: "completed-obligations", Class: "OUTCOME", Before: result.BeforeCompleted, After: result.AfterCompleted, Delta: result.CompletedDelta, Unit: "OBLIGATION"},
		{ID: "readiness-basis-points", Class: "OUTCOME", Before: result.BeforeBasisPoints, After: result.AfterBasisPoints, Delta: result.BasisPointsDelta, Unit: "BASIS_POINT"},
		{ID: "newly-satisfied", Class: "DRIVER", After: result.Gains, Delta: result.Gains, Unit: "OBLIGATION"},
		{ID: "regressions", Class: "GUARDRAIL", After: result.Regressions, Delta: result.Regressions, Unit: "OBLIGATION"},
		{ID: "unresolved-evidence", Class: "GUARDRAIL", Before: result.BeforeUnresolved, After: result.AfterUnresolved, Delta: result.AfterUnresolved - result.BeforeUnresolved, Unit: "OBLIGATION"},
	}
}

func proofs(comparable, arithmetic, resolved, regressionFree bool) []Proof {
	return []Proof{
		{ID: "fixed-contract", Choice: "FOUNDATION", Passed: comparable},
		{ID: "integer-arithmetic", Choice: "COHERENCE", Passed: arithmetic},
		{ID: "resolved-evidence", Choice: "FOUNDATION", Passed: resolved},
		{ID: "zero-regression", Choice: "REGRESSION", Passed: regressionFree},
	}
}

func seal(result Transition) Transition {
	result.Digest = ""
	payload, _ := json.Marshal(result)
	sum := sha256.Sum256(payload)
	result.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return result
}
