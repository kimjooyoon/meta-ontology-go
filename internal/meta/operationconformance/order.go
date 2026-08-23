package operationconformance

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
)

func observeOrder(evidence SplitGoEvidence) Decision {
	before, err := declarationSignature(evidence.Source)
	if err != nil || len(evidence.Candidates) == 0 {
		return DecisionFail
	}
	after := make([]string, 0)
	for _, candidate := range sortedCandidates(evidence.Candidates) {
		items, signatureErr := declarationSignature(candidate)
		if signatureErr != nil {
			return DecisionFail
		}
		after = append(after, items...)
	}
	if !sameStringsInOrder(before, after) {
		return DecisionFail
	}
	return DecisionPass
}

func declarationSignature(file FileEvidence) ([]string, error) {
	fset, parsed, err := parseEvidence(file)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(parsed.Decls))
	for _, declaration := range parsed.Decls {
		general, isGeneral := declaration.(*ast.GenDecl)
		if isGeneral && general.Tok == token.IMPORT {
			continue
		}
		var output bytes.Buffer
		if err := format.Node(&output, fset, declaration); err != nil {
			return nil, err
		}
		result = append(result, output.String())
	}
	return result, nil
}

func sameStringsInOrder(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
