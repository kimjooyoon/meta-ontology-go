package selectiveci

import (
	"math"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
)

func EvaluateObligationCoverage(input ObligationCoverageInput) ObligationCoverageResult {
	return ObserveObligationCoverage(input)
}

func ObserveObligationCoverage(input ObligationCoverageInput) ObligationCoverageResult {
	input = normalizeCoverageInput(input)
	result := coverageResultFor(input)
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return sealCoverage(result, CoverageDecisionUnknown, coverageInputReason(input))
	}
	result.InputDigest = digestBytes(canonical)
	if input.SchemaVersion != ObligationCoverageSchemaVersion {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonUnsupportedSchema)
	}
	if input.ChangedRootIDs == nil {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonMissingInput)
	}
	if !validDigest(input.SnapshotDigest) {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonInvalidSnapshot)
	}
	if reason := validateCoverageRegistry(input.Registry); reason != "" {
		return sealCoverage(result, CoverageDecisionUnknown, reason)
	}
	graph, err := input.Graph.Normalized()
	if err != nil {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonInvalidGraph)
	}
	result.GraphDigest = graph.Digest()
	if reason := validateCoverageBindings(graph, input.Registry); reason != "" {
		return sealCoverage(result, CoverageDecisionUnknown, reason)
	}
	if graph.SnapshotDigest != input.SnapshotDigest {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonStaleSnapshot)
	}
	if graph.RegistryDigest != input.Registry.Digest || graph.PolicyDigest != input.Registry.PolicyDigest {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonStaleGraph)
	}
	roots, reason := coverageRoots(input.ChangedRootIDs, graph)
	if reason != "" {
		return sealCoverage(result, CoverageDecisionUnknown, reason)
	}
	result.ChangedRootCount = uint64(len(roots))
	if len(roots) == 0 {
		return sealCoverage(result, CoverageDecisionExact, CoverageReasonNoChange)
	}
	required, uncovered := reachableCoverage(graph, input.Registry, roots)
	result.RequiredObligationCount = uint64(len(required))
	result.UncoveredRootIDs = uncovered
	result.UncoveredChangedRootCount = uint64(len(uncovered))
	result.CoveredChangedRootCount = result.ChangedRootCount - result.UncoveredChangedRootCount
	bound, scanned, reason := coverageCommandRecords(required, input.Registry)
	result.BoundCommandCount = bound
	work, ok := coverageWorkUnits(result.ChangedRootCount, result.RequiredObligationCount, scanned)
	if len(uncovered) != 0 {
		if ok {
			result.DeterministicWorkUnits = work
		}
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonMissingObligation)
	}
	if !ok {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonWorkOverflow)
	}
	result.DeterministicWorkUnits = work
	if reason != "" {
		return sealCoverage(result, CoverageDecisionUnknown, reason)
	}
	result.RequiredObligationIDs = required
	return sealCoverage(result, CoverageDecisionExact, CoverageReasonComplete)
}

func EvaluateCoverage(input ObligationCoverageInput) ObligationCoverageResult {
	return EvaluateObligationCoverage(input)
}

func ObserveCoverage(input ObligationCoverageInput) ObligationCoverageResult {
	return ObserveObligationCoverage(input)
}

func coverageResultFor(input ObligationCoverageInput) ObligationCoverageResult {
	result := ObligationCoverageResult{
		SchemaVersion:         ObligationCoverageSchemaVersion,
		UncoveredRootIDs:      []string{},
		RequiredObligationIDs: []string{},
	}
	result.SnapshotDigest = input.SnapshotDigest
	result.RegistryDigest = input.Registry.Digest
	result.GraphDigest = input.Graph.Digest()
	return result
}

func coverageInputReason(input ObligationCoverageInput) CoverageReason {
	if input.SchemaVersion != ObligationCoverageSchemaVersion {
		return CoverageReasonUnsupportedSchema
	}
	if input.ChangedRootIDs == nil {
		return CoverageReasonMissingInput
	}
	if !validDigest(input.SnapshotDigest) {
		return CoverageReasonInvalidSnapshot
	}
	if input.Registry.SchemaVersion != RegistrySchemaVersion {
		return CoverageReasonInvalidRegistry
	}
	if input.Graph.Version != impactgraph.SchemaVersion {
		return CoverageReasonInvalidGraph
	}
	return CoverageReasonInvalidInput
}

func validateCoverageRegistry(registry Registry) CoverageReason {
	if err := registry.Validate(); err != nil {
		reason := reasonFor(err)
		switch reason {
		case ReasonUnsupportedSchema:
			return CoverageReasonUnsupportedSchema
		case ReasonMismatchedDigest:
			return CoverageReasonStaleRegistry
		case ReasonMissingBinding, ReasonMissingCommand:
			return CoverageReasonMissingCommand
		case ReasonDanglingReference:
			return CoverageReasonDanglingCommand
		default:
			return CoverageReasonInvalidRegistry
		}
	}
	return ""
}

func validateCoverageBindings(graph impactgraph.Graph, registry Registry) CoverageReason {
	nodes := make(map[string]impactgraph.NodeKind, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node.Kind
	}
	bindings := make(map[string]struct{}, len(registry.Obligations))
	for _, binding := range registry.Obligations {
		bindings[binding.ID] = struct{}{}
		if nodes[binding.ID] != impactgraph.NodeKindObligation {
			return CoverageReasonStaleGraph
		}
	}
	for _, node := range graph.Nodes {
		if node.Kind == impactgraph.NodeKindObligation {
			if _, registered := bindings[node.ID]; !registered {
				return CoverageReasonStaleRegistry
			}
		}
	}
	return ""
}

func coverageRoots(raw []string, graph impactgraph.Graph) ([]string, CoverageReason) {
	roots := sortedCopy(raw)
	byID := make(map[string]impactgraph.NodeKind, len(graph.Nodes))
	for _, node := range graph.Nodes {
		byID[node.ID] = node.Kind
	}
	for index, root := range roots {
		if index > 0 && root == roots[index-1] {
			return nil, CoverageReasonDuplicateRoot
		}
		if root == "" || strings.TrimSpace(root) != root {
			return nil, CoverageReasonUnknownRoot
		}
		if byID[root] != impactgraph.NodeKindSemantic {
			return nil, CoverageReasonUnknownRoot
		}
	}
	return roots, ""
}

func reachableCoverage(graph impactgraph.Graph, registry Registry, roots []string) ([]string, []string) {
	byID := make(map[string]impactgraph.NodeKind, len(graph.Nodes))
	adjacency := make(map[string][]string, len(graph.Nodes))
	for _, node := range graph.Nodes {
		byID[node.ID] = node.Kind
	}
	for _, edge := range graph.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	registered := make(map[string]struct{}, len(registry.Obligations))
	for _, binding := range registry.Obligations {
		registered[binding.ID] = struct{}{}
	}
	requiredSet := make(map[string]struct{})
	uncovered := make([]string, 0)
	for _, root := range roots {
		reached := reachableFromCoverageRoot(root, adjacency, byID, registered)
		if len(reached) == 0 {
			uncovered = append(uncovered, root)
		}
		for _, obligation := range reached {
			requiredSet[obligation] = struct{}{}
		}
	}
	required := make([]string, 0, len(requiredSet))
	for obligation := range requiredSet {
		required = append(required, obligation)
	}
	sort.Strings(required)
	sort.Strings(uncovered)
	return required, uncovered
}

func reachableFromCoverageRoot(root string, adjacency map[string][]string, byID map[string]impactgraph.NodeKind, registered map[string]struct{}) []string {
	queue := []string{root}
	visited := map[string]struct{}{}
	reached := map[string]struct{}{}
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if _, seen := visited[current]; seen {
			continue
		}
		visited[current] = struct{}{}
		if byID[current] == impactgraph.NodeKindObligation {
			if _, ok := registered[current]; ok {
				reached[current] = struct{}{}
			}
		}
		queue = append(queue, adjacency[current]...)
	}
	result := make([]string, 0, len(reached))
	for obligation := range reached {
		result = append(result, obligation)
	}
	sort.Strings(result)
	return result
}

func coverageCommandRecords(required []string, registry Registry) (uint64, uint64, CoverageReason) {
	commands := make(map[string]struct{}, len(registry.Commands))
	for _, command := range registry.Commands {
		commands[command.ID] = struct{}{}
	}
	bindings := bindingIndex(registry.Obligations)
	bound := make(map[string]struct{})
	var scanned uint64
	for _, obligation := range required {
		binding, ok := bindings[obligation]
		if !ok {
			return uint64(len(bound)), scanned, CoverageReasonMissingObligation
		}
		if len(binding.CommandIDs) == 0 {
			return uint64(len(bound)), scanned, CoverageReasonMissingCommand
		}
		for _, command := range binding.CommandIDs {
			var ok bool
			scanned, ok = coverageAdd(scanned, 1)
			if !ok {
				return uint64(len(bound)), 0, CoverageReasonWorkOverflow
			}
			if _, registered := commands[command]; !registered {
				return uint64(len(bound)), scanned, CoverageReasonDanglingCommand
			}
			bound[command] = struct{}{}
		}
	}
	return uint64(len(bound)), scanned, ""
}

func coverageWorkUnits(roots, obligations, commandBindings uint64) (uint64, bool) {
	total, ok := coverageAdd(roots, obligations)
	if !ok {
		return 0, false
	}
	return coverageAdd(total, commandBindings)
}

func coverageAdd(left, right uint64) (uint64, bool) {
	if math.MaxUint64-left < right {
		return 0, false
	}
	return left + right, true
}

func sealCoverage(result ObligationCoverageResult, decision CoverageDecision, reason CoverageReason) ObligationCoverageResult {
	result.Decision = decision
	result.Reason = reason
	result.FullSuiteRequired = decision == CoverageDecisionUnknown
	if decision == CoverageDecisionUnknown {
		result.RequiredObligationIDs = []string{}
	}
	result = normalizeCoverageResult(result)
	result.OutputDigest = result.StableDigest()
	return result
}

func normalizeCoverageResult(result ObligationCoverageResult) ObligationCoverageResult {
	result.UncoveredRootIDs = sortedUnique(result.UncoveredRootIDs)
	result.RequiredObligationIDs = sortedUnique(result.RequiredObligationIDs)
	if result.UncoveredRootIDs == nil {
		result.UncoveredRootIDs = []string{}
	}
	if result.RequiredObligationIDs == nil {
		result.RequiredObligationIDs = []string{}
	}
	return result
}
