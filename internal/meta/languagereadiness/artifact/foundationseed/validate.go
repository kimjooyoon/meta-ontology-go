package foundationseed

import (
	"fmt"
	"reflect"
)

func Validate(report Report) error {
	if report.Schema != Schema || len(report.Indicators) != IndicatorCount ||
		len(report.Views) != 3 || len(report.Proofs) != 3 || report.Digest == "" {
		return fmt.Errorf("foundation seed contract malformed")
	}
	if !authorityDenied(report.Authority) ||
		report.Source.AuthorityDenied != authorityDenied(report.Authority) {
		return fmt.Errorf("foundation seed authority malformed")
	}
	expectedIndicators := indicators(report.Source)
	if !reflect.DeepEqual(report.Indicators, expectedIndicators) ||
		!reflect.DeepEqual(report.Summary, summarize(expectedIndicators)) {
		return fmt.Errorf("foundation seed indicators malformed")
	}
	if !reflect.DeepEqual(report.Views, views(expectedIndicators)) ||
		!reflect.DeepEqual(report.Proofs, proofs(report.Source, report.Authority)) ||
		!reflect.DeepEqual(report.NonClaims, fixedNonClaims) {
		return fmt.Errorf("foundation seed evidence malformed")
	}
	if report.Digest != seal(report).Digest {
		return fmt.Errorf("foundation seed digest mismatch")
	}
	switch report.Decision {
	case DecisionAuthorized:
		if report.Reason != ReasonExact || report.Resolution != ResolutionExact ||
			!report.Source.ExactExhaustion ||
			report.Summary.Satisfied != IndicatorCount ||
			report.Summary.Total != IndicatorCount ||
			report.Summary.BasisPoints != 10000 || !allProofsPassed(report.Proofs) {
			return fmt.Errorf("foundation seed authorization malformed")
		}
	case DecisionFailClosed:
		if report.Reason != ReasonUnknown || report.Resolution != ResolutionLower ||
			report.Source.ExactExhaustion {
			return fmt.Errorf("foundation seed rejection malformed")
		}
	default:
		return fmt.Errorf("foundation seed decision unknown")
	}
	return nil
}

func allProofsPassed(values []Proof) bool {
	for _, value := range values {
		if !value.Passed {
			return false
		}
	}
	return true
}
