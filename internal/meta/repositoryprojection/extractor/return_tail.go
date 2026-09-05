package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const returnTailStrategy = "return-preserving-terminal-tail"

func buildReturnTailCandidate(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, evidence typeEvidence, existing map[string]bool, preflight []renderedCapacityObservation) (*returnTailCandidate, error) {
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		return nil, fail("derive-recipe", "admit-return-tail", "OPERATION_INPUT_CONTRACT_MISSING", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	if contract.Operation == "" || contract.Activity != "ExtractFunction" || contract.InputEntity != "FunctionInput" ||
		contract.OutputEntity != "OperationResult" || contract.InputSubjectKind != sourcepolicy.SubjectKindFunction ||
		!contract.UsedInputFact || !contract.GeneratedOutputFact || contract.SourceDigest == "" || contract.SemanticDigest == "" {
		return nil, fail("derive-recipe", "admit-return-tail", "OPERATION_INPUT_CONTRACT_UNPROVEN", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	contractObligations, obligationsOK := returnTailContractObligations(contract.Obligations)
	if !obligationsOK {
		return nil, fail("derive-recipe", "admit-return-tail", "RETURN_TAIL_OBLIGATIONS_UNPROVEN", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	if !returnTailSignatureEvidence(function, evidence.info) {
		return nil, fail("derive-recipe", "type-check-return-tail", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
	}
	if !returnTailShapeEligible(function, evidence.info) {
		return nil, returnTailContradiction(obligationReturnShape, "function is not an unnamed single-error result")
	}
	statements := function.Body.List
	if len(statements) < 2 {
		return nil, returnTailContradiction(obligationRenderedCapacity, "terminal tail cannot make progress")
	}
	if _, ok := statements[len(statements)-1].(*ast.ReturnStmt); !ok {
		return nil, returnTailContradiction(obligationReturnShape, "function does not end in a return")
	}
	if hasReturnTailOuterHazard(statements, evidence.info) {
		return nil, returnTailContradiction(obligationControlFlow, "function contains unsupported control flow")
	}

	for startIndex := range slices.Backward(statements) {
		candidate, candidateErr := tryReturnTailStart(root, logical, source, fset, file, function, evidence, contract, contractObligations, existing, startIndex, preflight)
		if candidateErr != nil {
			if isKnownSuffixContradiction(candidateErr) {
				continue
			}
			return nil, candidateErr
		}
		if candidate != nil {
			return candidate, nil
		}
	}
	return nil, returnTailContradiction(obligationRenderedCapacity, "no terminal tail satisfies rendered capacity")
}

func returnTailShapeEligible(function *ast.FuncDecl, info *types.Info) bool {
	if function == nil || function.Recv != nil || function.Type == nil || function.Type.TypeParams != nil && len(function.Type.TypeParams.List) != 0 || function.Type.Results == nil || len(function.Type.Results.List) != 1 ||
		len(function.Type.Results.List[0].Names) != 0 || function.Body == nil || len(function.Body.List) == 0 {
		return false
	}
	if _, ok := function.Body.List[len(function.Body.List)-1].(*ast.ReturnStmt); !ok {
		return false
	}
	return isErrorType(info.TypeOf(function.Type.Results.List[0].Type))
}

func tryReturnTailStart(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, evidence typeEvidence, contract generation.OperationInputContractEvidence, contractObligations []ContractObligationEvidence, existing map[string]bool, startIndex int, preflight []renderedCapacityObservation) (*returnTailCandidate, error) {
	statements := function.Body.List[startIndex:]
	if len(statements) == 0 {
		return nil, returnTailContradiction(obligationRenderedCapacity, "terminal tail does not reduce the declaration")
	}
	if !returnTailReturnsCompatible(statements, evidence.info, evidence.info.TypeOf(function.Type.Results.List[0].Type)) {
		return nil, returnTailContradiction(obligationReturnShape, "terminal tail has a bare, multiple, or incompatible return")
	}
	first, last := statements[0], statements[len(statements)-1]
	start := fset.Position(first.Pos()).Offset
	end := fset.Position(last.End()).Offset
	start = includeLeadingComments(fset, file, first, start)
	if start < 0 || end < start || end > len(source) {
		return nil, returnTailContradiction(obligationRenderedCapacity, "terminal tail source coordinates are invalid")
	}
	proof := newReturnTailProofChain(contractObligations, source, source[start:end], contract.SourceDigest, contract.SemanticDigest)
	if err := proof.consume(0, returnTailPredicateResult{Status: "PASS", Payload: append([]byte(nil), source[start:end]...), CandidateDigest: proofDigest(source[start:end]), Detail: "unnamed single-error result with assignable selected returns"}); err != nil {
		return nil, err
	}
	if err := proof.consume(1, returnTailPredicateResult{Status: "PASS", Payload: append([]byte("control-flow\x00"), source[start:end]...), CandidateDigest: proofDigest(source[start:end]), Detail: "no escaping branch, defer, go, panic, or recover"}); err != nil {
		return nil, err
	}
	bindings, err := suffixBindings(statements, function, fset, evidence)
	if err != nil {
		return nil, err
	}
	if err := hasReturnTailBindingHazard(function.Body, statements, bindings, evidence.info); err != nil {
		return nil, err
	}
	if err := proof.consume(2, returnTailPredicateResult{Status: "PASS", Payload: proofBindingPayload(bindings), CandidateDigest: proofDigest(source[start:end]), Detail: "selected free bindings have no stale-copy, rebinding, address, or closure hazard"}); err != nil {
		return nil, err
	}
	if err := returnTailCalleeEffects(statements, evidence); err != nil {
		return nil, err
	}
	if err := proof.consume(3, returnTailPredicateResult{Status: "PASS", Payload: append([]byte("callee-effects\x00"), source[start:end]...), CandidateDigest: proofDigest(source[start:end]), Detail: "all selected calls have closed typed effect evidence"}); err != nil {
		return nil, err
	}

	name := stableReturnTailName(function.Name.Name, startIndex+1)
	if existing[name] {
		return nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{"helper=" + name})
	}
	helper, err := renderReturnTailHelper(fset, name, bindings, source[start:end])
	if err != nil {
		return nil, err
	}
	call := []byte("return " + name + "(" + bindingArguments(bindings) + ")")
	modified, err := replaceSource(source, start, end, call)
	if err != nil {
		return nil, err
	}
	formatted, err := format.Source(modified)
	if err != nil {
		return nil, returnTailContradiction(obligationProjectedConform, "outer source did not render")
	}
	afterFunctionLines, ok := namedFunctionLines(formatted, function.Name.Name)
	if !ok || afterFunctionLines > functionLineLimit {
		return nil, returnTailContradiction(obligationRenderedCapacity, "outer function remains over the declaration limit")
	}
	beforeOuterHelper, err := renderedFunctionHelper(source, function.Name.Name)
	if err != nil {
		return nil, err
	}
	renderedOuterHelper, err := renderedFunctionHelper(formatted, function.Name.Name)
	if err != nil {
		return nil, err
	}
	if physicalLines(renderedOuterHelper) > functionLineLimit || !renderedCapacityProgress(beforeOuterHelper, renderedOuterHelper) {
		return nil, returnTailContradiction(obligationRenderedCapacity, "rendered outer helper remains over the limit or made no capacity progress")
	}
	combined := append(bytes.TrimRight(formatted, "\n"), '\n', '\n')
	combined = append(combined, helper...)
	combined, err = format.Source(combined)
	if err != nil {
		return nil, returnTailContradiction(obligationProjectedConform, "combined source did not render")
	}
	renderedHelper, err := renderedFunctionHelper(combined, name)
	if err != nil {
		return nil, err
	}
	if physicalLines(renderedHelper) > functionLineLimit {
		return nil, returnTailContradiction(obligationRenderedCapacity, "rendered extracted helper including package and imports remains over the limit")
	}
	if !renderedCapacityProgress(beforeOuterHelper, renderedOuterHelper) {
		return nil, returnTailContradiction(obligationRenderedCapacity, "rendered outer helper made no capacity progress")
	}
	if bytes.Equal(combined, source) {
		return nil, returnTailContradiction(obligationRenderedCapacity, "terminal tail extraction made no progress")
	}
	if err := proof.consume(4, returnTailPredicateResult{Status: "PASS", Payload: append(append([]byte("rendered-capacity\x00"), renderedOuterHelper...), renderedHelper...), CandidateDigest: proofDigest(combined), Detail: fmt.Sprintf("outer_helper_lines=%d extracted_helper_lines=%d", physicalLines(renderedOuterHelper), physicalLines(renderedHelper))}); err != nil {
		return nil, err
	}
	return &returnTailCandidate{
		helperName: name,
		arguments:  bindings,
		helper:     helper,
		result:     combined,
		evidence: StrategyEvidence{
			Strategy:                 returnTailStrategy,
			Operation:                string(contract.Operation),
			ContractActivity:         contract.Activity,
			ContractInputEntity:      contract.InputEntity,
			ContractOutputEntity:     contract.OutputEntity,
			ContractInputSubjectKind: string(contract.InputSubjectKind),
			ContractSourceDigest:     contract.SourceDigest,
			ContractSemanticDigest:   contract.SemanticDigest,
			UsedInputFact:            contract.UsedInputFact,
			GeneratedOutputFact:      contract.GeneratedOutputFact,
			Subject:                  functionIdentity(fset, function),
			Helper:                   name,
			BeforeBytes:              len(source),
			AfterBytes:               len(combined),
			BeforeFunctionLines:      declarationLines(fset, function),
			AfterFunctionLines:       afterFunctionLines,
			RenderedHelperBytes:      len(renderedHelper),
			RenderedHelperLines:      physicalLines(renderedHelper),
			RenderedOuterHelperBytes: len(renderedOuterHelper),
			RenderedOuterHelperLines: physicalLines(renderedOuterHelper),
			Obligations:              obligationsFromProofStages(proof.stages),
			ContractObligations:      contractObligations,
			ProofStages:              proof.stages,
			PreflightObservations:    preflightObservationEvidence(contract, preflight),
		},
	}, nil
}

func preflightObservationEvidence(contract generation.OperationInputContractEvidence, observations []renderedCapacityObservation) []PreflightObservationEvidence {
	result := make([]PreflightObservationEvidence, 0, len(observations))
	for _, observation := range observations {
		item := PreflightObservationEvidence{
			Operation:              string(contract.Operation),
			Activity:               contract.Activity,
			InputEntity:            contract.InputEntity,
			InputSubjectKind:       string(contract.InputSubjectKind),
			Metric:                 sourcepolicy.DimensionFunctionLines,
			HelperMeasurementScope: renderedCapacityHelperMeasurementScope,
			Subject:                observation.subject,
			Receiver:               observation.receiver,
			FunctionStart:          observation.functionStart,
			FunctionEnd:            observation.functionEnd,
			DeclarationStart:       observation.declarationStart,
			DeclarationEnd:         observation.declarationEnd,
			SourceDigest:           observation.sourceDigest,
			ContractSourceDigest:   contract.SourceDigest,
			ContractSemanticDigest: contract.SemanticDigest,
			FunctionLines:          observation.functionLines,
			FunctionStatus:         string(observation.functionStatus),
			HelperStatus:           string(observation.helperStatus),
		}
		if observation.helperLines != nil {
			helperLines := *observation.helperLines
			item.HelperLines = &helperLines
		}
		if observation.helperFailure != nil {
			if failure, ok := errors.AsType[Failure](observation.helperFailure); ok {
				item.FailureReason = failure.Reason
				item.Failure = &PreflightFailureEvidence{
					Stage:         failure.Stage,
					Step:          failure.Step,
					Reason:        failure.Reason,
					UnknownClass:  failure.UnknownClass,
					NextOperation: failure.NextOperation,
					BlockedBy:     append([]string{}, failure.BlockedBy...),
					Diagnostics:   append([]string{}, failure.Diagnostics...),
				}
			} else {
				item.FailureReason = observation.helperFailure.Error()
			}
		}
		result = append(result, item)
	}
	return result
}

func returnTailContractObligations(values []generation.OperationInputContractObligationEvidence) ([]ContractObligationEvidence, bool) {
	if len(values) != len(returnTailObligations) {
		return nil, false
	}
	known := make(map[string]bool, len(values))
	result := make([]ContractObligationEvidence, 0, len(values))
	previousOutput := "FunctionInput"
	for _, value := range values {
		expectedActivity, expectedOutput := returnTailContractStage(value.Name)
		if known[value.Name] || !value.UsedInputFact || !value.GeneratedOutputFact || value.Activity != expectedActivity || value.InputEntity != previousOutput || value.OutputEntity != expectedOutput {
			return nil, false
		}
		known[value.Name] = true
		previousOutput = value.OutputEntity
		result = append(result, ContractObligationEvidence{
			Name: value.Name, Activity: value.Activity, InputEntity: value.InputEntity, OutputEntity: value.OutputEntity,
			UsedInputFact: value.UsedInputFact, GeneratedOutputFact: value.GeneratedOutputFact,
		})
	}
	for _, name := range returnTailObligations {
		if !known[name] {
			return nil, false
		}
	}
	if previousOutput != "ProjectedConformanceObligation" {
		return nil, false
	}
	return result, true
}

func returnTailContractStage(name string) (string, string) {
	switch name {
	case obligationReturnShape:
		return "ProveReturnShape", "ReturnShapeObligation"
	case obligationControlFlow:
		return "ProveControlFlow", "ControlFlowObligation"
	case obligationFreeBindings:
		return "ProveFreeBindings", "FreeBindingsObligation"
	case obligationCalleeEffects:
		return "ProveCalleeEffects", "CalleeEffectsObligation"
	case obligationRenderedCapacity:
		return "ProveRenderedCapacity", "RenderedCapacityObligation"
	case obligationProjectedConform:
		return "ProveProjectedConformance", "ProjectedConformanceObligation"
	default:
		return "", ""
	}
}

func stableReturnTailName(function string, suffix int) string {
	return fmt.Sprintf("%sExtractedReturnTail%02d", function, suffix)
}

func isErrorType(value types.Type) bool {
	return value != nil && types.Identical(value, types.Universe.Lookup("error").Type())
}

func returnTailSignatureEvidence(function *ast.FuncDecl, info *types.Info) bool {
	complete := true
	ast.Inspect(function.Type, func(node ast.Node) bool {
		identifier, ok := node.(*ast.Ident)
		if !ok || identifier.Name == "_" {
			return true
		}
		if info.Defs[identifier] == nil && info.Uses[identifier] == nil {
			complete = false
			return false
		}
		return true
	})
	return complete
}

func returnTailReturnsCompatible(statements []ast.Stmt, info *types.Info, resultType types.Type) bool {
	compatible := true
	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			returnStatement, ok := node.(*ast.ReturnStmt)
			if !ok {
				return true
			}
			if len(returnStatement.Results) != 1 {
				compatible = false
				return false
			}
			expressionType := info.TypeOf(returnStatement.Results[0])
			if expressionType == nil {
				identifier, isNil := returnStatement.Results[0].(*ast.Ident)
				if !isNil || identifier.Name != "nil" {
					compatible = false
					return false
				}
			} else if !types.AssignableTo(expressionType, resultType) {
				compatible = false
				return false
			}
			return true
		})
		if !compatible {
			return false
		}
	}
	return compatible
}

func hasReturnTailOuterHazard(statements []ast.Stmt, info *types.Info) bool {
	hazard := false
	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			if hazard {
				return false
			}
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			switch value := node.(type) {
			case *ast.DeferStmt, *ast.GoStmt, *ast.LabeledStmt, *ast.BranchStmt:
				hazard = true
			case *ast.CallExpr:
				if identifier, ok := value.Fun.(*ast.Ident); ok && (identifier.Name == "panic" || identifier.Name == "recover") {
					hazard = true
				}
			}
			return true
		})
		if hazard {
			return true
		}
	}
	return false
}

func hasReturnTailBindingHazard(functionBody *ast.BlockStmt, statements []ast.Stmt, bindings []suffixBinding, info *types.Info) error {
	free := make(map[types.Object]bool, len(bindings))
	for _, binding := range bindings {
		free[binding.object] = true
	}
	for _, statement := range statements {
		var hazard error
		ast.Inspect(statement, func(node ast.Node) bool {
			if hazard != nil {
				return false
			}
			switch value := node.(type) {
			case *ast.AssignStmt:
				for _, lhs := range value.Lhs {
					object, known := returnTailAssignedObject(lhs, info)
					if !known {
						hazard = failWithDiagnostics("derive-recipe", "prove-free-bindings", "FREE_BINDINGS_UNPROVEN", "DIRECT_MISSING", "restore-free-binding-evidence", []string{"obligation=" + obligationFreeBindings})
					} else if free[object] {
						hazard = returnTailContradiction(obligationFreeBindings, "terminal tail rebinds a free binding")
					}
				}
			case *ast.IncDecStmt:
				object, known := returnTailAssignedObject(value.X, info)
				if !known {
					hazard = failWithDiagnostics("derive-recipe", "prove-free-bindings", "FREE_BINDINGS_UNPROVEN", "DIRECT_MISSING", "restore-free-binding-evidence", []string{"obligation=" + obligationFreeBindings})
				} else if free[object] {
					hazard = returnTailContradiction(obligationFreeBindings, "terminal tail increments a free binding")
				}
			case *ast.UnaryExpr:
				if value.Op == token.AND {
					object, known := returnTailAssignedObject(value.X, info)
					if !known {
						hazard = failWithDiagnostics("derive-recipe", "prove-free-bindings", "FREE_BINDINGS_UNPROVEN", "DIRECT_MISSING", "restore-free-binding-evidence", []string{"obligation=" + obligationFreeBindings})
					} else if free[object] {
						hazard = returnTailContradiction(obligationFreeBindings, "terminal tail takes the address of a free binding")
					}
				}
			}
			return true
		})
		if hazard != nil {
			return hazard
		}
	}
	var hazard error
	ast.Inspect(functionBody, func(node ast.Node) bool {
		if hazard != nil {
			return false
		}
		if value, ok := node.(*ast.UnaryExpr); ok && value.Op == token.AND {
			object, known := returnTailAssignedObject(value.X, info)
			if !known {
				hazard = failWithDiagnostics("derive-recipe", "prove-free-bindings", "FREE_BINDINGS_UNPROVEN", "DIRECT_MISSING", "restore-free-binding-evidence", []string{"obligation=" + obligationFreeBindings})
			} else if free[object] {
				hazard = returnTailContradiction(obligationFreeBindings, "function takes the address of a selected free binding")
			}
			if hazard != nil {
				return false
			}
		}
		literal, ok := node.(*ast.FuncLit)
		if !ok {
			return true
		}
		ast.Inspect(literal.Body, func(inner ast.Node) bool {
			identifier, ok := inner.(*ast.Ident)
			if !ok || identifier.Name == "_" || identifier.Name == "nil" {
				return true
			}
			object := info.Uses[identifier]
			if object == nil && info.Defs[identifier] == nil {
				hazard = failWithDiagnostics("derive-recipe", "prove-free-bindings", "FREE_BINDINGS_UNPROVEN", "DIRECT_MISSING", "restore-free-binding-evidence", []string{"obligation=" + obligationFreeBindings})
				return false
			}
			if free[object] {
				hazard = returnTailContradiction(obligationFreeBindings, "closure captures a selected free binding")
				return false
			}
			return true
		})
		return false
	})
	return hazard
}

func returnTailAssignedObject(expression ast.Expr, info *types.Info) (types.Object, bool) {
	switch value := expression.(type) {
	case *ast.Ident:
		if value.Name == "_" {
			return nil, true
		}
		if object := info.Defs[value]; object != nil {
			return object, true
		}
		if object := info.Uses[value]; object != nil {
			return object, true
		}
		return nil, false
	case *ast.ParenExpr:
		return returnTailAssignedObject(value.X, info)
	case *ast.SelectorExpr:
		return returnTailAssignedObject(value.X, info)
	case *ast.IndexExpr:
		return returnTailAssignedObject(value.X, info)
	case *ast.IndexListExpr:
		return returnTailAssignedObject(value.X, info)
	case *ast.StarExpr:
		return returnTailAssignedObject(value.X, info)
	default:
		return nil, false
	}
}

func returnTailCalleeEffects(statements []ast.Stmt, evidence typeEvidence) error {
	var effectErr error
	ast.Inspect(&ast.BlockStmt{List: statements}, func(node ast.Node) bool {
		if effectErr != nil {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if returnTailTypedConversion(call, evidence.info) {
			return true
		}
		if !returnTailCalleeAllowed(call, evidence, map[*types.Func]bool{}) {
			effectErr = failWithDiagnostics("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", []string{"obligation=" + obligationCalleeEffects})
			return false
		}
		return true
	})
	return effectErr
}

func returnTailTypedConversion(call *ast.CallExpr, info *types.Info) bool {
	if call == nil || info == nil {
		return false
	}
	functionEvidence, functionKnown := info.Types[call.Fun]
	resultEvidence, resultKnown := info.Types[call]
	if !functionKnown || !resultKnown || !functionEvidence.IsType() || functionEvidence.Type == nil || resultEvidence.Type == nil {
		return false
	}
	return true
}

func returnTailCalleeAllowed(call *ast.CallExpr, evidence typeEvidence, visiting map[*types.Func]bool) bool {
	var object types.Object
	switch function := call.Fun.(type) {
	case *ast.Ident:
		object = evidence.info.Uses[function]
	case *ast.SelectorExpr:
		object = evidence.info.Uses[function.Sel]
	default:
		return false
	}
	switch value := object.(type) {
	case *types.Builtin:
		return value.Name() == "len"
	case *types.Func:
		if value.Pkg() != nil && value.Pkg().Path() == "fmt" && value.Name() == "Errorf" {
			return exactErrorfCall(value, call, evidence.info)
		}
		if value.Pkg() != nil && value.Pkg().Path() == "reflect" && value.Name() == "DeepEqual" {
			return exactDeepEqualSignature(value)
		}
		if value.Pkg() == nil || evidence.pkg == nil || value.Pkg() != evidence.pkg || value.Name() == "" {
			return false
		}
		declaration := evidence.funcs[value]
		if declaration == nil || declaration.Recv != nil || visiting[value] {
			return false
		}
		visiting[value] = true
		defer delete(visiting, value)
		return provenLocalPureFunction(declaration, evidence, visiting)
	default:
		return false
	}
}

func exactErrorfCall(function *types.Func, call *ast.CallExpr, info *types.Info) bool {
	signature, ok := function.Type().(*types.Signature)
	if !ok || !signature.Variadic() || signature.Params().Len() != 2 || signature.Results().Len() != 1 || !isErrorType(signature.Results().At(0).Type()) || len(call.Args) != 1 {
		return false
	}
	formatType := info.TypeOf(call.Args[0])
	return formatType != nil && types.AssignableTo(formatType, types.Typ[types.String]) && info.Types[call.Args[0]].Value != nil
}

func exactDeepEqualSignature(function *types.Func) bool {
	signature, ok := function.Type().(*types.Signature)
	return ok && !signature.Variadic() && signature.Params().Len() == 2 && signature.Results().Len() == 1 && signature.Results().At(0).Type() == types.Typ[types.Bool]
}

func provenLocalPureFunction(function *ast.FuncDecl, evidence typeEvidence, visiting map[*types.Func]bool) bool {
	local := make(map[types.Object]bool)
	ast.Inspect(function.Type, func(node ast.Node) bool {
		identifier, ok := node.(*ast.Ident)
		if ok {
			if object := evidence.info.Defs[identifier]; object != nil {
				local[object] = true
			}
		}
		return true
	})
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok {
			if object := evidence.info.Defs[identifier]; object != nil {
				local[object] = true
			}
		}
		return true
	})
	pure := true
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if !pure {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			pure = false
			return false
		}
		switch value := node.(type) {
		case *ast.DeferStmt, *ast.GoStmt, *ast.LabeledStmt, *ast.BranchStmt, *ast.SendStmt:
			pure = false
		case *ast.CallExpr:
			if !returnTailCalleeAllowed(value, evidence, visiting) {
				pure = false
			}
		case *ast.AssignStmt:
			for _, lhs := range value.Lhs {
				object, known := returnTailAssignedObject(lhs, evidence.info)
				if !known || !returnTailDirectLocalLValue(lhs, object, local) {
					pure = false
				}
			}
		case *ast.IncDecStmt:
			object, known := returnTailAssignedObject(value.X, evidence.info)
			if !known || !returnTailDirectLocalLValue(value.X, object, local) {
				pure = false
			}
		case *ast.UnaryExpr:
			if value.Op == token.AND {
				if _, ok := value.X.(*ast.CompositeLit); ok {
					break
				}
				pure = false
			}
		case *ast.Ident:
			if value.Name != "_" && value.Name != "nil" && evidence.info.Defs[value] == nil && evidence.info.Uses[value] == nil {
				pure = false
			}
			if object, ok := evidence.info.Uses[value].(*types.Var); ok && object.Parent() == evidence.pkg.Scope() {
				pure = false
			}
		}
		return true
	})
	return pure
}

func returnTailDirectLocalLValue(expression ast.Expr, object types.Object, local map[types.Object]bool) bool {
	if identifier, ok := expression.(*ast.Ident); ok && identifier.Name == "_" {
		return true
	}
	if parenthesized, ok := expression.(*ast.ParenExpr); ok {
		return returnTailDirectLocalLValue(parenthesized.X, object, local)
	}
	identifier, ok := expression.(*ast.Ident)
	return ok && object != nil && local[object] && identifier.Name != "_"
}

func returnTailContradiction(obligation, detail string) error {
	return knownSuffixContradiction("obligation=" + obligation + ": " + detail)
}
