package main

import (
	"errors"
	"reflect"
	"strings"
)

func validateCounterfactuals(source string, baselineCases []latticeCase, baselineClaims []claim, got []counterfactual) error {
	if len(got) != 2 {
		return errors.New("counterfactual denominator is not fixed")
	}
	semanticSource, err := replaceCaseField(source, "partial-invariant-descent", "observed", "3")
	if err != nil {
		return err
	}
	baselineDeclared, err := parseGoooCases(source)
	if err != nil {
		return err
	}
	nonSemanticSource := source + "\n// presentation-only lattice comment\n"
	wantSemantic, err := buildIndependentCounterfactual("observed-2-to-3", "SEMANTIC", semanticSource, baselineDeclared, baselineCases, baselineClaims, source)
	if err != nil {
		return err
	}
	wantPresentation, err := buildIndependentCounterfactual("comment-only", "NON_SEMANTIC", nonSemanticSource, baselineDeclared, baselineCases, baselineClaims, source)
	if err != nil {
		return err
	}
	want := []counterfactual{wantSemantic, wantPresentation}
	if !reflect.DeepEqual(got, want) {
		return errors.New("counterfactual receipt is not independently reconstructed")
	}
	return nil
}

func buildIndependentCounterfactual(id, kind, variantSource string, baselineDeclared []declaredCase, baselineCases []latticeCase, baselineClaims []claim, baselineSource string) (counterfactual, error) {
	declared, err := parseGoooCases(variantSource)
	if err != nil {
		return counterfactual{}, err
	}
	variantCases := make([]latticeCase, 0, len(declared))
	variantClaims := make([]claim, 0, len(declared))
	for _, item := range declared {
		reconstructed := reconstructCase(item)
		variantCases = append(variantCases, reconstructed)
		variantClaims = append(variantClaims, reconstructClaim(item, reconstructed.Transition))
	}
	baseCase, err := caseByID(baselineCases, "partial-invariant-descent")
	if err != nil {
		return counterfactual{}, err
	}
	variantCase, err := caseByID(variantCases, "partial-invariant-descent")
	if err != nil {
		return counterfactual{}, err
	}
	baseClaim, err := claimByID(baselineClaims, baseCase.ClaimID)
	if err != nil {
		return counterfactual{}, err
	}
	variantClaim, err := claimByID(variantClaims, variantCase.ClaimID)
	if err != nil {
		return counterfactual{}, err
	}
	baseSemantic := sourceSemanticDigest(baselineDeclared)
	variantSemantic := sourceSemanticDigest(declared)
	return counterfactual{
		ID: id, Kind: kind,
		BaselineDecision: baseCase.Decision, VariantDecision: variantCase.Decision,
		BaselineTransition: baseCase.Transition, VariantTransition: variantCase.Transition,
		BaselineClaim: baseClaim, VariantClaim: variantClaim,
		SourceDigestChanged:    digestText(baselineSource) != digestText(variantSource),
		SemanticDigestChanged:  baseSemantic != variantSemantic,
		DecisionChanged:        baseCase.Decision != variantCase.Decision || baseCase.Transition.Decision != variantCase.Transition.Decision,
		ClaimTransitionChanged: baseClaim.BeforeState != variantClaim.BeforeState || baseClaim.AfterState != variantClaim.AfterState,
	}, nil
}

func replaceCaseField(source, caseID, field, replacement string) (string, error) {
	marker := sourceCasePrefix + "id=" + caseID + ";"
	start := strings.Index(source, marker)
	if start < 0 {
		return "", errors.New("counterfactual case is missing")
	}
	lineEndOffset := strings.IndexByte(source[start:], '\n')
	if lineEndOffset < 0 {
		lineEndOffset = len(source) - start
	}
	line := source[start : start+lineEndOffset]
	fieldStart := strings.Index(line, field+"=")
	if fieldStart < 0 {
		return "", errors.New("counterfactual field is missing")
	}
	fieldStart += len(field) + 1
	fieldEnd := strings.IndexByte(line[fieldStart:], ';')
	if fieldEnd < 0 {
		return "", errors.New("counterfactual field is unterminated")
	}
	fieldEnd += fieldStart
	return source[:start+fieldStart] + replacement + source[start+fieldEnd:], nil
}

func caseByID(cases []latticeCase, id string) (latticeCase, error) {
	for _, item := range cases {
		if item.ID == id {
			return item, nil
		}
	}
	return latticeCase{}, errors.New("counterfactual case lookup failed")
}

func claimByID(claims []claim, id string) (claim, error) {
	for _, item := range claims {
		if item.ID == id {
			return item, nil
		}
	}
	return claim{}, errors.New("counterfactual claim lookup failed")
}
