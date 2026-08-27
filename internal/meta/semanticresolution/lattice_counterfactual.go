package semanticresolution

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func buildLatticeCounterfactuals(sourcePath, source string, baselineCases []LatticeCase, baselineClaims []ClaimRecord) ([]LatticeCounterfactual, error) {
	semanticSource, err := replaceCaseField(source, "partial-invariant-descent", "observed", "3")
	if err != nil {
		return nil, err
	}
	nonSemanticSource := source + "\n// presentation-only lattice comment\n"
	semantic, err := counterfactual("observed-2-to-3", "SEMANTIC", sourcePath, source, semanticSource, baselineCases, baselineClaims)
	if err != nil {
		return nil, err
	}
	presentation, err := counterfactual("comment-only", "NON_SEMANTIC", sourcePath, source, nonSemanticSource, baselineCases, baselineClaims)
	if err != nil {
		return nil, err
	}
	return []LatticeCounterfactual{semantic, presentation}, nil
}

func counterfactual(id, kind, sourcePath, baselineSource, variantSource string, baselineCases []LatticeCase, baselineClaims []ClaimRecord) (LatticeCounterfactual, error) {
	variantCases, variantClaims, variantSemanticDigest, err := casesFromGoooSource(sourcePath, variantSource)
	if err != nil {
		return LatticeCounterfactual{}, err
	}
	baselineCase, err := findCase(baselineCases, "partial-invariant-descent")
	if err != nil {
		return LatticeCounterfactual{}, err
	}
	variantCase, err := findCase(variantCases, "partial-invariant-descent")
	if err != nil {
		return LatticeCounterfactual{}, err
	}
	baselineClaim, err := findClaim(baselineClaims, baselineCase.ClaimID)
	if err != nil {
		return LatticeCounterfactual{}, err
	}
	variantClaim, err := findClaim(variantClaims, variantCase.ClaimID)
	if err != nil {
		return LatticeCounterfactual{}, err
	}
	baselineSemanticDigest := semanticDigestForSource(sourcePath, baselineSource)
	return LatticeCounterfactual{
		ID: id, Kind: kind,
		BaselineDecision: baselineCase.Decision, VariantDecision: variantCase.Decision,
		BaselineTransition: baselineCase.Transition, VariantTransition: variantCase.Transition,
		BaselineClaim: baselineClaim, VariantClaim: variantClaim,
		SourceDigestChanged:    digestSource(baselineSource) != digestSource(variantSource),
		SemanticDigestChanged:  baselineSemanticDigest != variantSemanticDigest,
		DecisionChanged:        baselineCase.Decision != variantCase.Decision || baselineCase.Transition.Decision != variantCase.Transition.Decision,
		ClaimTransitionChanged: baselineClaim.BeforeState != variantClaim.BeforeState || baselineClaim.AfterState != variantClaim.AfterState,
	}, nil
}

func replaceCaseField(source, caseID, field, replacement string) (string, error) {
	marker := latticeCaseProgramPrefix + "id=" + caseID + ";"
	start := strings.Index(source, marker)
	if start < 0 {
		return "", fmt.Errorf("case %q is missing from source", caseID)
	}
	end := strings.IndexByte(source[start:], '\n')
	if end < 0 {
		end = len(source) - start
	}
	lineEnd := start + end
	line := source[start:lineEnd]
	fieldStart := strings.Index(line, field+"=")
	if fieldStart < 0 {
		return "", fmt.Errorf("field %q is missing from case %q", field, caseID)
	}
	fieldStart += len(field) + 1
	fieldEnd := strings.IndexByte(line[fieldStart:], ';')
	if fieldEnd < 0 {
		return "", fmt.Errorf("field %q has no terminator", field)
	}
	fieldEnd += fieldStart
	return source[:start+fieldStart] + replacement + source[start+fieldEnd:], nil
}

func semanticDigestForSource(sourcePath, source string) string {
	items, err := parseGoooCases(sourcePath, source)
	if err != nil {
		return ""
	}
	return semanticDigest(items)
}

func digestSource(source string) string {
	digest := sha256.Sum256([]byte(source))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func findCase(cases []LatticeCase, id string) (LatticeCase, error) {
	for _, item := range cases {
		if item.ID == id {
			return item, nil
		}
	}
	return LatticeCase{}, fmt.Errorf("case %q is missing", id)
}

func findClaim(claims []ClaimRecord, id string) (ClaimRecord, error) {
	for _, item := range claims {
		if item.ID == id {
			return item, nil
		}
	}
	return ClaimRecord{}, fmt.Errorf("claim %q is missing", id)
}
