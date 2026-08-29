package operationconformance

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
)

func observeOrder(evidence SplitGoEvidence) Decision {
	before, err := declarationIdentities(evidence.Source)
	if err != nil || len(evidence.Candidates) == 0 {
		return DecisionFail
	}
	metadata := false
	for _, candidate := range evidence.Candidates {
		if len(candidate.DeclarationOrder) != 0 {
			metadata = true
			break
		}
	}
	if metadata {
		return observePartitionOrder(before, evidence.Candidates)
	}
	after := make([]string, 0)
	for _, candidate := range sortedCandidates(evidence.Candidates) {
		items, signatureErr := declarationIdentities(candidate)
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

func observePartitionOrder(before []string, candidates []FileEvidence) Decision {
	positions := make(map[string]int, len(before))
	for index, identity := range before {
		positions[identity] = index
	}
	type partition struct {
		first int
		items []string
	}
	partitions := make([]partition, 0, len(candidates))
	seen := make(map[string]bool, len(before))
	for _, candidate := range candidates {
		actual, err := declarationIdentities(candidate)
		if err != nil || len(candidate.DeclarationOrder) != len(actual) || len(actual) == 0 {
			return DecisionFail
		}
		first := len(before)
		for index, identity := range actual {
			if candidate.DeclarationOrder[index] != identity || seen[identity] {
				return DecisionFail
			}
			position, ok := positions[identity]
			if !ok {
				return DecisionFail
			}
			if position < first {
				first = position
			}
			seen[identity] = true
		}
		partitions = append(partitions, partition{first: first, items: candidate.DeclarationOrder})
	}
	for index := range before {
		if !seen[before[index]] {
			return DecisionFail
		}
	}
	for index := 0; index < len(partitions); index++ {
		for next := index + 1; next < len(partitions); next++ {
			if partitions[next].first < partitions[index].first {
				partitions[index], partitions[next] = partitions[next], partitions[index]
			}
		}
	}
	after := make([]string, 0, len(before))
	for _, item := range partitions {
		after = append(after, item.items...)
	}
	return decisionForOrder(before, after)
}

func decisionForOrder(before, after []string) Decision {
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

func declarationIdentities(file FileEvidence) ([]string, error) {
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
		sum := sha256.Sum256(output.Bytes())
		result = append(result, fmt.Sprintf("%x", sum))
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
