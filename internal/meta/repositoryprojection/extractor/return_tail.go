package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const returnTailStrategy = "return-preserving-terminal-tail"
const suffixStrategy = "suffix-extraction"

func preparationProgressEvidence(contract generation.OperationInputContractEvidence, source []byte, subject string, before, after int) *PreparationProgressEvidence {
	if after <= 0 || before <= after {
		return nil
	}
	return &PreparationProgressEvidence{
		Operation:              string(contract.Operation),
		Activity:               contract.Activity,
		InputEntity:            contract.InputEntity,
		InputSubjectKind:       string(contract.InputSubjectKind),
		ContractSourceDigest:   contract.SourceDigest,
		ContractSemanticDigest: contract.SemanticDigest,
		Subject:                subject,
		SourceDigest:           proofDigest(source),
		BeforeOverage:          before,
		AfterOverage:           after,
		Status:                 "PASS",
	}
}

func buildReturnTailCandidate(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, evidence typeEvidence, existing map[string]bool, preflight []renderedCapacityObservation, helperProofRegistry map[string]returnTailHelperProof) (*returnTailCandidate, error) {
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		return nil, fail("derive-recipe", "admit-return-tail", "OPERATION_INPUT_CONTRACT_MISSING", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	if contract.Operation == "" || contract.Activity != "ExtractFunction" || contract.InputEntity != "FunctionInput" ||
		contract.OutputEntity != "OperationResult" || contract.InputSubjectKind != sourcepolicy.SubjectKindFunction ||
		!contract.UsedInputFact || !contract.GeneratedOutputFact || contract.SourceDigest == "" || contract.SemanticDigest == "" {
		return nil, fail("derive-recipe", "admit-return-tail", "OPERATION_INPUT_CONTRACT_UNPROVEN", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	evidence.contractSourceDigest = contract.SourceDigest
	evidence.contractSemanticDigest = contract.SemanticDigest
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
		candidate, candidateErr := tryReturnTailStart(root, logical, source, fset, file, function, evidence, contract, contractObligations, existing, startIndex, preflight, helperProofRegistry)
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

func tryReturnTailStart(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, evidence typeEvidence, contract generation.OperationInputContractEvidence, contractObligations []ContractObligationEvidence, existing map[string]bool, startIndex int, preflight []renderedCapacityObservation, helperProofRegistry map[string]returnTailHelperProof) (*returnTailCandidate, error) {
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
	calleeDependencies, err := returnTailCalleeEffects(statements, evidence, helperProofRegistry)
	if err != nil {
		return nil, err
	}
	if err := proof.consume(3, returnTailPredicateResult{Status: "PASS", Payload: returnTailCalleeEffectsPayload(source[start:end], calleeDependencies), CandidateDigest: proofDigest(source[start:end]), Detail: "all selected calls have closed typed effect evidence"}); err != nil {
		return nil, err
	}

	name := stableReturnTailName(function.Name.Name, startIndex+1)
	if existing[name] {
		return nil, returnTailContradiction(obligationRenderedCapacity, "return-tail helper identity already exists")
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
	beforeMeasurement, err := canonicalRenderedCapacity(beforeOuterHelper)
	if err != nil {
		return nil, err
	}
	renderedOuterHelper, err := renderedFunctionHelper(formatted, function.Name.Name)
	if err != nil {
		return nil, err
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
	outerMeasurement, err := canonicalRenderedCapacity(renderedOuterHelper)
	if err != nil {
		return nil, err
	}
	helperMeasurement, err := canonicalRenderedCapacity(renderedHelper)
	if err != nil {
		return nil, err
	}
	beforeCapacity := renderedCapacitySnapshot{overage: beforeMeasurement.overage}
	afterCapacity := renderedCapacitySnapshot{overage: outerMeasurement.overage + helperMeasurement.overage}
	if !renderedCapacityProgress(beforeCapacity, afterCapacity) {
		return nil, returnTailContradiction(obligationRenderedCapacity, fmt.Sprintf("rendered capacity overage did not strictly decrease: before=%d after=%d", beforeCapacity.overage, afterCapacity.overage))
	}
	if bytes.Equal(combined, source) {
		return nil, returnTailContradiction(obligationRenderedCapacity, "terminal tail extraction made no progress")
	}
	helperProof, err := registerReturnTailHelperProof(root, logical, combined, name, contract, proof, evidence, helperProofRegistry)
	if err != nil {
		return nil, err
	}
	if !sameReturnTailDependencies(calleeDependencies, helperProof.dependencies) {
		return nil, fail("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", nil)
	}
	helperProofRegistry[name] = helperProof
	progress := preparationProgressEvidence(contract, source, functionIdentity(fset, function), beforeCapacity.overage, afterCapacity.overage)
	return &returnTailCandidate{
		helperName: name,
		arguments:  bindings,
		helper:     helper,
		result:     combined,
		evidence: StrategyEvidence{
			Strategy:                      returnTailStrategy,
			Operation:                     string(contract.Operation),
			ContractActivity:              contract.Activity,
			ContractInputEntity:           contract.InputEntity,
			ContractOutputEntity:          contract.OutputEntity,
			ContractInputSubjectKind:      string(contract.InputSubjectKind),
			ContractSourceDigest:          contract.SourceDigest,
			ContractSemanticDigest:        contract.SemanticDigest,
			UsedInputFact:                 contract.UsedInputFact,
			GeneratedOutputFact:           contract.GeneratedOutputFact,
			Subject:                       functionIdentity(fset, function),
			Helper:                        name,
			BeforeBytes:                   len(source),
			AfterBytes:                    len(combined),
			BeforeFunctionLines:           declarationLines(fset, function),
			AfterFunctionLines:            afterFunctionLines,
			BeforeRenderedCapacityOverage: beforeCapacity.overage,
			AfterRenderedCapacityOverage:  afterCapacity.overage,
			RenderedHelperBytes:           helperMeasurement.bytes,
			RenderedHelperLines:           helperMeasurement.lines,
			RenderedOuterHelperBytes:      outerMeasurement.bytes,
			RenderedOuterHelperLines:      outerMeasurement.lines,
			PreparationProgress:           progress,
			Obligations:                   obligationsFromProofStages(proof.stages),
			ContractObligations:           contractObligations,
			ProofStages:                   proof.stages,
			CalleeDependencies:            calleeDependencies,
			PreflightObservations:         preflightObservationEvidence(contract, preflight),
		},
	}, nil
}

func preflightObservationEvidence(contract generation.OperationInputContractEvidence, observations []renderedCapacityObservation) []PreflightObservationEvidence {
	result := make([]PreflightObservationEvidence, 0, len(observations))
	for _, observation := range observations {
		item := PreflightObservationEvidence{
			Operation:                       string(contract.Operation),
			Activity:                        contract.Activity,
			InputEntity:                     contract.InputEntity,
			InputSubjectKind:                string(contract.InputSubjectKind),
			Metric:                          sourcepolicy.DimensionFunctionLines,
			HelperMeasurementScope:          renderedCapacityHelperMeasurementScope,
			Subject:                         observation.subject,
			Receiver:                        observation.receiver,
			FunctionStart:                   observation.functionStart,
			FunctionEnd:                     observation.functionEnd,
			DeclarationStart:                observation.declarationStart,
			DeclarationEnd:                  observation.declarationEnd,
			SourceDigest:                    observation.sourceDigest,
			ContractSourceDigest:            contract.SourceDigest,
			ContractSemanticDigest:          contract.SemanticDigest,
			FunctionLines:                   observation.functionLines,
			FunctionRenderedCapacityOverage: observation.functionOverage,
			FunctionStatus:                  string(observation.functionStatus),
			HelperStatus:                    string(observation.helperStatus),
		}
		if observation.helperLines != nil {
			helperLines := *observation.helperLines
			item.HelperLines = &helperLines
		}
		if observation.helperOverage != nil {
			helperOverage := *observation.helperOverage
			item.HelperRenderedCapacityOverage = &helperOverage
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

func suffixStrategyEvidence(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, candidate *suffixCandidate, preflight []renderedCapacityObservation) (*StrategyEvidence, error) {
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		return nil, fail("derive-recipe", "admit-suffix", "OPERATION_INPUT_CONTRACT_MISSING", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	if contract.Operation == "" || contract.Activity != "ExtractFunction" || contract.InputEntity != "FunctionInput" ||
		contract.OutputEntity != "OperationResult" || contract.InputSubjectKind != sourcepolicy.SubjectKindFunction ||
		!contract.UsedInputFact || !contract.GeneratedOutputFact || contract.SourceDigest == "" || contract.SemanticDigest == "" {
		return nil, fail("derive-recipe", "admit-suffix", "OPERATION_INPUT_CONTRACT_UNPROVEN", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	afterFunctionLines, ok := namedFunctionLines(candidate.result, function.Name.Name)
	if !ok {
		return nil, fail("derive-recipe", "admit-suffix", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
	}
	progress := preparationProgressEvidence(contract, source, functionIdentity(fset, function), candidate.beforeRenderedCapacityOverage, candidate.afterRenderedCapacityOverage)
	return &StrategyEvidence{
		Strategy:                      suffixStrategy,
		Operation:                     string(contract.Operation),
		ContractActivity:              contract.Activity,
		ContractInputEntity:           contract.InputEntity,
		ContractOutputEntity:          contract.OutputEntity,
		ContractInputSubjectKind:      string(contract.InputSubjectKind),
		ContractSourceDigest:          contract.SourceDigest,
		ContractSemanticDigest:        contract.SemanticDigest,
		UsedInputFact:                 contract.UsedInputFact,
		GeneratedOutputFact:           contract.GeneratedOutputFact,
		Subject:                       functionIdentity(fset, function),
		Helper:                        candidate.helperName,
		BeforeBytes:                   len(source),
		AfterBytes:                    len(candidate.result),
		BeforeFunctionLines:           declarationLines(fset, function),
		AfterFunctionLines:            afterFunctionLines,
		BeforeRenderedCapacityOverage: candidate.beforeRenderedCapacityOverage,
		AfterRenderedCapacityOverage:  candidate.afterRenderedCapacityOverage,
		RenderedHelperBytes:           candidate.renderedHelper.bytes,
		RenderedHelperLines:           candidate.renderedHelper.lines,
		RenderedOuterHelperBytes:      candidate.renderedOuter.bytes,
		RenderedOuterHelperLines:      candidate.renderedOuter.lines,
		PreparationProgress:           progress,
		PreflightObservations:         preflightObservationEvidence(contract, preflight),
	}, nil
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
	return fmt.Sprintf("%sExtractedReturnTail%02d", safeGeneratedFunctionPrefix(function), suffix)
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

func returnTailCalleeEffects(statements []ast.Stmt, evidence typeEvidence, helperProofRegistry map[string]returnTailHelperProof) ([]CalleeDependencyEvidence, error) {
	context := returnTailValidationContext{
		visiting:        make(map[*types.Func]bool),
		memo:            make(map[*types.Func]returnTailValidation),
		proofBodyVisits: make(map[*types.Func]int),
	}
	return returnTailCalleeEffectsWithContext(statements, evidence, helperProofRegistry, &context)
}

func returnTailCalleeEffectsWithContext(statements []ast.Stmt, evidence typeEvidence, helperProofRegistry map[string]returnTailHelperProof, context *returnTailValidationContext) ([]CalleeDependencyEvidence, error) {
	if context == nil {
		context = &returnTailValidationContext{
			visiting:        make(map[*types.Func]bool),
			memo:            make(map[*types.Func]returnTailValidation),
			proofBodyVisits: make(map[*types.Func]int),
		}
	}
	dependencies := make([]CalleeDependencyEvidence, 0)
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
		validation := returnTailCalleeAllowed(call, evidence, helperProofRegistry, context)
		if !validation.valid {
			diagnostics := []string{"obligation=" + obligationCalleeEffects}
			if function, ok := returnTailCalleeFunction(call, evidence); ok {
				if proof, exists := helperProofRegistry[function.Name()]; exists {
					diagnostics = append(diagnostics, returnTailHelperProofDiagnostics(function, evidence, proof)...)
				}
			}
			effectErr = failWithDiagnostics("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", diagnostics)
			return false
		}
		if dependency, ok := returnTailCalleeDependency(call, evidence, helperProofRegistry); ok {
			dependencies = append(dependencies, dependency)
		}
		return true
	})
	sort.Slice(dependencies, func(left, right int) bool {
		if dependencies[left].Name != dependencies[right].Name {
			return dependencies[left].Name < dependencies[right].Name
		}
		return dependencies[left].EvidenceID < dependencies[right].EvidenceID
	})
	return dependencies, effectErr
}

func returnTailCalleeEffectsPayload(source []byte, dependencies []CalleeDependencyEvidence) []byte {
	payload := append([]byte("callee-effects\x00"), source...)
	for _, dependency := range dependencies {
		payload = append(payload, '\x00')
		payload = append(payload, dependency.Name...)
		payload = append(payload, '\x00')
		payload = append(payload, dependency.ObjectType...)
		payload = append(payload, '\x00')
		payload = append(payload, dependency.SignatureDigest...)
		payload = append(payload, '\x00')
		payload = append(payload, dependency.BodyDigest...)
		payload = append(payload, '\x00')
		payload = append(payload, dependency.EvidenceID...)
		payload = append(payload, '\x00')
		payload = append(payload, dependency.Declaration...)
	}
	return payload
}

func returnTailTypedConversion(call *ast.CallExpr, info *types.Info) bool {
	if call == nil || info == nil {
		return false
	}
	resultEvidence, resultKnown := info.Types[call]
	if !resultKnown || resultEvidence.Type == nil || !returnTailTypeExpression(call.Fun, info) {
		return false
	}
	return true
}

func returnTailTypeExpression(expression ast.Expr, info *types.Info) bool {
	evidence, known := info.Types[expression]
	if known && evidence.IsType() && evidence.Type != nil {
		return true
	}
	parenthesized, ok := expression.(*ast.ParenExpr)
	return ok && returnTailTypeExpression(parenthesized.X, info)
}

func returnTailCalleeAllowed(call *ast.CallExpr, evidence typeEvidence, helperProofRegistry map[string]returnTailHelperProof, context *returnTailValidationContext) returnTailValidation {
	var object types.Object
	switch function := call.Fun.(type) {
	case *ast.Ident:
		object = evidence.info.Uses[function]
	case *ast.SelectorExpr:
		object = evidence.info.Uses[function.Sel]
	default:
		return returnTailValidation{}
	}
	switch value := object.(type) {
	case *types.Builtin:
		return returnTailValidation{valid: value.Name() == "len"}
	case *types.Func:
		if value.Pkg() != nil && value.Pkg().Path() == "fmt" && value.Name() == "Errorf" {
			return returnTailValidation{valid: exactErrorfCall(value, call, evidence.info)}
		}
		if value.Pkg() != nil && value.Pkg().Path() == "reflect" && value.Name() == "DeepEqual" {
			return returnTailValidation{valid: exactDeepEqualSignature(value)}
		}
		if value.Pkg() == nil || evidence.pkg == nil || value.Pkg() != evidence.pkg || value.Name() == "" {
			return returnTailValidation{}
		}
		if context == nil {
			context = &returnTailValidationContext{
				visiting:        make(map[*types.Func]bool),
				memo:            make(map[*types.Func]returnTailValidation),
				proofBodyVisits: make(map[*types.Func]int),
			}
		}
		if result, ok := context.memo[value]; ok {
			return result
		}
		declaration := evidence.funcs[value]
		if declaration == nil || declaration.Recv != nil || context.visiting[value] {
			return returnTailValidation{}
		}
		helperProof, hasHelperProof := helperProofRegistry[value.Name()]
		if !hasHelperProof && returnTailGeneratedHelperName(value.Name()) {
			return returnTailValidation{}
		}
		allowedGlobals := map[types.Object]bool(nil)
		if hasHelperProof {
			if !returnTailHelperProofMatches(value, declaration, evidence, helperProof) {
				return returnTailValidation{}
			}
			var safe bool
			allowedGlobals, _, safe = returnTailGlobalReadEvidence(declaration, evidence)
			if !safe {
				return returnTailValidation{}
			}
		}
		context.visiting[value] = true
		defer delete(context.visiting, value)
		context.proofBodyVisits[value]++
		validation := provenLocalPureFunction(declaration, evidence, helperProofRegistry, allowedGlobals, context)
		if hasHelperProof && !sameReturnTailDependencies(validation.dependencies, helperProof.dependencies) {
			validation.valid = false
		}
		context.memo[value] = validation
		return validation
	default:
		return returnTailValidation{}
	}
}

func returnTailCalleeDependency(call *ast.CallExpr, evidence typeEvidence, helperProofRegistry map[string]returnTailHelperProof) (CalleeDependencyEvidence, bool) {
	if call == nil || evidence.info == nil || evidence.fset == nil {
		return CalleeDependencyEvidence{}, false
	}
	function, ok := returnTailCalleeFunction(call, evidence)
	if !ok {
		return CalleeDependencyEvidence{}, false
	}
	proof, ok := helperProofRegistry[function.Name()]
	if !ok {
		return CalleeDependencyEvidence{}, false
	}
	declaration := evidence.funcs[function]
	if declaration == nil || !returnTailHelperProofMatches(function, declaration, evidence, proof) {
		return CalleeDependencyEvidence{}, false
	}
	signatureDigest, bodyDigest, ok := returnTailFunctionDigests(evidence.fset, declaration)
	if !ok {
		return CalleeDependencyEvidence{}, false
	}
	return CalleeDependencyEvidence{
		Name:            function.Name(),
		ObjectType:      returnTailCanonicalType(function.Type()),
		SignatureDigest: signatureDigest,
		BodyDigest:      bodyDigest,
		EvidenceID:      proof.calleeEffectsEvidenceID,
		Declaration:     returnTailObjectIdentity(function, evidence),
	}, true
}

func returnTailCalleeFunction(call *ast.CallExpr, evidence typeEvidence) (*types.Func, bool) {
	if call == nil || evidence.info == nil {
		return nil, false
	}
	var object types.Object
	switch function := call.Fun.(type) {
	case *ast.Ident:
		object = evidence.info.Uses[function]
	case *ast.SelectorExpr:
		object = evidence.info.Uses[function.Sel]
	default:
		return nil, false
	}
	function, ok := object.(*types.Func)
	return function, ok
}

func returnTailHelperProofDiagnostics(function *types.Func, evidence typeEvidence, proof returnTailHelperProof) []string {
	declaration := evidence.funcs[function]
	signatureDigest, bodyDigest, digestOK := returnTailFunctionDigests(evidence.fset, declaration)
	_, globalReadIdentities, globalsSafe := returnTailGlobalReadEvidence(declaration, evidence)
	return []string{
		"callee=" + function.Name(),
		"proof-object=" + proof.helperObjectIdentity,
		"current-object=" + returnTailObjectIdentity(function, evidence),
		"proof-type=" + proof.helperType,
		"current-type=" + returnTailCanonicalType(function.Type()),
		"proof-signature=" + proof.helperSignatureDigest,
		"current-signature=" + signatureDigest,
		"proof-body=" + proof.helperBodyDigest,
		"current-body=" + bodyDigest,
		"digest-known=" + strconv.FormatBool(digestOK),
		"globals-safe=" + strconv.FormatBool(globalsSafe),
		"proof-globals=" + strings.Join(proof.globalReadIdentities, ","),
		"current-globals=" + strings.Join(globalReadIdentities, ","),
	}
}

func returnTailGeneratedHelperName(name string) bool {
	marker := "ExtractedReturnTail"
	index := strings.Index(name, marker)
	if index <= 0 || index+len(marker) >= len(name) {
		return false
	}
	for _, value := range name[index+len(marker):] {
		if value < '0' || value > '9' {
			return false
		}
	}
	return true
}

func sameReturnTailDependencies(left, right []CalleeDependencyEvidence) bool {
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

func provenLocalPureFunction(function *ast.FuncDecl, evidence typeEvidence, helperProofRegistry map[string]returnTailHelperProof, allowedGlobals map[types.Object]bool, context *returnTailValidationContext) returnTailValidation {
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
	validation := returnTailValidation{valid: true, dependencies: make([]CalleeDependencyEvidence, 0)}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if !validation.valid {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			validation.valid = false
			return false
		}
		switch value := node.(type) {
		case *ast.DeferStmt, *ast.GoStmt, *ast.LabeledStmt, *ast.BranchStmt, *ast.SendStmt:
			validation.valid = false
		case *ast.CallExpr:
			if returnTailTypedConversion(value, evidence.info) {
				break
			}
			calleeValidation := returnTailCalleeAllowed(value, evidence, helperProofRegistry, context)
			if !calleeValidation.valid {
				validation.valid = false
				break
			}
			if dependency, ok := returnTailCalleeDependency(value, evidence, helperProofRegistry); ok {
				validation.dependencies = append(validation.dependencies, dependency)
			}
		case *ast.AssignStmt:
			for _, lhs := range value.Lhs {
				object, known := returnTailAssignedObject(lhs, evidence.info)
				if !known || !returnTailDirectLocalLValue(lhs, object, local) {
					validation.valid = false
				}
			}
		case *ast.IncDecStmt:
			object, known := returnTailAssignedObject(value.X, evidence.info)
			if !known || !returnTailDirectLocalLValue(value.X, object, local) {
				validation.valid = false
			}
		case *ast.RangeStmt:
			if rangeType := evidence.info.TypeOf(value.X); rangeType != nil {
				switch rangeType.Underlying().(type) {
				case *types.Chan, *types.Signature:
					validation.valid = false
				}
			}
			if value.Tok == token.ASSIGN {
				for _, target := range []ast.Expr{value.Key, value.Value} {
					if target == nil {
						continue
					}
					object, known := returnTailAssignedObject(target, evidence.info)
					if !known || !returnTailDirectLocalLValue(target, object, local) {
						validation.valid = false
					}
				}
			}
		case *ast.UnaryExpr:
			if value.Op == token.ARROW {
				validation.valid = false
			} else if value.Op == token.AND {
				if _, ok := value.X.(*ast.CompositeLit); ok {
					break
				}
				validation.valid = false
			}
		case *ast.Ident:
			if value.Name != "_" && value.Name != "nil" && evidence.info.Defs[value] == nil && evidence.info.Uses[value] == nil {
				validation.valid = false
			}
			if object, ok := evidence.info.Uses[value].(*types.Var); ok && object.Parent() == evidence.pkg.Scope() && !allowedGlobals[object] {
				validation.valid = false
			}
		}
		return true
	})
	sort.Slice(validation.dependencies, func(left, right int) bool {
		if validation.dependencies[left].Name != validation.dependencies[right].Name {
			return validation.dependencies[left].Name < validation.dependencies[right].Name
		}
		return validation.dependencies[left].EvidenceID < validation.dependencies[right].EvidenceID
	})
	return validation
}

func registerReturnTailHelperProof(root, logical string, source []byte, helperName string, contract generation.OperationInputContractEvidence, proof returnTailProofChain, evidence typeEvidence, helperProofRegistry map[string]returnTailHelperProof) (returnTailHelperProof, error) {
	if helperProofRegistry == nil {
		return returnTailHelperProof{}, nil
	}
	helperSet := token.NewFileSet()
	helperFile, err := parser.ParseFile(helperSet, logical, source, parser.ParseComments)
	if err != nil {
		return returnTailHelperProof{}, fail("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", nil)
	}
	var helperDeclaration *ast.FuncDecl
	for _, declaration := range helperFile.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv == nil && function.Name != nil && function.Name.Name == helperName {
			helperDeclaration = function
			break
		}
	}
	if helperDeclaration == nil {
		return returnTailHelperProof{}, fail("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", nil)
	}
	helperEvidence, err := checkTypes(root, logical, helperSet, helperFile, helperDeclaration)
	if err != nil {
		return returnTailHelperProof{}, err
	}
	helperEvidence.contractSourceDigest = contract.SourceDigest
	helperEvidence.contractSemanticDigest = contract.SemanticDigest
	helperObject, ok := helperEvidence.info.Defs[helperDeclaration.Name].(*types.Func)
	if !ok || helperObject == nil || helperObject.Pkg() == nil || helperEvidence.pkg == nil || helperObject.Pkg() != helperEvidence.pkg {
		return returnTailHelperProof{}, fail("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", nil)
	}
	globalObjects, globalReadIdentities, safe := returnTailGlobalReadEvidence(helperDeclaration, helperEvidence)
	validationContext := &returnTailValidationContext{
		visiting:        make(map[*types.Func]bool),
		memo:            make(map[*types.Func]returnTailValidation),
		proofBodyVisits: make(map[*types.Func]int),
	}
	validation := provenLocalPureFunction(helperDeclaration, helperEvidence, helperProofRegistry, globalObjects, validationContext)
	if !safe || !validation.valid {
		return returnTailHelperProof{}, failWithDiagnostics("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", []string{"helper=" + helperName})
	}
	dependencies := validation.dependencies
	signatureDigest, bodyDigest, ok := returnTailFunctionDigests(helperSet, helperDeclaration)
	if !ok {
		return returnTailHelperProof{}, fail("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", nil)
	}
	calleeEvidenceID := ""
	if len(proof.stages) > 3 {
		calleeEvidenceID = proof.stages[3].OutputEvidenceID
	}
	if calleeEvidenceID == "" {
		return returnTailHelperProof{}, fail("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", nil)
	}
	return returnTailHelperProof{
		helperName:              helperName,
		helperObjectIdentity:    returnTailObjectIdentity(helperObject, helperEvidence),
		helperSignatureDigest:   signatureDigest,
		helperBodyDigest:        bodyDigest,
		helperType:              returnTailCanonicalType(helperObject.Type()),
		contractSourceDigest:    contract.SourceDigest,
		contractSemanticDigest:  contract.SemanticDigest,
		calleeEffectsEvidenceID: calleeEvidenceID,
		globalReadIdentities:    globalReadIdentities,
		dependencies:            dependencies,
	}, nil
}

func returnTailHelperProofMatches(function *types.Func, declaration *ast.FuncDecl, evidence typeEvidence, proof returnTailHelperProof) bool {
	if function == nil || declaration == nil || evidence.pkg == nil || function.Pkg() != evidence.pkg || proof.helperName != function.Name() ||
		proof.helperObjectIdentity == "" || proof.helperObjectIdentity != returnTailObjectIdentity(function, evidence) ||
		proof.calleeEffectsEvidenceID == "" || proof.contractSourceDigest == "" || proof.contractSemanticDigest == "" ||
		evidence.contractSourceDigest != proof.contractSourceDigest || evidence.contractSemanticDigest != proof.contractSemanticDigest ||
		returnTailCanonicalType(function.Type()) != proof.helperType {
		return false
	}
	signatureDigest, bodyDigest, ok := returnTailFunctionDigests(evidence.fset, declaration)
	if !ok || signatureDigest != proof.helperSignatureDigest || bodyDigest != proof.helperBodyDigest {
		return false
	}
	_, globalReadIdentities, safe := returnTailGlobalReadEvidence(declaration, evidence)
	return safe && sameReturnTailStrings(globalReadIdentities, proof.globalReadIdentities)
}

func returnTailFunctionDigests(fset *token.FileSet, function *ast.FuncDecl) (string, string, bool) {
	if fset == nil || function == nil || function.Type == nil || function.Body == nil {
		return "", "", false
	}
	var signature, body bytes.Buffer
	if err := format.Node(&signature, fset, function.Type); err != nil {
		return "", "", false
	}
	if err := format.Node(&body, fset, function.Body); err != nil {
		return "", "", false
	}
	return proofDigest(signature.Bytes()), proofDigest(body.Bytes()), true
}

func returnTailCanonicalType(value types.Type) string {
	if value == nil {
		return ""
	}
	return types.TypeString(value, func(pkg *types.Package) string {
		if pkg == nil {
			return ""
		}
		return pkg.Path()
	})
}

func returnTailGlobalReadEvidence(function *ast.FuncDecl, evidence typeEvidence) (map[types.Object]bool, []string, bool) {
	objects := make(map[types.Object]bool)
	if function == nil || evidence.info == nil || evidence.pkg == nil {
		return objects, nil, false
	}
	safe := true
	global := func(expression ast.Expr) (types.Object, bool) {
		for {
			parenthesized, ok := expression.(*ast.ParenExpr)
			if !ok {
				break
			}
			expression = parenthesized.X
		}
		identifier, ok := expression.(*ast.Ident)
		if !ok {
			return nil, false
		}
		object, ok := evidence.info.Uses[identifier].(*types.Var)
		return object, ok && object.Parent() == evidence.pkg.Scope()
	}
	containsGlobal := func(node ast.Node) bool {
		found := false
		ast.Inspect(node, func(inner ast.Node) bool {
			identifier, ok := inner.(*ast.Ident)
			if ok {
				if object, isGlobal := global(identifier); isGlobal {
					objects[object] = true
					found = true
				}
			}
			return !found
		})
		return found
	}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if !safe {
			return false
		}
		switch value := node.(type) {
		case *ast.FuncLit:
			safe = false
			return false
		case *ast.AssignStmt:
			for _, lhs := range value.Lhs {
				if object, isGlobal := global(lhs); isGlobal {
					objects[object] = true
					safe = false
				}
			}
			for _, rhs := range value.Rhs {
				if containsGlobal(rhs) {
					safe = false
				}
			}
		case *ast.IncDecStmt:
			if object, isGlobal := global(value.X); isGlobal {
				objects[object] = true
				safe = false
			}
		case *ast.UnaryExpr:
			if value.Op == token.ARROW {
				if object, isGlobal := global(value.X); isGlobal {
					objects[object] = true
					safe = false
				}
			} else if value.Op == token.AND {
				if object, isGlobal := global(value.X); isGlobal {
					objects[object] = true
					safe = false
				}
			}
		case *ast.RangeStmt:
			if object, isGlobal := global(value.X); isGlobal {
				objects[object] = true
				safe = false
			}
			if value.Tok == token.ASSIGN {
				if object, isGlobal := global(value.Key); isGlobal {
					objects[object] = true
					safe = false
				}
				if object, isGlobal := global(value.Value); isGlobal {
					objects[object] = true
					safe = false
				}
			}
		case *ast.ValueSpec:
			for _, expression := range value.Values {
				if containsGlobal(expression) {
					safe = false
				}
			}
		case *ast.CallExpr:
			if containsGlobal(value.Fun) {
				safe = false
			}
			for _, argument := range value.Args {
				if containsGlobal(argument) {
					safe = false
				}
			}
		case *ast.SelectorExpr:
			if containsGlobal(value.X) {
				safe = false
			}
		case *ast.IndexExpr:
			if containsGlobal(value.X) {
				safe = false
			}
		case *ast.IndexListExpr:
			if containsGlobal(value.X) {
				safe = false
			}
		case *ast.Ident:
			if object, isGlobal := global(value); isGlobal {
				objects[object] = true
			}
		}
		return true
	})
	identities := make([]string, 0, len(objects))
	for object := range objects {
		identities = append(identities, returnTailObjectIdentity(object, evidence))
	}
	sort.Strings(identities)
	return objects, identities, safe
}

func returnTailObjectIdentity(object types.Object, evidence typeEvidence) string {
	if object == nil {
		return ""
	}
	packagePath := ""
	if object.Pkg() != nil {
		packagePath = object.Pkg().Path()
	}
	return packagePath + "\x00" + object.Name() + "\x00" + returnTailCanonicalType(object.Type())
}

func returnTailPositionIdentity(root string, position token.Position) string {
	filename := filepath.ToSlash(position.Filename)
	if root != "" && filepath.IsAbs(filename) {
		if relative, err := filepath.Rel(root, filename); err == nil {
			filename = filepath.ToSlash(relative)
		}
	}
	return filename + ":" + strconv.Itoa(position.Line) + ":" + strconv.Itoa(position.Column)
}

func sameReturnTailStrings(left, right []string) bool {
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
