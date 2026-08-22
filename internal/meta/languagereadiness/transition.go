package languagereadiness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const (
	ImprovementSnapshotSchema   = "gooo/language-readiness-snapshot/v1"
	ImprovementTransitionSchema = "gooo/quantified-improvement-transition/v1"

	ImprovementSatisfied    ImprovementEvidenceStatus = "SATISFIED"
	ImprovementNotSatisfied ImprovementEvidenceStatus = "NOT_SATISFIED"
	ImprovementUnresolved   ImprovementEvidenceStatus = "UNRESOLVED"

	ImprovementImproved        ImprovementDecision = "IMPROVED"
	ImprovementNoChange        ImprovementDecision = "NO_CHANGE"
	ImprovementRegressed       ImprovementDecision = "REGRESSED"
	ImprovementLowerResolution ImprovementDecision = "LOWER_RESOLUTION"
)

type ImprovementEvidenceStatus string

type ImprovementDecision string

type ImprovementEvidence struct {
	ID     string                    `json:"id"`
	Status ImprovementEvidenceStatus `json:"status"`
}

// ImprovementSnapshot is a fixed-grain, integer-only view of readiness.
// Descriptive claims are deliberately absent: only registered evidence counts.
type ImprovementSnapshot struct {
	ContractSchema string                `json:"contract_schema"`
	RegistryDigest string                `json:"registry_digest"`
	Completed      int64                 `json:"completed"`
	Total          int64                 `json:"total"`
	BasisPoints    int64                 `json:"basis_points"`
	Evidence       []ImprovementEvidence `json:"evidence"`
}

type ImprovementIndicator struct {
	ID     string `json:"id"`
	Class  string `json:"class"`
	Before int64  `json:"before"`
	After  int64  `json:"after"`
	Delta  int64  `json:"delta"`
	Unit   string `json:"unit"`
}

type ImprovementProof struct {
	ID     string `json:"id"`
	Choice string `json:"choice"`
	Passed bool   `json:"passed"`
}

type ImprovementTransition struct {
	Schema             string                 `json:"schema"`
	Decision           ImprovementDecision    `json:"decision"`
	ReasonCode         string                 `json:"reason_code"`
	Comparable         bool                   `json:"comparable"`
	ContractSchema     string                 `json:"contract_schema"`
	RegistryDigest     string                 `json:"registry_digest"`
	BeforeCompleted    int64                  `json:"before_completed"`
	AfterCompleted     int64                  `json:"after_completed"`
	Total              int64                  `json:"total"`
	CompletedDelta     int64                  `json:"completed_delta"`
	BeforeBasisPoints  int64                  `json:"before_basis_points"`
	AfterBasisPoints   int64                  `json:"after_basis_points"`
	BasisPointsDelta   int64                  `json:"basis_points_delta"`
	Gains              int64                  `json:"gains"`
	Regressions        int64                  `json:"regressions"`
	BeforeUnresolved   int64                  `json:"before_unresolved"`
	AfterUnresolved    int64                  `json:"after_unresolved"`
	Indicators         []ImprovementIndicator `json:"indicators"`
	Proofs             []ImprovementProof     `json:"proofs"`
	Digest             string                 `json:"digest"`
}

type improvementInspection struct {
	statuses  map[string]ImprovementEvidenceStatus
	unresolved int64
	reason     string
}

// EvaluateImprovementTransition proves improvement only from comparable integer
// snapshots. It never promotes prose, confidence, or an unknown value to success.
func EvaluateImprovementTransition(before, after ImprovementSnapshot) ImprovementTransition {
	result := ImprovementTransition{
		Schema:            ImprovementTransitionSchema,
		Decision:          ImprovementLowerResolution,
		ReasonCode:        "SNAPSHOT_NOT_COMPARABLE",
		ContractSchema:    before.ContractSchema,
		RegistryDigest:    before.RegistryDigest,
		BeforeCompleted:   before.Completed,
		AfterCompleted:    after.Completed,
		Total:             before.Total,
		BeforeBasisPoints: before.BasisPoints,
		AfterBasisPoints:  after.BasisPoints,
		Proofs:            improvementProofs(false, false, false, false),
	}

	beforeInspection := inspectImprovementSnapshot(before)
	if beforeInspection.reason != "" {
		result.ReasonCode = "BEFORE_" + beforeInspection.reason
		return sealImprovementTransition(result)
	}
	afterInspection := inspectImprovementSnapshot(after)
	if afterInspection.reason != "" {
		result.ReasonCode = "AFTER_" + afterInspection.reason
		return sealImprovementTransition(result)
	}
	if before.ContractSchema != after.ContractSchema {
		result.ReasonCode = "CONTRACT_SCHEMA_MISMATCH"
		return sealImprovementTransition(result)
	}
	if before.RegistryDigest != after.RegistryDigest {
		result.ReasonCode = "REGISTRY_DIGEST_MISMATCH"
		return sealImprovementTransition(result)
	}
	if before.Total != after.Total {
		result.ReasonCode = "DENOMINATOR_MISMATCH"
		return sealImprovementTransition(result)
	}
	for id := range beforeInspection.statuses {
		if _, ok := afterInspection.statuses[id]; !ok {
			result.ReasonCode = "OBLIGATION_SET_MISMATCH"
			return sealImprovementTransition(result)
		}
	}

	result.Comparable = true
	result.CompletedDelta = after.Completed - before.Completed
	result.BasisPointsDelta = after.BasisPoints - before.BasisPoints
	result.BeforeUnresolved = beforeInspection.unresolved
	result.AfterUnresolved = afterInspection.unresolved
	for id, beforeStatus := range beforeInspection.statuses {
		afterStatus := afterInspection.statuses[id]
		if beforeStatus != ImprovementSatisfied && afterStatus == ImprovementSatisfied {
			result.Gains++
		}
		if beforeStatus == ImprovementSatisfied && afterStatus != ImprovementSatisfied {
			result.Regressions++
		}
	}
	result.Indicators = improvementIndicators(result)
	resolved := result.BeforeUnresolved == 0 && result.AfterUnresolved == 0
	regressionFree := result.Regressions == 0
	result.Proofs = improvementProofs(true, true, resolved, regressionFree)

	switch {
	case !resolved:
		result.ReasonCode = "UNRESOLVED_EVIDENCE"
	case result.Regressions > 0 || result.CompletedDelta < 0 || result.BasisPointsDelta < 0:
		result.Decision = ImprovementRegressed
		result.ReasonCode = "NUMERIC_REGRESSION"
	case result.CompletedDelta > 0 && result.BasisPointsDelta > 0 && result.Gains > 0:
		result.Decision = ImprovementImproved
		result.ReasonCode = "IMPROVEMENT_PROVEN"
	case result.CompletedDelta == 0 && result.BasisPointsDelta == 0 && result.Gains == 0:
		result.Decision = ImprovementNoChange
		result.ReasonCode = "NO_NUMERIC_CHANGE"
	default:
		result.ReasonCode = "POSITIVE_DELTA_NOT_PROVEN"
	}

	return sealImprovementTransition(result)
}

func inspectImprovementSnapshot(snapshot ImprovementSnapshot) improvementInspection {
	inspection := improvementInspection{statuses: make(map[string]ImprovementEvidenceStatus)}
	if snapshot.ContractSchema != ImprovementSnapshotSchema {
		inspection.reason = "CONTRACT_SCHEMA_UNKNOWN"
		return inspection
	}
	if !validImprovementDigest(snapshot.RegistryDigest) {
		inspection.reason = "REGISTRY_DIGEST_INVALID"
		return inspection
	}
	if snapshot.Total <= 0 || int64(len(snapshot.Evidence)) != snapshot.Total {
		inspection.reason = "DENOMINATOR_INVALID"
		return inspection
	}
	var completed int64
	for _, evidence := range snapshot.Evidence {
		if strings.TrimSpace(evidence.ID) == "" {
			inspection.reason = "EVIDENCE_ID_EMPTY"
			return inspection
		}
		if _, exists := inspection.statuses[evidence.ID]; exists {
			inspection.reason = "EVIDENCE_ID_DUPLICATE"
			return inspection
		}
		switch evidence.Status {
		case ImprovementSatisfied:
			completed++
		case ImprovementNotSatisfied:
		case ImprovementUnresolved:
			inspection.unresolved++
		default:
			inspection.reason = "EVIDENCE_STATUS_UNKNOWN"
			return inspection
		}
		inspection.statuses[evidence.ID] = evidence.Status
	}
	if snapshot.Completed != completed {
		inspection.reason = "COMPLETED_COUNT_INVALID"
		return inspection
	}
	if snapshot.BasisPoints != completed*10_000/snapshot.Total {
		inspection.reason = "BASIS_POINTS_INVALID"
	}
	return inspection
}

func improvementIndicators(result ImprovementTransition) []ImprovementIndicator {
	return []ImprovementIndicator{
		{ID: "completed-obligations", Class: "OUTCOME", Before: result.BeforeCompleted, After: result.AfterCompleted, Delta: result.CompletedDelta, Unit: "OBLIGATION"},
		{ID: "readiness-basis-points", Class: "OUTCOME", Before: result.BeforeBasisPoints, After: result.AfterBasisPoints, Delta: result.BasisPointsDelta, Unit: "BASIS_POINT"},
		{ID: "newly-satisfied", Class: "DRIVER", After: result.Gains, Delta: result.Gains, Unit: "OBLIGATION"},
		{ID: "regressions", Class: "GUARDRAIL", After: result.Regressions, Delta: result.Regressions, Unit: "OBLIGATION"},
		{ID: "unresolved-evidence", Class: "GUARDRAIL", Before: result.BeforeUnresolved, After: result.AfterUnresolved, Delta: result.AfterUnresolved - result.BeforeUnresolved, Unit: "OBLIGATION"},
	}
}

func improvementProofs(comparable, arithmetic, resolved, regressionFree bool) []ImprovementProof {
	return []ImprovementProof{
		{ID: "fixed-contract", Choice: "FOUNDATION", Passed: comparable},
		{ID: "integer-arithmetic", Choice: "COHERENCE", Passed: arithmetic},
		{ID: "resolved-evidence", Choice: "FOUNDATION", Passed: resolved},
		{ID: "zero-regression", Choice: "REGRESSION", Passed: regressionFree},
	}
}

func validImprovementDigest(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+sha256.Size*2 {
		return false
	}
	payload := strings.TrimPrefix(value, prefix)
	if payload != strings.ToLower(payload) {
		return false
	}
	_, err := hex.DecodeString(payload)
	return err == nil
}

func sealImprovementTransition(result ImprovementTransition) ImprovementTransition {
	result.Digest = ""
	payload, _ := json.Marshal(result)
	sum := sha256.Sum256(payload)
	result.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return result
}
