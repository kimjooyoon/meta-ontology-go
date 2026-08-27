package hygienicoriginidentity

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// Validate checks the fixed contract and the sealed digest of an independent
// receipt. It does not reinterpret the source through the production parser.
func Validate(report Report, expectedUnknown bool, expectedHead string) error {
	if report.SchemaVersion != SchemaVersion || report.Producer != Producer || report.Consumer != Consumer ||
		report.MetaOperation != MetaOperation || report.ProofChoice != ProofChoice {
		return fmt.Errorf("receipt identity contract mismatch")
	}
	if report.Source.HeadSHA != expectedHead || !validSHA(expectedHead) {
		return fmt.Errorf("receipt head is not the expected exact subject")
	}
	if len(report.Cases) != ExpectedCaseTotal || len(report.Claims) != ExpectedClaimTotal {
		return fmt.Errorf("receipt denominator changed: cases %d/%d claims %d/%d", len(report.Cases), ExpectedCaseTotal, len(report.Claims), ExpectedClaimTotal)
	}
	if report.Metrics.FixedCaseDenominator != ExpectedCaseTotal || report.Metrics.FixedClaimDenominator != ExpectedClaimTotal ||
		report.Metrics.ObservedCaseTotal != ExpectedCaseTotal || report.Metrics.ObservedClaimTotal != ExpectedClaimTotal ||
		report.Metrics.SameSpellingCaseTotal != ExpectedCaseTotal || report.Metrics.CapturedCaseTotal != 1 ||
		report.Metrics.NonCapturedCaseTotal != 1 || report.Metrics.ClassifiedClaimTotal != ExpectedClaimTotal ||
		report.Metrics.DischargedClaimTotal != 2 || report.Metrics.RefutedClaimTotal != 2 ||
		report.Metrics.OpenClaimTotal != 0 || report.Metrics.ClassificationCoverageBPS != 10000 ||
		report.Metrics.PreservationSatisfactionBPS != 5000 {
		return fmt.Errorf("receipt metrics do not match fixed denominator")
	}
	if err := validateCases(report.Cases); err != nil {
		return err
	}
	if err := validateClaims(report.Claims); err != nil {
		return err
	}
	if report.Authority.RepositoryWrites != 0 || report.Authority.RepositoryMutationAuthorized {
		return fmt.Errorf("read-only authority was not preserved")
	}
	if expectedUnknown {
		if report.Decision != DecisionUnknown || report.Resolution != ResolutionLower || len(report.Unknowns) != ExpectedUnknownPath || report.Metrics.UnknownPathTotal != ExpectedUnknownPath {
			return fmt.Errorf("expected one explicit UNKNOWN path")
		}
		for _, unknown := range report.Unknowns {
			if unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" {
				return fmt.Errorf("UNKNOWN path lost stage/step/reason")
			}
		}
	} else if report.Decision != DecisionPass || report.Resolution != ResolutionExact || len(report.Unknowns) != 0 || report.Metrics.UnknownPathTotal != 0 {
		return fmt.Errorf("expected exact PASS without unknown paths")
	}
	if digest := report.ReceiptDigest; digest == "" || digest != sealedDigest(report) {
		return fmt.Errorf("receipt digest is not sealed")
	}
	return nil
}

func validateCases(cases []Case) error {
	byID := make(map[string]Case, len(cases))
	for _, item := range cases {
		if _, exists := byID[item.ID]; exists {
			return fmt.Errorf("duplicate receipt case %q", item.ID)
		}
		byID[item.ID] = item
	}
	captured, capturedOK := byID["captured"]
	hygienic, hygienicOK := byID["hygienic"]
	if !capturedOK || !hygienicOK {
		return fmt.Errorf("receipt cases are not the fixed captured/hygienic pair")
	}
	if err := validateCase(captured, false, true, ConsumerBinding, ConsumerCallSite); err != nil {
		return err
	}
	return validateCase(hygienic, true, false, ProducerExpansion, FreshProducerScope)
}

func validateCase(item Case, originPreserved, captured bool, expectedIdentity, expectedScope string) error {
	if item.Spelling != "tmp" || !item.SameSpelling || item.Captured != captured ||
		item.OriginIdentityPreserved != originPreserved || item.ScopeProvenancePreserved != originPreserved ||
		item.OriginIdentity != expectedIdentity || item.ResolvedIdentity != expectedIdentity || item.ScopeProvenance != expectedScope {
		return fmt.Errorf("case %q does not preserve the origin/scope contrast", item.ID)
	}
	return nil
}

func validateClaims(claims []Claim) error {
	statusByID := make(map[string]string, len(claims))
	for _, claim := range claims {
		if claim.Status != StatusOpen && claim.Status != StatusDischarged && claim.Status != StatusRefuted {
			return fmt.Errorf("claim %q has unknown status %q", claim.ID, claim.Status)
		}
		if _, exists := statusByID[claim.ID]; exists {
			return fmt.Errorf("duplicate receipt claim %q", claim.ID)
		}
		statusByID[claim.ID] = claim.Status
	}
	want := map[string]string{
		"captured.origin-identity":  StatusRefuted,
		"captured.scope-provenance": StatusRefuted,
		"hygienic.origin-identity":  StatusDischarged,
		"hygienic.scope-provenance": StatusDischarged,
	}
	if len(statusByID) != len(want) {
		return fmt.Errorf("claim denominator changed")
	}
	for id, expected := range want {
		if statusByID[id] != expected {
			return fmt.Errorf("claim %q status=%q want %q", id, statusByID[id], expected)
		}
	}
	return nil
}

func sealedDigest(report Report) string {
	report.ReceiptDigest = ""
	return digestJSON(report)
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

// Encode is the only receipt serialization path used by the command and CI.
func Encode(report Report) ([]byte, error) {
	value, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode receipt: %w", err)
	}
	return append(value, '\n'), nil
}
