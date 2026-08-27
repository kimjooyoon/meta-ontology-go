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
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const canonicalLowering = "syntax.ParseFile->bidir.Lower"

func Evaluate(contractRaw, receiptRaw, source, effectsRaw []byte) (Result, error) {
	var policy contract
	var report receipt
	if err := decode(contractRaw, &policy); err != nil {
		return Result{}, err
	}
	if err := decode(receiptRaw, &report); err != nil {
		return Result{}, err
	}
	observedEffects, effectsErr := decodeEffects(effectsRaw)
	result := Result{Schema: JudgeSchema, SubjectSHA: report.SubjectSHA, ContractID: report.ContractID,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ReportDigest: report.Digest,
		SourceDigest: report.Source.Digest, EffectsArtifactDigest: report.Effects.ArtifactDigest,
		Denominator: report.Summary.Denominator, Numerator: report.Summary.Numerator,
		ForbiddenProducerImports: 0, AllowedProducerImports: 0, RepositoryWrites: report.Effects.RepositoryWrites,
		MutationAuthority: report.Effects.MutationAuthority, MutationAuthorityResolution: report.Effects.MutationAuthorityResolution}
	add := func(id string, passed bool, step string) {
		status := "PASS"
		if !passed {
			status = "FAIL"
		}
		result.Checks = append(result.Checks, check{ID: id, Status: status,
			Coordinate: coordinate{Stage: "ambiguity-budget-judge", Step: step, Reason: status}})
	}

	add("contract-identity", validContract(policy), "contract")
	observed, sourceErr := observeSource(policy.SourcePath, source, policy.BudgetPolicy)
	add("source-syntax-and-lowering", sourceErr == nil && observed.Lowering == canonicalLowering, "parse-lower")
	add("source-binding", sourceErr == nil && observed.Path == report.Source.Path && reflect.DeepEqual(observed, report.Source) &&
		observed.Package == policy.SourcePackage && observed.Namespace == policy.SourceNamespace, "source")
	budget, budgetOK := findBudget(observed, policy.BudgetActivity)
	add("gooo-graph-cardinalities", sourceErr == nil && budgetOK && budget.Program == budgetBinding(policy.BudgetPolicy) &&
		budget.ID == policy.BudgetPolicy.Version && validCasePrograms(observed, policy), "graph")
	add("receipt-identity", report.Schema == ReceiptSchema && report.ContractID == policy.ID && report.Producer == Producer &&
		report.Consumer == Consumer && report.MetaOperation == MetaOperation && report.ProofChoice == "FOUNDATION" &&
		reflect.DeepEqual(report.BudgetPolicy, policy.BudgetPolicy) && report.BudgetBinding == budgetBinding(policy.BudgetPolicy) &&
		report.BudgetAuthority == "CONTRACT_POLICY", "identity")
	add("versioned-denominator", report.Summary.Denominator == policy.Denominator && validDenominator(policy.Denominator) &&
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
			want := makeCaseReceipt(observed.Digest, observed.SemanticDigest, program, policy.BudgetPolicy)
			casePass = casePass && reflect.DeepEqual(report.Cases[index], want)
			wantClaims = append(wantClaims, want.Claim)
			wantIndicators = append(wantIndicators, indicatorsFor(observed.SemanticDigest, program, policy.BudgetPolicy)...)
		}
	}
	add("subject-case-vector", casePass, "cases")
	add("claim-propositions-and-transitions", casePass && reflect.DeepEqual(report.Claims, wantClaims), "claims")
	add("integer-set-observations", casePass && reflect.DeepEqual(report.Indicators, wantIndicators), "indicators")

	wantSubjectDecision, wantSubjectResolution, wantSubjectReason := subjectVector(report.Cases)
	add("subject-resolution-separation", report.SubjectDecision == wantSubjectDecision && report.SubjectResolution == wantSubjectResolution &&
		report.SubjectReason == wantSubjectReason && report.SubjectResolution == "LOWER_RESOLUTION" &&
		report.SubjectCoordinate == (coordinate{Stage: "ambiguity-budget", Step: "subject-resolution", Reason: wantSubjectReason}), "subject")

	wantSummary := summarize(report.Cases, report.Interventions, policy.Denominator)
	add("case-cardinalities-and-numerators", reflect.DeepEqual(report.Summary, wantSummary), "summary")

	wantInterventions := buildInterventions(policy.Interventions, source, observed, policy.BudgetPolicy)
	add("intervention-separation", sourceErr == nil && reflect.DeepEqual(report.Interventions, wantInterventions) && allInterventionsSatisfied(wantInterventions), "interventions")
	add("dependency-guard", result.ForbiddenProducerImports == 0 && result.AllowedProducerImports == 0, "dependency")
	add("effect-artifact-binding", effectsErr == nil && reflect.DeepEqual(report.Effects, observedEffects) &&
		report.Effects.RepositoryWrites == 0 && report.Effects.WriteSetEqual && report.Effects.MutationAuthority == "UNKNOWN" &&
		report.Effects.MutationAuthorityResolution == "NOT_OBSERVED", "effects")
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
		policy.SourceNamespace == "" || policy.BudgetActivity == "" || !validPolicy(policy.BudgetPolicy) || !validDenominator(policy.Denominator) ||
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
		if item.ID == "" || item.TargetActivity == "" || interventions[item.ID] || (item.Kind != "SEMANTIC" && item.Kind != "NONSEMANTIC") {
			return false
		}
		interventions[item.ID] = true
	}
	return true
}

func validPolicy(policy budgetPolicy) bool {
	if policy.Schema != PolicySchema || policy.ID == "" || policy.Version == "" || policy.Authority != "CONTRACT_POLICY" || len(policy.Dimensions) != IntegerDimensionTotal {
		return false
	}
	seen := map[string]bool{}
	for _, dimension := range policy.Dimensions {
		if !contains(integerDimensions[:], dimension.ID) || seen[dimension.ID] || dimension.Limit < 0 {
			return false
		}
		seen[dimension.ID] = true
	}
	for _, dimension := range integerDimensions {
		if !seen[dimension] {
			return false
		}
	}
	return true
}

func validDenominator(value denominator) bool {
	return value.Schema == DenominatorSchema && value.Version != "" && value.Cases == CaseTotal &&
		value.IntegerObservations == CaseTotal*IntegerDimensionTotal && value.Claims == CaseTotal &&
		value.Interventions == InterventionTotal && value.AuthorityObservations == 1
}

func expectedMinimum() integerSet {
	return integerSet{InterpretationCandidates: 1, EvidencePaths: 1}
}

func policyCounts(policy budgetPolicy) integerSet {
	var counts integerSet
	for _, dimension := range policy.Dimensions {
		switch dimension.ID {
		case "interpretation_candidates":
			counts.InterpretationCandidates = dimension.Limit
		case "unresolved_branches":
			counts.UnresolvedBranches = dimension.Limit
		case "evidence_paths":
			counts.EvidencePaths = dimension.Limit
		}
	}
	return counts
}

func budgetBinding(policy budgetPolicy) string {
	return "ambiguity-budget:budget-policy:" + policy.Version
}

func policyLimit(policy budgetPolicy, id string) int {
	for _, dimension := range policy.Dimensions {
		if dimension.ID == id {
			return dimension.Limit
		}
	}
	return 0
}

func validUnobserved(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if !contains(integerDimensions[:], value) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func derivedClass(program programObservation, budget integerSet) string {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN"
	}
	if program.Counts == expectedMinimum() {
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

func inputState(program programObservation) string {
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
	Activity string
	Text     string
	Kind     string
	ID       string
}

func parseComputesProgram(activity, text string) (computesProgram, error) {
	if activity == "" || text == "" || text != strings.TrimSpace(text) || strings.ContainsAny(text, " \t\r\n") {
		return computesProgram{}, fmt.Errorf("non-canonical computes program")
	}
	parts := strings.Split(text, ":")
	if len(parts) == 3 && parts[0] == "ambiguity-budget" && parts[1] == "budget-policy" && parts[2] != "" {
		return computesProgram{Activity: activity, Text: text, Kind: "BUDGET", ID: parts[2]}, nil
	}
	if len(parts) != 3 || parts[0] != "ambiguity-budget" || parts[1] != "case" || parts[2] == "" {
		return computesProgram{}, fmt.Errorf("unsupported computes program")
	}
	return computesProgram{Activity: activity, Text: text, Kind: "CASE", ID: parts[2]}, nil
}

func observeSource(path string, raw []byte, policy budgetPolicy) (sourceObservation, error) {
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
	index := indexGraph(ir)
	observation := sourceObservation{Path: path, Digest: digestBytes(raw), SemanticDigest: "sha256:" + ir.StableHash(), Lowering: canonicalLowering,
		Package: ir.Package, Namespace: ir.Namespace.String()}
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			observation.Entities++
		case semantic.Activity:
			observation.Activities++
			if node.ValueProgram == "" {
				continue
			}
			program, err := parseComputesProgram(node.Name, node.ValueProgram)
			if err != nil {
				return observation, err
			}
			if program.Kind == "CASE" {
				result, err := observeCaseProgram(index, node, program, policy)
				if err != nil {
					return observation, err
				}
				observation.Programs = append(observation.Programs, result)
			} else {
				observation.Programs = append(observation.Programs, programObservation{Activity: program.Activity, Program: program.Text, ProgramKind: program.Kind,
					ID: program.ID, Digest: digestBytes([]byte(program.Text)), ActivitySemanticDigest: activitySemanticDigest(index, node)})
			}
		}
	}
	sort.Slice(observation.Programs, func(i, j int) bool { return observation.Programs[i].Activity < observation.Programs[j].Activity })
	return observation, nil
}

type graphIndex struct {
	nodes       map[string]semantic.Node
	used        map[string][]string
	generatedBy map[string][]string
}

func indexGraph(ir semantic.IR) graphIndex {
	index := graphIndex{nodes: map[string]semantic.Node{}, used: map[string][]string{}, generatedBy: map[string][]string{}}
	for _, node := range ir.Graph.Nodes() {
		index.nodes[node.ID.String()] = node
	}
	for _, fact := range ir.Graph.AllFacts() {
		subject, object := fact.Subject.String(), fact.Object.String()
		switch fact.Predicate {
		case semantic.Used:
			index.used[subject] = append(index.used[subject], object)
		case semantic.WasGeneratedBy:
			index.generatedBy[subject] = append(index.generatedBy[subject], object)
		}
	}
	for activity := range index.used {
		index.used[activity] = uniqueStrings(index.used[activity])
	}
	for entity := range index.generatedBy {
		index.generatedBy[entity] = uniqueStrings(index.generatedBy[entity])
	}
	return index
}

func observeCaseProgram(index graphIndex, activityNode semantic.Node, program computesProgram, policy budgetPolicy) (programObservation, error) {
	anchor := ""
	for _, candidate := range index.used[activityNode.ID.String()] {
		if node, ok := index.nodes[candidate]; ok && node.Kind == semantic.Entity && strings.Contains(candidate, "/case/"+program.ID) {
			if anchor != "" {
				return programObservation{}, fmt.Errorf("activity %q has multiple case anchors", activityNode.Name)
			}
			anchor = candidate
		}
	}
	if anchor == "" {
		return programObservation{}, fmt.Errorf("activity %q has no case anchor", activityNode.Name)
	}
	elements := ambiguityElements{}
	for entityID, node := range index.nodes {
		if node.Kind != semantic.Entity || !generatedForCase(index, entityID, anchor) {
			continue
		}
		if strings.Contains(entityID, "/candidate/") {
			elements.CandidateIDs = append(elements.CandidateIDs, entityID)
		}
		if strings.Contains(entityID, "/branch/resolved/") {
			elements.ResolvedBranchIDs = append(elements.ResolvedBranchIDs, entityID)
		}
		if strings.Contains(entityID, "/branch/unresolved/") {
			elements.UnresolvedBranchIDs = append(elements.UnresolvedBranchIDs, entityID)
		}
		if strings.Contains(entityID, "/branch-observation/") {
			elements.BranchObservationIDs = append(elements.BranchObservationIDs, entityID)
		}
		if strings.Contains(entityID, "/evidence-path/") && pathUsesCandidate(index, entityID, anchor) {
			elements.EvidencePathIDs = append(elements.EvidencePathIDs, entityID)
		}
	}
	normalizeElements(&elements)
	counts := elementCounts(elements)
	activityDigest := activitySemanticDigest(index, activityNode)
	elementDigest := digestValue(elements)
	unobserved, gaps := observationGaps(program.ID, anchor, activityDigest, elements)
	result := programObservation{Activity: program.Activity, Program: program.Text, ProgramKind: program.Kind, ID: program.ID, Counts: counts,
		UnobservedDimensions: unobserved, ObservationGaps: gaps, Elements: elements, ElementDigest: elementDigest,
		ActivitySemanticDigest: activityDigest, SemanticDigest: digestValue(struct {
			Activity string
			Program  string
			Elements ambiguityElements
			Counts   integerSet
		}{program.Activity, program.Text, elements, counts}), Digest: digestBytes([]byte(program.Text))}
	result.Class = derivedClass(result, policyCounts(policy))
	result.InputState = inputState(result)
	return result, nil
}

func generatedForCase(index graphIndex, entityID, anchor string) bool {
	for _, activityID := range index.generatedBy[entityID] {
		if contains(index.used[activityID], anchor) {
			return true
		}
	}
	return false
}

func pathUsesCandidate(index graphIndex, entityID, anchor string) bool {
	for _, activityID := range index.generatedBy[entityID] {
		used := index.used[activityID]
		if !contains(used, anchor) {
			continue
		}
		for _, entityID := range used {
			if entityID != anchor && strings.Contains(entityID, "/candidate/") {
				return true
			}
		}
	}
	return false
}

func activitySemanticDigest(index graphIndex, node semantic.Node) string {
	generated := make([]string, 0)
	for entityID, activities := range index.generatedBy {
		if contains(activities, node.ID.String()) {
			generated = append(generated, entityID)
		}
	}
	used := append([]string(nil), index.used[node.ID.String()]...)
	sort.Strings(used)
	sort.Strings(generated)
	return digestValue(struct {
		ID        string
		Used      []string
		Generated []string
	}{node.ID.String(), used, generated})
}

func observationGaps(caseID, anchor, activityDigest string, elements ambiguityElements) ([]string, []observationGap) {
	missing := make([]string, 0, IntegerDimensionTotal)
	gaps := make([]observationGap, 0, IntegerDimensionTotal)
	if len(elements.CandidateIDs) == 0 {
		missing = append(missing, "interpretation_candidates")
	}
	if len(elements.BranchObservationIDs) == 0 {
		missing = append(missing, "unresolved_branches")
	}
	if len(elements.EvidencePathIDs) == 0 {
		missing = append(missing, "evidence_paths")
	}
	for _, dimension := range missing {
		coord := coordinate{Stage: "AMBIGUITY_OBSERVATION", Step: dimension, Reason: "AMBIGUITY_COORDINATE_UNOBSERVED"}
		gaps = append(gaps, observationGap{Dimension: dimension, Coordinate: coord, EvidenceDigest: digestValue(struct {
			CaseID         string
			Anchor         string
			Dimension      string
			ActivityDigest string
		}{caseID, anchor, dimension, activityDigest})})
	}
	return missing, gaps
}

func normalizeElements(elements *ambiguityElements) {
	elements.CandidateIDs = uniqueStrings(elements.CandidateIDs)
	elements.ResolvedBranchIDs = uniqueStrings(elements.ResolvedBranchIDs)
	elements.UnresolvedBranchIDs = uniqueStrings(elements.UnresolvedBranchIDs)
	elements.EvidencePathIDs = uniqueStrings(elements.EvidencePathIDs)
	elements.BranchObservationIDs = uniqueStrings(elements.BranchObservationIDs)
}

func uniqueStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	out := result[:0]
	for _, value := range result {
		if len(out) == 0 || out[len(out)-1] != value {
			out = append(out, value)
		}
	}
	return out
}

func elementCounts(elements ambiguityElements) integerSet {
	return integerSet{InterpretationCandidates: len(elements.CandidateIDs), UnresolvedBranches: len(elements.UnresolvedBranchIDs), EvidencePaths: len(elements.EvidencePathIDs)}
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
	if len(source.Programs) != CaseTotal+1 || source.Activities != SourceActivityTotal || source.Entities != SourceEntityTotal {
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

func makeCaseReceipt(sourceDigest, sourceSemanticDigest string, program programObservation, policy budgetPolicy) caseReceipt {
	decision, resolution, reason := subjectDecision(program, policyCounts(policy))
	prop := proposition(program.ID, policy)
	propDigest := digestBytes([]byte(prop))
	evidence := digestValue(struct {
		PropositionDigest      string
		SourceSemanticDigest   string
		ActivitySemanticDigest string
		ProgramSemanticDigest  string
		ElementDigest          string
		Counts                 integerSet
		UnobservedDimensions   []string
		ObservationGaps        []observationGap
		Decision               string
		Resolution             string
		Reason                 string
	}{propDigest, sourceSemanticDigest, program.ActivitySemanticDigest, program.SemanticDigest, program.ElementDigest,
		program.Counts, program.UnobservedDimensions, program.ObservationGaps, decision, resolution, reason})
	coord := caseCoordinate(program, reason)
	claim := transition{CaseID: program.ID, Proposition: prop, PropositionDigest: propDigest, From: "OPEN", To: claimTarget(decision),
		Stage: coord.Stage, Step: coord.Step, Reason: reason, EvidenceDigest: evidence}
	return caseReceipt{ID: program.ID, Activity: program.Activity, RawSourceDigest: sourceDigest, Class: program.Class,
		InputState: program.InputState, Program: program.Program, ProgramDigest: program.Digest,
		ProgramSemanticDigest: program.SemanticDigest, ActivitySemanticDigest: program.ActivitySemanticDigest,
		Elements: program.Elements, ElementDigest: program.ElementDigest, Counts: program.Counts,
		UnobservedDimensions: append([]string(nil), program.UnobservedDimensions...), ObservationGaps: append([]observationGap(nil), program.ObservationGaps...),
		Decision: decision, Resolution: resolution, Reason: reason, Coordinate: coord, Proposition: prop,
		PropositionDigest: propDigest, Claim: claim, EvidenceDigest: evidence, Conformance: "MATCH"}
}

func caseCoordinate(program programObservation, reason string) coordinate {
	if len(program.ObservationGaps) > 0 {
		return program.ObservationGaps[0].Coordinate
	}
	return coordinate{Stage: "AMBIGUITY_BUDGET", Step: "case:" + program.ID, Reason: reason}
}

func proposition(caseID string, policy budgetPolicy) string {
	return "counts-within-budget(case:" + caseID + ",budget:" + policy.ID + ")"
}

func exceeds(value, budget integerSet) bool {
	return value.InterpretationCandidates > budget.InterpretationCandidates || value.UnresolvedBranches > budget.UnresolvedBranches || value.EvidencePaths > budget.EvidencePaths
}

func subjectDecision(program programObservation, budget integerSet) (string, string, string) {
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

func indicatorsFor(sourceSemanticDigest string, program programObservation, policy budgetPolicy) []indicator {
	values := []struct {
		metric, dimension, proof string
		observed, limit          int
	}{
		{"gooo.metric.ambiguity-budget.candidate-cardinality.v3", "interpretation_candidates", "FOUNDATION", program.Counts.InterpretationCandidates, policyLimit(policy, "interpretation_candidates")},
		{"gooo.metric.ambiguity-budget.unresolved-branch-cardinality.v3", "unresolved_branches", "COHERENCE", program.Counts.UnresolvedBranches, policyLimit(policy, "unresolved_branches")},
		{"gooo.metric.ambiguity-budget.evidence-path-cardinality.v3", "evidence_paths", "REGRESSION", program.Counts.EvidencePaths, policyLimit(policy, "evidence_paths")},
	}
	result := make([]indicator, 0, len(values))
	for _, value := range values {
		observed := !contains(program.UnobservedDimensions, value.dimension)
		evaluation := "UNOBSERVED"
		if observed {
			evaluation = "WITHIN_LIMIT"
			if value.observed > value.limit {
				evaluation = "EXCEEDS_LIMIT"
			}
		}
		evidence := digestValue(struct {
			SourceSemanticDigest   string
			ActivitySemanticDigest string
			ProgramSemanticDigest  string
			ElementDigest          string
			Dimension              string
			Observed               int
			CoordinateObserved     bool
			Budget                 int
		}{sourceSemanticDigest, program.ActivitySemanticDigest, program.SemanticDigest, program.ElementDigest,
			value.dimension, value.observed, observed, value.limit})
		result = append(result, indicator{MetricID: value.metric, CaseID: program.ID, Dimension: value.dimension, ProofChoice: value.proof,
			Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, Observed: value.observed, CoordinateObserved: observed,
			Budget: value.limit, Relation: "<=", Evaluation: evaluation, EvidenceDigest: evidence})
	}
	return result
}

func summarize(cases []caseReceipt, interventions []interventionReceipt, denominator denominator) summary {
	result := summary{CasesTotal: len(cases), IntegerDimensions: IntegerDimensionTotal, Denominator: denominator}
	for _, item := range cases {
		if item.InputState == "KNOWN" {
			result.KnownCases++
		}
		switch item.Class {
		case "ZERO":
			result.ZeroAmbiguityCases++
		case "BOUNDARY":
			result.BoundaryCases++
		case "OVER":
			result.OverBudgetCases++
		case "UNKNOWN":
			result.UnknownCases++
		}
		if item.Resolution == "LOWER_RESOLUTION" {
			result.LowerResolutionCases++
		}
		switch item.Claim.To {
		case "DISCHARGED":
			result.Numerator.ClaimsDischarged++
		case "REFUTED":
			result.Numerator.ClaimsRefuted++
		case "OPEN":
			result.OpenClaims++
			result.Numerator.ClaimsOpen++
		}
		for _, dimension := range integerDimensions {
			if !contains(item.UnobservedDimensions, dimension) {
				result.Numerator.IntegerObservationsObserved++
			}
		}
	}
	result.Numerator.CasesConforming = len(cases)
	for _, item := range interventions {
		if item.Satisfied {
			result.Numerator.InterventionsSatisfied++
		}
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

func buildInterventions(specs []interventionContract, raw []byte, base sourceObservation, policy budgetPolicy) []interventionReceipt {
	result := make([]interventionReceipt, 0, len(specs))
	for _, spec := range specs {
		mutated, err := interventionSource(raw, spec)
		if err != nil {
			result = append(result, interventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity})
			continue
		}
		next, err := observeSource(base.Path, mutated, policy)
		beforeProgram, beforeOK := findCase(base, spec.TargetActivity)
		afterProgram, afterOK := findCase(next, spec.TargetActivity)
		if err != nil || !beforeOK || !afterOK {
			result = append(result, interventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity})
			continue
		}
		before := makeCaseReceipt(base.Digest, base.SemanticDigest, beforeProgram, policy)
		after := makeCaseReceipt(next.Digest, next.SemanticDigest, afterProgram, policy)
		satisfied := interventionSatisfied(spec.Kind, before, after, base, next)
		evidence := digestValue(struct {
			ID     string
			Before caseReceipt
			After  caseReceipt
			Base   sourceObservation
			Next   sourceObservation
		}{spec.ID, before, after, base, next})
		result = append(result, interventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity,
			SourceDigestBefore: base.Digest, SourceDigestAfter: next.Digest, SemanticDigestBefore: base.SemanticDigest, SemanticDigestAfter: next.SemanticDigest,
			ElementsBefore: before.Elements, ElementsAfter: after.Elements, CountsBefore: before.Counts, CountsAfter: after.Counts,
			UnobservedBefore: append([]string(nil), before.UnobservedDimensions...), UnobservedAfter: append([]string(nil), after.UnobservedDimensions...),
			ClassBefore: before.Class, ClassAfter: after.Class, InputStateBefore: before.InputState, InputStateAfter: after.InputState,
			ClaimBefore: before.Claim, ClaimAfter: after.Claim, DecisionBefore: before.Decision, ResolutionBefore: before.Resolution, ReasonBefore: before.Reason,
			DecisionAfter: after.Decision, ResolutionAfter: after.Resolution, ReasonAfter: after.Reason, Satisfied: satisfied, EvidenceDigest: evidence})
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
	if spec.Kind != "SEMANTIC" || spec.TargetActivity != "BoundaryAmbiguity" {
		return nil, fmt.Errorf("unsupported intervention")
	}
	needle := "activity BoundaryAmbiguity(BoundaryCase) -> AmbiguityReceipt computes \"ambiguity-budget:case:boundary-ambiguity\""
	addition := "entity BoundaryCandidateC id \"gooo://ambiguity-budget/case/boundary-ambiguity/candidate/c\"\n" +
		"entity BoundaryEvidencePathC id \"gooo://ambiguity-budget/case/boundary-ambiguity/evidence-path/c\"\n" +
		"activity ObserveBoundaryCandidateC(BoundaryCase) -> BoundaryCandidateC\n" +
		"activity ObserveBoundaryEvidencePathC(BoundaryCase, BoundaryCandidateC) -> BoundaryEvidencePathC\n"
	if !strings.Contains(source, needle) {
		return nil, fmt.Errorf("semantic target not found")
	}
	return []byte(strings.Replace(source, needle, addition+needle, 1)), nil
}

func interventionSatisfied(kind string, before, after caseReceipt, base, next sourceObservation) bool {
	if kind == "SEMANTIC" {
		return base.Digest != next.Digest && base.SemanticDigest != next.SemanticDigest && !reflect.DeepEqual(before.Elements, after.Elements) && before.Counts != after.Counts &&
			before.Decision == "PASS" && before.Resolution == "EXACT" && before.Claim.From == "OPEN" && before.Claim.To == "DISCHARGED" &&
			after.Decision == "FAIL_CLOSED" && after.Resolution == "LOWER_RESOLUTION" && after.Reason == "AMBIGUITY_BUDGET_EXCEEDED" &&
			after.Claim.From == "OPEN" && after.Claim.To == "REFUTED" && before.Proposition == after.Proposition && before.Claim.PropositionDigest == after.Claim.PropositionDigest
	}
	return base.Digest != next.Digest && base.SemanticDigest == next.SemanticDigest && reflect.DeepEqual(before.Elements, after.Elements) && before.Counts == after.Counts &&
		before.Class == after.Class && before.InputState == after.InputState && before.Decision == after.Decision && before.Resolution == after.Resolution && before.Reason == after.Reason &&
		reflect.DeepEqual(before.UnobservedDimensions, after.UnobservedDimensions) && before.Proposition == after.Proposition && before.PropositionDigest == after.PropositionDigest && before.Claim == after.Claim
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
		if value.Choice != want[index] || value.Producer != Producer || value.Consumer != Consumer || value.MetaOperation == "" || !value.Passed || !validDigest(value.EvidenceDigest) {
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

func decodeEffects(raw []byte) (effects, error) {
	var value effects
	if err := decode(raw, &value); err != nil {
		return effects{}, err
	}
	if value.Schema != EffectsSchema || value.Version != "v1" || !value.TrackedAndUntracked || !validDigest(value.SnapshotBeforeDigest) ||
		!validDigest(value.SnapshotAfterDigest) || value.RepositoryWrites != 0 || !value.WriteSetEqual || value.MutationAuthority != "UNKNOWN" ||
		value.MutationAuthorityResolution != "NOT_OBSERVED" {
		return effects{}, fmt.Errorf("invalid workspace effects")
	}
	value.ArtifactDigest = digestBytes(raw)
	return value, nil
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
