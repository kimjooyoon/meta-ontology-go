package extractor

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
)

const mapLiteralStrategy = "caller-evaluated-map-construction"

// Only data construction moves. Value expressions, including opaque calls,
// remain arguments in the original function, in their original lexical order.
func buildMapLiteralCandidate(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, evidence typeEvidence, preflight []renderedCapacityObservation, previous Failure) (*returnTailCandidate, error) {
	for _, statement := range function.Body.List {
		literal, keys := eligibleMapLiteral(statement, file, evidence)
		if literal == nil {
			continue
		}
		candidate, err := renderMapLiteralCandidate(source, fset, file, function, literal, keys, evidence.pkg)
		if err != nil {
			if isKnownSuffixContradiction(err) {
				continue
			}
			return nil, err
		}
		if err := projectedFinalConformance(root, logical, map[string][]byte{logical: candidate.result}); err != nil {
			return nil, err
		}
		proof, err := mapLiteralStrategyEvidence(root, logical, source, fset, file, function, candidate, preflight, previous)
		if err != nil {
			return nil, err
		}
		return &returnTailCandidate{helperName: candidate.helperName, helper: candidate.helper, result: candidate.result, evidence: *proof}, nil
	}
	return nil, nil
}

func eligibleMapLiteral(statement ast.Stmt, file *ast.File, evidence typeEvidence) (*ast.CompositeLit, []string) {
	assignment, ok := statement.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		return nil, nil
	}
	if _, ok := assignment.Lhs[0].(*ast.Ident); !ok {
		return nil, nil
	}
	literal, ok := assignment.Rhs[0].(*ast.CompositeLit)
	if !ok || evidence.info == nil || evidence.pkg == nil {
		return nil, nil
	}
	if _, ok := literal.Type.(*ast.MapType); !ok {
		return nil, nil
	}
	empty := types.NewInterfaceType(nil, nil).Complete()
	if !types.Identical(evidence.info.TypeOf(literal), types.NewMap(types.Typ[types.String], empty)) {
		return nil, nil
	}
	if evidence.pkg.Scope().Lookup("string") != nil || evidence.pkg.Scope().Lookup("any") != nil {
		return nil, nil
	}
	for _, comment := range file.Comments {
		if comment.Pos() < literal.End() && comment.End() > literal.Pos() {
			return nil, nil
		}
	}
	keys, ok := mapLiteralKeys(literal)
	if !ok || len(keys) < 2 {
		return nil, nil
	}
	return literal, keys
}

func mapLiteralKeys(literal *ast.CompositeLit) ([]string, bool) {
	keys := make([]string, 0, len(literal.Elts))
	seen := make(map[string]bool, len(literal.Elts))
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return nil, false
		}
		key, ok := pair.Key.(*ast.BasicLit)
		if !ok || key.Kind != token.STRING {
			return nil, false
		}
		value, err := strconv.Unquote(key.Value)
		if err != nil || seen[value] {
			return nil, false
		}
		seen[value] = true
		keys = append(keys, value)
	}
	return keys, true
}

func mapLiteralHelperName(file *ast.File, pkg *types.Package, function string) string {
	used := make(map[string]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok {
			used[identifier.Name] = true
		}
		return true
	})
	for index := 1; ; index++ {
		name := fmt.Sprintf("%sExtractedMap%02d", safeGeneratedFunctionPrefix(function), index)
		if !used[name] && pkg.Scope().Lookup(name) == nil {
			return name
		}
	}
}
