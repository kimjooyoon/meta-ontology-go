package semanticdeltareceiptconsumer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
)

const (
	ClaimExpectationPath          = "examples/semantic-delta-receipt/claim-transition-expectations.json"
	claimExpectationSchema        = "gooo/semantic-delta-claim-transition-expectations/v1"
	claimExpectationDenominatorID = "gooo://semantic-delta-receipt-denominator/v2"
	claimExpectationFixedTotal    = 5
	claimIdentityStage            = "claim-identity"
	claimIdentityStep             = "compare-fixed-expectation"
	claimIdentityExactReason      = "FIXED_CLAIM_IDENTITY_EXACT"
	claimIdentityInventoryReason  = "CLAIM_ID_INVENTORY_MISMATCH"
	claimIdentityDigestReason     = "CLAIM_TRANSITION_IDENTITY_DIGEST_MISMATCH"
	claimExpectationReason        = "CLAIM_ID_EXPECTATION_UNAVAILABLE"
	claimExpectationContractError = "CLAIM_ID_EXPECTATION_INVENTORY_MISMATCH"
)

var claimExpectationCaseIDs = []string{"equivalent", "semantic-change", "value-program-change", "indeterminate", "ambiguous-match"}

type claimIdentityExpectation struct {
	ID                               string   `json:"id"`
	ExpectedClaimIDs                 []string `json:"expected_claim_ids"`
	ExpectedTransitionIdentityDigest string   `json:"expected_transition_identity_digest"`
}

type claimIdentityExpectationContract struct {
	Schema         string                     `json:"schema"`
	DenominatorID  string                     `json:"denominator_id"`
	FixedCaseTotal int                        `json:"fixed_case_total"`
	Cases          []claimIdentityExpectation `json:"cases"`
}

// ClaimIdentityComparison is validator evidence. Its observed values are
// reconstructed from raw source by this package; the producer receipt is not
// an input to the comparison.
type ClaimIdentityComparison struct {
	CaseID                           string   `json:"case_id"`
	DenominatorID                    string   `json:"denominator_id"`
	ExpectationPath                  string   `json:"expectation_path"`
	ExpectedClaimIDs                 []string `json:"expected_claim_ids"`
	ObservedClaimIDs                 []string `json:"observed_claim_ids"`
	ExpectedTransitionIdentityDigest string   `json:"expected_transition_identity_digest"`
	ObservedTransitionIdentityDigest string   `json:"observed_transition_identity_digest"`
	Decision                         string   `json:"decision"`
	Resolution                       string   `json:"resolution"`
	Stage                            string   `json:"stage"`
	Step                             string   `json:"step"`
	Reason                           string   `json:"reason"`
	FixedTotal                       int      `json:"fixed_total"`
	CoverageBPS                      int      `json:"coverage_bps"`
}

func ValidateFixedClaimIdentity(input Input) ClaimIdentityComparison {
	comparison := ClaimIdentityComparison{CaseID: input.CaseID, DenominatorID: claimExpectationDenominatorID, ExpectationPath: ClaimExpectationPath, Decision: decisionFailClosed, Resolution: resolutionLower, Stage: claimIdentityStage, Step: claimIdentityStep, Reason: claimExpectationReason}
	contract, err := readClaimIdentityExpectations()
	if err != nil {
		return comparison
	}
	expectation, ok := expectationForCase(contract, input.CaseID)
	if !ok {
		comparison.Reason = claimExpectationContractError
		return comparison
	}
	comparison.ExpectedClaimIDs = append([]string(nil), expectation.ExpectedClaimIDs...)
	comparison.ExpectedTransitionIdentityDigest = expectation.ExpectedTransitionIdentityDigest
	comparison.FixedTotal = len(expectation.ExpectedClaimIDs)
	beforeRaw, beforeErr := os.ReadFile(input.BeforePath)
	afterRaw, afterErr := os.ReadFile(input.AfterPath)
	if beforeErr != nil || afterErr != nil {
		return comparison
	}
	reconstructed := reconstructReceipt(input, beforeRaw, afterRaw)
	comparison.ObservedClaimIDs = append([]string(nil), reconstructed.ClaimIDInventory...)
	comparison.ObservedTransitionIdentityDigest = reconstructed.ClaimTransitionIdentityDigest
	if err := exactClaimIDInventory(comparison.ExpectedClaimIDs, comparison.ObservedClaimIDs); err != nil {
		comparison.Reason = claimIdentityInventoryReason
		return comparison
	}
	if comparison.ExpectedTransitionIdentityDigest != comparison.ObservedTransitionIdentityDigest {
		comparison.Reason = claimIdentityDigestReason
		return comparison
	}
	comparison.Decision, comparison.Resolution, comparison.Reason, comparison.CoverageBPS = "EXACT", resolutionExact, claimIdentityExactReason, 10000
	return comparison
}

func readClaimIdentityExpectations() (claimIdentityExpectationContract, error) {
	paths := []string{ClaimExpectationPath}
	if _, filename, _, ok := runtime.Caller(0); ok {
		paths = append(paths, filepath.Join(filepath.Dir(filename), "../../../../", ClaimExpectationPath))
	}
	var raw []byte
	var err error
	for _, path := range paths {
		raw, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return claimIdentityExpectationContract{}, err
	}
	var contract claimIdentityExpectationContract
	if err := json.Unmarshal(raw, &contract); err != nil {
		return claimIdentityExpectationContract{}, err
	}
	if contract.Schema != claimExpectationSchema || contract.DenominatorID != claimExpectationDenominatorID || contract.FixedCaseTotal != claimExpectationFixedTotal || exactStringInventory(claimExpectationCaseIDs, expectationIDs(contract.Cases)) != nil {
		return claimIdentityExpectationContract{}, fmt.Errorf("fixed claim expectation contract drift")
	}
	for _, expectation := range contract.Cases {
		if len(expectation.ExpectedClaimIDs) == 0 || exactClaimIDInventory(expectation.ExpectedClaimIDs, expectation.ExpectedClaimIDs) != nil || !sha256DigestPattern.MatchString(expectation.ExpectedTransitionIdentityDigest) {
			return claimIdentityExpectationContract{}, fmt.Errorf("invalid fixed claim expectation %q", expectation.ID)
		}
	}
	return contract, nil
}

func expectationForCase(contract claimIdentityExpectationContract, caseID string) (claimIdentityExpectation, bool) {
	for _, expectation := range contract.Cases {
		if expectation.ID == caseID {
			return expectation, true
		}
	}
	return claimIdentityExpectation{}, false
}

func expectationIDs(cases []claimIdentityExpectation) []string {
	ids := make([]string, 0, len(cases))
	for _, expectation := range cases {
		ids = append(ids, expectation.ID)
	}
	return ids
}

var sha256DigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func exactStringInventory(expected, observed []string) error {
	if len(expected) != len(observed) {
		return fmt.Errorf("inventory length mismatch")
	}
	counts := make(map[string]int, len(observed))
	for _, id := range observed {
		counts[id]++
	}
	for _, id := range expected {
		if counts[id] != 1 {
			return fmt.Errorf("missing, duplicate, or replaced inventory item %q", id)
		}
		counts[id] = 0
	}
	for _, count := range counts {
		if count != 0 {
			return fmt.Errorf("extra inventory item")
		}
	}
	return nil
}

func exactClaimIDInventory(expected, observed []string) error {
	if err := exactStringInventory(expected, observed); err != nil {
		return err
	}
	left, right := append([]string(nil), expected...), append([]string(nil), observed...)
	sort.Strings(left)
	sort.Strings(right)
	for index := range left {
		if left[index] != right[index] {
			return fmt.Errorf("claim inventory mismatch")
		}
	}
	return nil
}
