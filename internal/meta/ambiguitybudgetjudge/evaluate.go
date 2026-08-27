package ambiguitybudgetjudge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const canonicalLowering = "syntax.ParseFile->bidir.Lower"

func Evaluate(contractRaw, receiptRaw, source []byte) (Result, error) {
	var policy contract
	var report receipt
	if err := decode(contractRaw, &policy); err != nil {
		return Result{}, err
	}
	if err := decode(receiptRaw, &report); err != nil {
		return Result{}, err
	}
	result := Result{Schema: JudgeSchema, SubjectSHA: report.SubjectSHA, ContractID: report.ContractID,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ReportDigest: report.Digest,
		SourceDigest: report.Source.Digest, FixedDenominator: report.Summary.FixedDenominator,
		CasesTotal: report.Summary.CasesTotal, InterventionsTotal: report.Summary.InterventionsTotal,
		ForbiddenProducerImports: 0, AllowedProducerImports: 0, RepositoryWrites: report.Effects.RepositoryWrites,
		MutationAuthority: report.Effects.MutationAuthority}
	add := func(id string, passed bool, step string) {
		status := "PASS"
		if !passed {
			status = "FAIL"
		}
		result.Checks = append(result.Checks, check{ID: id, Status: status,
			Coordinate: coordinate{Stage: "ambiguity-budget-judge", Step: step, Reason: status}})
	}

	add("contract-identity", validContract(policy), "contract")
	observed, sourceErr := observeSource(policy.SourcePath, source)
	add("source-syntax-and-lowering", sourceErr == nil && observed.Lowering == canonicalLowering, "parse-lower")
	add("source-binding", sourceErr == nil && observed.Path == report.Source.Path && reflect.DeepEqual(observed, report.Source) &&
		observed.Package == policy.SourcePackage && observed.Namespace == policy.SourceNamespace, "source")
	budget, budgetOK := findBudget(observed, policy.BudgetActivity)
	add("gooo-computes-semantics", sourceErr == nil && budgetOK && budget.Counts == expectedBudget() && validCasePrograms(observed, policy), "computes")
	add("receipt-identity", report.Schema == ReceiptSchema && report.ContractID == policy.ID && report.Producer == Producer &&
		report.Consumer == Consumer && report.MetaOperation == MetaOperation && report.ProofChoice == "FOUNDATION" &&
		report.Budget == expectedBudget(), "identity")
	add("fixed-denominator", report.Summary.FixedDenominator == FixedDenominator && report.Summary.InterventionsTotal == InterventionTotal &&
		report.Summary.IntegerDimensions == IntegerDimensionTotal, "denominator")

	casePass := sourceErr == nil && budgetOK && len(report.Cases) == CaseTotal
	wantClaims := make([]transition, 0, CaseTotal)
	wantIndicators := make([]indicator, 0, CaseTotal*IntegerDimensionTotal)
	if casePass {
		for index, spec := range policy.Cases {
			program, ok := findCase(observed, spec.Activity)
			if !ok || program.ID != spec.ID {
				casePass = false
				continue
			}
			want := makeCaseReceipt(observed.Digest, program, budget.Counts)
			casePass = casePass && report.Cases[index] == want
			wantClaims = append(wantClaims, want.Claim)
			wantIndicators = append(wantIndicators, indicatorsFor(observed.Digest, program, budget.Counts)...)
		}
	}
	add("subject-case-vector", casePass, "cases")
	add("claim-transitions", casePass && reflect.DeepEqual(report.Claims, wantClaims), "claims")
	add("integer-set-observations", casePass && reflect.DeepEqual(report.Indicators, wantIndicators), "indicators")

	wantSubjectDecision, wantSubjectResolution, wantSubjectReason := subjectVector(report.Cases)
	add("subject-resolution-separation", report.SubjectDecision == wantSubjectDecision && report.SubjectResolution == wantSubjectResolution &&
		report.SubjectReason == wantSubjectReason && report.SubjectResolution == "LOWER_RESOLUTION" &&
		report.SubjectCoordinate == (coordinate{Stage: "ambiguity-budget", Step: "subject-resolution", Reason: wantSubjectReason}), "subject")

	wantSummary := summary{CasesTotal: CaseTotal, KnownCases: 3, ZeroAmbiguityCases: 1, BoundaryCases: 1,
		OverBudgetCases: 1, UnknownCases: 1, LowerResolutionCases: 2, OpenClaims: 1,
		IntegerDimensions: IntegerDimensionTotal, InterventionsTotal: InterventionTotal, FixedDenominator: FixedDenominator}
	add("case-cardinalities", report.Summary == wantSummary, "summary")

	wantInterventions := buildInterventions(policy.Interventions, source, observed, budget.Counts)
	add("intervention-separation", sourceErr == nil && reflect.DeepEqual(report.Interventions, wantInterventions) && allInterventionsSatisfied(wantInterventions), "interventions")
	add("dependency-guard", result.ForbiddenProducerImports == 0 && result.AllowedProducerImports == 0, "dependency")
	add("effect-guard", report.Effects == (effects{}) && result.RepositoryWrites == 0 && !result.MutationAuthority, "effects")
	add("proof-bindings", proofsPass(report.Proofs), "proofs")
	add("receipt-digest", report.Digest == digestReport(report), "digest")

	result.ConformanceDecision, result.ConformanceResolution, result.ConformanceReason = "PASS", "EXACT", "CONFORMANCE_CASES_MATCHED"
	if !allChecksPass(result.Checks) {
		result.ConformanceDecision, result.ConformanceResolution, result.ConformanceReason = "FAIL_CLOSED", "LOWER_RESOLUTION", "CONFORMANCE_RECEIPT_REJECTED"
	}
	result.SubjectDecision, result.SubjectResolution = report.SubjectDecision, report.SubjectResolution
	result.Digest = resultDigest(result)
	return result, nil
}

func validContract(policy contract) bool {
	if policy.Schema != ContractSchema || policy.ID == "" || policy.SourcePath == "" || policy.SourcePackage == "" ||
		policy.SourceNamespace == "" || policy.BudgetActivity == "" || policy.FixedDenominator != FixedDenominator ||
		len(policy.Cases) != CaseTotal || len(policy.Interventions) != InterventionTotal || len(policy.NotClaimed) != 4 {
		return false
	}
	ids, activities := map[string]bool{}, map[string]bool{}
	for _, item := range policy.Cases {
		if item.ID == "" || item.Activity == "" || ids[item.ID] || activities[item.Activity] {
			return false
		}
		ids[item.ID], activities[item.Activity] = true, true
	}
	interventions := map[string]bool{}
	for _, item := range policy.Interventions {
		if item.ID == "" || item.TargetActivity == "" || interventions[item.ID] ||
			(item.Kind != "SEMANTIC" && item.Kind != "NONSEMANTIC") {
			return false
		}
		interventions[item.ID] = true
	}
	return true
}

func expectedBudget() integerSet {
	return integerSet{InterpretationCandidates: 2, UnresolvedBranches: 1, EvidencePaths: 2}
}

func validUnobserved(unobserved []string) bool {
	seen := make(map[string]bool, len(unobserved))
	for _, dimension := range unobserved {
		if !contains(integerDimensions[:], dimension) || seen[dimension] {
			return false
		}
		seen[dimension] = true
	}
	return true
}

func derivedClass(program computesProgram, budget integerSet) string {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN"
	}
	if program.Counts == (integerSet{InterpretationCandidates: 1, UnresolvedBranches: 0, EvidencePaths: 1}) {
		return "ZERO"
	}
	if program.Counts == budget {
		return "BOUNDARY"
	}
	if exceeds(program.Counts, budget) {
		return "OVER"
	}
	return "WITHIN"
}

func inputState(program computesProgram) string {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN"
	}
	return "KNOWN"
}

func claimTarget(decision string) string {
	switch decision {
	case "PASS":
		return "DISCHARGED"
	case "FAIL_CLOSED":
		return "REFUTED"
	default:
		return "OPEN"
	}
}

type computesProgram struct {
	Activity             string
	Text                 string
	Kind                 string
	ID                   string
	Counts               integerSet
	UnobservedDimensions []string
}

func parseComputesProgram(activity, text string) (computesProgram, error) {
	if activity == "" || text == "" || text != strings.TrimSpace(text) || strings.ContainsAny(text, " \t\r\n") {
		return computesProgram{}, fmt.Errorf("non-canonical computes program")
	}
	parts := strings.Split(text, ":")
	if len(parts) == 3 && parts[0] == "ambiguity-budget" && parts[1] == "budget" {
		counts, unobserved, err := parseIntegerSet(parts[2])
		if err != nil {
			return computesProgram{}, err
		}
		if len(unobserved) != 0 {
			return computesProgram{}, fmt.Errorf("budget coordinates must be observed")
		}
		return computesProgram{Activity: activity, Text: text, Kind: "BUDGET", ID: "budget", Counts: counts}, nil
	}
	if len(parts) != 4 || parts[0] != "ambiguity-budget" || parts[1] != "case" || parts[2] == "" {
		return computesProgram{}, fmt.Errorf("unsupported computes program")
	}
	counts, unobserved, err := parseIntegerSet(parts[3])
	if err != nil {
		return computesProgram{}, err
	}
	return computesProgram{Activity: activity, Text: text, Kind: "CASE", ID: parts[2], Counts: counts, UnobservedDimensions: unobserved}, nil
}

func parseIntegerSet(text string) (integerSet, []string, error) {
	parts := strings.Split(text, ",")
	if len(parts) != IntegerDimensionTotal {
		return integerSet{}, nil, fmt.Errorf("integer set coordinate count mismatch")
	}
	values := [IntegerDimensionTotal]int{}
	unobserved := make([]string, 0, IntegerDimensionTotal)
	for index, part := range parts {
		if part == "" || part != strings.TrimSpace(part) {
			return integerSet{}, nil, fmt.Errorf("non-canonical integer coordinate")
		}
		if part == "?" {
			unobserved = append(unobserved, integerDimensions[index])
			continue
		}
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return integerSet{}, nil, fmt.Errorf("invalid integer coordinate")
		}
		values[index] = value
	}
	return integerSet{InterpretationCandidates: values[0], UnresolvedBranches: values[1], EvidencePaths: values[2]}, unobserved, nil
}

var integerDimensions = [...]string{"interpretation_candidates", "unresolved_branches", "evidence_paths"}

func observeSource(path string, raw []byte) (sourceObservation, error) {
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return sourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, fmt.Errorf("source syntax is unknown")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, err
	}
	if err := ir.Validate(); err != nil {
		return sourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, err
	}
	observation := sourceObservation{Path: path, Digest: digestBytes(raw), SemanticDigest: "sha256:" + ir.StableHash(), Lowering: canonicalLowering,
		Package: ir.Package, Namespace: ir.Namespace.String()}
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			observation.Entities++
		case semantic.Activity:
			observation.Activities++
			program, err := parseComputesProgram(node.Name, node.ValueProgram)
			if err != nil {
				return observation, err
			}
			if program.Kind == "CASE" {
				program.Class = derivedClass(program, expectedBudget())
				program.InputState = inputState(program)
			}
			observation.Programs = append(observation.Programs, programObservation{Activity: program.Activity, Program: program.Text,
				ProgramKind: program.Kind, ID: program.ID, Class: program.Class, InputState: program.InputState, Counts: program.Counts,
				UnobservedDimensions: append([]string(nil), program.UnobservedDimensions...),
				Digest:               digestBytes([]byte(program.Text))})
		}
	}
	sort.Slice(observation.Programs, func(i, j int) bool { return observation.Programs[i].Activity < observation.Programs[j].Activity })
	return observation, nil
}

func findBudget(source sourceObservation, activity string) (programObservation, bool) {
	for _, program := range source.Programs {
		if program.ProgramKind == "BUDGET" && program.Activity == activity {
			return program, true
		}
	}
	return programObservation{}, false
}

func findCase(source sourceObservation, activity string) (programObservation, bool) {
	for _, program := range source.Programs {
		if program.ProgramKind == "CASE" && program.Activity == activity {
			return program, true
		}
	}
	return programObservation{}, false
}

func validCasePrograms(source sourceObservation, policy contract) bool {
	if len(source.Programs) != CaseTotal+1 || source.Activities != CaseTotal+1 {
		return false
	}
	classes := map[string]bool{}
	for _, spec := range policy.Cases {
		program, ok := findCase(source, spec.Activity)
		if !ok || program.ID != spec.ID || !validUnobserved(program.UnobservedDimensions) || classes[program.Class] {
			return false
		}
		classes[program.Class] = true
	}
	return len(classes) == CaseTotal && classes["ZERO"] && classes["BOUNDARY"] && classes["OVER"] && classes["UNKNOWN"]
}

func makeCaseReceipt(sourceDigest string, program programObservation, budget integerSet) caseReceipt {
	parsed := computesProgram{Activity: program.Activity, Text: program.Program, Kind: program.ProgramKind,
		ID: program.ID, Counts: program.Counts, UnobservedDimensions: append([]string(nil), program.UnobservedDimensions...)}
	decision, resolution, reason := subjectDecision(parsed, budget)
	evidence := digestValue(struct {
		SourceDigest         string
		Activity             string
		Program              string
		Counts               integerSet
		UnobservedDimensions []string
	}{sourceDigest, program.Activity, program.Program, program.Counts, program.UnobservedDimensions})
	coord := caseCoordinate(parsed, reason)
	claim := transition{CaseID: program.ID, From: "OPEN", To: claimTarget(decision), Stage: coord.Stage, Step: coord.Step,
		Reason: reason, EvidenceDigest: evidence}
	return caseReceipt{ID: program.ID, Activity: program.Activity, Class: program.Class, InputState: program.InputState,
		Program: program.Program, ProgramDigest: program.Digest, Counts: program.Counts,
		UnobservedDimensions: append([]string(nil), program.UnobservedDimensions...), Decision: decision, Resolution: resolution,
		Reason: reason, Coordinate: coord, Claim: claim, EvidenceDigest: evidence, Conformance: "MATCH"}
}

func caseCoordinate(program computesProgram, reason string) coordinate {
	if len(program.UnobservedDimensions) > 0 {
		return coordinate{Stage: "AMBIGUITY_OBSERVATION", Step: program.UnobservedDimensions[0], Reason: "AMBIGUITY_COORDINATE_UNOBSERVED"}
	}
	return coordinate{Stage: "AMBIGUITY_BUDGET", Step: "case:" + program.ID, Reason: reason}
}

func exceeds(value integerSet, budget integerSet) bool {
	return value.InterpretationCandidates > budget.InterpretationCandidates ||
		value.UnresolvedBranches > budget.UnresolvedBranches || value.EvidencePaths > budget.EvidencePaths
}

func subjectDecision(program computesProgram, budget integerSet) (string, string, string) {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COORDINATE_UNOBSERVED"
	}
	if program.Counts.InterpretationCandidates < 1 || program.Counts.EvidencePaths < 1 {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COUNT_UNKNOWN"
	}
	if exceeds(program.Counts, budget) {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED"
	}
	return "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"
}

func indicatorsFor(sourceDigest string, program programObservation, budget integerSet) []indicator {
	values := []struct {
		metric, dimension, proof string
		observed, limit          int
	}{
		{"gooo.metric.ambiguity-budget.candidate-count.v2", "interpretation_candidates", "FOUNDATION", program.Counts.InterpretationCandidates, budget.InterpretationCandidates},
		{"gooo.metric.ambiguity-budget.unresolved-branches.v2", "unresolved_branches", "COHERENCE", program.Counts.UnresolvedBranches, budget.UnresolvedBranches},
		{"gooo.metric.ambiguity-budget.evidence-paths.v2", "evidence_paths", "REGRESSION", program.Counts.EvidencePaths, budget.EvidencePaths},
	}
	result := make([]indicator, 0, len(values))
	for _, value := range values {
		coordinateObserved := !contains(program.UnobservedDimensions, value.dimension)
		evaluation := "UNOBSERVED"
		if coordinateObserved {
			evaluation = "WITHIN_LIMIT"
			if value.observed > value.limit {
				evaluation = "EXCEEDS_LIMIT"
			}
		}
		evidence := digestValue(struct {
			SourceDigest       string
			Activity           string
			Dimension          string
			Observed           int
			CoordinateObserved bool
			Budget             int
		}{sourceDigest, program.Activity, value.dimension, value.observed, coordinateObserved, value.limit})
		result = append(result, indicator{MetricID: value.metric, CaseID: program.ID, Dimension: value.dimension, ProofChoice: value.proof,
			Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, Observed: value.observed, Budget: value.limit,
			CoordinateObserved: coordinateObserved, Relation: "<=", Evaluation: evaluation, EvidenceDigest: evidence})
	}
	return result
}

func subjectVector(cases []caseReceipt) (string, string, string) {
	for _, item := range cases {
		if item.Resolution == "LOWER_RESOLUTION" {
			return "MIXED", "LOWER_RESOLUTION", "AMBIGUITY_CASE_VECTOR_CONTAINS_LOWER_RESOLUTION"
		}
	}
	return "EXACT", "EXACT", "AMBIGUITY_CASE_VECTOR_EXACT"
}

func buildInterventions(specs []interventionContract, raw []byte, base sourceObservation, budget integerSet) []interventionReceipt {
	result := make([]interventionReceipt, 0, len(specs))
	for _, spec := range specs {
		mutated, err := interventionSource(raw, spec)
		if err != nil {
			result = append(result, interventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity})
			continue
		}
		next, err := observeSource(base.Path, mutated)
		beforeProgram, beforeOK := findCase(base, spec.TargetActivity)
		afterProgram, afterOK := findCase(next, spec.TargetActivity)
		if err != nil || !beforeOK || !afterOK {
			result = append(result, interventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity})
			continue
		}
		before := makeCaseReceipt(base.Digest, beforeProgram, budget)
		after := makeCaseReceipt(next.Digest, afterProgram, budget)
		satisfied := interventionSatisfied(spec.Kind, before, after, base, next)
		evidence := digestValue(struct {
			ID     string
			Before caseReceipt
			After  caseReceipt
			Base   sourceObservation
			Next   sourceObservation
		}{spec.ID, before, after, base, next})
		result = append(result, interventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity,
			SourceDigestBefore: base.Digest, SourceDigestAfter: next.Digest, SemanticDigestBefore: base.SemanticDigest,
			SemanticDigestAfter: next.SemanticDigest, CountsBefore: before.Counts, CountsAfter: after.Counts,
			UnobservedBefore: append([]string(nil), before.UnobservedDimensions...), UnobservedAfter: append([]string(nil), after.UnobservedDimensions...),
			ClassBefore: before.Class, ClassAfter: after.Class, InputStateBefore: before.InputState, InputStateAfter: after.InputState,
			ClaimBefore: before.Claim, ClaimAfter: after.Claim,
			DecisionBefore: before.Decision, ResolutionBefore: before.Resolution, ReasonBefore: before.Reason,
			DecisionAfter: after.Decision, ResolutionAfter: after.Resolution, ReasonAfter: after.Reason,
			Satisfied: satisfied, EvidenceDigest: evidence})
	}
	return result
}

func interventionSource(raw []byte, spec interventionContract) ([]byte, error) {
	source := string(raw)
	if spec.Kind == "NONSEMANTIC" {
		needle := "activity " + spec.TargetActivity + "("
		if !strings.Contains(source, needle) {
			return nil, fmt.Errorf("target activity not found")
		}
		return []byte(strings.Replace(source, needle, "// ambiguity-budget: comment-only intervention\n"+needle, 1)), nil
	}
	return mutateActivityProgram(raw, spec.TargetActivity, func(program computesProgram) (string, error) {
		if program.Kind != "CASE" {
			return "", fmt.Errorf("target is not a case")
		}
		program.Counts.InterpretationCandidates++
		return "ambiguity-budget:case:" + program.ID + ":" + formatIntegerSet(program.Counts, program.UnobservedDimensions), nil
	})
}

func mutateActivityProgram(raw []byte, target string, transform func(computesProgram) (string, error)) ([]byte, error) {
	source := string(raw)
	lines := strings.SplitAfter(source, "\n")
	for index, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 || fields[0] != "activity" {
			continue
		}
		name, _, _ := strings.Cut(fields[1], "(")
		if name != target {
			continue
		}
		marker := ` computes "`
		start := strings.Index(line, marker)
		if start < 0 {
			return nil, fmt.Errorf("target computes program not found")
		}
		start += len(marker)
		end := strings.LastIndex(line, `"`)
		if end <= start {
			return nil, fmt.Errorf("target computes program malformed")
		}
		program, err := parseComputesProgram(target, line[start:end])
		if err != nil {
			return nil, err
		}
		replacement, err := transform(program)
		if err != nil {
			return nil, err
		}
		lines[index] = line[:start] + replacement + line[end:]
		return []byte(strings.Join(lines, "")), nil
	}
	return nil, fmt.Errorf("target activity not found")
}

func formatIntegerSet(value integerSet, unobserved []string) string {
	values := []string{strconv.Itoa(value.InterpretationCandidates), strconv.Itoa(value.UnresolvedBranches), strconv.Itoa(value.EvidencePaths)}
	for index, dimension := range integerDimensions {
		if contains(unobserved, dimension) {
			values[index] = "?"
		}
	}
	return strings.Join(values, ",")
}

func interventionSatisfied(kind string, before, after caseReceipt, base, next sourceObservation) bool {
	if kind == "SEMANTIC" {
		return base.Digest != next.Digest && base.SemanticDigest != next.SemanticDigest && before.Counts != after.Counts && before.Decision == "PASS" &&
			before.Resolution == "EXACT" && before.Claim.From == "OPEN" && before.Claim.To == "DISCHARGED" &&
			after.Decision == "FAIL_CLOSED" && after.Resolution == "LOWER_RESOLUTION" && after.Reason == "AMBIGUITY_BUDGET_EXCEEDED" &&
			after.Claim.From == "OPEN" && after.Claim.To == "REFUTED"
	}
	return base.Digest != next.Digest && base.SemanticDigest == next.SemanticDigest && before.Counts == after.Counts && before.Decision == after.Decision &&
		before.Class == after.Class && before.InputState == after.InputState && before.Resolution == after.Resolution && before.Reason == after.Reason &&
		reflect.DeepEqual(before.UnobservedDimensions, after.UnobservedDimensions) && before.Claim == after.Claim
}

func allInterventionsSatisfied(values []interventionReceipt) bool {
	if len(values) != InterventionTotal {
		return false
	}
	for _, value := range values {
		if !value.Satisfied {
			return false
		}
	}
	return true
}

func proofsPass(values []proof) bool {
	if len(values) != 3 {
		return false
	}
	want := []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
	for index, value := range values {
		if value.Choice != want[index] || value.Producer != Producer || value.Consumer != Consumer || value.MetaOperation == "" ||
			!value.Passed || !validDigest(value.EvidenceDigest) {
			return false
		}
	}
	return true
}

func allChecksPass(values []check) bool {
	for _, value := range values {
		if value.Status != "PASS" {
			return false
		}
	}
	return true
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func decode(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return err
	}
	return nil
}

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(raw)
}

func digestReport(report receipt) string {
	report.Digest = ""
	return digestValue(report)
}

func resultDigest(result Result) string {
	result.Digest = ""
	return digestValue(result)
}

func validDigest(value string) bool {
	return len(value) == len("sha256:")+64 && strings.HasPrefix(value, "sha256:") && strings.Trim(value[len("sha256:"):], "0123456789abcdef") == ""
}
