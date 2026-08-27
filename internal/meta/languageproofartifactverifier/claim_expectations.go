package languageproofartifactverifier

import (
	"encoding/json"
	"fmt"
)

// ClaimStateExpectation is validator-owned expectation data. Producer
// observation never reads this table: it records the states it actually
// derives, while the verifier compares those observations with this fixed
// external contract.
type ClaimStateExpectation struct {
	CaseID string   `json:"case_id"`
	States []string `json:"states"`
}

// ClaimStateExpectationPhase is one complete, phase-indexed 16 x 5 contract.
// The denominator is deliberately carried with the table so a missing case
// cannot silently shrink the measurement.
type ClaimStateExpectationPhase struct {
	FixedDenominator int                     `json:"fixed_denominator"`
	Cases            []ClaimStateExpectation `json:"cases"`
	Totals           claimStateTotalsJSON    `json:"totals"`
}

// ClaimStatePhaseTransition records the one evidence-time reclassification
// between the two fixed tables. PRELIMINARY has not supplied the external
// attachments for the byte-only negative case, so its recipe claim is OPEN;
// FINAL observes that absence at the independent recheck boundary and marks
// the same claim REFUTED. The case coordinate is preserved in both reports.
type ClaimStatePhaseTransition struct {
	FromPhase  string     `json:"from_phase"`
	ToPhase    string     `json:"to_phase"`
	CaseID     string     `json:"case_id"`
	ClaimID    string     `json:"claim_id"`
	From       string     `json:"from"`
	To         string     `json:"to"`
	Coordinate Coordinate `json:"coordinate"`
	Basis      string     `json:"basis"`
}

// ClaimStateExpectationDocument is the portable validator contract emitted
// for CI. It contains both phase tables and the exact cross-phase transition;
// it is not an input to Evaluate.
type ClaimStateExpectationDocument struct {
	Schema             string                                `json:"schema"`
	Version            int                                   `json:"version"`
	FixedDenominator   int                                   `json:"fixed_denominator"`
	Phases             map[string]ClaimStateExpectationPhase `json:"phases"`
	PhaseTransitions   []ClaimStatePhaseTransition           `json:"phase_transitions"`
	ClaimAdjudications []ClaimAdjudicationExpectation        `json:"claim_adjudications"`
}

type claimStateTotals struct {
	Discharged int
	Open       int
	Refuted    int
}

type claimStateMismatch struct {
	CaseID   string `json:"case_id"`
	ClaimID  string `json:"claim_id"`
	Actual   string `json:"actual"`
	Expected string `json:"expected"`
}

// ClaimAdjudicationExpectation fixes the causal coordinates for the two
// evidence-local cases where the same report contains multiple non-discharged
// claims. EvidenceDigestCount checks that the claim keeps its declared
// evidence links while the kernel decides whether those attachments are
// actually present and valid.
type ClaimAdjudicationExpectation struct {
	Phase               string     `json:"phase"`
	CaseID              string     `json:"case_id"`
	ClaimID             string     `json:"claim_id"`
	Status              string     `json:"status"`
	Resolution          string     `json:"resolution"`
	Reason              string     `json:"reason"`
	Coordinate          Coordinate `json:"coordinate"`
	EvidenceDigestCount int        `json:"evidence_digest_count"`
}

type claimAdjudicationMismatch struct {
	CaseID                string     `json:"case_id"`
	ClaimID               string     `json:"claim_id"`
	ActualStatus          string     `json:"actual_status"`
	ExpectedStatus        string     `json:"expected_status"`
	ActualResolution      string     `json:"actual_resolution"`
	ExpectedResolution    string     `json:"expected_resolution"`
	ActualReason          string     `json:"actual_reason"`
	ExpectedReason        string     `json:"expected_reason"`
	ActualCoordinate      Coordinate `json:"actual_coordinate"`
	ExpectedCoordinate    Coordinate `json:"expected_coordinate"`
	ActualEvidenceDigests []string   `json:"actual_evidence_digests"`
	ExpectedEvidenceCount int        `json:"expected_evidence_count"`
}

func fixedClaimAdjudicationTable(phase string) []ClaimAdjudicationExpectation {
	if phase != ProofPhasePreliminary && phase != ProofPhaseFinal {
		return nil
	}
	return []ClaimAdjudicationExpectation{
		{Phase: phase, CaseID: "missing-operation-evidence", ClaimID: "operation-receipt-bound", Status: "REFUTED", Resolution: "INVARIANT_ONLY", Reason: "CLAIM_REFUTED", Coordinate: Coordinate{"CONSUME_EVIDENCE", "operation-evidence", "PROOF_EVIDENCE_MISSING"}, EvidenceDigestCount: 1},
		{Phase: phase, CaseID: "missing-operation-evidence", ClaimID: "recipe-match", Status: "REFUTED", Resolution: "INVARIANT_ONLY", Reason: "CLAIM_REFUTED", Coordinate: Coordinate{"CONSUME_RECIPE", "recipe-evidence", "RECIPE_OPERATION_EVIDENCE_MISSING"}, EvidenceDigestCount: 3},
		{Phase: phase, CaseID: "unrelated-evidence-tamper", ClaimID: "no-byte-authority", Status: "REFUTED", Resolution: "INVARIANT_ONLY", Reason: "CLAIM_REFUTED", Coordinate: Coordinate{"CONSUME_INVARIANT", "invariant-evidence", "INVARIANT_EVIDENCE_NOT_PRESERVED"}, EvidenceDigestCount: 1},
		{Phase: phase, CaseID: "unrelated-evidence-tamper", ClaimID: "recipe-match", Status: "OPEN", Resolution: "LOWER_RESOLUTION", Reason: "CLAIM_PENDING", Coordinate: Coordinate{"CONSUME_RECIPE", "recipe-evidence", "RECIPE_INVARIANT_EVIDENCE_NOT_RESOLVED"}, EvidenceDigestCount: 3},
	}
}

type claimStateExpectationDetail struct {
	Phase            string               `json:"phase"`
	FixedDenominator int                  `json:"fixed_denominator"`
	Mismatches       []claimStateMismatch `json:"mismatches"`
	ActualTotals     claimStateTotalsJSON `json:"actual_totals"`
	ExpectedTotals   claimStateTotalsJSON `json:"expected_totals"`
}

type claimStateTotalsJSON struct {
	Discharged int `json:"discharged"`
	Open       int `json:"open"`
	Refuted    int `json:"refuted"`
}

// The FINAL table is the fixed 80-instance contract used after the
// independent consumer recheck. The PRELIMINARY table is independently
// enumerated rather than derived from observed totals; it differs only at
// the declared evidence-time transition above.
var fixedClaimStateExpectationsFinal = []ClaimStateExpectation{
	{CaseID: "valid-proof-carrying-artifact", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED"}},
	{CaseID: "tampered-evidence", States: []string{"REFUTED", "OPEN", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "coherent-tamper-reconstruction", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "missing-operation-evidence", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "bytes-only-no-authority", States: []string{"OPEN", "OPEN", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "independent-recipe-mismatch", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "recipe-only-mismatch", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "missing-attachment", States: []string{"DISCHARGED", "OPEN", "DISCHARGED", "OPEN", "OPEN"}},
	{CaseID: "wrong-attachment-digest", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "unrelated-evidence-tamper", States: []string{"DISCHARGED", "DISCHARGED", "REFUTED", "OPEN", "OPEN"}},
	{CaseID: "stale-head", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED"}},
	{CaseID: "unauthorized-consumer", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED"}},
	{CaseID: "coherent-claim-proposition-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-dependency-tamper", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-proof-choice-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-target-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
}

var fixedClaimStateExpectationsPreliminary = []ClaimStateExpectation{
	{CaseID: "valid-proof-carrying-artifact", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED"}},
	{CaseID: "tampered-evidence", States: []string{"REFUTED", "OPEN", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "coherent-tamper-reconstruction", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "missing-operation-evidence", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "bytes-only-no-authority", States: []string{"OPEN", "OPEN", "DISCHARGED", "OPEN", "OPEN"}},
	{CaseID: "independent-recipe-mismatch", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "recipe-only-mismatch", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "missing-attachment", States: []string{"DISCHARGED", "OPEN", "DISCHARGED", "OPEN", "OPEN"}},
	{CaseID: "wrong-attachment-digest", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "unrelated-evidence-tamper", States: []string{"DISCHARGED", "DISCHARGED", "REFUTED", "OPEN", "OPEN"}},
	{CaseID: "stale-head", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED"}},
	{CaseID: "unauthorized-consumer", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED"}},
	{CaseID: "coherent-claim-proposition-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-dependency-tamper", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-proof-choice-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-target-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
}

func fixedClaimStateTable(phase string) []ClaimStateExpectation {
	var source []ClaimStateExpectation
	switch phase {
	case ProofPhasePreliminary:
		source = fixedClaimStateExpectationsPreliminary
	case ProofPhaseFinal:
		source = fixedClaimStateExpectationsFinal
	default:
		return nil
	}
	result := make([]ClaimStateExpectation, len(source))
	for index, item := range source {
		result[index] = ClaimStateExpectation{CaseID: item.CaseID, States: append([]string(nil), item.States...)}
	}
	return result
}

func phaseClaimStateTransitions() []ClaimStatePhaseTransition {
	return []ClaimStatePhaseTransition{{
		FromPhase: ProofPhasePreliminary, ToPhase: ProofPhaseFinal,
		CaseID: "bytes-only-no-authority", ClaimID: "recipe-match", From: "OPEN", To: "REFUTED",
		Coordinate: Coordinate{"CONSUME_INPUT", "external-evidence", "ARTIFACT_BYTES_NOT_AUTHORITY"},
		Basis:      "FINAL_RECHECK_OBSERVES_MISSING_EXTERNAL_EVIDENCE",
	}}
}

// ClaimStateExpectations exposes both validator-owned phase tables and the
// declared phase transition for CI evidence. Evaluate never calls it.
func ClaimStateExpectations() ClaimStateExpectationDocument {
	phases := map[string]ClaimStateExpectationPhase{}
	claimAdjudications := make([]ClaimAdjudicationExpectation, 0, 8)
	for _, phase := range []string{ProofPhasePreliminary, ProofPhaseFinal} {
		table := fixedClaimStateTable(phase)
		totals := fixedClaimStateTotals(phase)
		phases[phase] = ClaimStateExpectationPhase{FixedDenominator: CaseTotal * ClaimTemplateTotal, Cases: table, Totals: claimStateTotalsJSON{Discharged: totals.Discharged, Open: totals.Open, Refuted: totals.Refuted}}
		claimAdjudications = append(claimAdjudications, fixedClaimAdjudicationTable(phase)...)
	}
	return ClaimStateExpectationDocument{Schema: "gooo/language-proof-carrying-artifact-claim-state-expectations/v1", Version: 1, FixedDenominator: CaseTotal * ClaimTemplateTotal, Phases: phases, PhaseTransitions: phaseClaimStateTransitions(), ClaimAdjudications: claimAdjudications}
}

func fixedClaimStateTotals(phase string) claimStateTotals {
	var totals claimStateTotals
	for _, item := range fixedClaimStateTable(phase) {
		for _, state := range item.States {
			switch state {
			case "DISCHARGED":
				totals.Discharged++
			case "OPEN":
				totals.Open++
			case "REFUTED":
				totals.Refuted++
			default:
				panic("invalid fixed claim state expectation")
			}
		}
	}
	return totals
}

func fixedClaimStateTotalsJSON(phase string) claimStateTotalsJSON {
	totals := fixedClaimStateTotals(phase)
	return claimStateTotalsJSON{Discharged: totals.Discharged, Open: totals.Open, Refuted: totals.Refuted}
}

func observedClaimStateTotals(cases []CaseResult) claimStateTotalsJSON {
	var totals claimStateTotalsJSON
	for _, item := range cases {
		for _, claim := range item.Claims {
			switch claim.Status {
			case "DISCHARGED":
				totals.Discharged++
			case "OPEN":
				totals.Open++
			case "REFUTED":
				totals.Refuted++
			}
		}
	}
	return totals
}

func validateClaimStateExpectations(phase string, cases []CaseResult) error {
	expectations := fixedClaimStateTable(phase)
	if len(expectations) != CaseTotal || len(expectations)*ClaimTemplateTotal != CaseTotal*ClaimTemplateTotal {
		return fmt.Errorf("fixed %s claim state expectation inventory has denominator %d, want %d", phase, len(expectations)*ClaimTemplateTotal, CaseTotal*ClaimTemplateTotal)
	}
	if len(cases) != CaseTotal {
		return fmt.Errorf("observed %s claim state inventory has denominator %d, want %d", phase, len(cases)*ClaimTemplateTotal, CaseTotal*ClaimTemplateTotal)
	}
	claimIDs := make([]string, 0, ClaimTemplateTotal)
	for _, spec := range claimSpecs() {
		claimIDs = append(claimIDs, spec.ID)
	}
	mismatches := make([]claimStateMismatch, 0)
	for caseIndex, expected := range expectations {
		if expected.CaseID != CaseIDs()[caseIndex] || len(expected.States) != ClaimTemplateTotal || cases[caseIndex].ID != expected.CaseID {
			return fmt.Errorf("fixed %s claim state expectation case inventory mismatch at index %d", phase, caseIndex)
		}
		for claimIndex, expectedState := range expected.States {
			if cases[caseIndex].Claims[claimIndex].ID != claimIDs[claimIndex] {
				return fmt.Errorf("fixed %s claim state expectation claim inventory mismatch: %s", phase, expected.CaseID)
			}
			if cases[caseIndex].Claims[claimIndex].Status != expectedState {
				mismatches = append(mismatches, claimStateMismatch{CaseID: expected.CaseID, ClaimID: claimIDs[claimIndex], Actual: cases[caseIndex].Claims[claimIndex].Status, Expected: expectedState})
			}
		}
	}
	if len(mismatches) == 0 {
		return nil
	}
	detail, err := json.Marshal(claimStateExpectationDetail{Phase: phase, FixedDenominator: CaseTotal * ClaimTemplateTotal, Mismatches: mismatches, ActualTotals: observedClaimStateTotals(cases), ExpectedTotals: fixedClaimStateTotalsJSON(phase)})
	if err != nil {
		return fmt.Errorf("claim state expectation mismatch")
	}
	return &ValidationError{Coordinate: Coordinate{"VERIFY_CLAIM_STATES", "case-claim-state", "CLAIM_STATE_EXPECTATION_MISMATCH"}, Detail: string(detail)}
}

func validateClaimAdjudicationExpectations(phase string, cases []CaseResult) error {
	mismatches := make([]claimAdjudicationMismatch, 0)
	for _, expected := range fixedClaimAdjudicationTable(phase) {
		var actual *ClaimResult
		for caseIndex := range cases {
			if cases[caseIndex].ID != expected.CaseID {
				continue
			}
			for claimIndex := range cases[caseIndex].Claims {
				if cases[caseIndex].Claims[claimIndex].ID == expected.ClaimID {
					actual = &cases[caseIndex].Claims[claimIndex]
					break
				}
			}
		}
		if actual == nil {
			mismatches = append(mismatches, claimAdjudicationMismatch{CaseID: expected.CaseID, ClaimID: expected.ClaimID, ExpectedStatus: expected.Status, ExpectedResolution: expected.Resolution, ExpectedReason: expected.Reason, ExpectedCoordinate: expected.Coordinate, ExpectedEvidenceCount: expected.EvidenceDigestCount})
			continue
		}
		validEvidence := len(actual.EvidenceDigests) == expected.EvidenceDigestCount
		for _, digest := range actual.EvidenceDigests {
			validEvidence = validEvidence && validDigest(digest)
		}
		if actual.Status != expected.Status || actual.Resolution != expected.Resolution || actual.Reason != expected.Reason || actual.Coordinate != expected.Coordinate || !validEvidence {
			mismatches = append(mismatches, claimAdjudicationMismatch{
				CaseID: expected.CaseID, ClaimID: expected.ClaimID,
				ActualStatus: actual.Status, ExpectedStatus: expected.Status,
				ActualResolution: actual.Resolution, ExpectedResolution: expected.Resolution,
				ActualReason: actual.Reason, ExpectedReason: expected.Reason,
				ActualCoordinate: actual.Coordinate, ExpectedCoordinate: expected.Coordinate,
				ActualEvidenceDigests: actual.EvidenceDigests, ExpectedEvidenceCount: expected.EvidenceDigestCount,
			})
		}
	}
	if len(mismatches) == 0 {
		return nil
	}
	detail, err := json.Marshal(struct {
		Phase      string                      `json:"phase"`
		Mismatches []claimAdjudicationMismatch `json:"mismatches"`
	}{Phase: phase, Mismatches: mismatches})
	if err != nil {
		return fmt.Errorf("claim adjudication expectation mismatch")
	}
	return &ValidationError{Coordinate: Coordinate{"VERIFY_CLAIM_STATES", "claim-adjudication", "CLAIM_ADJUDICATION_EXPECTATION_MISMATCH"}, Detail: string(detail)}
}

func validatePhaseState(cases []CaseResult, phase string) error {
	for _, transition := range phaseClaimStateTransitions() {
		want := transition.To
		if phase == transition.FromPhase {
			want = transition.From
		}
		for _, item := range cases {
			if item.ID != transition.CaseID {
				continue
			}
			for _, claim := range item.Claims {
				if claim.ID == transition.ClaimID && claim.Status != want {
					return &ValidationError{Coordinate: Coordinate{"VERIFY_CLAIM_STATES", "phase-transition", "CLAIM_PHASE_TRANSITION_MISMATCH"}, Detail: fmt.Sprintf(`{"phase":%q,"case_id":%q,"claim_id":%q,"actual":%q,"expected":%q,"stage":%q,"step":%q,"reason":%q}`, phase, transition.CaseID, transition.ClaimID, claim.Status, want, transition.Coordinate.Stage, transition.Coordinate.Step, transition.Coordinate.Reason)}
				}
			}
		}
	}
	return nil
}

func projectCasesForPhase(cases []CaseResult, phase string) []CaseResult {
	projected := make([]CaseResult, len(cases))
	copy(projected, cases)
	for index := range projected {
		projected[index].Claims = append([]ClaimResult(nil), cases[index].Claims...)
	}
	for _, transition := range phaseClaimStateTransitions() {
		want := transition.To
		if phase == transition.FromPhase {
			want = transition.From
		}
		for caseIndex := range projected {
			if projected[caseIndex].ID != transition.CaseID {
				continue
			}
			for claimIndex := range projected[caseIndex].Claims {
				claim := &projected[caseIndex].Claims[claimIndex]
				if claim.ID != transition.ClaimID {
					continue
				}
				claim.Status = want
				switch want {
				case "OPEN":
					claim.Resolution, claim.Reason, claim.Provenance = "LOWER_RESOLUTION", "CLAIM_PENDING", "consumer-observation"
				case "REFUTED":
					claim.Resolution, claim.Reason, claim.Provenance = "INVARIANT_ONLY", "CLAIM_REFUTED", "consumer-observation"
				case "DISCHARGED":
					claim.Resolution, claim.Reason, claim.Provenance = "EXACT", "CLAIM_DISCHARGED", "consumer-canonical-recipe-v2"
				}
				claim.Coordinate = transition.Coordinate
				claim.StateDigest = claimStateDigest(*claim)
			}
		}
	}
	return projected
}

func validatePhaseTransitionPair(preliminary, final []CaseResult) error {
	if len(preliminary) != CaseTotal || len(final) != CaseTotal {
		return &ValidationError{Coordinate: Coordinate{"VERIFY_CLAIM_STATES", "phase-transition", "CLAIM_PHASE_TRANSITION_MISMATCH"}, Detail: "phase transition comparison requires 80 claim instances in each phase"}
	}
	transitions := phaseClaimStateTransitions()
	if len(transitions) != 1 {
		return &ValidationError{Coordinate: Coordinate{"VERIFY_CLAIM_STATES", "phase-transition", "CLAIM_PHASE_TRANSITION_MISMATCH"}, Detail: "phase transition inventory is not exactly one declared transition"}
	}
	declared := transitions[0]
	for caseIndex := range preliminary {
		if preliminary[caseIndex].ID != final[caseIndex].ID || len(preliminary[caseIndex].Claims) != ClaimTemplateTotal || len(final[caseIndex].Claims) != ClaimTemplateTotal {
			return &ValidationError{Coordinate: Coordinate{"VERIFY_CLAIM_STATES", "phase-transition", "CLAIM_PHASE_TRANSITION_MISMATCH"}, Detail: "phase transition case or claim inventory mismatch"}
		}
		for claimIndex := range preliminary[caseIndex].Claims {
			before := preliminary[caseIndex].Claims[claimIndex]
			after := final[caseIndex].Claims[claimIndex]
			expectedBefore, expectedAfter := before.Status, before.Status
			if preliminary[caseIndex].ID == declared.CaseID && before.ID == declared.ClaimID {
				expectedBefore, expectedAfter = declared.From, declared.To
			}
			if before.Status != expectedBefore || after.Status != expectedAfter ||
				(preliminary[caseIndex].ID == declared.CaseID && before.ID == declared.ClaimID &&
					(preliminary[caseIndex].Coordinate != declared.Coordinate || final[caseIndex].Coordinate != declared.Coordinate)) {
				return &ValidationError{Coordinate: Coordinate{"VERIFY_CLAIM_STATES", "phase-transition", "CLAIM_PHASE_TRANSITION_MISMATCH"}, Detail: fmt.Sprintf(`{"from_phase":%q,"to_phase":%q,"case_id":%q,"claim_id":%q,"actual_from":%q,"actual_to":%q,"expected_from":%q,"expected_to":%q,"stage":%q,"step":%q,"reason":%q}`, declared.FromPhase, declared.ToPhase, declared.CaseID, declared.ClaimID, before.Status, after.Status, expectedBefore, expectedAfter, declared.Coordinate.Stage, declared.Coordinate.Step, declared.Coordinate.Reason)}
			}
		}
	}
	return nil
}
