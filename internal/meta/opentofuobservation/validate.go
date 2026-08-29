package opentofuobservation

import "sort"

type ValidationError struct {
	Decision        string
	Resolution      string
	Reason          string
	GraphDiagnostic *GraphValidationDiagnostic
}

func (err *ValidationError) Error() string { return err.Reason }

func refuted(reason string) error {
	return &ValidationError{Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: reason}
}

func unavailable(reason string) error {
	return &ValidationError{Decision: DecisionUnknown, Resolution: ResolutionLower, Reason: reason}
}

func malformed(reason string) error {
	return &ValidationError{Decision: DecisionFailClosed, Resolution: ResolutionLower, Reason: reason}
}

const graphSchema = "gooo-graph/v1"

func ValidateObservation(observation Observation) error {
	if observation.Schema != ObservationSchema || observation.ContractID == "" || len(observation.SubjectSHA) != 40 {
		return malformed("OBSERVATION_SCHEMA_OR_IDENTITY_MALFORMED")
	}
	if len(observation.UserPaths) != len(fixedPaths) {
		return malformed("USER_PATH_DENOMINATOR_INVALID")
	}
	for index, path := range fixedPaths {
		if observation.UserPaths[index] != path {
			return malformed("USER_PATH_IDENTITY_INVALID")
		}
	}
	if observation.Release.AssetURL == "" || observation.Release.AssetSHA256 == "" || observation.Release.ChecksumsSHA256 == "" {
		return unavailable("ASSET_CHECKSUM_EVIDENCE_UNAVAILABLE")
	}
	if observation.Release.AssetURL != ExpectedAssetURL || observation.Release.AssetSHA256 != ExpectedAssetSHA || observation.Release.ChecksumsSHA256 != ExpectedSumsSHA {
		return refuted("ASSET_CHECKSUM_MISMATCH")
	}
	if observation.Release.AssetBytes == 0 || observation.Release.ReleaseID == "" {
		return unavailable("RELEASE_IDENTITY_EVIDENCE_UNAVAILABLE")
	}
	if observation.Release.AssetBytes != ExpectedAssetSize || observation.Release.ReleaseID != ExpectedReleaseID {
		return refuted("RELEASE_IDENTITY_MISMATCH")
	}
	if observation.Release.Version == "" || observation.Release.Platform == "" || observation.Release.VersionJSON == nil || !validDigest(observation.Release.VersionJSONSHA) {
		return unavailable("CLI_VERSION_JSON_UNAVAILABLE")
	}
	if observation.Release.Version != "1.12.6" || observation.Release.Platform != "linux_amd64" {
		return refuted("CLI_VERSION_JSON_MISMATCH")
	}
	if observation.Release.Command.ExitCode != 0 || !observation.Release.Command.Executed || observation.Release.Command.StdoutBytes <= 0 {
		return unavailable("CLI_VERSION_COMMAND_RECEIPT_MISSING")
	}
	if observation.ObserverGoVersion == "" || observation.ObserverGOVERSION == "" {
		return unavailable("OBSERVER_GO_TOOLCHAIN_UNAVAILABLE")
	}
	if observation.ObserverGOVERSION != ExpectedGo {
		return refuted("OBSERVER_GO_TOOLCHAIN_MISMATCH")
	}
	if len(observation.FixtureFiles) == 0 || !validDigest(observation.FixtureDigest) || observation.FixturePhysicalLines <= 0 {
		return unavailable("FIXTURE_INPUT_UNAVAILABLE")
	}
	if observation.RepositoryWrites != 0 || observation.LocalTestExecutions != 0 {
		return refuted("READ_ONLY_OR_LOCAL_TEST_BOUNDARY_VIOLATED")
	}
	if len(observation.Executions) != 2 {
		return unavailable("EXECUTION_RECEIPT_MISSING")
	}
	if err := validateExecutions(observation.Executions, observation.FixtureDigest); err != nil {
		return err
	}
	if err := validateReuse(observation); err != nil {
		return err
	}
	if !validCellEvidenceDigests(observation.CellEvidenceProjections, observation.CellEvidenceDigests) {
		return malformed("CELL_EVIDENCE_DIGEST_BINDING_INVALID")
	}
	if err := validateGraph(observation.Graph); err != nil {
		return err
	}
	return validateRuntime(observation)
}

func validateExecutions(runs []ExecutionRun, fixtureDigest string) error {
	seen := map[int]bool{}
	for _, run := range runs {
		if run.Index < 1 || run.Index > 2 || seen[run.Index] {
			return malformed("EXECUTION_RECEIPT_MALFORMED")
		}
		seen[run.Index] = true
		if run.FixtureDigest == "" || run.PlanJSONDigest == "" || run.TestEventDigest == "" {
			return unavailable("EXECUTION_DIGEST_EVIDENCE_UNAVAILABLE")
		}
		if run.FixtureDigest != fixtureDigest {
			return refuted("FIXTURE_EXECUTION_BINDING_MISMATCH")
		}
		if !validDigest(run.PlanJSONDigest) || !validDigest(run.TestEventDigest) {
			return malformed("EXECUTION_DIGEST_MALFORMED")
		}
		if run.PlanJSONBytes <= 0 || !run.PlanSchemaValid || !run.TestEventsValid || !validTestInventory(run) {
			return unavailable("OPENTOFU_JSON_EVIDENCE_INCOMPLETE")
		}
		if len(run.Commands) != 4 || !validCommands(run.Commands) {
			return unavailable("COMMAND_RUNTIME_RECEIPT_MISSING")
		}
	}
	if !seen[1] || !seen[2] {
		return malformed("EXECUTION_INDEX_SET_MISMATCH")
	}
	return nil
}

func validTestInventory(run ExecutionRun) bool {
	if run.TestEventCount != 5 || len(run.TestTypeCounts) != 5 || run.TestAbstractDiscovered != 1 || run.TestRunExecuted != 1 ||
		run.TestSummaryPassed != 1 || run.TestSummaryFailed != 0 || run.TestSummaryErrored != 0 || run.TestSummarySkipped != 0 {
		return false
	}
	for _, eventType := range []string{"version", "test_abstract", "test_file", "test_run", "test_summary"} {
		if run.TestTypeCounts[eventType] != 1 {
			return false
		}
	}
	return true
}

func validCommands(commands []CommandReceipt) bool {
	for _, command := range commands {
		if !command.Executed || command.ExitCode != 0 || command.WallMS <= 0 || command.PeakRSSKiB <= 0 || command.CwdRole == "" || len(command.Command) == 0 {
			return false
		}
		if !validDigest(command.StdoutDigest) || !validDigest(command.StderrDigest) || command.StdoutBytes < 0 || command.StderrBytes < 0 {
			return false
		}
		for _, argument := range command.Command {
			if argument == "" || argument[0] == '/' {
				return false
			}
		}
	}
	return true
}

func validateReuse(observation Observation) error {
	reuse := observation.Reuse
	if reuse.Discovered != reuse.Executed+reuse.Reused+reuse.Skipped || reuse.PriorCandidates != reuse.Reused+reuse.Invalidated {
		return refuted("REUSE_ACCOUNTING_CONTRADICTION")
	}
	if reuse.Decision != "NOT_REUSED_FIRST_RUN" || reuse.Reason != "NO_PRIOR_RECEIPT" || reuse.Reused != 0 || reuse.PriorReceiptDigest != "" {
		return refuted("REUSE_FIRST_RUN_CLAIM_INVALID")
	}
	for _, digest := range []string{reuse.SourceDigest, reuse.FixtureDigest, reuse.ArgumentDigest, reuse.EnvironmentDigest, reuse.ReleaseDigest, reuse.ToolchainDigest, reuse.DependencyGraphDigest, reuse.ExpectedResultDigest} {
		if !validDigest(digest) {
			return unavailable("REUSE_ELIGIBILITY_EVIDENCE_MISSING")
		}
	}
	return nil
}

func validateGraph(graph GraphObservation) error {
	diagnostic := graphValidationDiagnostic(graph)
	if !validGraphDenominator(graph, diagnostic) {
		return graphValidationError(DecisionFailClosed, ResolutionLower, "GOOO_GRAPH_DENOMINATOR_INVALID", diagnostic)
	}
	nodes, duplicate := graphNodes(graph.Nodes)
	if duplicate != "" {
		diagnostic.FirstDuplicate = duplicate
		return graphValidationError(DecisionRefuted, ResolutionExact, "GOOO_GRAPH_EDGE_BINDING_CONTRADICTION", diagnostic)
	}
	seenCells, seenActivities := map[string]bool{}, map[string]bool{}
	for _, binding := range graph.Bindings {
		if seenCells[binding.CellID] || seenActivities[binding.ActivityID] {
			diagnostic.FirstDuplicate = binding.CellID
			return graphValidationError(DecisionRefuted, ResolutionExact, "GOOO_GRAPH_EDGE_BINDING_CONTRADICTION", diagnostic)
		}
		seenCells[binding.CellID], seenActivities[binding.ActivityID] = true, true
		missingNode, missingEdge := graphBindingGaps(graph, nodes, binding)
		if missingNode != "" || missingEdge != "" {
			diagnostic.FirstMissingNode, diagnostic.FirstMissingEdge = missingNode, missingEdge
			return graphValidationError(DecisionRefuted, ResolutionExact, "GOOO_GRAPH_EDGE_BINDING_CONTRADICTION", diagnostic)
		}
	}
	return nil
}

func graphValidationDiagnostic(graph GraphObservation) *GraphValidationDiagnostic {
	activityCount := 0
	for _, node := range graph.Nodes {
		if node.Kind == "Activity" {
			activityCount++
		}
	}
	return &GraphValidationDiagnostic{SchemaObserved: graph.Schema, SchemaExpected: graphSchema,
		NodeCountObserved: len(graph.Nodes), NodeCountExpected: 36, ActivityCountObserved: activityCount, ActivityCountExpected: 12,
		RelationCountObserved: len(graph.Relations), RelationCountExpected: 24, BindingCountObserved: len(graph.Bindings), BindingCountExpected: 12,
		DeclaredActivityCountObserved: graph.ActivityCount, DeclaredEdgeCountObserved: graph.EdgeCount}
}

func validGraphDenominator(graph GraphObservation, diagnostic *GraphValidationDiagnostic) bool {
	return graph.Schema == graphSchema && graph.ActivityCount == diagnostic.ActivityCountExpected && graph.EdgeCount == diagnostic.RelationCountExpected &&
		len(graph.Nodes) == diagnostic.NodeCountExpected && diagnostic.ActivityCountObserved == diagnostic.ActivityCountExpected &&
		len(graph.Relations) == diagnostic.RelationCountExpected && len(graph.Bindings) == diagnostic.BindingCountExpected &&
		validDigest(graph.ProgramDigest) && graph.GraphHash != ""
}

func graphNodes(nodes []GraphNode) (map[string]GraphNode, string) {
	byID := make(map[string]GraphNode, len(nodes))
	for _, node := range nodes {
		if node.ID == "" {
			return byID, "empty-node-id"
		}
		if _, exists := byID[node.ID]; exists {
			return byID, node.ID
		}
		byID[node.ID] = node
	}
	return byID, ""
}

func graphBindingGaps(graph GraphObservation, nodes map[string]GraphNode, binding GraphBinding) (string, string) {
	for _, id := range []string{binding.ActivityID, binding.InputID, binding.OutputID} {
		if id == "" {
			return "empty-binding-node", ""
		}
		if _, ok := nodes[id]; !ok {
			return id, ""
		}
	}
	used := graphRelationCount(graph.Relations, binding.ActivityID, "used", binding.InputID)
	generated := graphRelationCount(graph.Relations, binding.OutputID, "wasGeneratedBy", binding.ActivityID)
	if binding.UsedEdgeCount != 1 || used != 1 {
		return "", "used:" + binding.ActivityID + "->" + binding.InputID
	}
	if binding.GeneratedCount != 1 || generated != 1 {
		return "", "wasGeneratedBy:" + binding.OutputID + "->" + binding.ActivityID
	}
	return "", ""
}

func graphRelationCount(relations []GraphRelation, subject, predicate, object string) int {
	count := 0
	for _, relation := range relations {
		if relation.Status == "deterministic" && relation.Subject == subject && relation.Predicate == predicate && relation.Object == object {
			count++
		}
	}
	return count
}

func graphValidationError(decision, resolution, reason string, diagnostic *GraphValidationDiagnostic) error {
	return &ValidationError{Decision: decision, Resolution: resolution, Reason: reason, GraphDiagnostic: diagnostic}
}

func validateRuntime(observation Observation) error {
	values := []int{observation.Runtime.ConsumerBuildMS, observation.Runtime.ConsumerBuildPeakRSS, observation.Runtime.TofuInitMS, observation.Runtime.TofuInitPeakRSS, observation.Runtime.TofuPlanMS, observation.Runtime.TofuPlanPeakRSS, observation.Runtime.TofuShowMS, observation.Runtime.TofuShowPeakRSS, observation.Runtime.TofuTestMS, observation.Runtime.TofuTestPeakRSS, observation.Runtime.TotalWallMS, observation.Runtime.MaxPeakRSSKiB}
	for _, value := range values {
		if value <= 0 {
			return unavailable("RUNTIME_OBSERVATION_MISSING")
		}
	}
	return nil
}

func sortedCellIDs() []string {
	ids := make([]string, 0, len(fixedCells))
	for _, cell := range fixedCells {
		ids = append(ids, cell.ID)
	}
	sort.Strings(ids)
	return ids
}
