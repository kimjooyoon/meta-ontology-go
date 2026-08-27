package verify

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/hygienicoriginidentity/consumer"
)

const (
	ExpectationPass    = "pass"
	ExpectationUnknown = "unknown"
)

// Validate reconstructs the report from the source before checking the seal.
// This rejects a coherent reseal of a tampered report, not only a bit flip.
func Validate(files fs.FS, report consumer.Report, expectation, expectedHead string) error {
	if err := validateIdentity(report, expectedHead); err != nil {
		return err
	}
	rebuilt, err := consumer.Evaluate(files, report.Source.Path, expectedHead, consumer.SnapshotPair{})
	if err != nil {
		return fmt.Errorf("reconstruct source report: %w", err)
	}
	if report.Source.RawDigest != rebuilt.Source.RawDigest || report.Source.SemanticDigest != rebuilt.Source.SemanticDigest {
		return fmt.Errorf("report source digest does not match reconstructed source")
	}
	if consumer.ContentDigest(report) != consumer.ContentDigest(rebuilt) {
		return fmt.Errorf("report content does not match reconstructed semantic report")
	}
	if report.ReceiptDigest == "" || report.ReceiptDigest != consumer.SealedDigest(report) {
		return fmt.Errorf("receipt digest is not sealed")
	}
	if err := validateAuthority(report.Authority); err != nil {
		return err
	}
	if err := validateCommonMetrics(report.Metrics); err != nil {
		return err
	}
	if err := validateTransitions(report); err != nil {
		return err
	}
	switch expectation {
	case ExpectationPass:
		return validatePass(report)
	case ExpectationUnknown:
		return validateUnknown(report)
	default:
		return fmt.Errorf("unknown validation expectation %q", expectation)
	}
}

func ValidateIntervention(files fs.FS, report, baseline consumer.Report, expectedHead string) error {
	if err := validateIdentity(report, expectedHead); err != nil {
		return err
	}
	rebuilt, err := consumer.Evaluate(files, report.Source.Path, expectedHead, consumer.SnapshotPair{})
	if err != nil {
		return fmt.Errorf("reconstruct intervention report: %w", err)
	}
	if consumer.ContentDigest(report) != consumer.ContentDigest(rebuilt) || report.Source.RawDigest != rebuilt.Source.RawDigest || report.Source.SemanticDigest != rebuilt.Source.SemanticDigest {
		return fmt.Errorf("intervention report does not match reconstructed source")
	}
	if report.ReceiptDigest == "" || report.ReceiptDigest != consumer.SealedDigest(report) {
		return fmt.Errorf("intervention receipt digest is not sealed")
	}
	if report.Source.SemanticDigest == baseline.Source.SemanticDigest || report.ReceiptDigest == baseline.ReceiptDigest {
		return fmt.Errorf("semantic intervention did not change semantic and receipt digests")
	}
	if err := validateAuthority(report.Authority); err != nil {
		return err
	}
	if err := validateCommonMetrics(report.Metrics); err != nil {
		return err
	}
	if report.Decision != consumer.DecisionRefuted || baseline.Decision != consumer.DecisionPass {
		return fmt.Errorf("intervention decision transition was not PASS to REFUTED")
	}
	if err := validateTransitions(report); err != nil {
		return err
	}
	if report.Metrics.TargetPreservationDischarged != 1 || report.Metrics.TargetPreservationRefuted != 1 || report.Metrics.TargetPreservationOpen != 0 {
		return fmt.Errorf("intervention target claims did not transition 1 discharged/1 refuted")
	}
	if report.Metrics.TargetPreservationBPS != 5000 {
		return fmt.Errorf("intervention target preservation did not fall to 5000 bps")
	}
	if len(report.Cases) != 2 || len(report.Claims) != 4 || report.Metrics.CapturedCaseTotal != 2 || report.Metrics.NonCapturedCaseTotal != 0 {
		return fmt.Errorf("intervention did not turn the target into a capture")
	}
	if err := validateInterventionCases(report.Cases); err != nil {
		return err
	}
	if err := validateInterventionClaims(report.Claims); err != nil {
		return fmt.Errorf("intervention claims: %w", err)
	}
	return nil
}

func validateInterventionClaims(claims []consumer.Claim) error {
	want := map[string]string{
		"captured.origin-identity":  consumer.StatusRefuted,
		"captured.scope-provenance": consumer.StatusRefuted,
		"hygienic.origin-identity":  consumer.StatusDischarged,
		"hygienic.scope-provenance": consumer.StatusRefuted,
	}
	if len(claims) != len(want) {
		return fmt.Errorf("intervention claim denominator changed")
	}
	for _, claim := range claims {
		if claim.ID == "" || claim.CaseID == "" || claim.Proposition == "" || claim.EvidenceDigest == "" || claim.Provenance == "" || claim.Status != want[claim.ID] {
			return fmt.Errorf("intervention claim %q is not the expected transition", claim.ID)
		}
	}
	return nil
}

func validateIdentity(report consumer.Report, expectedHead string) error {
	if report.SchemaVersion != consumer.SchemaVersion || report.Producer != consumer.Producer || report.Consumer != consumer.Consumer || report.MetaOperation != consumer.MetaOperation || report.ProofChoice != consumer.ProofChoice {
		return fmt.Errorf("receipt identity contract mismatch")
	}
	if report.Source.Path == "" || report.Source.HeadSHA != expectedHead || !validSHA(expectedHead) || !validDigest(report.Source.RawDigest) || !validDigest(report.Source.SemanticDigest) {
		return fmt.Errorf("receipt source subject is not exact")
	}
	return nil
}

func validateAuthority(authority consumer.Authority) error {
	if authority.RepositoryWrites != 0 || authority.RepositoryMutationAuthorized || !authority.SnapshotsEqual || authority.MutationCount != 0 || authority.PromotionCount != 0 || !validDigest(authority.BeforeSnapshotDigest) || !validDigest(authority.AfterSnapshotDigest) {
		return fmt.Errorf("CI before/after snapshot did not prove read-only execution")
	}
	return nil
}

func validateCommonMetrics(metrics consumer.Metrics) error {
	if metrics.FixedCaseDenominator != consumer.ExpectedCaseTotal || metrics.FixedClaimDenominator != consumer.ExpectedClaimTotal || metrics.FixedTargetPreservationDenominator != consumer.ExpectedTargetTotal || metrics.ObservedCaseTotal != consumer.ExpectedCaseTotal || metrics.SameSpellingCaseTotal != 2 || metrics.SourceCasesObserved != 2 || metrics.SourceCasesExpected != 2 || metrics.ProducerImportsObserved != 0 || metrics.ProducerImportsExpected != 0 || metrics.SemanticCausalityObserved != 1 || metrics.SemanticCausalityExpected != 1 || metrics.CommentInvarianceObserved != 1 || metrics.CommentInvarianceExpected != 1 || metrics.NonsemanticPreservationObserved != 1 || metrics.NonsemanticPreservationExpected != 1 || metrics.CanonicalConformanceObserved != 1 || metrics.CanonicalConformanceExpected != 1 || metrics.SubjectResolutionObserved != 1 || metrics.SubjectResolutionExpected != 1 || metrics.ControlCaptureObserved != 1 || metrics.ControlCaptureExpected != 1 || metrics.HygienicNonCaptureExpected != 1 || metrics.TargetPreservationExpected != 2 {
		return fmt.Errorf("fixed denominator metrics changed")
	}
	return nil
}

func validatePass(report consumer.Report) error {
	if report.Decision != consumer.DecisionPass || report.Resolution != consumer.ResolutionExact || len(report.Cases) != 2 || len(report.Claims) != 4 || len(report.Unknowns) != 0 || report.Metrics.CapturedCaseTotal != 1 || report.Metrics.NonCapturedCaseTotal != 1 || report.Metrics.ClassifiedClaimTotal != 4 || report.Metrics.DischargedClaimTotal != 2 || report.Metrics.RefutedClaimTotal != 2 || report.Metrics.OpenClaimTotal != 0 || report.Metrics.TargetPreservationObserved != 2 || report.Metrics.TargetPreservationDischarged != 2 || report.Metrics.TargetPreservationRefuted != 0 || report.Metrics.TargetPreservationOpen != 0 || report.Metrics.TargetPreservationBPS != 10000 {
		return fmt.Errorf("PASS report does not separate controls from target preservation")
	}
	if err := validateFixedCases(report.Cases, false); err != nil {
		return err
	}
	return validateFixedClaims(report.Claims, false)
}

func validateUnknown(report consumer.Report) error {
	if report.Decision != consumer.DecisionUnknown || report.Resolution != consumer.ResolutionLower || len(report.Cases) != 2 || len(report.Claims) != 5 || len(report.Unknowns) != 1 || report.Metrics.CapturedCaseTotal != 1 || report.Metrics.NonCapturedCaseTotal != 1 || report.Metrics.ClassifiedClaimTotal != 3 || report.Metrics.DischargedClaimTotal != 1 || report.Metrics.RefutedClaimTotal != 2 || report.Metrics.OpenClaimTotal != 2 || report.Metrics.TargetPreservationObserved != 2 || report.Metrics.TargetPreservationDischarged != 1 || report.Metrics.TargetPreservationOpen != 1 {
		return fmt.Errorf("UNKNOWN report does not retain the semantic provenance gap")
	}
	unknown := report.Unknowns[0]
	if unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || !validDigest(unknown.EvidenceDigest) || unknown.Provenance == "" {
		return fmt.Errorf("UNKNOWN coordinates/evidence/provenance are incomplete")
	}
	if err := validateFixedCases(report.Cases, true); err != nil {
		return err
	}
	return validateFixedClaims(report.Claims, true)
}

func validateFixedCases(cases []consumer.Case, unknown bool) error {
	byID := make(map[string]consumer.Case, len(cases))
	for _, item := range cases {
		if _, exists := byID[item.ID]; exists {
			return fmt.Errorf("duplicate case %q", item.ID)
		}
		byID[item.ID] = item
	}
	captured, capturedOK := byID["captured"]
	hygienic, hygienicOK := byID["hygienic"]
	if !capturedOK || !hygienicOK {
		return fmt.Errorf("fixed captured/hygienic cases are missing")
	}
	if captured.Spelling != "tmp" || !captured.SameSpelling || !captured.Control || captured.Target || !captured.Captured || captured.OriginIdentity != consumer.ConsumerBinding || captured.DefinitionScope != consumer.ConsumerCallSite || captured.UseScope != consumer.ConsumerCallSite || captured.ResolvedUseScope != consumer.ConsumerCallSite || captured.ResolvedIdentity != consumer.ConsumerBinding {
		return fmt.Errorf("captured control case is not the fixed semantic record")
	}
	if hygienic.Spelling != "tmp" || !hygienic.SameSpelling || hygienic.Control || !hygienic.Target || hygienic.OriginIdentity != consumer.ProducerExpansion || hygienic.DefinitionScope != consumer.FreshProducerScope || hygienic.UseScope != consumer.FreshProducerScope {
		return fmt.Errorf("hygienic target case is not the fixed semantic record")
	}
	if unknown {
		if hygienic.ResolvedUseScope != "" || hygienic.ResolvedIdentity != "" || hygienic.Captured || hygienic.ScopeProvenancePreserved {
			return fmt.Errorf("UNKNOWN case unexpectedly resolved provenance")
		}
		return nil
	}
	if hygienic.ResolvedUseScope != consumer.FreshProducerScope || hygienic.ResolvedIdentity != consumer.ProducerExpansion || hygienic.Captured || !hygienic.OriginIdentityPreserved || !hygienic.ScopeProvenancePreserved {
		return fmt.Errorf("hygienic target case is not the fixed non-capture record")
	}
	return nil
}

func validateInterventionCases(cases []consumer.Case) error {
	for _, item := range cases {
		if item.ID == "captured" {
			if !item.Captured || item.ResolvedIdentity != consumer.ConsumerBinding || item.ResolvedUseScope != consumer.ConsumerCallSite {
				return fmt.Errorf("intervention control case changed unexpectedly")
			}
			continue
		}
		if item.ID != "hygienic" || !item.Captured || item.ResolvedIdentity != consumer.ConsumerBinding || item.DefinitionScope != consumer.FreshProducerScope || item.UseScope != consumer.FreshProducerScope || item.ResolvedUseScope != consumer.ConsumerCallSite || item.ScopeProvenancePreserved {
			return fmt.Errorf("intervention did not change hygienic non-capture to consumer capture")
		}
	}
	return nil
}

func validateFixedClaims(claims []consumer.Claim, unknown bool) error {
	status := map[string]string{}
	for _, claim := range claims {
		if claim.ID == "" || claim.CaseID == "" || claim.Proposition == "" || claim.EvidenceDigest == "" || claim.Provenance == "" {
			return fmt.Errorf("claim %q is missing evidence or provenance", claim.ID)
		}
		if _, exists := status[claim.ID]; exists {
			return fmt.Errorf("duplicate claim %q", claim.ID)
		}
		status[claim.ID] = claim.Status
	}
	want := map[string]string{
		"captured.origin-identity":  consumer.StatusRefuted,
		"captured.scope-provenance": consumer.StatusRefuted,
		"hygienic.origin-identity":  consumer.StatusDischarged,
		"hygienic.scope-provenance": consumer.StatusDischarged,
	}
	if unknown {
		want["hygienic.scope-provenance"] = consumer.StatusOpen
		want["unknown.scope-provenance"] = consumer.StatusOpen
	}
	if len(status) != len(want) {
		return fmt.Errorf("claim denominator changed")
	}
	for id, expected := range want {
		if status[id] != expected {
			return fmt.Errorf("claim %q status=%q want %q", id, status[id], expected)
		}
	}
	return nil
}

func validateTransitions(report consumer.Report) error {
	if len(report.Transitions) != len(report.Claims) {
		return fmt.Errorf("claim transition ledger is not append-only complete")
	}
	status := map[string]string{}
	for _, claim := range report.Claims {
		status[claim.ID] = claim.Status
	}
	for index, transition := range report.Transitions {
		if transition.Sequence != index+1 || transition.Before != consumer.StatusOpen || transition.After != status[transition.ClaimID] || transition.Reason == "" || !validDigest(transition.EvidenceDigest) || transition.Provenance == "" {
			return fmt.Errorf("invalid append-only claim transition at sequence %d", transition.Sequence)
		}
	}
	return nil
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validDigest(value string) bool {
	raw := strings.TrimPrefix(value, "sha256:")
	if len(raw) != 64 || raw != strings.ToLower(raw) {
		return false
	}
	_, err := hex.DecodeString(raw)
	return err == nil
}
