package domaincapability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type DomainMeasurementStatus string

const (
	DomainMeasurementUnknown  DomainMeasurementStatus = "UNKNOWN"
	DomainMeasurementDeferred DomainMeasurementStatus = "DEFERRED"
	DomainMeasurementBound    DomainMeasurementStatus = "BOUND"
)

type DomainInvestmentDecision string

const (
	DomainDecisionBindEvidence      DomainInvestmentDecision = "BIND_EVIDENCE"
	DomainDecisionAddUseCase        DomainInvestmentDecision = "ADD_USE_CASE"
	DomainDecisionDeferInvestment   DomainInvestmentDecision = "DEFER_INVESTMENT"
	DomainDecisionHoldDeclaredScope DomainInvestmentDecision = "HOLD_DECLARED_SCOPE"
)

type DomainMeasurementInput struct {
	Domain                    string
	ScopeDigest               string
	SourceDigest              string
	CapabilityDigest          string
	ObservationEvidenceDigest string
	ObservedUseCases          int
	RequiredUseCases          int
	InvestmentBudget          int
	EstimatedInvestment       int
}

type DomainMeasurementPlan struct {
	Status                    DomainMeasurementStatus
	Domain                    string
	ScopeDigest               string
	SourceDigest              string
	CapabilityDigest          string
	ObservationEvidenceDigest string
	ObservedUseCases          int
	RequiredUseCases          int
	RemainingUseCases         int
	InvestmentBudget          int
	EstimatedInvestment       int
	Decision                  DomainInvestmentDecision
	MetricMeaning             string
	ScopePolicy               string
	FirstBoundary             string
	Reason                    string
	ReviewOnly                bool
	PlanDigest                string
}

// MeasureDomain binds an explicit domain scope to observable evidence and an
// investment decision. Counts describe evidence coverage for that scope; they
// are not a claim that the language is complete.
func MeasureDomain(input DomainMeasurementInput) DomainMeasurementPlan {
	plan := DomainMeasurementPlan{
		Status:                    DomainMeasurementUnknown,
		Domain:                    strings.TrimSpace(input.Domain),
		ScopeDigest:               input.ScopeDigest,
		SourceDigest:              input.SourceDigest,
		CapabilityDigest:          input.CapabilityDigest,
		ObservationEvidenceDigest: input.ObservationEvidenceDigest,
		ObservedUseCases:          input.ObservedUseCases,
		RequiredUseCases:          input.RequiredUseCases,
		InvestmentBudget:          input.InvestmentBudget,
		EstimatedInvestment:       input.EstimatedInvestment,
		Decision:                  DomainDecisionBindEvidence,
		MetricMeaning:             "evidence coverage for a declared domain scope; not language completeness",
		ScopePolicy:               "declared_scope_only",
		FirstBoundary:             "domain",
		Reason:                    "an explicit domain scope is required before measurement",
		ReviewOnly:                true,
	}
	switch {
	case plan.Domain == "":
		plan.FirstBoundary = "domain"
		plan.Reason = "an explicit domain name is required"
	case !validReverseObservationDigest(plan.ScopeDigest):
		plan.Status = DomainMeasurementDeferred
		plan.FirstBoundary = "scope"
		plan.Reason = "the declared domain scope must be provenance-bound"
	case !validReverseObservationDigest(plan.SourceDigest):
		plan.Status = DomainMeasurementDeferred
		plan.FirstBoundary = "source"
		plan.Reason = "the measured source must be provenance-bound"
	case !validReverseObservationDigest(plan.CapabilityDigest):
		plan.Status = DomainMeasurementDeferred
		plan.FirstBoundary = "capability"
		plan.Reason = "the capability surface must be provenance-bound"
	case !validReverseObservationDigest(plan.ObservationEvidenceDigest):
		plan.Status = DomainMeasurementDeferred
		plan.FirstBoundary = "observation"
		plan.Reason = "observed use-case evidence must be provenance-bound"
	case plan.RequiredUseCases <= 0:
		plan.Status = DomainMeasurementDeferred
		plan.FirstBoundary = "measurement_policy"
		plan.Reason = "the declared scope must contain a positive use-case target"
	case plan.ObservedUseCases < 0 || plan.InvestmentBudget < 0 || plan.EstimatedInvestment < 0:
		plan.Status = DomainMeasurementDeferred
		plan.FirstBoundary = "measurement_values"
		plan.Reason = "measurement values cannot be negative"
	default:
		plan.Status = DomainMeasurementBound
		plan.FirstBoundary = ""
		plan.RemainingUseCases = plan.RequiredUseCases - plan.ObservedUseCases
		if plan.RemainingUseCases < 0 {
			plan.RemainingUseCases = 0
		}
		switch {
		case plan.EstimatedInvestment > plan.InvestmentBudget:
			plan.Decision = DomainDecisionDeferInvestment
			plan.Reason = "the evidence is bound but the proposed investment exceeds the declared budget"
		case plan.ObservedUseCases < plan.RequiredUseCases:
			plan.Decision = DomainDecisionAddUseCase
			plan.Reason = "the evidence is bound and more in-scope use cases are needed before review"
		default:
			plan.Decision = DomainDecisionHoldDeclaredScope
			plan.Reason = "the evidence is bound for the declared scope and should be reviewed before expansion"
		}
	}
	plan.PlanDigest = plan.digest()
	return plan
}

func (plan DomainMeasurementPlan) Validate() error {
	if !plan.ReviewOnly {
		return fmt.Errorf("domain measurement plan is not review-only")
	}
	if plan.MetricMeaning != "evidence coverage for a declared domain scope; not language completeness" {
		return fmt.Errorf("domain measurement plan has an invalid metric meaning")
	}
	if plan.ScopePolicy != "declared_scope_only" {
		return fmt.Errorf("domain measurement plan escaped its declared scope")
	}
	if !validReverseObservationDigest(plan.PlanDigest) {
		return fmt.Errorf("domain measurement plan digest is invalid")
	}
	switch plan.Status {
	case DomainMeasurementUnknown, DomainMeasurementDeferred:
		if strings.TrimSpace(plan.FirstBoundary) == "" || strings.TrimSpace(plan.Reason) == "" {
			return fmt.Errorf("unresolved domain measurement lost its first boundary")
		}
	case DomainMeasurementBound:
		if plan.FirstBoundary != "" {
			return fmt.Errorf("bound domain measurement retained an unresolved boundary")
		}
		for name, value := range map[string]string{
			"scope":       plan.ScopeDigest,
			"source":      plan.SourceDigest,
			"capability":  plan.CapabilityDigest,
			"observation": plan.ObservationEvidenceDigest,
		} {
			if !validReverseObservationDigest(value) {
				return fmt.Errorf("bound domain measurement %s digest is invalid", name)
			}
		}
		if plan.RequiredUseCases <= 0 || plan.ObservedUseCases < 0 || plan.InvestmentBudget < 0 || plan.EstimatedInvestment < 0 {
			return fmt.Errorf("bound domain measurement has invalid values")
		}
		remaining := plan.RequiredUseCases - plan.ObservedUseCases
		if remaining < 0 {
			remaining = 0
		}
		if plan.RemainingUseCases != remaining {
			return fmt.Errorf("bound domain measurement has an inconsistent remaining-use-case count")
		}
	default:
		return fmt.Errorf("domain measurement status %q is invalid", plan.Status)
	}
	switch plan.Decision {
	case DomainDecisionBindEvidence, DomainDecisionAddUseCase, DomainDecisionDeferInvestment, DomainDecisionHoldDeclaredScope:
	default:
		return fmt.Errorf("domain measurement decision %q is invalid", plan.Decision)
	}
	if plan.digest() != plan.PlanDigest {
		return fmt.Errorf("domain measurement plan digest does not match")
	}
	return nil
}

func (plan DomainMeasurementPlan) digest() string {
	parts := []string{
		"meta-ontology-domain-measurement",
		string(plan.Status),
		plan.Domain,
		plan.ScopeDigest,
		plan.SourceDigest,
		plan.CapabilityDigest,
		plan.ObservationEvidenceDigest,
		strconv.Itoa(plan.ObservedUseCases),
		strconv.Itoa(plan.RequiredUseCases),
		strconv.Itoa(plan.RemainingUseCases),
		strconv.Itoa(plan.InvestmentBudget),
		strconv.Itoa(plan.EstimatedInvestment),
		string(plan.Decision),
		plan.MetricMeaning,
		plan.ScopePolicy,
		plan.FirstBoundary,
		plan.Reason,
		strconv.FormatBool(plan.ReviewOnly),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(sum[:])
}
