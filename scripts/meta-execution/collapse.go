package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type collapseSourceInspection struct {
	Subject              string
	Receiver             string
	Signature            string
	OutsideDeclarations  []string
	CommentGroups        []string
	StartLine            int
	EndLine              int
	AssignmentReturn     bool
	SingleReturn         bool
	ReturnExpression     string
	CommentsPreserved    bool
}

type collapseInstanceEvidence struct {
	Operation                     string                  `json:"operation"`
	Subject                       string                  `json:"subject"`
	Receiver                      string                  `json:"receiver"`
	InputContractSourceDigest     string                  `json:"input_contract_source_digest"`
	InputContractSemanticDigest   string                  `json:"input_contract_semantic_digest"`
	BeforeSignature               string                  `json:"before_signature"`
	AfterSignature                string                  `json:"after_signature"`
	BeforeCommentGroups           []string                `json:"before_comment_groups"`
	AfterCommentGroups            []string                `json:"after_comment_groups"`
	BeforeOutsideDeclarations     []string                `json:"before_outside_declarations"`
	AfterOutsideDeclarations      []string                `json:"after_outside_declarations"`
	BeforeSourceDigest             string                  `json:"before_source_digest"`
	AfterSourceDigest              string                  `json:"after_source_digest"`
	Before                         []byte                  `json:"before"`
	After                          []byte                  `json:"after"`
	ChangedFiles                   []string                `json:"changed_files"`
	PreflightCount                 int                     `json:"preflight_count"`
	ApplyCount                     int                     `json:"apply_count"`
	Process                        operationReplayEvidence `json:"process"`
}

func validateCollapseAction(action generation.Action) *operationError {
	binding, ok := generation.BindingForOperation(generation.DefaultRegistry(), sourcepolicy.OperationCollapseAssign)
	if !ok {
		return newOperationError("validate-action", "resolve-collapse-binding", "ACTION_BINDING_UNAVAILABLE", "DIRECT_MISSING", "restore-operation-contract")
	}
	expectedIndicatorID, indicatorIDOK := collapseIndicatorID(action.SourceIndicator)
	valid := indicatorIDOK && action.IndicatorID == expectedIndicatorID && action.Subject != "" &&
		action.MetricID == sourcepolicy.DimensionRefactorAssign &&
		action.Applicability == sourcepolicy.ApplicabilityApplicable &&
		action.ApplicabilityRule == sourcepolicy.ApplicabilityRuleDefault &&
		action.ApplicabilityReason == sourcepolicy.ApplicabilityReasonCatalogApplicable &&
		!action.Blocking &&
		action.Operation == binding.Operation && action.Activity == binding.Activity && action.Output == binding.Output &&
		action.SubjectKind == binding.InputSubjectKind && action.InputSubjectKind == binding.InputSubjectKind &&
		action.InputContractSourceDigest == binding.InputContractSourceDigest &&
		action.InputContractSemanticDigest == binding.InputContractSemanticDigest &&
		action.IndependenceGroupID == binding.IndependenceGroupID &&
		action.ProofChoice == binding.ProofChoice && string(action.MetricProofChoice) == string(binding.ProofChoice) &&
		action.MetricProducer == "linecaps.Analyze" && action.MetricConsumer == "refactor-planner" &&
		action.Executor == binding.Executor && action.Evaluator == binding.Evaluator &&
		slices.Equal(action.RequiredIndicatorIDs, binding.RequiredIndicatorIDs) &&
		action.ReceiptRequired == binding.ReceiptRequired && action.Priority == binding.Priority &&
		action.SourceIndicator.Subject == action.Subject &&
		action.SourceIndicator.SubjectKind == binding.InputSubjectKind &&
		action.SourceIndicator.Proof == sourcepolicy.ProofRegression &&
		action.SourceIndicator.Producer == "linecaps.Analyze" &&
		action.SourceIndicator.Consumer == "refactor-planner" &&
		action.SourceIndicator.Operation == binding.Operation &&
		action.SourceIndicator.MetricID == sourcepolicy.DimensionRefactorAssign &&
		action.SourceIndicator.Applicability == sourcepolicy.ApplicabilityApplicable &&
		action.SourceIndicator.ApplicabilityRule == sourcepolicy.ApplicabilityRuleDefault &&
		action.SourceIndicator.ApplicabilityReason == sourcepolicy.ApplicabilityReasonCatalogApplicable &&
		!action.SourceIndicator.Blocking &&
		!action.SourceIndicator.Satisfied && action.IndicatorOutcome == action.SourceIndicator.Outcome() && action.IndicatorOutcome.Actionable()
	if !valid {
		return newOperationError("validate-action", "bind-collapse-action", "ACTION_BINDING_INVALID", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	return nil
}

func executeCollapse(workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action, trace metaExecutionTrace) (operationMaterialization, *operationError) {
	subject, err := sourcepolicy.ParseSourceSubject(action.Subject)
	if err != nil {
		return operationMaterialization{}, newOperationError("observe-plan", "parse-collapse-subject", "SUBJECT_COORDINATE_MALFORMED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	first, firstErr := materializeCollapse(workspace, gitDir, metricsPath, plan, action, subject, trace, "first")
	if firstErr != nil {
		second, secondErr := materializeCollapse(workspace, gitDir, metricsPath, plan, action, subject, trace, "replay")
		if len(first.Canonical) == 0 || secondErr == nil || len(second.Canonical) == 0 {
			return first, firstErr
		}
		if !sameOperationError(firstErr, secondErr) || !bytes.Equal(first.Canonical, second.Canonical) || first.InstanceDigest != second.InstanceDigest {
			return first, newOperationError("replay-operation", "compare-instance-evidence", "REPLAY_EVIDENCE_MISMATCH", "KNOWN_CONTRADICTION", "report-counterexample")
		}
		return first, firstErr
	}
	second, secondErr := materializeCollapse(workspace, gitDir, metricsPath, plan, action, subject, trace, "replay")
	if secondErr != nil {
		return second, secondErr
	}
	if !bytes.Equal(first.Canonical, second.Canonical) || first.InstanceDigest != second.InstanceDigest {
		return first, newOperationError("replay-operation", "compare-instance-evidence", "REPLAY_EVIDENCE_MISMATCH", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	first.Verifier = second.Verifier
	return first, nil
}

func materializeCollapse(workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action, subject sourcepolicy.SourceSubject, trace metaExecutionTrace, pass string) (operationMaterialization, *operationError) {
	temporary, err := copyWorkspace(workspace)
	if err != nil {
		return operationMaterialization{}, newOperationError("prepare-workspace", "materialize-disposable-workspace", "WORKSPACE_MATERIALIZATION_FAILED", "DIRECT_MISSING", "restore-workspace")
	}
	defer os.RemoveAll(temporary)
	snapshot, snapshotErr := readOnlyGitSnapshot(gitDir, plan.HeadSHA)
	if snapshotErr != nil {
		return operationMaterialization{}, newOperationError("prepare-workspace", "isolate-git-context", "GIT_SNAPSHOT_UNAVAILABLE", "DIRECT_MISSING", "restore-git-context")
	}
	defer os.RemoveAll(filepath.Dir(snapshot))
	environment := childEnvironment(snapshot, temporary)
	sourcePath := filepath.Join(temporary, filepath.FromSlash(subject.Path))
	beforeFiles, err := snapshotWorkspaceFiles(temporary)
	if err != nil {
		return operationMaterialization{}, newOperationError("evaluate-operation", "snapshot-collapse-input", "INPUT_EVIDENCE_UNAVAILABLE", "DIRECT_MISSING", "restore-operation-evidence")
	}
	before, err := os.ReadFile(sourcePath)
	if err != nil {
		return operationMaterialization{}, newOperationError("observe-plan", "read-collapse-source", "SOURCE_UNAVAILABLE", "DIRECT_MISSING", "restore-source")
	}
	beforeInspection, err := inspectCollapseSource(before, subject)
	if err != nil || !beforeInspection.AssignmentReturn || !beforeInspection.CommentsPreserved {
		return operationMaterialization{}, newOperationError("evaluate-operation", "validate-collapse-input", "INPUT_CANDIDATE_INVALID", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	contractDigest, ok := collapseContractDigest(action)
	if !ok {
		return operationMaterialization{}, newOperationError("evaluate-operation", "bind-collapse-contract", "CONTRACT_EVIDENCE_MALFORMED", "MALFORMED_EVIDENCE", "restore-operation-contract")
	}

	checkDescriptor := []string{"go", "run", "./scripts/refactor-metrics", "-root", "<workspace>", "-metrics", "<source-metrics>", "-sha", plan.HeadSHA, "-subject", action.Subject, "-check"}
	checkActual := []string{"go", "run", "./scripts/refactor-metrics", "-root", temporary, "-metrics", metricsPath, "-sha", plan.HeadSHA, "-subject", action.Subject, "-check"}
	preflight, preflightErr := runProcessObserved(temporary, environment, checkDescriptor, checkActual, &trace, pass, "evaluator")
	if preflightErr != nil || preflight.Observation.ExitCode != 0 || !collapseSummaryMatches(preflight.Stdout, false) {
		return operationMaterialization{Executor: preflight.Observation}, newOperationError("evaluate-operation", "preflight-collapse-candidate", "PREFLIGHT_CONFORMANCE_FAILED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	applyDescriptor := []string{"go", "run", "./scripts/refactor-metrics", "-root", "<workspace>", "-metrics", "<source-metrics>", "-sha", plan.HeadSHA, "-subject", action.Subject}
	applyActual := []string{"go", "run", "./scripts/refactor-metrics", "-root", temporary, "-metrics", metricsPath, "-sha", plan.HeadSHA, "-subject", action.Subject}
	executor, executorErr := runProcessObserved(temporary, environment, applyDescriptor, applyActual, &trace, pass, "executor")
	materialized := operationMaterialization{Executor: executor.Observation, Evaluator: preflight.Observation}
	if executorErr != nil || executor.Observation.ExitCode != 0 || !collapseSummaryMatches(executor.Stdout, true) {
		return materialized, newOperationError("execute-operation", "apply-collapse-candidate", "EXECUTOR_PROCESS_FAILED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	afterApplied, err := os.ReadFile(sourcePath)
	if err != nil {
		return materialized, newOperationError("evaluate-operation", "read-collapse-output", "OUTPUT_EVIDENCE_UNAVAILABLE", "DIRECT_MISSING", "restore-operation-evidence")
	}
	afterAppliedInspection, err := inspectCollapseSource(afterApplied, subject)
	if err != nil {
		return materialized, newOperationError("evaluate-operation", "validate-collapse-output", "OUTPUT_EVIDENCE_MALFORMED", "MALFORMED_EVIDENCE", "restore-operation-evidence")
	}
	if failure := validateCollapseOutput(beforeInspection, afterAppliedInspection, afterApplied); failure != nil {
		return materialized, failure
	}
	verifier := runGoTestObserved(temporary, environment, &trace, pass)
	materialized.Verifier = verifier.Observation
	after, err := os.ReadFile(sourcePath)
	if err != nil {
		return materialized, newOperationError("evaluate-operation", "read-collapse-final-source", "OUTPUT_EVIDENCE_UNAVAILABLE", "DIRECT_MISSING", "restore-operation-evidence")
	}
	afterInspection, err := inspectCollapseSource(after, subject)
	if err != nil {
		return materialized, newOperationError("evaluate-operation", "validate-collapse-final-output", "OUTPUT_EVIDENCE_MALFORMED", "MALFORMED_EVIDENCE", "restore-operation-evidence")
	}
	if !bytes.Equal(afterApplied, after) {
		return materialized, newOperationError("evaluate-operation", "compare-collapse-final-source", "WORKSPACE_EFFECT_OUT_OF_SCOPE", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	if failure := validateCollapseOutput(beforeInspection, afterInspection, after); failure != nil {
		return materialized, failure
	}
	changedFiles, err := changedWorkspaceFiles(beforeFiles, temporary, subject.Path)
	if err != nil {
		return materialized, newOperationError("evaluate-operation", "compare-collapse-workspace", "WORKSPACE_EFFECT_UNAVAILABLE", "DIRECT_MISSING", "restore-operation-evidence")
	}
	if len(changedFiles) != 1 || changedFiles[0] != subject.Path {
		return materialized, newOperationError("evaluate-operation", "compare-collapse-workspace", "WORKSPACE_EFFECT_OUT_OF_SCOPE", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	if verifier.Observation.ExitCode != 0 {
		return materialized, newOperationError("verify-operation", "go-test-transformed-workspace", "PROJECTED_COMPILE_OR_TEST_FAILED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	indicators, allPassed := collapseIndicatorReceipts(action, plan.HeadSHA, beforeInspection, afterInspection)
	if indicators == nil {
		return materialized, newOperationError("evaluate-operation", "bind-indicator-observations", "INSTANCE_INDICATOR_MISSING", "DIRECT_MISSING", "restore-operation-evidence")
	}
	if !allPassed {
		failure := newOperationError("evaluate-operation", "validate-collapse-indicators", "INDICATOR_OBSERVATION_FAILED", "KNOWN_CONTRADICTION", "report-counterexample")
		failure.evidence = collapseIndicatorFailureEvidence(indicators)
		return materialized, failure
	}

	canonicalValue := collapseInstanceEvidence{
		Operation:                   string(sourcepolicy.OperationCollapseAssign),
		Subject:                     subject.String(),
		Receiver:                    beforeInspection.Receiver,
		InputContractSourceDigest:   action.InputContractSourceDigest,
		InputContractSemanticDigest: action.InputContractSemanticDigest,
		BeforeSignature:             beforeInspection.Signature,
		AfterSignature:              afterInspection.Signature,
		BeforeCommentGroups:         beforeInspection.CommentGroups,
		AfterCommentGroups:          afterInspection.CommentGroups,
		BeforeOutsideDeclarations:   beforeInspection.OutsideDeclarations,
		AfterOutsideDeclarations:    afterInspection.OutsideDeclarations,
		BeforeSourceDigest:          digestBytes(before),
		AfterSourceDigest:           digestBytes(after),
		Before:                      before,
		After:                       after,
		ChangedFiles:                changedFiles,
		PreflightCount:              1,
		ApplyCount:                  1,
		Process:                     collapseReplayEvidence(executor.Observation, preflight.Observation, verifier.Observation),
	}
	canonical, _ := json.Marshal(canonicalValue)
	materialized.OperationID = string(sourcepolicy.OperationCollapseAssign)
	materialized.ContractDigest = contractDigest
	materialized.InstanceDigest = digestBytes(canonical)
	materialized.Indicators = indicators
	materialized.Canonical = canonical
	return materialized, nil
}

func collapseContractDigest(action generation.Action) (string, bool) {
	if !validBareDigest(action.InputContractSourceDigest) {
		return "", false
	}
	return "sha256:" + action.InputContractSourceDigest, true
}

func collapseIndicatorID(indicator sourcepolicy.Indicator) (string, bool) {
	payload, err := json.Marshal(indicator)
	if err != nil {
		return "", false
	}
	return digestBytes(payload), true
}

func validBareDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !('0' <= character && character <= '9') && !('a' <= character && character <= 'f') {
			return false
		}
	}
	return true
}

func collapseSummaryMatches(output []byte, write bool) bool {
	want := fmt.Sprintf("refactor-metrics: checked=1 operation=%s write=%t", sourcepolicy.OperationCollapseAssign, write)
	return strings.TrimSpace(string(output)) == want
}

func collapseSummaryOutput(write bool) []byte {
	return []byte(fmt.Sprintf("refactor-metrics: checked=1 operation=%s write=%t\n", sourcepolicy.OperationCollapseAssign, write))
}

func collapseReplayEvidence(executor, evaluator, verifier generation.ProcessObservation) operationReplayEvidence {
	return operationReplayEvidence{
		Executor:  collapseReplayProcess(executor, collapseSummaryOutput(true), nil),
		Evaluator: collapseReplayProcess(evaluator, collapseSummaryOutput(false), nil),
		Verifier:  collapseReplayProcess(verifier, nil, nil),
	}
}

func collapseReplayProcess(observation generation.ProcessObservation, stdout, stderr []byte) replayProcessObservation {
	projected := replayProcess(observation)
	projected.StdoutBytes = len(stdout)
	projected.StdoutDigest = digestBytes(stdout)
	projected.StderrBytes = len(stderr)
	projected.StderrDigest = digestBytes(stderr)
	return projected
}

func inspectCollapseSource(source []byte, subject sourcepolicy.SourceSubject) (collapseSourceInspection, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, subject.Path, source, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return collapseSourceInspection{}, err
	}
	var matches []*ast.FuncDecl
	ast.Inspect(file, func(node ast.Node) bool {
		function, ok := node.(*ast.FuncDecl)
		if !ok {
			return true
		}
		name, line, identityOK := linecaps.FunctionIdentity(fset, function)
		if identityOK && name == subject.Name && line == subject.Line {
			matches = append(matches, function)
		}
		return true
	})
	if len(matches) != 1 {
		return collapseSourceInspection{}, fmt.Errorf("subject %q matched %d functions", subject.String(), len(matches))
	}
	target := matches[0]
	if target.Body == nil {
		return collapseSourceInspection{}, fmt.Errorf("subject %q has no function body", subject.String())
	}
	receiver := ""
	if target.Recv != nil {
		receiver, err = renderCollapseReceiver(fset, target.Recv)
		if err != nil {
			return collapseSourceInspection{}, err
		}
	}
	signature, err := renderCollapseNode(fset, target.Type)
	if err != nil {
		return collapseSourceInspection{}, err
	}
	outsideDeclarations := make([]string, 0, len(file.Decls)-1)
	for _, declaration := range file.Decls {
		if declaration == target {
			continue
		}
		rendered, err := renderCollapseNode(fset, declaration)
		if err != nil {
			return collapseSourceInspection{}, err
		}
		outsideDeclarations = append(outsideDeclarations, rendered)
	}
	commentGroups := make([]string, 0, len(file.Comments))
	for _, group := range file.Comments {
		comments := make([]string, 0, len(group.List))
		for _, comment := range group.List {
			comments = append(comments, comment.Text)
		}
		commentGroups = append(commentGroups, strings.Join(comments, "\n"))
	}
	inspection := collapseSourceInspection{
		Subject:             subject.String(),
		Receiver:            receiver,
		Signature:           receiver + "|" + signature,
		OutsideDeclarations: outsideDeclarations,
		CommentGroups:       commentGroups,
		StartLine:           fset.Position(target.Pos()).Line,
		EndLine:             fset.Position(target.End()).Line,
	}
	if len(target.Body.List) == 1 {
		returnExpression, ok := target.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(returnExpression.Results) != 1 {
			return collapseSourceInspection{}, fmt.Errorf("subject %q is not a single-return function", subject.String())
		}
		expression, err := renderCollapseExpression(fset, returnExpression.Results[0])
		if err != nil {
			return collapseSourceInspection{}, err
		}
		inspection.SingleReturn = true
		inspection.ReturnExpression = expression
		return inspection, nil
	}
	if len(target.Body.List) != 2 {
		return collapseSourceInspection{}, fmt.Errorf("subject %q has %d body statements", subject.String(), len(target.Body.List))
	}
	assignment, ok := target.Body.List[0].(*ast.AssignStmt)
	if !ok || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		return collapseSourceInspection{}, fmt.Errorf("subject %q has no single assignment", subject.String())
	}
	lhs, ok := assignment.Lhs[0].(*ast.Ident)
	if !ok || lhs.Name == "_" {
		return collapseSourceInspection{}, fmt.Errorf("subject %q has no named assignment", subject.String())
	}
	returnStatement, ok := target.Body.List[1].(*ast.ReturnStmt)
	if !ok || len(returnStatement.Results) != 1 {
		return collapseSourceInspection{}, fmt.Errorf("subject %q has no single return", subject.String())
	}
	returned, ok := returnStatement.Results[0].(*ast.Ident)
	if !ok || returned.Name != lhs.Name {
		return collapseSourceInspection{}, fmt.Errorf("subject %q does not return its assignment", subject.String())
	}
	expression, err := renderCollapseExpression(fset, assignment.Rhs[0])
	if err != nil {
		return collapseSourceInspection{}, err
	}
	inspection.AssignmentReturn = true
	inspection.ReturnExpression = expression
	inspection.CommentsPreserved = true
	for _, comment := range file.Comments {
		if comment.Pos() >= assignment.Pos() && comment.Pos() <= returnStatement.End() {
			inspection.CommentsPreserved = false
			break
		}
	}
	return inspection, nil
}

func renderCollapseExpression(fset *token.FileSet, expression ast.Expr) (string, error) {
	return renderCollapseNode(fset, expression)
}

func renderCollapseNode(fset *token.FileSet, node ast.Node) (string, error) {
	var rendered bytes.Buffer
	if err := format.Node(&rendered, fset, node); err != nil {
		return "", err
	}
	return rendered.String(), nil
}

func renderCollapseReceiver(fset *token.FileSet, receiver *ast.FieldList) (string, error) {
	if len(receiver.List) != 1 {
		return "", fmt.Errorf("receiver has %d fields", len(receiver.List))
	}
	field := receiver.List[0]
	typeName, err := renderCollapseNode(fset, field.Type)
	if err != nil {
		return "", err
	}
	if len(field.Names) == 0 {
		return "(" + typeName + ")", nil
	}
	names := make([]string, 0, len(field.Names))
	for _, name := range field.Names {
		names = append(names, name.Name)
	}
	return "(" + strings.Join(names, ", ") + " " + typeName + ")", nil
}

func validateCollapseOutput(before, after collapseSourceInspection, source []byte) *operationError {
	if !after.SingleReturn || after.ReturnExpression != before.ReturnExpression || after.Receiver != before.Receiver || after.Signature != before.Signature ||
		!slices.Equal(after.OutsideDeclarations, before.OutsideDeclarations) || !slices.Equal(after.CommentGroups, before.CommentGroups) ||
		after.StartLine != before.StartLine || after.EndLine >= before.EndLine {
		return newOperationError("evaluate-operation", "validate-collapse-output", "OUTPUT_IDENTITY_MISMATCH", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	formatted, err := format.Source(source)
	if err != nil || !bytes.Equal(formatted, source) {
		return newOperationError("evaluate-operation", "validate-collapse-output", "FORMAT_FIXED_POINT_FAILED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	return nil
}

func snapshotWorkspaceFiles(root string) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if relative == ".git" || strings.HasPrefix(relative, ".git"+string(filepath.Separator)) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("workspace symlink is not allowed: %s", relative)
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(relative)] = digestBytes(data)
		return nil
	})
	return result, err
}

func changedWorkspaceFiles(before map[string]string, root, subject string) ([]string, error) {
	after, err := snapshotWorkspaceFiles(root)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(before)+len(after))
	for path := range before {
		seen[path] = struct{}{}
	}
	for path := range after {
		seen[path] = struct{}{}
	}
	changed := make([]string, 0)
	for path := range seen {
		if before[path] != after[path] {
			changed = append(changed, path)
		}
	}
	sort.Strings(changed)
	_ = subject
	return changed, nil
}

func collapseIndicatorReceipts(action generation.Action, headSHA string, before, after collapseSourceInspection) ([]generation.IndicatorReceipt, bool) {
	commentsPreserved := before.CommentsPreserved && slices.Equal(before.CommentGroups, after.CommentGroups)
	values := map[string]bool{
		"go.ast.single-match/v1":    before.AssignmentReturn,
		"go.comments.preserved/v1":  commentsPreserved,
		"go.format.fixed-point/v1":  after.SingleReturn && after.EndLine < before.EndLine,
	}
	result := make([]generation.IndicatorReceipt, 0, len(action.RequiredIndicatorIDs))
	allPassed := true
	transformed := before.Subject + "=>" + after.Subject
	for _, identifier := range action.RequiredIndicatorIDs {
		value, ok := values[identifier]
		if !ok {
			return nil, false
		}
		actual := 0
		verdict := generation.IndicatorVerdictFail
		if value {
			actual = 1
			verdict = generation.IndicatorVerdictPass
		} else {
			allPassed = false
		}
		result = append(result, makeIndicatorReceipt(identifier, action.Subject, headSHA, string(sourcepolicy.OperationCollapseAssign), actual, 1, before.EndLine-before.StartLine+1, after.EndLine-after.StartLine+1, transformed, verdict, action.ProofChoice))
	}
	return result, allPassed
}

func collapseIndicatorFailureEvidence(indicators []generation.IndicatorReceipt) []generation.ObservationFailureEvidence {
	result := make([]generation.ObservationFailureEvidence, 0, len(indicators))
	for _, indicator := range indicators {
		observed := 0
		if indicator.Observation != nil {
			observed = indicator.Observation.ActualValue
		}
		result = append(result, generation.ObservationFailureEvidence{
			IndicatorID: indicator.ID, Observed: observed, Expected: 1,
			Decision: string(indicator.Verdict), Counterexample: "collapse-indicator-observation",
		})
	}
	return result
}
