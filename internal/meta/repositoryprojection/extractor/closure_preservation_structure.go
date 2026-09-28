package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
)

type closureProofCandidate struct {
	source   []byte
	fset     *token.FileSet
	file     *ast.File
	helper   *ast.FuncDecl
	literal  *ast.FuncLit
	call     *ast.CallExpr
	evidence typeEvidence
}

type closureProofBinding struct {
	input    types.Object
	output   types.Object
	identity string
	slot     string
}

func closureProofOutput(root, logical string, original *ast.File, target *ast.FuncDecl, candidate CallbackPreviewCandidate) (closureProofCandidate, error) {
	output := closureProofCandidate{source: []byte(candidate.CandidateSource), fset: token.NewFileSet()}
	file, err := parser.ParseFile(output.fset, logical, output.source, parser.ParseComments)
	if err != nil {
		return output, err
	}
	output.file, output.helper = file, callbackPreviewFunction(file, candidate.HelperName)
	fresh := candidate.HelperName != ""
	ast.Inspect(original, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok && identifier.Name == candidate.HelperName {
			fresh = false
		}
		return true
	})
	if !fresh || output.helper == nil || output.helper.Recv != nil || output.helper.Doc != nil || output.helper.Body == nil || len(output.helper.Body.List) != 1 {
		return output, closureProofFailure("factory is not a fresh, single-return function")
	}
	returned, ok := output.helper.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(returned.Results) != 1 {
		return output, closureProofFailure("factory return is missing")
	}
	output.literal, ok = returned.Results[0].(*ast.FuncLit)
	if !ok {
		return output, closureProofFailure("factory does not return the callback literal")
	}
	outputTarget := callbackPreviewFunction(file, target.Name.Name)
	if outputTarget == nil {
		return output, closureProofFailure("candidate lost its original function")
	}
	output.evidence, err = checkTypes(root, logical, output.fset, file, outputTarget)
	if err != nil {
		return output, err
	}
	count := 0
	ast.Inspect(outputTarget.Body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if name, ok := call.Fun.(*ast.Ident); ok && output.evidence.info.Uses[name] == output.evidence.info.Defs[output.helper.Name] {
				output.call, count = call, count+1
			}
		}
		return true
	})
	if count != 1 {
		return output, closureProofFailure("factory construction is not bound exactly once")
	}
	metadata, err := parser.ParseFile(token.NewFileSet(), "helper.go", "package proof\n"+candidate.HelperSource, parser.ParseComments)
	if err != nil {
		return output, err
	}
	wrapper, err := parser.ParseExpr(candidate.WrapperSource)
	if err != nil {
		return output, err
	}
	if !closureASTEqual(output.helper, callbackPreviewFunction(metadata, candidate.HelperName)) || !closureASTEqual(output.call, wrapper) {
		return output, closureProofFailure("helper or construction metadata differs from candidate source")
	}
	return output, nil
}

func closureProofBindings(fset *token.FileSet, file *ast.File, callback *ast.FuncLit, evidence typeEvidence, captures []CallbackPreviewCapture, output closureProofCandidate) ([]closureProofBinding, error) {
	if len(output.helper.Type.Params.List) != len(captures) || len(output.call.Args) != len(captures) {
		return nil, closureProofFailure("factory environment arity differs from source captures")
	}
	objects := map[string]types.Object{}
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok {
			if object, ok := evidence.info.Uses[identifier].(*types.Var); ok {
				objects[callbackPreviewObjectIdentity(object, fset)] = object
			}
		}
		return true
	})
	bindings := make([]closureProofBinding, 0, len(captures))
	for index, capture := range captures {
		original := objects[capture.ObjectIdentity]
		field := output.helper.Type.Params.List[index]
		if original == nil || len(field.Names) != 1 {
			return nil, closureProofFailure("factory parameter has no source object")
		}
		parameter := output.evidence.info.Defs[field.Names[0]]
		pointer, ok := output.evidence.info.TypeOf(field.Type).(*types.Pointer)
		if !ok || parameter == nil || closureProofType(pointer.Elem()) != closureProofType(original.Type()) {
			return nil, closureProofFailure("factory parameter copies or changes a captured type")
		}
		address, ok := output.call.Args[index].(*ast.UnaryExpr)
		if !ok || address.Op != token.AND {
			return nil, closureProofFailure("factory does not capture the variable address")
		}
		name, ok := address.X.(*ast.Ident)
		if !ok || name.Name != original.Name() {
			return nil, closureProofFailure("factory captures a different variable")
		}
		actual := output.evidence.info.Uses[name]
		if actual == nil || closureProofType(actual.Type()) != closureProofType(original.Type()) {
			return nil, closureProofFailure("factory capture type identity differs")
		}
		slot := closureProofFreshName(fmt.Sprintf("goooPreservationCapture%d", index), file, output.file)
		bindings = append(bindings, closureProofBinding{input: original, output: parameter, identity: capture.ObjectIdentity, slot: slot})
	}
	return bindings, nil
}

func closureProofBodies(source []byte, fset *token.FileSet, callback *ast.FuncLit, evidence typeEvidence, output closureProofCandidate, bindings []closureProofBinding) ([]byte, []byte, []ClosureCaptureReference, error) {
	before := map[types.Object]int{}
	after := map[types.Object]int{}
	for index, binding := range bindings {
		before[binding.input], after[binding.output] = index, index
	}
	inputEdits, inputCounts, err := closureProofReferenceEdits(fset, callback, evidence.info, before, bindings, false)
	if err != nil {
		return nil, nil, nil, err
	}
	outputEdits, outputCounts, err := closureProofReferenceEdits(output.fset, output.literal, output.evidence.info, after, bindings, true)
	if err != nil {
		return nil, nil, nil, err
	}
	references := make([]ClosureCaptureReference, 0, len(bindings))
	for index, binding := range bindings {
		if inputCounts[index] == 0 || inputCounts[index] != outputCounts[index] {
			return nil, nil, nil, closureProofFailure("captured reference multiplicity differs")
		}
		references = append(references, ClosureCaptureReference{ObjectIdentity: binding.identity, References: inputCounts[index]})
	}
	a, err := closureProofNormalized(source, fset, callback, inputEdits, true)
	if err != nil {
		return nil, nil, nil, err
	}
	b, err := closureProofNormalized(output.source, output.fset, output.literal, outputEdits, true)
	return a, b, references, err
}

func closureProofReferenceEdits(fset *token.FileSet, literal *ast.FuncLit, info *types.Info, objects map[types.Object]int, bindings []closureProofBinding, indirect bool) ([]edit, []int, error) {
	var edits []edit
	counts := make([]int, len(bindings))
	var invalid bool
	ast.Inspect(literal.Body, func(node ast.Node) bool {
		if indirect {
			if parenthesized, ok := node.(*ast.ParenExpr); ok {
				if dereference, ok := parenthesized.X.(*ast.StarExpr); ok {
					if name, ok := dereference.X.(*ast.Ident); ok {
						if index, ok := objects[info.Uses[name]]; ok {
							edits = append(edits, edit{start: fset.Position(node.Pos()).Offset, end: fset.Position(node.End()).Offset, replacement: []byte(bindings[index].slot)})
							counts[index]++
							return false
						}
					}
				}
			}
		}
		if identifier, ok := node.(*ast.Ident); ok {
			if index, ok := objects[info.Uses[identifier]]; ok {
				if indirect {
					invalid = true
					return false
				}
				edits = append(edits, edit{start: fset.Position(node.Pos()).Offset, end: fset.Position(node.End()).Offset, replacement: []byte(bindings[index].slot)})
				counts[index]++
			}
		}
		return true
	})
	if invalid {
		return nil, nil, closureProofFailure("factory environment pointer escapes its canonical reference")
	}
	return edits, counts, nil
}

func closureProofContexts(source []byte, fset *token.FileSet, callback *ast.FuncLit, output closureProofCandidate) ([]byte, []byte, error) {
	marker := []byte(closureProofFreshName("goooPreservationCallback", output.file))
	before, err := applyEdits(source, []edit{{start: fset.Position(callback.Pos()).Offset, end: fset.Position(callback.End()).Offset, replacement: marker}})
	if err != nil {
		return nil, nil, err
	}
	after, err := applyEdits(output.source, []edit{
		{start: output.fset.Position(output.call.Pos()).Offset, end: output.fset.Position(output.call.End()).Offset, replacement: marker},
		{start: output.fset.Position(output.helper.Pos()).Offset, end: output.fset.Position(output.helper.End()).Offset},
	})
	if err != nil {
		return nil, nil, err
	}
	a, err := closureProofParse(before, false)
	if err != nil {
		return nil, nil, err
	}
	b, err := closureProofParse(after, false)
	return a, b, err
}

func closureProofNormalized(source []byte, fset *token.FileSet, node ast.Node, edits []edit, expression bool) ([]byte, error) {
	start, end := fset.Position(node.Pos()).Offset, fset.Position(node.End()).Offset
	for index := range edits {
		edits[index].start -= start
		edits[index].end -= start
	}
	normalized, err := applyEdits(source[start:end], edits)
	if err != nil {
		return nil, err
	}
	return closureProofParse(normalized, expression)
}

func closureProofParse(source []byte, expression bool) ([]byte, error) {
	fset := token.NewFileSet()
	var node ast.Node
	var err error
	if expression {
		node, err = parser.ParseExprFrom(fset, "proof.go", source, parser.ParseComments)
	} else {
		node, err = parser.ParseFile(fset, "proof.go", source, parser.ParseComments)
	}
	if err != nil {
		return nil, err
	}
	return closureASTEncoding(node), nil
}

func closureASTEncoding(node ast.Node) []byte {
	var output bytes.Buffer
	_ = ast.Fprint(&output, token.NewFileSet(), node, func(name string, value reflect.Value) bool {
		return name != "Obj" && name != "Scope" && name != "Unresolved" && value.Type() != reflect.TypeFor[token.Pos]()
	})
	return output.Bytes()
}

func closureASTEqual(left, right ast.Node) bool {
	return bytes.Equal(closureASTEncoding(left), closureASTEncoding(right))
}

func closureProofType(value types.Type) string {
	return types.TypeString(value, func(pkg *types.Package) string { return pkg.Path() })
}

func closureProofFreshName(prefix string, files ...*ast.File) string {
	used := map[string]bool{}
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			if identifier, ok := node.(*ast.Ident); ok {
				used[identifier.Name] = true
			}
			return true
		})
	}
	for used[prefix] {
		prefix += "_"
	}
	return prefix
}
