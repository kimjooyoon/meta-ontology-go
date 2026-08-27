package evidencequorumconsumer

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.Scope != Scope || !validDigest(report.SourceRawDigest) ||
		!validDigest(report.SourceSemanticDigest) || !validDigest(report.PolicySemanticDigest) ||
		report.Decision != DecisionPass || report.Summary.CasesSatisfied != report.Summary.CasesTotal ||
		report.Summary.CasesTotal == 0 || report.Summary.ConfidenceAggregated || report.RepositoryWrites != 0 || report.MutationAuthority {
		return fmt.Errorf("consumer report invariant failed")
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("consumer report digest mismatch")
	}
	for _, item := range report.Cases {
		if item.Status != "SATISFIED" || item.ConformanceDecision != DecisionPass || len(item.Claims) != 1 {
			return fmt.Errorf("case %q is not conformant", item.ID)
		}
		claim := item.Claims[0]
		if claim.ID == "" || claim.Producer == "" || claim.Consumer == "" || claim.MetaOperation == "" || claim.ProofChoice == "" ||
			claim.State != item.SubjectState || claim.SubjectDecision != item.SubjectDecision || claim.ObservationState != item.ObservationState ||
			len(claim.Transitions) != 1 {
			return fmt.Errorf("case %q claim transition invariant failed", item.ID)
		}
		transition := claim.Transitions[0]
		if transition.From != "OPEN" || transition.To != claim.State || !validDigest(transition.PreviousDigest) ||
			transition.Stage == "" || transition.Step == "" || transition.Reason == "" || len(transition.Provenance) == 0 {
			return fmt.Errorf("case %q append-only transition invariant failed", item.ID)
		}
		if item.ObservationState == ObservationUnknown && (item.SubjectState != StatusOpen || item.Resolution != ResolutionLower ||
			item.Stage != "UNKNOWN" || item.Step != "UNKNOWN") {
			return fmt.Errorf("case %q unknown resolution invariant failed", item.ID)
		}
	}
	return nil
}

func ValidatePolicy(policy evidencequorumpolicy.Policy) error {
	if !validDigest(policy.SemanticDigest) || policy.SourcePath == "" || policy.SourceEntry == "" || policy.Threshold < 1 ||
		policy.CaseDenominator < 1 || len(policy.RequiredRoles) == 0 || len(policy.RequiredPredicates) == 0 || policy.PriorClaimState != "OPEN" {
		return fmt.Errorf("policy invariant failed")
	}
	return nil
}

func ProvenanceGroupCount(receipts [][]byte) int {
	groups := map[string]bool{}
	for _, raw := range receipts {
		receipt, err := DecodeReceipt(raw)
		if err == nil && evidencequorumwire.Verify(receipt) {
			groups[receipt.ExecutableDigest+"|"+receipt.DependencyDigest] = true
		}
	}
	return len(groups)
}
