package roundtrip

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Verify checks both projection directions, the full round-trip, and any
// supplied generated-source locality witness.
func Verify(observation Observation) Report {
	var report Report
	report.merge(CheckDSLToIR(observation.DSL, observation.IR))
	report.merge(CheckGoToIR(observation.IR, observation.GoIR))
	report.merge(CheckRoundTrip(observation.DSL, observation.GoIR))
	if observation.BeforeGo != nil || observation.AfterGo != nil {
		allowed := append([]semantic.ID(nil), observation.AllowedIDs...)
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
func CheckDSLToIR(dsl, lowered semantic.IR) Report {
	return comparePair(RuleDSLToIR, "dsl", dsl, "ir", lowered)
}

// CheckGoToIR checks the canonical IR against facts lifted from Go.
func CheckGoToIR(ir, lifted semantic.IR) Report {
	return comparePair(RuleGoToIR, "ir", ir, "go-ir", lifted)
}

// CheckRoundTrip checks DSL → IR → Go → IR semantic stability.
func CheckRoundTrip(dsl, lifted semantic.IR) Report {
	return comparePair(RuleRoundTrip, "dsl", dsl, "go-ir", lifted)
}

// SemanticDelta computes presentation-insensitive changes between snapshots.
func SemanticDelta(before, after semantic.IR) (Delta, error) {
	left, err := before.Normalized()
	if err != nil {
		return Delta{}, fmt.Errorf("before IR: %w", err)
	}
	right, err := after.Normalized()
	if err != nil {
		return Delta{}, fmt.Errorf("after IR: %w", err)
	}
	delta := Delta{
		MetadataChanged: left.Version != right.Version || left.Package != right.Package || left.Namespace != right.Namespace,
	}
	diffNodes(&delta, left.Graph.Nodes(), right.Graph.Nodes())
	diffFacts(&delta, left.Graph.DeterministicFacts(), right.Graph.DeterministicFacts())
	delta.TouchedIDs = touchedIDs(delta)
	delta.AffectedIDs = affectedIDs(delta, left, right)
	return delta, nil
}

func comparePair(rule, expectedPath string, expected semantic.IR, actualPath string, actual semantic.IR) Report {
	var report Report
	comparison := semantic.CompareIR(expected, actual)
	if !comparison.LeftValid {
		report.add(snapshotViolation(expectedPath, errorString(comparison.LeftError)))
	}
	if !comparison.RightValid {
		report.add(snapshotViolation(actualPath, errorString(comparison.RightError)))
	}
	if !report.OK() {
		report.normalize()
		return report
	}
	if comparison.SemanticEqual {
		return report
	}
	delta, err := SemanticDelta(expected, actual)
	if err != nil {
		report.add(snapshotViolation("semantic", err))
		return report
	}
	if delta.MetadataChanged {
		report.add(semanticViolation(rule, "ir-metadata", "unchanged", "changed", "semantic IR metadata changed"))
	}
	for _, node := range delta.RemovedNodes {
		report.add(semanticViolation(rule, node.ID.String(), "present", "removed", "semantic node was removed or changed"))
	}
	for _, node := range delta.AddedNodes {
		report.add(semanticViolation(rule, node.ID.String(), "absent", "added", "semantic node was added or changed"))
	}
	for _, fact := range delta.RemovedFacts {
		report.add(semanticViolation(rule, factIdentity(fact), "present", "removed", "semantic fact was removed or changed"))
	}
	for _, fact := range delta.AddedFacts {
		report.add(semanticViolation(rule, factIdentity(fact), "absent", "added", "semantic fact was added or changed"))
	}
	if len(report.Violations) == 0 {
		report.add(semanticViolation(rule, "ir", "equivalent", "different", "semantic digest differs without an explainable delta"))
	}
	report.normalize()
	return report
}

func errorString(message string) error {
	if message == "" {
		return fmt.Errorf("semantic snapshot is invalid")
	}
	return fmt.Errorf("%s", message)
}

func diffNodes(delta *Delta, before, after []semantic.Node) {
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

func diffFacts(delta *Delta, before, after []semantic.Fact) {
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
	sort.Slice(delta.AddedFacts, func(i, j int) bool { return factIdentity(delta.AddedFacts[i]) < factIdentity(delta.AddedFacts[j]) })
	sort.Slice(delta.RemovedFacts, func(i, j int) bool { return factIdentity(delta.RemovedFacts[i]) < factIdentity(delta.RemovedFacts[j]) })
}

func nodeMap(nodes []semantic.Node) map[semantic.ID]semantic.Node {
	result := make(map[semantic.ID]semantic.Node, len(nodes))
	for _, node := range nodes {
		result[node.ID] = node
	}
	return result
}

func factMap(facts []semantic.Fact) map[semantic.FactKey]semantic.Fact {
	result := make(map[semantic.FactKey]semantic.Fact, len(facts))
	for _, fact := range facts {
		result[fact.Key()] = fact
	}
	return result
}

func sameNode(left, right semantic.Node) bool {
	return left.SemanticCanonical() == right.SemanticCanonical()
}

func sameFact(left, right semantic.Fact) bool {
	return left.SemanticCanonical() == right.SemanticCanonical()
}

func factIdentity(fact semantic.Fact) string {
	key := fact.Key()
	return key.Subject.String() + "\x00" + key.Predicate.String() + "\x00" + key.Object.String()
}

func touchedIDs(delta Delta) []semantic.ID {
	values := make(map[semantic.ID]struct{})
	for _, node := range append(delta.AddedNodes, delta.RemovedNodes...) {
		values[node.ID] = struct{}{}
	}
	for _, fact := range append(delta.AddedFacts, delta.RemovedFacts...) {
		values[fact.Subject] = struct{}{}
		values[fact.Object] = struct{}{}
	}
	return sortedIDs(values)
}

func affectedIDs(delta Delta, before, after semantic.IR) []semantic.ID {
	values := make(map[semantic.ID]struct{})
	for _, id := range delta.TouchedIDs {
		values[id] = struct{}{}
	}
	facts := append(before.Graph.DeterministicFacts(), after.Graph.DeterministicFacts()...)
	for _, fact := range facts {
		if containsID(delta.TouchedIDs, fact.Subject) {
			values[fact.Object] = struct{}{}
		}
		if containsID(delta.TouchedIDs, fact.Object) {
			values[fact.Subject] = struct{}{}
		}
	}
	return sortedIDs(values)
}

func changedLocality(before, after semantic.IR) []semantic.ID {
	delta, err := SemanticDelta(before, after)
	if err != nil {
		return nil
	}
	return append([]semantic.ID(nil), delta.AffectedIDs...)
}

func hasIR(left, right semantic.IR) bool {
	return len(left.Graph.Nodes()) > 0 || len(left.Graph.DeterministicFacts()) > 0 ||
		len(right.Graph.Nodes()) > 0 || len(right.Graph.DeterministicFacts()) > 0
}

func containsID(values []semantic.ID, target semantic.ID) bool {
	index := sort.Search(len(values), func(index int) bool { return values[index] >= target })
	return index < len(values) && values[index] == target
}

func sortedIDs(values map[semantic.ID]struct{}) []semantic.ID {
	result := make([]semantic.ID, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
