package verticalsliceclosureeligibility

import (
	"fmt"
	"reflect"
)

func validAssurance(value assuranceCapsule) bool {
	return value.Schema == "gooo/language-assurance-report/v1" &&
		value.SubjectSHA == AssuranceEvidenceSubject &&
		value.DenominatorID == AssuranceDenominatorID &&
		value.DenominatorDigest == AssuranceDenominatorDigest &&
		value.ReportDigest == AssuranceReportDigest &&
		value.Summary.DenominatorTotal == 12 && value.Summary.Operating == 10 &&
		value.Summary.NotImplemented == 2 && value.Summary.ImplementationCoverageBPS == 8333 &&
		value.Summary.UnknownTopDecisions == 0 && value.Summary.RepositoryWrites == 0
}

func evidenceLinked(assurance assuranceCapsule, shadow shadowCapsule) bool {
	return shadow.AssuranceSubjectSHA == assurance.SubjectSHA &&
		shadow.AssuranceDigest == AssuranceCapsuleDigest &&
		shadow.DenominatorDigest == ShadowDenominatorDigest
}

func validShadow(value shadowCapsule) bool {
	if value.Schema != "gooo/vertical-slice-closure-shadow/v1" || value.MetricID != MetricID ||
		value.MetaOperation != "close-vertical-slice" || value.Decision != "SHADOW_PASS" ||
		value.Reason != "VERTICAL_SLICE_CLOSURE_SHADOW_PROVEN" || value.Resolution != ResolutionExact ||
		value.EnforcementEffect != EffectNone || value.HeadSHA != ShadowEvidenceHead ||
		value.ReportDigest != ShadowReportDigest || value.RepositoryWrites != 0 || value.PromotionApplied != 0 {
		return false
	}
	summary := value.Summary
	if summary.DenominatorTotal != 12 || summary.BeforeOperating != 10 ||
		summary.ProjectedOperating != 11 || summary.BeforeCoverageBPS != 8333 ||
		summary.ProjectedCoverageBPS != 9166 || summary.BoundariesTotal != 6 ||
		summary.BoundariesSatisfied != 6 || summary.UnknownBoundaries != 0 ||
		summary.BlockedBoundaries != 0 || summary.LinksTotal != 12 ||
		summary.LinksSatisfied != 12 || summary.EvidenceAvailable != 6 ||
		summary.UnknownTopDecisions != 0 || summary.KnownFailures != 0 ||
		summary.ObservedRepositoryWrites != 0 {
		return false
	}
	return validBoundaries(value.Boundaries) && validSourceIndicators(value.Indicators)
}

func validBoundaries(values []boundary) bool {
	expected := map[string][3]int{
		"syntax": {17, 17, 1}, "semantics": {20, 20, 2}, "binding": {12, 12, 2},
		"use-cases": {3, 3, 1}, "toolchain": {156, 156, 3}, "release": {20, 20, 3},
	}
	if len(values) != len(expected) {
		return false
	}
	for _, value := range values {
		target, ok := expected[value.ID]
		if !ok || value.Value != target[0] || value.Target != target[1] ||
			value.LinksSatisfied != target[2] || value.LinksTotal != target[2] ||
			value.Status != "SATISFIED" || value.Resolution != ResolutionExact ||
			value.HeadSHA != "64a529d71d2fc76000e345b4dd86ad982ebb679e" ||
			!value.EvidenceAvailable || value.UnknownTopDecision || value.KnownFailure ||
			value.RepositoryWrites != 0 {
			return false
		}
	}
	return true
}

func validSourceIndicators(values []sourceIndicator) bool {
	classes, proofs := map[string]int{}, map[string]int{}
	for _, value := range values {
		if !value.Satisfied {
			return false
		}
		classes[value.Class]++
		proofs[value.ProofChoice]++
	}
	return len(values) == 6 && classes["OUTCOME"] == 1 && classes["DRIVER"] == 2 &&
		classes["GUARDRAIL"] == 3 && proofs["FOUNDATION"] == 3 &&
		proofs["COHERENCE"] == 2 && proofs["REGRESSION"] == 1
}

func Validate(report Report, input Input) error {
	if report.Schema != ReportSchema || !reflect.DeepEqual(report, Evaluate(input)) {
		return fmt.Errorf("vertical slice eligibility report does not replay")
	}
	return nil
}
