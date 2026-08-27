package ambiguitybudget

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const canonicalLowering = "syntax.ParseFile->bidir.Lower"

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

func observeSource(path string, raw []byte, policy BudgetPolicy) (SourceObservation, error) {
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return SourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, fmt.Errorf("source syntax is unknown")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, fmt.Errorf("source semantic lowering is unknown: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return SourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, fmt.Errorf("source semantic validation is unknown: %w", err)
	}

	index := indexGraph(ir)
	observation := SourceObservation{Path: path, Digest: digestBytes(raw), SemanticDigest: "sha256:" + ir.StableHash(), Lowering: canonicalLowering,
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
			program, parseErr := parseComputesProgram(node.Name, node.ValueProgram)
			if parseErr != nil {
				return observation, fmt.Errorf("activity %q computes: %w", node.Name, parseErr)
			}
			if program.Kind == "CASE" {
				programObservation, observeErr := observeCaseProgram(index, node, program, policy)
				if observeErr != nil {
					return observation, observeErr
				}
				observation.Programs = append(observation.Programs, programObservation)
				continue
			}
			observation.Programs = append(observation.Programs, ProgramObservation{Activity: program.Activity, Program: program.Text,
				ProgramKind: program.Kind, ID: program.ID, Digest: digestBytes([]byte(program.Text)),
				ActivitySemanticDigest: activitySemanticDigest(index, node)})
		}
	}
	sort.Slice(observation.Programs, func(i, j int) bool { return observation.Programs[i].Activity < observation.Programs[j].Activity })
	return observation, nil
}

func observeCaseProgram(index graphIndex, activityNode semantic.Node, program computesProgram, policy BudgetPolicy) (ProgramObservation, error) {
	used := index.used[activityNode.ID.String()]
	anchor := ""
	for _, candidate := range used {
		if node, ok := index.nodes[candidate]; ok && node.Kind == semantic.Entity && strings.Contains(candidate, "/case/"+program.ID) {
			if anchor != "" {
				return ProgramObservation{}, fmt.Errorf("activity %q has multiple case anchors", activityNode.Name)
			}
			anchor = candidate
		}
	}
	if anchor == "" {
		return ProgramObservation{}, fmt.Errorf("activity %q has no case anchor", activityNode.Name)
	}

	elements := AmbiguityElements{}
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
		if strings.Contains(entityID, "/branch-observation") {
			elements.BranchObservationIDs = append(elements.BranchObservationIDs, entityID)
		}
		if strings.Contains(entityID, "/evidence-path/") && pathUsesCandidate(index, entityID, anchor) {
			elements.EvidencePathIDs = append(elements.EvidencePathIDs, entityID)
		}
	}
	normalizeElements(&elements)
	counts := IntegerSet{InterpretationCandidates: len(elements.CandidateIDs), UnresolvedBranches: len(elements.UnresolvedBranchIDs), EvidencePaths: len(elements.EvidencePathIDs)}
	activityDigest := activitySemanticDigest(index, activityNode)
	elementDigest := digestValue(elements)
	unobserved, gaps := observationGaps(program.ID, anchor, activityDigest, elements)
	result := ProgramObservation{Activity: program.Activity, Program: program.Text, ProgramKind: program.Kind, ID: program.ID,
		Counts: counts, UnobservedDimensions: unobserved, ObservationGaps: gaps, Elements: elements, ElementDigest: elementDigest,
		ActivitySemanticDigest: activityDigest, SemanticDigest: digestValue(struct {
			Activity string
			Program  string
			Elements AmbiguityElements
			Counts   IntegerSet
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
		for _, entity := range used {
			if entity != anchor && strings.Contains(entity, "/candidate/") {
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

func observationGaps(caseID, anchor, activityDigest string, elements AmbiguityElements) ([]string, []ObservationGap) {
	var missing []string
	var gaps []ObservationGap
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
		coordinate := Coordinate{Stage: "AMBIGUITY_OBSERVATION", Step: dimension, Reason: "AMBIGUITY_COORDINATE_UNOBSERVED"}
		gaps = append(gaps, ObservationGap{Dimension: dimension, Coordinate: coordinate,
			EvidenceDigest: digestValue(struct {
				CaseID         string
				Anchor         string
				Dimension      string
				ActivityDigest string
			}{caseID, anchor, dimension, activityDigest})})
	}
	return missing, gaps
}

func normalizeElements(elements *AmbiguityElements) {
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

func findBudget(source SourceObservation, activity string) (ProgramObservation, bool) {
	for _, program := range source.Programs {
		if program.ProgramKind == "BUDGET" && program.Activity == activity {
			return program, true
		}
	}
	return ProgramObservation{}, false
}

func findCase(source SourceObservation, activity string) (ProgramObservation, bool) {
	for _, program := range source.Programs {
		if program.ProgramKind == "CASE" && program.Activity == activity {
			return program, true
		}
	}
	return ProgramObservation{}, false
}
