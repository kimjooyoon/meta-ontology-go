package extractor

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const (
	callbackPreviewSchema        = "gooo.callback-preview/v1"
	callbackPreviewTarget        = "TestPaginationFixturesExecuteParserAndHTTPClient"
	callbackPreviewStateUnknown  = "UNKNOWN"
	callbackPreviewPromotionNone = "NONE"
)

// CallbackPreviewResult is deliberately not Result. Its candidate is caller-
// owned preview data and cannot enter the normal OperationResult or staging
// path.
type CallbackPreviewResult struct {
	Schema                   string                    `json:"schema"`
	LogicalPath              string                    `json:"logical_path"`
	Subject                  string                    `json:"subject"`
	SourceDigest             string                    `json:"source_digest"`
	ContractSourceDigest     string                    `json:"contract_source_digest"`
	ContractSemanticDigest   string                    `json:"contract_semantic_digest"`
	State                    string                    `json:"state"`
	Reason                   string                    `json:"reason"`
	Candidate                *CallbackPreviewCandidate `json:"candidate,omitempty"`
	Captures                 []CallbackPreviewCapture  `json:"captures,omitempty"`
	PendingEffects           []CallbackPreviewEffect   `json:"pending_effects,omitempty"`
	Evidence                 CallbackPreviewEvidence   `json:"evidence"`
	OperationResultAdmission string                    `json:"operation_result_admission"`
	ApplyPermission          string                    `json:"apply_permission"`
}

type CallbackPreviewCandidate struct {
	CandidateIdentity   string `json:"candidate_identity"`
	SourceDigest        string `json:"source_digest"`
	CandidateDigest     string `json:"candidate_digest"`
	HelperName          string `json:"helper_name"`
	WrapperSource       string `json:"wrapper_source"`
	HelperSource        string `json:"helper_source"`
	CandidateSource     string `json:"candidate_source"`
	HelperBytes         int    `json:"helper_bytes"`
	HelperLines         int    `json:"helper_lines"`
	ParentFunctionLines int    `json:"parent_function_lines"`
	CaptureCount        int    `json:"capture_count"`
	PendingEffectCount  int    `json:"pending_effect_count"`
	State               string `json:"state"`
	Promotion           string `json:"promotion"`
}

type CallbackPreviewCapture struct {
	Name           string `json:"name"`
	ObjectIdentity string `json:"object_identity"`
	ObjectType     string `json:"object_type"`
	BindingMode    string `json:"binding_mode"`
}

type CallbackPreviewEffect struct {
	CallIdentity string `json:"call_identity"`
	Symbol       string `json:"symbol"`
	Signature    string `json:"signature"`
	ReceiverType string `json:"receiver_type"`
	EffectKind   string `json:"effect_kind"`
	State        string `json:"state"`
	StartOffset  int    `json:"start_offset"`
	EndOffset    int    `json:"end_offset"`
}

type CallbackPreviewEvidence struct {
	CandidateIdentity        string `json:"candidate_identity"`
	SourceDigest             string `json:"source_digest"`
	CandidateDigest          string `json:"candidate_digest"`
	State                    string `json:"state"`
	CaptureCount             int    `json:"capture_count"`
	PendingEffectCount       int    `json:"pending_effect_count"`
	ResolvedEffectCount      int    `json:"resolved_effect_count"`
	HelperLines              int    `json:"helper_lines"`
	ParentFunctionLines      int    `json:"parent_function_lines"`
	OperationResultAdmission string `json:"operation_result_admission"`
	ApplyPermission          string `json:"apply_permission"`
}

// PreviewBoundedPaginationCallback observes and renders the exact callback
// shape used by the pagination fixture. It never writes the repository and
// never returns an accepted extraction Result.
func PreviewBoundedPaginationCallback(root, logical string) (CallbackPreviewResult, error) {
	contract, err := generation.LoadCallbackPreviewContract()
	if err != nil {
		return CallbackPreviewResult{}, err
	}
	source, fset, file, err := readParsedSource(root, logical)
	if err != nil {
		return CallbackPreviewResult{}, err
	}
	base := CallbackPreviewResult{Schema: callbackPreviewSchema, LogicalPath: logical, Subject: "func:" + callbackPreviewTarget, SourceDigest: callbackPreviewDigest(source), ContractSourceDigest: contract.SourceDigest, ContractSemanticDigest: contract.SemanticDigest, State: callbackPreviewStateUnknown, OperationResultAdmission: "FORBIDDEN", ApplyPermission: "FORBIDDEN"}
	target := callbackPreviewFunction(file, callbackPreviewTarget)
	if target == nil {
		base.Reason = "CALLBACK_TARGET_MISSING"
		base.Evidence = CallbackPreviewEvidence{SourceDigest: base.SourceDigest, State: base.State, OperationResultAdmission: base.OperationResultAdmission, ApplyPermission: base.ApplyPermission}
		return base, nil
	}
	typeEvidence, err := checkTypes(root, logical, fset, file, target)
	if err != nil {
		base.Reason = "TYPE_EVIDENCE_MISSING"
		base.Evidence = CallbackPreviewEvidence{SourceDigest: base.SourceDigest, State: base.State, OperationResultAdmission: base.OperationResultAdmission, ApplyPermission: base.ApplyPermission}
		return base, nil
	}
	callback, err := callbackPreviewFuncLit(target, typeEvidence.info)
	if err != nil {
		base.Reason = err.Error()
		base.Evidence = CallbackPreviewEvidence{SourceDigest: base.SourceDigest, State: base.State, OperationResultAdmission: base.OperationResultAdmission, ApplyPermission: base.ApplyPermission}
		return base, nil
	}
	captures := callbackPreviewCaptures(callback, typeEvidence, fset)
	effects := callbackPreviewEffects(callback, typeEvidence, fset)
	candidate, err := callbackPreviewCandidate(source, fset, file, target, callback, captures, effects)
	if err != nil {
		return CallbackPreviewResult{}, err
	}
	base.Candidate = &candidate
	base.Captures = captures
	base.PendingEffects = effects
	base.Reason = "PENDING_TYPED_CALLBACK_EFFECTS"
	if candidate.HelperLines > functionLineLimit || candidate.ParentFunctionLines > functionLineLimit {
		base.Reason = "CALLBACK_CANDIDATE_OVER_CAPACITY"
	}
	base.Evidence = CallbackPreviewEvidence{CandidateIdentity: candidate.CandidateIdentity, SourceDigest: candidate.SourceDigest, CandidateDigest: candidate.CandidateDigest, State: base.State, CaptureCount: len(captures), PendingEffectCount: len(effects), ResolvedEffectCount: 0, HelperLines: candidate.HelperLines, ParentFunctionLines: candidate.ParentFunctionLines, OperationResultAdmission: base.OperationResultAdmission, ApplyPermission: base.ApplyPermission}
	return base, nil
}

func callbackPreviewFunction(file *ast.File, name string) *ast.FuncDecl {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name != nil && function.Name.Name == name {
			return function
		}
	}
	return nil
}

func callbackPreviewFuncLit(function *ast.FuncDecl, info *types.Info) (*ast.FuncLit, error) {
	var found *ast.FuncLit
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.X == nil {
			return true
		}
		object, ok := info.Uses[selector.Sel].(*types.TypeName)
		if !ok || object.Pkg() == nil || object.Pkg().Path() != "net/http" || object.Name() != "HandlerFunc" {
			return true
		}
		literal, ok := call.Args[0].(*ast.FuncLit)
		if !ok || literal.Type == nil || literal.Type.Params == nil || len(literal.Type.Params.List) != 2 || literal.Type.Results != nil {
			return true
		}
		if !callbackPreviewHTTPParam(literal.Type.Params.List[0].Type, info, "ResponseWriter") || !callbackPreviewHTTPRequestParam(literal.Type.Params.List[1].Type, info) {
			return true
		}
		found = literal
		return false
	})
	if found == nil {
		return nil, fmt.Errorf("CALLBACK_TARGET_SHAPE_UNSUPPORTED")
	}
	return found, nil
}

func callbackPreviewHTTPParam(expression ast.Expr, info *types.Info, name string) bool {
	typeValue := info.TypeOf(expression)
	named, ok := typeValue.(*types.Named)
	return ok && named.Obj() != nil && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "net/http" && named.Obj().Name() == name
}

func callbackPreviewHTTPRequestParam(expression ast.Expr, info *types.Info) bool {
	pointer, ok := info.TypeOf(expression).(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := pointer.Elem().(*types.Named)
	return ok && named.Obj() != nil && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "net/http" && named.Obj().Name() == "Request"
}

func callbackPreviewCaptures(callback *ast.FuncLit, evidence typeEvidence, fset *token.FileSet) []CallbackPreviewCapture {
	local := make(map[types.Object]bool)
	selectorNames := make(map[*ast.Ident]bool)
	ast.Inspect(callback.Type, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok {
			if object := evidence.info.Defs[identifier]; object != nil {
				local[object] = true
			}
		}
		return true
	})
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.Ident:
			if object := evidence.info.Defs[value]; object != nil {
				local[object] = true
			}
		case *ast.SelectorExpr:
			selectorNames[value.Sel] = true
		}
		return true
	})
	objects := make(map[types.Object]bool)
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		identifier, ok := node.(*ast.Ident)
		if !ok || selectorNames[identifier] {
			return true
		}
		object := evidence.info.Uses[identifier]
		if object == nil || local[object] {
			return true
		}
		if _, ok := object.(*types.PkgName); ok {
			return true
		}
		objects[object] = true
		return true
	})
	captures := make([]CallbackPreviewCapture, 0, len(objects))
	for object := range objects {
		mode := "pointer-identity"
		if variable, ok := object.(*types.Var); !ok || variable.Parent() == evidence.pkg.Scope() {
			mode = "typed-reference"
		}
		captures = append(captures, CallbackPreviewCapture{Name: object.Name(), ObjectIdentity: callbackPreviewObjectIdentity(object, fset), ObjectType: callbackPreviewTypeString(object.Type(), evidence.pkg), BindingMode: mode})
	}
	sort.Slice(captures, func(left, right int) bool { return captures[left].Name < captures[right].Name })
	return captures
}

func callbackPreviewEffects(callback *ast.FuncLit, evidence typeEvidence, fset *token.FileSet) []CallbackPreviewEffect {
	effects := make([]CallbackPreviewEffect, 0)
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		object, receiverType, kind := callbackPreviewCallIdentity(call, evidence)
		if object == nil && kind == "typed-conversion" {
			return true
		}
		if object == nil && kind == "known-pure-conversion" {
			return true
		}
		symbol, signature := "<unknown>", "<unknown>"
		if object != nil {
			symbol, signature = callbackPreviewObjectIdentity(object, fset), callbackPreviewTypeString(object.Type(), evidence.pkg)
		}
		start := fset.Position(call.Pos()).Offset
		end := fset.Position(call.End()).Offset
		effectKind := kind
		if callbackPreviewFrameSensitive(object) {
			effectKind = "caller-frame-sensitive"
		}
		if effectKind == "" {
			effectKind = "unresolved-callee-effect"
		}
		effects = append(effects, CallbackPreviewEffect{CallIdentity: fmt.Sprintf("%s@%d:%d", symbol, start, end), Symbol: symbol, Signature: signature, ReceiverType: receiverType, EffectKind: effectKind, State: callbackPreviewStateUnknown, StartOffset: start, EndOffset: end})
		return true
	})
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.FuncLit:
			effects = append(effects, CallbackPreviewEffect{CallIdentity: fmt.Sprintf("nested-func-lit@%d:%d", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset), Symbol: "<func-lit>", Signature: "<unknown>", EffectKind: "nested-function-retention", State: callbackPreviewStateUnknown, StartOffset: fset.Position(value.Pos()).Offset, EndOffset: fset.Position(value.End()).Offset})
		case *ast.GoStmt:
			effects = append(effects, CallbackPreviewEffect{CallIdentity: fmt.Sprintf("go@%d:%d", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset), Symbol: "<go>", EffectKind: "async-retention", State: callbackPreviewStateUnknown, StartOffset: fset.Position(value.Pos()).Offset, EndOffset: fset.Position(value.End()).Offset})
		case *ast.DeferStmt:
			effects = append(effects, CallbackPreviewEffect{CallIdentity: fmt.Sprintf("defer@%d:%d", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset), Symbol: "<defer>", EffectKind: "deferred-effect", State: callbackPreviewStateUnknown, StartOffset: fset.Position(value.Pos()).Offset, EndOffset: fset.Position(value.End()).Offset})
		case *ast.RangeStmt:
			if rangeType := evidence.info.TypeOf(value.X); rangeType != nil {
				if _, ok := rangeType.Underlying().(*types.Signature); ok {
					effects = append(effects, CallbackPreviewEffect{CallIdentity: fmt.Sprintf("range-func@%d:%d", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset), Symbol: "<range-func>", Signature: callbackPreviewTypeString(rangeType, evidence.pkg), EffectKind: "function-iterator", State: callbackPreviewStateUnknown, StartOffset: fset.Position(value.Pos()).Offset, EndOffset: fset.Position(value.End()).Offset})
				}
			}
		}
		return true
	})
	sort.Slice(effects, func(left, right int) bool { return effects[left].StartOffset < effects[right].StartOffset })
	return effects
}

func callbackPreviewCallIdentity(call *ast.CallExpr, evidence typeEvidence) (types.Object, string, string) {
	if call == nil || evidence.info == nil {
		return nil, "", "unknown-call"
	}
	if tv, ok := evidence.info.Types[call.Fun]; ok && tv.IsType() {
		return nil, "", "typed-conversion"
	}
	switch function := call.Fun.(type) {
	case *ast.Ident:
		object := evidence.info.Uses[function]
		if object == nil {
			if evidence.info.TypeOf(function) != nil {
				return nil, callbackPreviewTypeString(evidence.info.TypeOf(function), evidence.pkg), "function-valued-call"
			}
			return nil, "", "unknown-call"
		}
		return object, "", callbackPreviewObjectKind(object, call, evidence)
	case *ast.SelectorExpr:
		if selection := evidence.info.Selections[function]; selection != nil {
			return selection.Obj(), callbackPreviewTypeString(selection.Recv(), evidence.pkg), callbackPreviewObjectKind(selection.Obj(), call, evidence)
		}
		return evidence.info.Uses[function.Sel], "", callbackPreviewObjectKind(evidence.info.Uses[function.Sel], call, evidence)
	default:
		return nil, "", "function-valued-call"
	}
}

func callbackPreviewObjectKind(object types.Object, call *ast.CallExpr, evidence typeEvidence) string {
	if object == nil {
		return "unknown-call"
	}
	if _, ok := object.(*types.Var); ok {
		if signature, ok := object.Type().Underlying().(*types.Signature); ok && signature != nil {
			return "function-valued-call"
		}
	}
	if function, ok := object.(*types.Func); ok {
		if function.Pkg() != nil && function.Pkg().Path() == "net/http" && (function.Name() == "Error" || function.Name() == "NotFound") {
			return "dynamic-interface-argument"
		}
		if selection := callbackPreviewSelection(call, evidence.info); selection != nil {
			if _, ok := selection.Recv().Underlying().(*types.Interface); ok {
				return "dynamic-interface-method"
			}
			return "typed-method"
		}
		if function.Pkg() != nil && evidence.pkg != nil && function.Pkg() != evidence.pkg {
			return "external-function"
		}
		return "local-function"
	}
	return "unresolved-callee-effect"
}

func callbackPreviewSelection(call *ast.CallExpr, info *types.Info) *types.Selection {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || info == nil {
		return nil
	}
	return info.Selections[selector]
}

func callbackPreviewFrameSensitive(object types.Object) bool {
	function, ok := object.(*types.Func)
	return ok && function.Pkg() != nil && (function.Pkg().Path() == "runtime" || function.Pkg().Path() == "runtime/debug") && (function.Name() == "Caller" || function.Name() == "Callers" || function.Name() == "CallersFrames" || function.Name() == "Stack")
}

func callbackPreviewCandidate(source []byte, fset *token.FileSet, file *ast.File, target *ast.FuncDecl, callback *ast.FuncLit, captures []CallbackPreviewCapture, effects []CallbackPreviewEffect) (CallbackPreviewCandidate, error) {
	usedNames := make(map[string]bool)
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name != nil {
			usedNames[function.Name.Name] = true
		}
	}
	helperName := target.Name.Name + "BoundedCallbackPreview"
	if usedNames[helperName] {
		return CallbackPreviewCandidate{}, fmt.Errorf("callback preview helper name collides: %s", helperName)
	}
	params := make([]*ast.Field, 0, len(callback.Type.Params.List)+len(captures))
	for _, field := range callback.Type.Params.List {
		params = append(params, field)
	}
	for _, capture := range captures {
		expression, err := parser.ParseExpr("*" + capture.ObjectType)
		if err != nil {
			return CallbackPreviewCandidate{}, fmt.Errorf("parse capture type %s: %w", capture.ObjectType, err)
		}
		params = append(params, &ast.Field{Names: []*ast.Ident{ast.NewIdent(capture.Name)}, Type: expression})
	}
	helper := &ast.FuncDecl{Name: ast.NewIdent(helperName), Type: &ast.FuncType{Params: &ast.FieldList{List: params}}, Body: callback.Body}
	var helperBuffer bytes.Buffer
	if err := format.Node(&helperBuffer, fset, helper); err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("format callback helper: %w", err)
	}
	wrapper := &ast.FuncLit{Type: callback.Type, Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent(helperName), Args: callbackPreviewCallArguments(callback, captures)}}}}}
	var wrapperBuffer bytes.Buffer
	if err := format.Node(&wrapperBuffer, fset, wrapper); err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("format callback wrapper: %w", err)
	}
	start := fset.Position(callback.Pos()).Offset
	end := fset.Position(callback.End()).Offset
	if start < 0 || end < start || end > len(source) {
		return CallbackPreviewCandidate{}, fmt.Errorf("callback preview source span is invalid")
	}
	modified := make([]byte, 0, len(source)+len(helperBuffer.Bytes())+len(wrapperBuffer.Bytes()))
	modified = append(modified, source[:start]...)
	modified = append(modified, wrapperBuffer.Bytes()...)
	modified = append(modified, source[end:]...)
	modified = append(modified, '\n', '\n')
	modified = append(modified, helperBuffer.Bytes()...)
	formatted, err := format.Source(modified)
	if err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("format callback candidate: %w", err)
	}
	candidateFset := token.NewFileSet()
	candidateFile, err := parser.ParseFile(candidateFset, "callback-preview.go", formatted, parser.ParseComments)
	if err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("parse callback candidate: %w", err)
	}
	parentLines := 0
	if candidateTarget := callbackPreviewFunction(candidateFile, target.Name.Name); candidateTarget != nil {
		parentLines = candidateFset.Position(candidateTarget.End()).Line - candidateFset.Position(candidateTarget.Pos()).Line + 1
	}
	helperLines := physicalLines(helperBuffer.Bytes())
	identity := fmt.Sprintf("%s#%s@%d:%d", filepath.ToSlash("cmd/language-readiness-witness/predecessor-selection/pagination_test.go"), target.Name.Name, start, end)
	candidateDigest := callbackPreviewDigest(formatted)
	return CallbackPreviewCandidate{CandidateIdentity: identity, SourceDigest: callbackPreviewDigest(source), CandidateDigest: candidateDigest, HelperName: helperName, WrapperSource: wrapperBuffer.String(), HelperSource: helperBuffer.String(), CandidateSource: string(formatted), HelperBytes: len(helperBuffer.Bytes()), HelperLines: helperLines, ParentFunctionLines: parentLines, CaptureCount: len(captures), PendingEffectCount: len(effects), State: callbackPreviewStateUnknown, Promotion: callbackPreviewPromotionNone}, nil
}

func callbackPreviewCallArguments(callback *ast.FuncLit, captures []CallbackPreviewCapture) []ast.Expr {
	arguments := make([]ast.Expr, 0, 2+len(captures))
	for _, field := range callback.Type.Params.List {
		for _, name := range field.Names {
			arguments = append(arguments, ast.NewIdent(name.Name))
		}
	}
	for _, capture := range captures {
		arguments = append(arguments, &ast.UnaryExpr{Op: token.AND, X: ast.NewIdent(capture.Name)})
	}
	return arguments
}

func callbackPreviewObjectIdentity(object types.Object, fset *token.FileSet) string {
	if object == nil {
		return ""
	}
	position := fset.Position(object.Pos())
	packagePath := ""
	if object.Pkg() != nil {
		packagePath = object.Pkg().Path()
	}
	return fmt.Sprintf("%s:%s:%d:%d", packagePath, object.Name(), position.Line, position.Column)
}

func callbackPreviewTypeString(typeValue types.Type, current *types.Package) string {
	return types.TypeString(typeValue, func(pkg *types.Package) string {
		if pkg == nil {
			return ""
		}
		if current != nil && pkg.Path() == current.Path() {
			return ""
		}
		return pkg.Name()
	})
}

func callbackPreviewDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + fmt.Sprintf("%x", sum[:])
}
