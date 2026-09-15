package roundtrip

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

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
