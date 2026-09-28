package domaincapability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// InvestmentDisposition is a bounded action signal, not a semantic-completeness score.
type InvestmentDisposition string

const (
	InvestmentUnknown     InvestmentDisposition = "UNKNOWN"
	InvestmentDeferred    InvestmentDisposition = "DEFERRED"
	InvestmentInvestigate InvestmentDisposition = "INVESTIGATE"
	InvestmentInvest      InvestmentDisposition = "INVEST"
	InvestmentMaintain    InvestmentDisposition = "MAINTAIN"
)

// InvestmentInput makes scope, target policy, and evidence quality explicit.
// TargetNumerator/TargetDenominator is a declared investment target, not a
// claim that the domain is correct or complete.
type InvestmentInput struct {
	DomainID                  string
	ScopeDigest               string
	EvidenceDigest            string
	DeclaredCapabilityCount   int
	ObservedCapabilityCount   int
	UnresolvedCapabilityCount int
	TargetNumerator           int
	TargetDenominator         int
	EvidenceComplete          bool
	Comparable                bool
}

// InvestmentPlan explains where additional work is justified by the supplied
// evidence while preserving UNKNOWN and DEFERRED boundaries.
type InvestmentPlan struct {
	DomainID                  string
	Disposition               InvestmentDisposition
	CoverageNumerator         int
	CoverageDenominator       int
	TargetNumerator           int
	TargetDenominator         int
	GapNumerator              int
	UnresolvedCapabilityCount int
	FirstBoundary             string
	Reason                    string
	ScopeDigest               string
	EvidenceDigest            string
	PlanDigest                string
}

func PlanDomainInvestment(input InvestmentInput) InvestmentPlan {
	plan := InvestmentPlan{
		DomainID:                  strings.TrimSpace(input.DomainID),
		Disposition:               InvestmentUnknown,
		CoverageNumerator:         input.ObservedCapabilityCount,
		CoverageDenominator:       input.DeclaredCapabilityCount,
		TargetNumerator:           input.TargetNumerator,
		TargetDenominator:         input.TargetDenominator,
		UnresolvedCapabilityCount: input.UnresolvedCapabilityCount,
		ScopeDigest:               input.ScopeDigest,
		EvidenceDigest:            input.EvidenceDigest,
		Reason:                    "investment inputs are not sufficiently bound",
		FirstBoundary:             "investment_input",
	}

	switch {
	case plan.DomainID == "":
		plan.FirstBoundary = "domain_identity"
		plan.Reason = "a domain identity is required before allocating investment"
	case strings.TrimSpace(input.ScopeDigest) == "":
		plan.FirstBoundary = "scope_evidence"
		plan.Reason = "a stable scope digest is required before comparing capability coverage"
	case input.DeclaredCapabilityCount < 0 || input.ObservedCapabilityCount < 0 || input.UnresolvedCapabilityCount < 0:
		plan.FirstBoundary = "capability_counts"
		plan.Reason = "capability counts cannot be negative"
	case input.ObservedCapabilityCount > input.DeclaredCapabilityCount || input.UnresolvedCapabilityCount > input.DeclaredCapabilityCount:
		plan.FirstBoundary = "capability_counts"
		plan.Reason = "observed and unresolved capabilities cannot exceed declared scope"
	case input.TargetDenominator <= 0 || input.TargetNumerator < 0 || input.TargetNumerator > input.TargetDenominator:
		plan.FirstBoundary = "investment_target"
		plan.Reason = "the investment target must be an explicit ratio between zero and one"
	case !input.Comparable:
		plan.Disposition = InvestmentDeferred
		plan.FirstBoundary = "evidence_comparability"
		plan.Reason = "incomparable scope or evidence cannot justify an investment decision"
	case !input.EvidenceComplete:
		plan.Disposition = InvestmentDeferred
		plan.FirstBoundary = "evidence_completeness"
		plan.Reason = "incomplete evidence must remain deferred rather than become a score"
	case input.UnresolvedCapabilityCount > 0:
		plan.Disposition = InvestmentInvestigate
		plan.FirstBoundary = "unresolved_capability_boundary"
		plan.Reason = "unresolved capability boundaries require investigation before broadening scope"
	default:
		plan.GapNumerator = input.TargetNumerator*input.DeclaredCapabilityCount - input.ObservedCapabilityCount*input.TargetDenominator
		if plan.GapNumerator > 0 {
			plan.Disposition = InvestmentInvest
			plan.FirstBoundary = "observed_coverage_gap"
			plan.Reason = "observed capability coverage is below the declared investment target"
		} else {
			plan.Disposition = InvestmentMaintain
			plan.FirstBoundary = "none"
			plan.Reason = "observed capability coverage meets the declared investment target"
		}
	}
	plan.PlanDigest = investmentPlanDigest(plan)
	return plan
}

func (plan InvestmentPlan) Validate() error {
	if plan.Disposition != InvestmentUnknown && plan.Disposition != InvestmentDeferred && plan.Disposition != InvestmentInvestigate && plan.Disposition != InvestmentInvest && plan.Disposition != InvestmentMaintain {
		return fmt.Errorf("invalid investment disposition %q", plan.Disposition)
	}
	if plan.CoverageNumerator < 0 || plan.CoverageDenominator < 0 || plan.TargetNumerator < 0 || plan.TargetDenominator < 0 || plan.UnresolvedCapabilityCount < 0 {
		return fmt.Errorf("investment plan contains negative counts")
	}
	if plan.CoverageNumerator > plan.CoverageDenominator || plan.UnresolvedCapabilityCount > plan.CoverageDenominator {
		return fmt.Errorf("investment plan exceeds declared scope")
	}
	if plan.TargetDenominator == 0 || plan.TargetNumerator > plan.TargetDenominator {
		return fmt.Errorf("investment plan target is not a ratio")
	}
	if strings.TrimSpace(plan.FirstBoundary) == "" || strings.TrimSpace(plan.Reason) == "" {
		return fmt.Errorf("investment plan boundary or reason is missing")
	}
	if plan.PlanDigest != investmentPlanDigest(plan) {
		return fmt.Errorf("investment plan digest mismatch")
	}
	return nil
}

func investmentPlanDigest(plan InvestmentPlan) string {
	parts := []string{
		"meta-ontology-investment-plan",
		plan.DomainID,
		string(plan.Disposition),
		strconv.Itoa(plan.CoverageNumerator),
		strconv.Itoa(plan.CoverageDenominator),
		strconv.Itoa(plan.TargetNumerator),
		strconv.Itoa(plan.TargetDenominator),
		strconv.Itoa(plan.GapNumerator),
		strconv.Itoa(plan.UnresolvedCapabilityCount),
		plan.FirstBoundary,
		plan.Reason,
		plan.ScopeDigest,
		plan.EvidenceDigest,
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(digest[:])
}
