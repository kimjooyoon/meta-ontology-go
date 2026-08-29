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
	before, err := declarationDigests(evidence.Source)
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
		return observePartitionOrder(evidence)
	}
	after := make([]string, 0)
	for _, candidate := range sortedCandidates(evidence.Candidates) {
		items, signatureErr := declarationDigests(candidate)
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

func observePartitionOrder(evidence SplitGoEvidence) Decision {
	before, err := declarationOrders(evidence.Source)
	if err != nil {
		return DecisionFail
	}
	type partition struct {
		first int
		items []DeclarationOrder
	}
	partitions := make([]partition, 0, len(evidence.Candidates))
	seen := make(map[int]bool, len(before))
	for _, candidate := range evidence.Candidates {
		actual, err := declarationDigests(candidate)
		if err != nil || len(candidate.DeclarationOrder) != len(actual) || len(actual) == 0 {
			return DecisionFail
		}
		first := len(before)
		for index, digest := range actual {
			entry := candidate.DeclarationOrder[index]
			if entry.Ordinal < 0 || entry.Ordinal >= len(before) || entry.Digest != digest || seen[entry.Ordinal] {
				return DecisionFail
			}
			if before[entry.Ordinal].Digest != digest {
				return DecisionFail
			}
			if entry.Ordinal < first {
				first = entry.Ordinal
			}
			seen[entry.Ordinal] = true
		}
		partitions = append(partitions, partition{first: first, items: append([]DeclarationOrder{}, candidate.DeclarationOrder...)})
	}
	for index := range before {
		if !seen[index] {
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
	after := make([]DeclarationOrder, 0, len(before))
	for _, item := range partitions {
		after = append(after, item.items...)
	}
	if len(after) != len(before) {
		return DecisionFail
	}
	for index, entry := range after {
		if entry.Ordinal != before[index].Ordinal || entry.Digest != before[index].Digest {
			return DecisionFail
		}
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

func declarationDigests(file FileEvidence) ([]string, error) {
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

func declarationOrders(file FileEvidence) ([]DeclarationOrder, error) {
	digests, err := declarationDigests(file)
	if err != nil {
		return nil, err
	}
	result := make([]DeclarationOrder, len(digests))
	for index, digest := range digests {
		result[index] = DeclarationOrder{Ordinal: index, Digest: digest}
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
