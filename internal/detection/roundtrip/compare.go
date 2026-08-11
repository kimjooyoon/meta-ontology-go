package roundtrip

import (
	"fmt"
	"sort"
)

// Verify checks both projection directions, the full round-trip, and any
// supplied generated-source locality witness.
func Verify(observation Observation) Report {
	var report Report
	report.merge(CheckDSLToIR(observation.DSL, observation.IR))
	report.merge(CheckGoToIR(observation.IR, observation.GoIR))
	report.merge(CheckRoundTrip(observation.DSL, observation.GoIR))
	if observation.BeforeGo != nil || observation.AfterGo != nil {
		allowed := observation.AllowedIDs
		if len(allowed) == 0 && hasIR(observation.BeforeIR, observation.AfterIR) {
			allowed = changedLocality(observation.BeforeIR, observation.AfterIR)
		}
		report.merge(CheckLocality(LocalityInput{
			Before:     observation.BeforeGo,
			After:      observation.AfterGo,
			AllowedIDs: allowed,
		}))
	}
	report.normalize()
	return report
}

// CheckDSLToIR checks the authoritative DSL semantic view against its lowered
// IR projection.
func CheckDSLToIR(dsl, lowered IR) Report {
	return comparePair(RuleDSLToIR, "dsl", dsl, "ir", lowered)
}

// CheckGoToIR checks the canonical IR against facts lifted from Go.
func CheckGoToIR(ir, lifted IR) Report {
	return comparePair(RuleGoToIR, "ir", ir, "go-ir", lifted)
}

// CheckRoundTrip checks DSL → IR → Go → IR semantic stability.
func CheckRoundTrip(dsl, lifted IR) Report {
	return comparePair(RuleRoundTrip, "dsl", dsl, "go-ir", lifted)
}

// SemanticDelta computes presentation-insensitive changes between snapshots.
func SemanticDelta(before, after IR) (Delta, error) {
	if err := before.Validate(); err != nil {
		return Delta{}, fmt.Errorf("before IR: %w", err)
	}
	if err := after.Validate(); err != nil {
		return Delta{}, fmt.Errorf("after IR: %w", err)
	}
	delta := Delta{}
	diffNodes(&delta, before.normalized().Nodes, after.normalized().Nodes)
	diffFacts(&delta, before.normalized().Facts, after.normalized().Facts)
	delta.TouchedIDs = touchedIDs(delta)
	delta.AffectedIDs = affectedIDs(delta, before, after)
	return delta, nil
}

func comparePair(rule, expectedPath string, expected IR, actualPath string, actual IR) Report {
	var report Report
	if err := expected.Validate(); err != nil {
		report.add(snapshotViolation(expectedPath, err))
	}
	if err := actual.Validate(); err != nil {
		report.add(snapshotViolation(actualPath, err))
	}
	if !report.OK() {
		report.normalize()
		return report
	}
	if Equivalent(expected, actual) {
		return report
	}
	delta, err := SemanticDelta(expected, actual)
	if err != nil {
		report.add(snapshotViolation("semantic", err))
		return report
	}
	for _, node := range delta.RemovedNodes {
		report.add(semanticViolation(rule, node.ID, "present", "removed", "semantic node was removed"))
	}
	for _, node := range delta.AddedNodes {
		report.add(semanticViolation(rule, node.ID, "absent", "added", "semantic node was added or changed"))
	}
	for _, fact := range delta.RemovedFacts {
		report.add(semanticViolation(rule, factKey(fact), "present", "removed", "semantic fact was removed"))
	}
	for _, fact := range delta.AddedFacts {
		report.add(semanticViolation(rule, factKey(fact), "absent", "added", "semantic fact was added or changed"))
	}
	report.normalize()
	return report
}

func diffNodes(delta *Delta, before, after []Node) {
	left, right := nodeMap(before), nodeMap(after)
	for id, node := range left {
		other, exists := right[id]
		if !exists || !sameNode(node, other) {
			delta.RemovedNodes = append(delta.RemovedNodes, node)
		}
	}
	for id, node := range right {
		other, exists := left[id]
		if !exists || !sameNode(node, other) {
			delta.AddedNodes = append(delta.AddedNodes, node)
		}
	}
	sort.Slice(delta.AddedNodes, func(i, j int) bool { return delta.AddedNodes[i].ID < delta.AddedNodes[j].ID })
	sort.Slice(delta.RemovedNodes, func(i, j int) bool { return delta.RemovedNodes[i].ID < delta.RemovedNodes[j].ID })
}

func diffFacts(delta *Delta, before, after []Fact) {
	left, right := factMap(before), factMap(after)
	for key, fact := range left {
		other, exists := right[key]
		if !exists || !sameFact(fact, other) {
			delta.RemovedFacts = append(delta.RemovedFacts, fact)
		}
	}
	for key, fact := range right {
		other, exists := left[key]
		if !exists || !sameFact(fact, other) {
			delta.AddedFacts = append(delta.AddedFacts, fact)
		}
	}
	sort.Slice(delta.AddedFacts, func(i, j int) bool { return factKey(delta.AddedFacts[i]) < factKey(delta.AddedFacts[j]) })
	sort.Slice(delta.RemovedFacts, func(i, j int) bool { return factKey(delta.RemovedFacts[i]) < factKey(delta.RemovedFacts[j]) })
}

func nodeMap(nodes []Node) map[string]Node {
	result := make(map[string]Node, len(nodes))
	for _, node := range nodes {
		result[node.ID] = node
	}
	return result
}

func factMap(facts []Fact) map[string]Fact {
	result := make(map[string]Fact, len(facts))
	for _, fact := range facts {
		result[factKey(fact)] = fact
	}
	return result
}

func touchedIDs(delta Delta) []string {
	values := make(map[string]struct{})
	for _, node := range append(delta.AddedNodes, delta.RemovedNodes...) {
		values[node.ID] = struct{}{}
	}
	for _, fact := range append(delta.AddedFacts, delta.RemovedFacts...) {
		values[fact.Subject] = struct{}{}
		values[fact.Object] = struct{}{}
	}
	return sortedIDs(values)
}

func affectedIDs(delta Delta, before, after IR) []string {
	values := make(map[string]struct{})
	for _, id := range delta.TouchedIDs {
		values[id] = struct{}{}
	}
	for _, fact := range append(before.Facts, after.Facts...) {
		if containsID(delta.TouchedIDs, fact.Subject) {
			values[fact.Object] = struct{}{}
		}
		if containsID(delta.TouchedIDs, fact.Object) {
			values[fact.Subject] = struct{}{}
		}
	}
	return sortedIDs(values)
}

func changedLocality(before, after IR) []string {
	delta, err := SemanticDelta(before, after)
	if err != nil {
		return nil
	}
	return delta.AffectedIDs
}

func hasIR(left, right IR) bool {
	return len(left.Nodes) > 0 || len(left.Facts) > 0 || len(right.Nodes) > 0 || len(right.Facts) > 0
}

func containsID(values []string, target string) bool {
	return sort.SearchStrings(values, target) < len(values) && values[sort.SearchStrings(values, target)] == target
}

func sortedIDs(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
