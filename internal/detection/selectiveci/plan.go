package selectiveci

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func Plan(input Input) PlanResult {
	result := PlanResult{SchemaVersion: SchemaVersion, BaseSnapshotDigest: input.Base.Digest, HeadSnapshotDigest: input.Head.Digest}
	if err := input.Validate(); err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	changed, err := changedSemanticIDs(input.Base, input.Head)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	result.ChangedSemanticIDs = changed
	graph, err := buildGraph(input, changed)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	required, err := applicableObligations(graph, changed)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	commands, guards, err := selectedCommands(input.Registry, required)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	frontier, selected, err := commandFrontier(input, commands, guards)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	receiptDigests, pathIDs, err := validateSelectedEvidence(input, selected)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	result = fillSelection(result, selected, frontier, guards)
	result.ResourceReceiptDigests = receiptDigests
	result.ProvenancePathIDs = pathIDs
	result.Status = StatusSelective
	result.ReasonCode = ReasonNone
	return sealResult(result)
}

func Evaluate(input Input) PlanResult { return Plan(input) }

func PlanJSON(data []byte) PlanResult {
	input, err := DecodeJSON(data)
	if err != nil {
		return sealResult(fallback(PlanResult{SchemaVersion: SchemaVersion}, reasonFor(err)))
	}
	return Plan(input)
}

func PlanJSONWithError(data []byte) (PlanResult, error) {
	input, err := DecodeJSON(data)
	if err != nil {
		return sealResult(fallback(PlanResult{SchemaVersion: SchemaVersion}, reasonFor(err))), err
	}
	return Plan(input), nil
}

func fallback(result PlanResult, reason string) PlanResult {
	result.Status = StatusFullSuiteFallback
	result.ReasonCode = reason
	result.SelectedCommandIDs = nil
	result.SelectedGuardCommandIDs = nil
	result.SelectedWorkIDs = nil
	result.ResourceReceiptDigests = nil
	result.ProvenancePathIDs = nil
	return result
}

func changedSemanticIDs(base, head SnapshotManifest) ([]string, error) {
	baseFiles, headFiles := manifestFiles(base), manifestFiles(head)
	ids := map[string]struct{}{}
	paths := map[string]struct{}{}
	for path := range baseFiles {
		paths[path] = struct{}{}
	}
	for path := range headFiles {
		paths[path] = struct{}{}
	}
	for path := range paths {
		before, beforeOK := baseFiles[path]
		after, afterOK := headFiles[path]
		if beforeOK && afterOK && before.BlobDigest == after.BlobDigest && equalStrings(before.SemanticIDs, after.SemanticIDs) {
			continue
		}
		for _, id := range append(append([]string{}, before.SemanticIDs...), after.SemanticIDs...) {
			ids[id] = struct{}{}
		}
		if len(before.SemanticIDs) == 0 && len(after.SemanticIDs) == 0 {
			return nil, failure(ReasonUnknownPath, "changed path has no stable semantic ID")
		}
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
}

func manifestFiles(manifest SnapshotManifest) map[string]SnapshotFile {
	result := make(map[string]SnapshotFile, len(manifest.Files))
	for _, file := range manifest.Files {
		result[file.Path] = file
	}
	return result
}

func equalStrings(left, right []string) bool {
	left, right = sortedCopy(left), sortedCopy(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func buildGraph(input Input, changed []string) (impactgraph.Graph, error) {
	registry := input.Registry
	nodes := append([]impactgraph.Node(nil), registry.Nodes...)
	byID := map[string]impactgraph.NodeKind{}
	for _, node := range nodes {
		byID[node.ID] = node.Kind
	}
	for _, file := range append(append([]SnapshotFile{}, input.Base.Files...), input.Head.Files...) {
		for _, id := range file.SemanticIDs {
			if err := addNode(byID, &nodes, impactgraph.Node{ID: id, Kind: impactgraph.NodeKindSemantic}); err != nil {
				return impactgraph.Graph{}, err
			}
		}
	}
	edges := make([]impactgraph.Edge, 0, len(registry.DependencyEdges)+len(registry.Obligations))
	for _, edge := range registry.DependencyEdges {
		edges = append(edges, impactgraph.Edge{From: edge.From, To: edge.To, Kind: edge.Kind})
	}
	for _, binding := range registry.Obligations {
		if err := addNode(byID, &nodes, impactgraph.Node{ID: binding.ID, Kind: impactgraph.NodeKindObligation}); err != nil {
			return impactgraph.Graph{}, err
		}
		if err := addNode(byID, &nodes, impactgraph.Node{ID: binding.Subject, Kind: impactgraph.NodeKindSemantic}); err != nil {
			return impactgraph.Graph{}, err
		}
		edges = append(edges, impactgraph.Edge{From: binding.Subject, To: binding.ID, Kind: impactgraph.EdgeKindAffects})
	}
	graph := impactgraph.Graph{Version: impactgraph.SchemaVersion, SnapshotDigest: input.Head.Digest, RegistryDigest: registry.Digest, PolicyDigest: registry.PolicyDigest, Nodes: nodes, Edges: edges}
	if err := graph.Validate(); err != nil {
		return impactgraph.Graph{}, failure(graphFailureReason(err.Error()), err.Error())
	}
	return graph, nil
}

func addNode(byID map[string]impactgraph.NodeKind, nodes *[]impactgraph.Node, node impactgraph.Node) error {
	if kind, exists := byID[node.ID]; exists {
		if kind != node.Kind {
			return failure(ReasonEvaluatorError, "stable ID has conflicting node kinds")
		}
		return nil
	}
	byID[node.ID] = node.Kind
	*nodes = append(*nodes, node)
	return nil
}

func applicableObligations(graph impactgraph.Graph, changed []string) ([]string, error) {
	if len(changed) == 0 {
		return nil, nil
	}
	executed := make([]string, 0)
	for _, node := range graph.Nodes {
		if node.Kind == impactgraph.NodeKindObligation {
			executed = append(executed, node.ID)
		}
	}
	evaluation := impactgraph.Evaluate(graph, changed, executed)
	if evaluation.Decision != impactgraph.PASS {
		return nil, failure(ReasonEvaluatorError, evaluation.FailureCode)
	}
	return evaluation.Required, nil
}

func selectedCommands(registry Registry, obligations []string) ([]Command, []Command, error) {
	commands := commandIndex(registry.Commands)
	guards := append([]Command(nil), registry.GlobalGuardCommands...)
	bindings := bindingIndex(registry.Obligations)
	selected := map[string]Command{}
	for _, obligationID := range obligations {
		binding, ok := bindings[obligationID]
		if !ok {
			return nil, nil, failure(ReasonMissingBinding, "required obligation is not bound")
		}
		for _, commandID := range binding.CommandIDs {
			command, ok := commands[commandID]
			if !ok {
				return nil, nil, failure(ReasonDanglingReference, "obligation command is not registered")
			}
			selected[command.ID] = command
		}
	}
	return sortedCommands(selected), guards, nil
}

func commandIndex(commands []Command) map[string]Command {
	result := make(map[string]Command, len(commands))
	for _, command := range commands {
		result[command.ID] = command
	}
	return result
}

func bindingIndex(bindings []ObligationBinding) map[string]ObligationBinding {
	result := make(map[string]ObligationBinding, len(bindings))
	for _, binding := range bindings {
		result[binding.ID] = binding
	}
	return result
}

func sortedCommands(commands map[string]Command) []Command {
	result := make([]Command, 0, len(commands))
	for _, command := range commands {
		result = append(result, command)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

type selectedPath struct {
	command    Command
	obligation string
	guard      bool
}

func commandFrontier(input Input, commands, guards []Command) (workfrontier.Result, []selectedPath, error) {
	paths := make([]workfrontier.RepairPath, 0, len(commands)+len(guards))
	selected := make([]selectedPath, 0, len(commands)+len(guards))
	for _, command := range commands {
		selected = append(selected, selectedPath{command: command})
	}
	for _, command := range guards {
		selected = append(selected, selectedPath{command: command, guard: true})
	}
	states := make([]workfrontier.ObligationState, 0, len(selected))
	pressures := make([]workfrontier.Pressure, 0, len(selected)*2)
	for index := range selected {
		entry := &selected[index]
		if !entry.guard {
			entry.obligation = firstObligationForCommand(input.Registry.Obligations, entry.command.ID)
		} else {
			entry.obligation = "guard/" + entry.command.ID
		}
		pressures = append(pressures, workfrontier.Pressure{StableID: entry.obligation}, workfrontier.Pressure{StableID: entry.command.ID})
		states = append(states, workfrontier.ObligationState{ObligationID: entry.obligation, Status: "PENDING"})
		paths = append(paths, workfrontier.RepairPath{StableID: entry.command.ID, ObligationID: entry.obligation, ReadSet: []string{entry.obligation}, WriteSet: []string{entry.command.ID}, RequiredPressureIDs: []string{entry.obligation, entry.command.ID}, CPUCoreNSUpperBound: entry.command.CPUWorkUnits})
	}
	frontier := workfrontier.Select(workfrontier.Input{SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: input.Head.Digest, PolicyDigest: input.Registry.PolicyDigest, RegistryDigest: input.Registry.Digest, MinimumSelectedPressures: 2, Capacity: workfrontier.Capacity{CPUCoreNS: input.CPUCapacity}, Pressures: pressures, States: states, Paths: paths})
	if frontier.Status == workfrontier.DecisionUnknown {
		return frontier, nil, failure(ReasonEvaluatorError, "work frontier returned UNKNOWN")
	}
	if frontier.Status == workfrontier.DecisionBlocked || len(frontier.SelectedIDs) != len(paths) {
		return frontier, nil, failure(ReasonFrontierBlocked, "work frontier did not select every command")
	}
	byID := map[string]selectedPath{}
	for _, entry := range selected {
		byID[entry.command.ID] = entry
	}
	ordered := make([]selectedPath, 0, len(frontier.SelectedIDs))
	for _, id := range frontier.SelectedIDs {
		ordered = append(ordered, byID[id])
	}
	return frontier, ordered, nil
}

func firstObligationForCommand(bindings []ObligationBinding, commandID string) string {
	for _, binding := range bindings {
		for _, candidate := range binding.CommandIDs {
			if candidate == commandID {
				return binding.ID
			}
		}
	}
	return ""
}

func fillSelection(result PlanResult, selected []selectedPath, frontier workfrontier.Result, guards []Command) PlanResult {
	workByCommand := map[string]string{}
	for i, pathID := range frontier.SelectedIDs {
		if i < len(frontier.WorkIDs) {
			workByCommand[pathID] = frontier.WorkIDs[i]
		}
	}
	for _, entry := range selected {
		if entry.guard {
			result.SelectedGuardCommandIDs = append(result.SelectedGuardCommandIDs, entry.command.ID)
		} else {
			result.SelectedCommandIDs = append(result.SelectedCommandIDs, entry.command.ID)
		}
		result.SelectedWorkIDs = append(result.SelectedWorkIDs, workByCommand[entry.command.ID])
	}
	return result
}

func validateSelectedEvidence(input Input, selected []selectedPath) ([]string, []string, error) {
	receiptDigests := make([]string, 0, len(selected))
	pathIDs := make([]string, 0, len(selected))
	receipts := map[string]Receipt{}
	for _, receipt := range input.Receipts {
		receipts[receipt.CommandID] = receipt
	}
	paths := map[string]ProvenancePath{}
	for _, path := range input.ProvenancePaths {
		paths[path.CommandID] = path
	}
	for _, entry := range selected {
		receipt, ok := receipts[entry.command.ID]
		if !ok {
			return nil, nil, failure(ReasonResourceReceipt, "selected command has no resource receipt")
		}
		if receipt.SnapshotDigest != input.Head.Digest {
			return nil, nil, failure(ReasonMismatchedDigest, "resource receipt snapshot does not match head")
		}
		resource := resourceenvelope.Evaluate(receipt.Envelope)
		if resource.Status != resourceenvelope.PASS {
			return nil, nil, resourceFailure(resource.ReasonCode)
		}
		if resource.CPUCoreNS > entry.command.CPUWorkUnits || resource.PeakRSSBytes > entry.command.MemoryBytes {
			return nil, nil, failure(ReasonResourceLimit, "resource receipt exceeds command ceiling")
		}
		path, ok := paths[entry.command.ID]
		if !ok {
			return nil, nil, failure(ReasonAmbiguousPath, "selected command has no provenance path")
		}
		if err := evaluatePath(path); err != nil {
			return nil, nil, err
		}
		receiptDigests = append(receiptDigests, digestBytes([]byte(entry.command.ID+"\x00"+receipt.SnapshotDigest+"\x00"+resource.Canonical())))
		pathIDs = append(pathIDs, path.Requirement.PathID)
	}
	return sortedUnique(receiptDigests), sortedUnique(pathIDs), nil
}

func resourceFailure(reason string) error {
	if reason == "cpu-arithmetic" {
		return failure(ReasonResourceArithmetic, reason)
	}
	return failure(ReasonResourceReceipt, reason)
}

func evaluatePath(path ProvenancePath) error {
	requirement, err := pathRequirement(path.Requirement)
	if err != nil {
		return failure(ReasonAmbiguousPath, err.Error())
	}
	if normalized, normalizeErr := path.Path.Normalized(); normalizeErr == nil {
		if topologyErr := validatePathTopology(normalized, requirement); topologyErr != nil {
			return topologyErr
		}
	}
	evaluation := pathclosure.Evaluate(path.Path, []pathclosure.Requirement{requirement})
	if evaluation.Status == pathclosure.PASS && len(evaluation.Complete) == 1 {
		return nil
	}
	if evaluation.Code == pathclosure.CodeDuplicate {
		return failure(ReasonDuplicateID, evaluation.Code)
	}
	if evaluation.Code == pathclosure.CodeMissingRecord {
		return failure(ReasonUnknownPath, evaluation.Code)
	}
	if evaluation.Code == pathclosure.CodeMissingEvidence || evaluation.Code == pathclosure.CodeMissingSnapshot {
		return failure(ReasonDanglingReference, evaluation.Code)
	}
	if strings.Contains(strings.ToLower(evaluation.Code), "malformed") {
		return failure(ReasonAmbiguousPath, evaluation.Code)
	}
	return failure(ReasonEvaluatorError, evaluation.Code)
}

func validatePathTopology(path semantic.InferencePathV1, requirement pathclosure.Requirement) error {
	byID := make(map[semantic.ID]semantic.InferenceEdge, len(path.Edges))
	for _, edge := range path.Edges {
		byID[edge.RecordID] = edge
	}
	edges := make([]semantic.InferenceEdge, 0, len(requirement.RecordIDs))
	for _, recordID := range requirement.RecordIDs {
		edge, ok := byID[recordID]
		if !ok {
			return nil
		}
		edges = append(edges, edge)
	}
	if _, err := semantic.NewInferencePathChain(edges...); err != nil {
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "cycle") || strings.Contains(message, "path_orphan") {
			return failure(ReasonCycle, err.Error())
		}
		if strings.Contains(message, "path_ambiguity") {
			return failure(ReasonAmbiguousPath, err.Error())
		}
	}
	return nil
}

func graphFailureReason(message string) string {
	message = strings.ToLower(message)
	if strings.Contains(message, "duplicate") {
		return ReasonDuplicateID
	}
	if strings.Contains(message, "endpoint") || strings.Contains(message, "registered") {
		return ReasonDanglingReference
	}
	return ReasonEvaluatorError
}

func pathRequirement(raw PathRequirement) (pathclosure.Requirement, error) {
	pathID, err := semantic.ParseIdentity(raw.PathID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	start, err := semantic.ParseIdentity(raw.StartID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	end, err := semantic.ParseIdentity(raw.EndID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	records := make([]semantic.ID, len(raw.RecordIDs))
	kinds := make([]semantic.InferenceKind, len(raw.ExpectedKinds))
	for i := range raw.RecordIDs {
		records[i], err = semantic.ParseIdentity(raw.RecordIDs[i])
		if err != nil {
			return pathclosure.Requirement{}, err
		}
		kinds[i] = semantic.InferenceKind(raw.ExpectedKinds[i])
	}
	return pathclosure.Requirement{PathID: pathID, RecordIDs: records, ExpectedKinds: kinds, StartID: start, EndID: end}, nil
}
