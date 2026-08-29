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
	before, err := initializationUnits(evidence.Source)
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
	candidates := evidence.Candidates
	if !metadata {
		candidates = sortedCandidates(candidates)
	}
	after := make([]initializationUnit, 0)
	for _, candidate := range candidates {
		items, candidateErr := initializationUnits(candidate)
		if candidateErr != nil {
			return DecisionFail
		}
		after = append(after, items...)
	}
	if !sameInitializationUnits(before, after) {
		return DecisionFail
	}
	return DecisionPass
}

type initializationUnit struct {
	Ordinal    int
	Digest     string
	HasOrdinal bool
}

func initializationUnits(file FileEvidence) ([]initializationUnit, error) {
	fset, parsed, err := parseEvidence(file)
	if err != nil {
		return nil, err
	}
	units := make([]initializationUnit, 0)
	declarationIndex := 0
	for _, declaration := range parsed.Decls {
		general, isImport := declaration.(*ast.GenDecl)
		if isImport && general.Tok == token.IMPORT {
			continue
		}
		digest, digestErr := digestDeclaration(fset, declaration)
		if digestErr != nil {
			return nil, digestErr
		}
		if isInitializationDeclaration(declaration) {
			unit := initializationUnit{Ordinal: declarationIndex, Digest: digest}
			if len(file.DeclarationOrder) != 0 {
				if declarationIndex >= len(file.DeclarationOrder) {
					return nil, fmt.Errorf("initialization declaration metadata is incomplete")
				}
				entry := file.DeclarationOrder[declarationIndex]
				if entry.Digest != digest || entry.Ordinal < 0 {
					return nil, fmt.Errorf("initialization declaration metadata mismatch")
				}
				unit.Ordinal, unit.HasOrdinal = entry.Ordinal, true
			}
			units = append(units, unit)
		}
		declarationIndex++
	}
	return units, nil
}

func isInitializationDeclaration(declaration ast.Decl) bool {
	if general, ok := declaration.(*ast.GenDecl); ok && general.Tok == token.VAR {
		for _, specification := range general.Specs {
			if value, ok := specification.(*ast.ValueSpec); ok && len(value.Values) != 0 {
				return true
			}
		}
	}
	function, ok := declaration.(*ast.FuncDecl)
	return ok && function.Recv == nil && function.Name != nil && function.Name.Name == "init"
}

func digestDeclaration(fset *token.FileSet, declaration ast.Decl) (string, error) {
	var output bytes.Buffer
	if err := format.Node(&output, fset, declaration); err != nil {
		return "", err
	}
	sum := sha256.Sum256(output.Bytes())
	return fmt.Sprintf("%x", sum), nil
}

func sameInitializationUnits(before, after []initializationUnit) bool {
	if len(before) != len(after) {
		return false
	}
	for index := range before {
		if before[index].Digest != after[index].Digest ||
			after[index].HasOrdinal && before[index].Ordinal != after[index].Ordinal {
			return false
		}
	}
	return true
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
