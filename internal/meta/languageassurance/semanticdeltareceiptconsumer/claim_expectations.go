package semanticdeltareceiptconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
)

const (
	ClaimExpectationPath           = "examples/semantic-delta-receipt/claim-transition-expectations.json"
	claimExpectationSchema         = "gooo/semantic-delta-claim-transition-expectations/v1"
	claimExpectationDenominatorID  = "gooo://semantic-delta-receipt-denominator/v2"
	claimCountContractVersion      = "v1"
	claimCountEquivalent           = 7
	claimCountSemanticChange       = 7
	claimCountValueProgramChange   = 7
	claimCountIndeterminate        = 3
	claimCountAmbiguousMatch       = 7
	claimExpectationFixedTotal     = claimCountEquivalent + claimCountSemanticChange + claimCountValueProgramChange + claimCountIndeterminate + claimCountAmbiguousMatch
	denominatorEvolutionRequired   = "REQUIRED_FOR_FIXED_CLAIM_COUNT_CHANGE"
	claimExpectationFixedCaseTotal = 5
	claimIdentityStage             = "claim-identity"
	claimIdentityStep              = "compare-fixed-expectation"
	claimIdentityPass              = "PASS"
	claimIdentityExactReason       = "FIXED_CLAIM_IDENTITY_EXACT"
	claimIdentityInventoryReason   = "CLAIM_ID_INVENTORY_MISMATCH"
	claimIdentityDigestReason      = "CLAIM_TRANSITION_IDENTITY_DIGEST_MISMATCH"
	claimExpectationReadStep       = "read-expectation-artifact"
	claimExpectationDecodeStep     = "decode-expectation-artifact"
	claimExpectationContractStep   = "validate-expectation-contract"
	claimExpectationMissingReason  = "CLAIM_EXPECTATION_ARTIFACT_UNAVAILABLE"
	claimExpectationInvalidReason  = "CLAIM_EXPECTATION_ARTIFACT_INVALID"
	claimExpectationContractError  = "CLAIM_EXPECTATION_CONTRACT_MISMATCH"
	claimSourceBeforeStep          = "read-before"
	claimSourceAfterStep           = "read-after"
	claimSourceBeforeReason        = "SOURCE_BEFORE_UNAVAILABLE"
	claimSourceAfterReason         = "SOURCE_AFTER_UNAVAILABLE"
	claimReconstructionStep        = "reconstruct-raw-pair"
	claimReconstructionReason      = "CLAIM_IDENTITY_RECONSTRUCTION_UNKNOWN"
)

var claimExpectationCaseIDs = []string{"equivalent", "semantic-change", "value-program-change", "indeterminate", "ambiguous-match"}

type claimIdentityExpectation struct {
	ID                               string   `json:"id"`
	ExpectedClaimIDs                 []string `json:"expected_claim_ids"`
	ExpectedTransitionIdentityDigest string   `json:"expected_transition_identity_digest"`
	ExpectedClaimTotal               int      `json:"expected_claim_total"`
	CaseRowDigest                    string   `json:"case_row_digest,omitempty"`
}

type claimIdentityExpectationContract struct {
	Schema                      string                     `json:"schema"`
	DenominatorID               string                     `json:"denominator_id"`
	ClaimCountContractVersion   string                     `json:"claim_count_contract_version"`
	FixedClaimTotal             int                        `json:"fixed_claim_total"`
	DenominatorEvolutionReceipt string                     `json:"denominator_evolution_receipt"`
	FixedCaseTotal              int                        `json:"fixed_case_total"`
	Cases                       []claimIdentityExpectation `json:"cases"`
}

type SourcePairObservation struct {
	BeforePath           string `json:"before_path"`
	AfterPath            string `json:"after_path"`
	BeforeRawDigest      string `json:"before_raw_digest"`
	AfterRawDigest       string `json:"after_raw_digest"`
	BeforeSemanticDigest string `json:"before_semantic_digest"`
	AfterSemanticDigest  string `json:"after_semantic_digest"`
}

// ClaimIdentityComparison is validator evidence. Its observed values are
// reconstructed from raw source by this package; the producer receipt is not
// an input to the comparison.
type ClaimIdentityComparison struct {
	CaseID                           string                `json:"case_id"`
	DenominatorID                    string                `json:"denominator_id"`
	ExpectationPath                  string                `json:"expectation_path"`
	ExpectationArtifactDigest        string                `json:"expectation_artifact_digest"`
	ExpectationCaseRowDigest         string                `json:"expectation_case_row_digest"`
	ExpectedClaimIDs                 []string              `json:"expected_claim_ids"`
	ObservedClaimIDs                 []string              `json:"observed_claim_ids"`
	ExpectedTransitionIdentityDigest string                `json:"expected_transition_identity_digest"`
	ObservedTransitionIdentityDigest string                `json:"observed_transition_identity_digest"`
	ObservedSourcePair               SourcePairObservation `json:"observed_source_pair"`
	Decision                         string                `json:"decision"`
	Resolution                       string                `json:"resolution"`
	Stage                            string                `json:"stage"`
	Step                             string                `json:"step"`
	Reason                           string                `json:"reason"`
	FixedTotal                       int                   `json:"fixed_total"`
	ExpectedClaimTotal               int                   `json:"expected_claim_total"`
	CoverageBPS                      int                   `json:"coverage_bps"`
}

func ValidateFixedClaimIdentity(input Input) ClaimIdentityComparison {
	comparison := ClaimIdentityComparison{CaseID: input.CaseID, DenominatorID: claimExpectationDenominatorID, ExpectationPath: ClaimExpectationPath, Decision: decisionFailClosed, Resolution: resolutionLower, Stage: claimIdentityStage, Step: claimExpectationReadStep, Reason: claimExpectationMissingReason}
	contract, artifactRaw, err := readClaimIdentityExpectations()
	if err != nil {
		if failure, ok := err.(claimExpectationFailure); ok {
			comparison.Stage, comparison.Step, comparison.Reason = failure.Stage, failure.Step, failure.Reason
		}
		return comparison
	}
	comparison.ExpectationArtifactDigest = digestBytes(artifactRaw)
	expectation, ok := expectationForCase(contract, input.CaseID)
	if !ok {
		comparison.Step, comparison.Reason = claimExpectationContractStep, claimExpectationContractError
		return comparison
	}
	comparison.ExpectedClaimIDs = append([]string(nil), expectation.ExpectedClaimIDs...)
	comparison.ExpectedTransitionIdentityDigest = expectation.ExpectedTransitionIdentityDigest
	comparison.ExpectationCaseRowDigest = expectation.CaseRowDigest
	comparison.FixedTotal = expectation.ExpectedClaimTotal
	comparison.ExpectedClaimTotal = expectation.ExpectedClaimTotal
	comparison.Step = claimIdentityStep
	comparison.ObservedSourcePair.BeforePath = input.BeforePath
	comparison.ObservedSourcePair.AfterPath = input.AfterPath
	beforeRaw, beforeErr := os.ReadFile(input.BeforePath)
	afterRaw, afterErr := os.ReadFile(input.AfterPath)
	if beforeErr != nil {
		comparison.Stage, comparison.Step, comparison.Reason = "source-pair", claimSourceBeforeStep, claimSourceBeforeReason
		return comparison
	}
	if afterErr != nil {
		comparison.Stage, comparison.Step, comparison.Reason = "source-pair", claimSourceAfterStep, claimSourceAfterReason
		return comparison
	}
	reconstructed := reconstructReceipt(input, beforeRaw, afterRaw)
	comparison.ObservedClaimIDs = append([]string(nil), reconstructed.ClaimIDInventory...)
	comparison.ObservedTransitionIdentityDigest = reconstructed.ClaimTransitionIdentityDigest
	comparison.ObservedSourcePair = SourcePairObservation{BeforePath: input.BeforePath, AfterPath: input.AfterPath, BeforeRawDigest: reconstructed.Before.SourceDigest, AfterRawDigest: reconstructed.After.SourceDigest, BeforeSemanticDigest: reconstructed.Before.SemanticDigest, AfterSemanticDigest: reconstructed.After.SemanticDigest}
	if comparison.ObservedTransitionIdentityDigest == "" {
		comparison.Stage, comparison.Step, comparison.Reason = claimIdentityStage, claimReconstructionStep, claimReconstructionReason
		return comparison
	}
	if err := exactClaimIDInventory(comparison.ExpectedClaimIDs, comparison.ObservedClaimIDs); err != nil {
		comparison.Reason = claimIdentityInventoryReason
		return comparison
	}
	if comparison.ExpectedTransitionIdentityDigest != comparison.ObservedTransitionIdentityDigest {
		comparison.Reason = claimIdentityDigestReason
		return comparison
	}
	comparison.Decision, comparison.Resolution, comparison.Reason, comparison.CoverageBPS = claimIdentityPass, resolutionExact, claimIdentityExactReason, 10000
	return comparison
}

type claimExpectationFailure struct {
	Stage  string
	Step   string
	Reason string
}

func (failure claimExpectationFailure) Error() string {
	return failure.Reason
}

func readClaimIdentityExpectations() (claimIdentityExpectationContract, []byte, error) {
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
		return claimIdentityExpectationContract{}, nil, claimExpectationFailure{Stage: claimIdentityStage, Step: claimExpectationReadStep, Reason: claimExpectationMissingReason}
	}
	contract, err := decodeClaimIdentityExpectations(raw)
	if err != nil {
		return claimIdentityExpectationContract{}, nil, claimExpectationFailure{Stage: claimIdentityStage, Step: claimExpectationDecodeStep, Reason: claimExpectationInvalidReason}
	}
	if err := validateClaimExpectationContract(contract); err != nil {
		return claimIdentityExpectationContract{}, nil, claimExpectationFailure{Stage: claimIdentityStage, Step: claimExpectationContractStep, Reason: claimExpectationContractError}
	}
	return contract, raw, nil
}

func decodeClaimIdentityExpectations(raw []byte) (claimIdentityExpectationContract, error) {
	var contract claimIdentityExpectationContract
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return claimIdentityExpectationContract{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return claimIdentityExpectationContract{}, fmt.Errorf("expectation artifact has trailing data")
	}
	return contract, nil
}

func validateClaimExpectationContract(contract claimIdentityExpectationContract) error {
	if contract.Schema != claimExpectationSchema || contract.DenominatorID != claimExpectationDenominatorID || contract.ClaimCountContractVersion != claimCountContractVersion || contract.FixedClaimTotal != claimExpectationFixedTotal || contract.DenominatorEvolutionReceipt != denominatorEvolutionRequired || contract.FixedCaseTotal != claimExpectationFixedCaseTotal || exactStringInventory(claimExpectationCaseIDs, expectationIDs(contract.Cases)) != nil {
		return fmt.Errorf("fixed claim expectation contract drift")
	}
	observedTotal := 0
	for _, expectation := range contract.Cases {
		fixedTotal, ok := fixedClaimCountForCase(expectation.ID)
		if !ok || expectation.ExpectedClaimTotal != fixedTotal || len(expectation.ExpectedClaimIDs) != fixedTotal || exactClaimIDInventory(expectation.ExpectedClaimIDs, expectation.ExpectedClaimIDs) != nil || !sha256DigestPattern.MatchString(expectation.ExpectedTransitionIdentityDigest) || expectation.CaseRowDigest != caseRowDigest(expectation) || !sha256DigestPattern.MatchString(expectation.CaseRowDigest) {
			return fmt.Errorf("invalid fixed claim expectation %q", expectation.ID)
		}
		observedTotal += expectation.ExpectedClaimTotal
	}
	if observedTotal != claimExpectationFixedTotal {
		return fmt.Errorf("fixed claim expectation total drift: got %d", observedTotal)
	}
	return nil
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

func FixedClaimCountForCase(caseID string) (int, bool) {
	return fixedClaimCountForCase(caseID)
}

func FixedClaimTotal() int {
	return claimExpectationFixedTotal
}

func ClaimCountContractVersion() string {
	return claimCountContractVersion
}

func fixedClaimCountForCase(caseID string) (int, bool) {
	switch caseID {
	case "equivalent":
		return claimCountEquivalent, true
	case "semantic-change":
		return claimCountSemanticChange, true
	case "value-program-change":
		return claimCountValueProgramChange, true
	case "indeterminate":
		return claimCountIndeterminate, true
	case "ambiguous-match":
		return claimCountAmbiguousMatch, true
	default:
		return 0, false
	}
}

func caseRowDigest(expectation claimIdentityExpectation) string {
	expectation.CaseRowDigest = ""
	expectation.ExpectedClaimIDs = append([]string(nil), expectation.ExpectedClaimIDs...)
	sort.Strings(expectation.ExpectedClaimIDs)
	return digestValue(expectation)
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
