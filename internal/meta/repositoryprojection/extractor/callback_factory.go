package extractor

import (
	"bytes"
	"cmp"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"
)

// The factory runs when the closure is created, not when it is invoked. Its
// returned literal retains the callback body and addresses of shared variables.
// This is a second candidate construction, not discharge of pending effects.
func callbackPreviewFactoryCandidate(root string, source []byte, logical string, fset *token.FileSet, file *ast.File, target *ast.FuncDecl, callback *ast.FuncLit, evidence typeEvidence, captures []CallbackPreviewCapture, effects []CallbackPreviewEffect) (CallbackPreviewCandidate, error) {
	used := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok {
			used[identifier.Name] = true
		}
		return true
	})
	name := "boundedCallbackFactoryPreview" + target.Name.Name
	if used[name] {
		return CallbackPreviewCandidate{}, fmt.Errorf("callback factory identity collides: %s", name)
	}
	parameters, arguments, replacements, err := callbackFactoryBindings(callback, evidence, fset, captures, used)
	if err != nil {
		return CallbackPreviewCandidate{}, err
	}
	literal, err := callbackFactoryLiteral(source, fset, callback, evidence.info, replacements)
	if err != nil {
		return CallbackPreviewCandidate{}, err
	}
	signature := types.TypeString(evidence.info.TypeOf(callback), func(pkg *types.Package) string {
		if pkg == evidence.pkg {
			return ""
		}
		return packageAlias(evidence.info, pkg)
	})
	helper, err := format.Source(fmt.Appendf(nil, "func %s(%s) %s {\nreturn %s\n}\n", name, strings.Join(parameters, ", "), signature, literal))
	if err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("format callback factory: %w", err)
	}
	replacement := name + "(" + strings.Join(arguments, ", ") + ")"
	start, end := fset.Position(callback.Pos()).Offset, fset.Position(callback.End()).Offset
	modified, err := replaceSource(source, start, end, []byte(replacement))
	if err != nil {
		return CallbackPreviewCandidate{}, err
	}
	modified = append(append(modified, '\n', '\n'), helper...)
	formatted, err := format.Source(modified)
	if err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("format callback factory candidate: %w", err)
	}
	candidateSet := token.NewFileSet()
	candidateFile, err := parser.ParseFile(candidateSet, logical, formatted, parser.ParseComments)
	if err != nil {
		return CallbackPreviewCandidate{}, err
	}
	candidateTarget := callbackPreviewFunction(candidateFile, target.Name.Name)
	if candidateTarget == nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("callback factory lost its source target")
	}
	if _, err := checkTypes(root, logical, candidateSet, candidateFile, candidateTarget); err != nil {
		return CallbackPreviewCandidate{}, err
	}
	return CallbackPreviewCandidate{
		CandidateIdentity: fmt.Sprintf("%s#%s@%d:%d", filepath.ToSlash(logical), target.Name.Name, start, end),
		SourceDigest:      callbackPreviewDigest(source), CandidateDigest: callbackPreviewDigest(formatted), HelperName: name,
		WrapperSource: replacement, HelperSource: string(helper), CandidateSource: string(formatted),
		HelperBytes: len(helper), HelperLines: physicalLines(helper), ParentFunctionLines: declarationLines(candidateSet, candidateTarget),
		CaptureCount: len(captures), PendingEffectCount: len(effects), State: callbackPreviewStateUnknown, Promotion: callbackPreviewPromotionNone,
	}, nil
}

func callbackFactoryBindings(callback *ast.FuncLit, evidence typeEvidence, fset *token.FileSet, captures []CallbackPreviewCapture, used map[string]bool) ([]string, []string, map[types.Object]string, error) {
	objects := map[string]types.Object{}
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok {
			if object, ok := evidence.info.Uses[identifier].(*types.Var); ok {
				objects[callbackPreviewObjectIdentity(object, fset)] = object
			}
		}
		return true
	})
	parameters, arguments := make([]string, 0, len(captures)), make([]string, 0, len(captures))
	replacements := make(map[types.Object]string, len(captures))
	for index, capture := range captures {
		object := objects[capture.ObjectIdentity]
		if object == nil || replacements[object] != "" || object.Name() != capture.Name ||
			callbackPreviewTypeString(object.Type(), evidence.pkg) != capture.ObjectType || capture.BindingMode != "pointer-identity" {
			return nil, nil, nil, fmt.Errorf("callback factory capture is not bound: %s", capture.ObjectIdentity)
		}
		name := fmt.Sprintf("goooCapture%d", index)
		for used[name] {
			name += "_"
		}
		used[name] = true
		typeName := types.TypeString(object.Type(), func(pkg *types.Package) string {
			if pkg == evidence.pkg {
				return ""
			}
			return packageAlias(evidence.info, pkg)
		})
		parameters = append(parameters, name+" *"+typeName)
		arguments = append(arguments, "&"+capture.Name)
		replacements[object] = "(*" + name + ")"
	}
	return parameters, arguments, replacements, nil
}

type callbackFactoryEdit struct {
	start       int
	end         int
	replacement string
}

func callbackFactoryLiteral(source []byte, fset *token.FileSet, callback *ast.FuncLit, info *types.Info, replacements map[types.Object]string) (string, error) {
	start, end := fset.Position(callback.Pos()).Offset, fset.Position(callback.End()).Offset
	if start < 0 || end < start || end > len(source) {
		return "", fmt.Errorf("callback factory literal has invalid source coordinates")
	}
	var edits []callbackFactoryEdit
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok {
			if replacement := replacements[info.Uses[identifier]]; replacement != "" {
				edits = append(edits, callbackFactoryEdit{fset.Position(identifier.Pos()).Offset, fset.Position(identifier.End()).Offset, replacement})
			}
		}
		return true
	})
	slices.SortFunc(edits, func(left, right callbackFactoryEdit) int { return cmp.Compare(left.start, right.start) })
	var output bytes.Buffer
	cursor := start
	for _, edit := range edits {
		if edit.start < cursor || edit.end < edit.start || edit.end > end {
			return "", fmt.Errorf("callback factory capture edits overlap or escape their literal")
		}
		output.Write(source[cursor:edit.start])
		output.WriteString(edit.replacement)
		cursor = edit.end
	}
	output.Write(source[cursor:end])
	return output.String(), nil
}

func validateCallbackPreviewLowering(result CallbackPreviewResult) error {
	candidate := result.Candidate
	expression, err := parser.ParseExpr(candidate.WrapperSource)
	if err != nil {
		return fmt.Errorf("callback preview lowering expression: %w", err)
	}
	if result.LoweringStrategy == callbackPreviewWrapperLowering {
		literal, ok := expression.(*ast.FuncLit)
		if !ok || len(literal.Body.List) != 1 {
			return fmt.Errorf("wrapper lowering does not contain its callback call")
		}
		statement, ok := literal.Body.List[0].(*ast.ExprStmt)
		if !ok {
			return fmt.Errorf("wrapper lowering does not invoke its helper")
		}
		expression = statement.X
	} else {
		file, err := parser.ParseFile(token.NewFileSet(), "factory.go", "package preview\n"+candidate.HelperSource, 0)
		if err != nil {
			return fmt.Errorf("callback factory helper: %w", err)
		}
		helper := callbackPreviewFunction(file, candidate.HelperName)
		if helper == nil || helper.Body == nil || len(helper.Body.List) != 1 {
			return fmt.Errorf("callback factory must only return a literal")
		}
		returned, ok := helper.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(returned.Results) != 1 {
			return fmt.Errorf("callback factory return is missing")
		}
		if _, ok := returned.Results[0].(*ast.FuncLit); !ok {
			return fmt.Errorf("callback factory does not return a literal")
		}
	}
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("callback preview lowering has no construction call")
	}
	name, ok := call.Fun.(*ast.Ident)
	if !ok || name.Name != candidate.HelperName {
		return fmt.Errorf("callback preview lowering calls a different helper")
	}
	if result.LoweringStrategy == callbackPreviewFactoryLowering {
		if len(call.Args) != len(result.Captures) {
			return fmt.Errorf("callback factory capture count differs from its call")
		}
		for index, argument := range call.Args {
			address, ok := argument.(*ast.UnaryExpr)
			if !ok || address.Op != token.AND {
				return fmt.Errorf("callback factory capture is not passed by address")
			}
			variable, ok := address.X.(*ast.Ident)
			if !ok || variable.Name != result.Captures[index].Name {
				return fmt.Errorf("callback factory capture identity differs from its call")
			}
		}
	}
	return nil
}
