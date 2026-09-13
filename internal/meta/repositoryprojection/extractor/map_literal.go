package extractor

import (
	"errors"
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
		candidate, err := renderMapLiteralCandidate(source, fset, file, function, literal, keys, evidence)
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

// A rejected relocation is not a rejection of a caller-local transformation.
// Missing contract/type evidence must never grant an alternative strategy.
func mapLiteralPriorFailure(err error) (Failure, bool) {
	if failure, ok := errors.AsType[Failure](err); ok {
		return failure, failure.Reason == "CALLEE_EFFECTS_UNPROVEN"
	}
	rejection, ok := err.(suffixContradiction)
	if !ok || rejection.obligation != obligationControlFlow {
		return Failure{}, false
	}
	return Failure{
		Stage:         "derive-recipe",
		Step:          "admit-return-tail",
		Reason:        "RETURN_TAIL_CONTROL_FLOW_UNSUPPORTED",
		UnknownClass:  "KNOWN_CONTRADICTION",
		NextOperation: "select-caller-preserving-alternative",
		BlockedBy:     []string{},
		Diagnostics: []string{
			"strategy=" + returnTailStrategy,
			"obligation=" + rejection.obligation,
			"original_rejection=" + rejection.message,
		},
	}, true
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

func findMapReceiverFusion(source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, evidence typeEvidence) *mapReceiverFusion {
	for index := 0; index+1 < len(function.Body.List); index++ {
		first, firstOK := function.Body.List[index].(*ast.AssignStmt)
		next, nextOK := function.Body.List[index+1].(*ast.AssignStmt)
		if !firstOK || !nextOK {
			continue
		}
		receiver := mapReceiverFusionBinding(first, next, evidence)
		if receiver == nil || mapReceiverFusionHasComments(fset, file, first, next) {
			continue
		}
		return renderMapReceiverFusion(source, fset, first, next, receiver)
	}
	return nil
}

func mapReceiverFusionBinding(first, next *ast.AssignStmt, evidence typeEvidence) *ast.Ident {
	if evidence.info == nil || evidence.pkg == nil || first.Tok != token.DEFINE ||
		len(first.Lhs) != 1 || len(first.Rhs) != 1 || len(next.Rhs) != 1 ||
		(next.Tok != token.DEFINE && next.Tok != token.ASSIGN) {
		return nil
	}
	binding, ok := first.Lhs[0].(*ast.Ident)
	if !ok || binding.Name == "_" {
		return nil
	}
	initializer, ok := first.Rhs[0].(*ast.CallExpr)
	if !ok || evidence.info.TypeOf(initializer) == nil {
		return nil
	}
	if _, ok := evidence.info.TypeOf(initializer).Underlying().(*types.Pointer); !ok {
		return nil
	}
	call, ok := next.Rhs[0].(*ast.CallExpr)
	if !ok {
		return nil
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	receiver, ok := selector.X.(*ast.Ident)
	object := evidence.info.Defs[binding]
	if !ok || object == nil || evidence.info.Uses[receiver] != object {
		return nil
	}
	method, _, _ := types.LookupFieldOrMethod(evidence.info.TypeOf(initializer), false, evidence.pkg, selector.Sel.Name)
	if _, ok := method.(*types.Func); !ok || !mapReceiverFusionSingleUse(next, evidence.info, object) {
		return nil
	}
	return receiver
}

func mapReceiverFusionSingleUse(next *ast.AssignStmt, info *types.Info, object types.Object) bool {
	for _, target := range next.Lhs {
		if _, ok := target.(*ast.Ident); !ok {
			return false
		}
	}
	uses := 0
	for _, used := range info.Uses {
		if used == object {
			uses++
		}
	}
	return uses == 1
}

func mapReceiverFusionHasComments(fset *token.FileSet, file *ast.File, first, next ast.Stmt) bool {
	startLine := fset.Position(first.Pos()).Line
	for _, comment := range file.Comments {
		if comment.Pos() < next.End() && fset.Position(comment.End()).Line >= startLine-1 {
			return true
		}
	}
	return false
}
