package opentofuobservation

import (
	"sort"
	"strings"
)

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
	if observation.Release.Command.ExitCode != 0 || !observation.Release.Command.Executed || observation.Release.Command.Phase == "" || observation.Release.Command.StdoutBytes <= 0 {
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
	if observation.Inventory.InputRegularFiles <= 0 || observation.Inventory.InputPhysicalLines <= 0 || observation.Inventory.OutputArtifactFiles <= 0 || observation.Inventory.ReusableArtifactFiles < 0 {
		return malformed("ARTIFACT_INVENTORY_MALFORMED")
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
		if !command.Executed || command.ExitCode != 0 || command.WallMS <= 0 || command.PeakRSSKiB <= 0 || command.CwdRole == "" || command.Phase == "" || len(command.Command) == 0 {
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
	if !physicalReuseAccountingMatches(reuse, observation) {
		return refuted("REUSE_PHYSICAL_ACCOUNTING_MISMATCH")
	}
	if reuse.RequestMode == "REUSE" && reuse.Invalidated == 1 {
		if reuse.Discovered != reuse.Executed+reuse.Reused+reuse.Skipped+reuse.Invalidated || reuse.PriorCandidates != reuse.Reused+reuse.Invalidated {
			return refuted("REUSE_ACCOUNTING_CONTRADICTION")
		}
		return validateInvalidatedReuse(reuse, observation)
	}
	if reuse.Discovered != reuse.Executed+reuse.Reused+reuse.Skipped || reuse.PriorCandidates != reuse.Reused+reuse.Invalidated {
		return refuted("REUSE_ACCOUNTING_CONTRADICTION")
	}
	for _, digest := range reuseDigests(reuse) {
		if !validDigest(digest) {
			return unavailable("REUSE_ELIGIBILITY_EVIDENCE_MISSING")
		}
	}
	switch reuse.RequestMode {
	case "BASELINE":
		return validateBaselineReuse(reuse, observation.PriorReceipt)
	case "REUSE":
		return validateReuseRequest(reuse, observation)
	default:
		return malformed("REUSE_REQUEST_MODE_MALFORMED")
	}
}

func reuseDigests(reuse ReuseAccounting) []string {
	return []string{reuse.SourceDigest, reuse.FixtureDigest, reuse.ArgumentDigest, reuse.EnvironmentDigest,
		reuse.ReleaseDigest, reuse.ToolchainDigest, reuse.DependencyGraphDigest, reuse.ExpectedResultDigest}
}

func physicalReuseAccountingMatches(reuse ReuseAccounting, observation Observation) bool {
	baselineCommands, baselineTests := countPhysicalOpenTofuCommands(observation)
	reuseCommands, reuseTests := countPhysicalReuseCommands(observation)
	return reuse.BaselinePhysicalCommandExecutions == baselineCommands &&
		reuse.BaselinePhysicalTestExecutions == baselineTests &&
		reuse.ReusePhysicalCommandExecutions == reuseCommands &&
		reuse.ReusePhysicalTestExecutions == reuseTests
}

func countPhysicalOpenTofuCommands(observation Observation) (commands, tests int) {
	if observation.Release.Command.Executed && strings.HasPrefix(observation.Release.Command.Name, "tofu-") && !strings.HasPrefix(observation.Release.Command.Phase, "reuse") {
		commands++
	}
	for _, run := range observation.Executions {
		for _, command := range run.Commands {
			if command.Executed && strings.HasPrefix(command.Name, "tofu-") && !strings.HasPrefix(command.Phase, "reuse") {
				commands++
				if command.Name == "tofu-test" {
					tests++
				}
			}
		}
	}
	return commands, tests
}

func countPhysicalReuseCommands(observation Observation) (commands, tests int) {
	for _, run := range observation.Executions {
		for _, command := range run.Commands {
			if command.Executed && strings.HasPrefix(command.Name, "tofu-") && strings.HasPrefix(command.Phase, "reuse") {
				commands++
				if command.Name == "tofu-test" {
					tests++
				}
			}
		}
	}
	return commands, tests
}

func validateBaselineReuse(reuse ReuseAccounting, prior *PriorReceipt) error {
	if prior != nil || reuse.Requests != 1 || reuse.Discovered != 1 || reuse.Executed != 1 || reuse.Reused != 0 || reuse.Skipped != 0 || reuse.PriorCandidates != 0 || reuse.Invalidated != 0 || reuse.Decision != "EXECUTE" || reuse.Reason != "NO_PRIOR_RECEIPT" || !reuse.RequiresExecution || reuse.PriorReceiptsValid != 0 || len(reuse.ReusedArtifactFiles) != 0 || reuse.PriorReceiptDigest != "" || reuse.PriorReceiptFileDigest != "" || reuse.PriorArtifactManifestDigest != "" {
		return refuted("REUSE_FIRST_RUN_CLAIM_INVALID")
	}
	if reuse.BaselinePhysicalCommandExecutions <= 0 || reuse.BaselinePhysicalTestExecutions <= 0 || reuse.ReusePhysicalCommandExecutions != 0 || reuse.ReusePhysicalTestExecutions != 0 || reuse.DecisionWallMS <= 0 || reuse.DecisionPeakRSSKiB <= 0 {
		return malformed("REUSE_BASELINE_ACCOUNTING_MALFORMED")
	}
	return nil
}

func validateReuseRequest(reuse ReuseAccounting, observation Observation) error {
	if reuse.Invalidated == 1 {
		return validateInvalidatedReuse(reuse, observation)
	}
	if reuse.Requests != 2 || reuse.Discovered != 2 || reuse.Executed != 1 || reuse.Reused != 1 || reuse.Skipped != 0 || reuse.PriorCandidates != 1 || reuse.Invalidated != 0 || reuse.Decision != "REUSED" || reuse.Reason != "EXACT_PRIOR_RECEIPT" || reuse.RequiresExecution || reuse.PriorReceiptsValid != 1 || reuse.DecisionWallMS <= 0 || reuse.DecisionPeakRSSKiB <= 0 {
		return refuted("REUSE_ACCOUNTING_CONTRADICTION")
	}
	if !physicalReuseAccountingMatches(reuse, observation) {
		return refuted("REUSE_PHYSICAL_ACCOUNTING_MISMATCH")
	}
	if observation.PriorReceipt == nil {
		if reuse.PriorReceiptDigest != "" || reuse.PriorReceiptFileDigest != "" || reuse.PriorArtifactManifestDigest != "" || len(reuse.ReusedArtifactFiles) > 0 || observation.Inventory.ReusableArtifactFiles != 0 {
			return refuted("CACHE_MARKER_WITHOUT_EXACT_DIGESTS")
		}
		return unavailable("PRIOR_RECEIPT_MISSING")
	}
	if err := validatePriorReceipt(observation.SubjectSHA, observation.ContractID, observation.PriorReceipt); err != nil {
		return err
	}
	prior := observation.PriorReceipt
	if observation.Inventory.ReusableArtifactFiles != len(prior.ArtifactFiles) {
		return refuted("REUSABLE_ARTIFACT_COUNT_MISMATCH")
	}
	if reuse.PriorReceiptDigest != prior.Report.ReportDigest || reuse.PriorReceiptFileDigest != prior.ReceiptFileDigest || reuse.PriorArtifactManifestDigest != prior.ArtifactManifestDigest || !sameStrings(reuse.ReusedArtifactFiles, prior.ArtifactFiles) {
		return refuted("REUSE_INPUT_DIGEST_MISMATCH")
	}
	if !sameReuseDigests(reuse, prior.Report.Reuse) {
		return refuted("REUSE_INPUT_DIGEST_MISMATCH")
	}
	return nil
}

func validateInvalidatedReuse(reuse ReuseAccounting, observation Observation) error {
	if reuse.Requests != 2 || reuse.Discovered != 2 || reuse.Executed != 0 || reuse.Reused != 0 || reuse.Skipped != 1 || reuse.PriorCandidates != 1 || reuse.Invalidated != 1 || reuse.Decision != "REQUIRES_EXECUTION" || reuse.Reason != "REUSE_INPUT_DIGEST_MISMATCH" || !reuse.RequiresExecution || reuse.PriorReceiptsValid != 0 || reuse.DecisionWallMS <= 0 || reuse.DecisionPeakRSSKiB <= 0 || observation.PriorReceipt == nil {
		return refuted("REUSE_ACCOUNTING_CONTRADICTION")
	}
	if err := validatePriorReceipt(observation.SubjectSHA, observation.ContractID, observation.PriorReceipt); err != nil {
		return err
	}
	if sameReuseDigests(reuse, observation.PriorReceipt.Report.Reuse) {
		return refuted("REUSE_ELIGIBILITY_DIGEST_MISMATCH")
	}
	return refuted("REUSE_INPUT_DIGEST_MISMATCH")
}

func validatePriorReceipt(subject, contract string, prior *PriorReceipt) error {
	if prior.Report.Decision == DecisionRefuted || prior.Report.Decision == DecisionFailClosed {
		return refuted("PRIOR_RECEIPT_NOT_REUSABLE")
	}
	if prior.Report.Decision != DecisionPass || prior.Report.Resolution != ResolutionExact {
		return malformed("PRIOR_RECEIPT_TOP_DECISION_INVALID")
	}
	if err := ValidateReport(prior.Report, subject, contract); err != nil {
		return malformed("PRIOR_RECEIPT_MALFORMED")
	}
	if prior.Report.Summary.ClosedCells != len(fixedCells) || prior.Report.Summary.ReplayMatches != 1 || prior.Report.Release.AssetSHA256 != ExpectedAssetSHA || prior.Report.Release.ChecksumsSHA256 != ExpectedSumsSHA || prior.Report.RepositoryWrites != 0 || !allTestsPassed(prior.Report.Executions) {
		return refuted("PRIOR_RECEIPT_RESULT_NOT_REUSABLE")
	}
	if !validDigest(prior.ReceiptFileDigest) || !validArtifactManifest(prior.ArtifactFiles, prior.ArtifactDigests, prior.ArtifactManifestDigest) {
		return malformed("PRIOR_RECEIPT_ARTIFACT_MANIFEST_MALFORMED")
	}
	if prior.Report.Reuse.RequestMode != "BASELINE" || prior.Report.Reuse.Decision != "EXECUTE" {
		return refuted("PRIOR_RECEIPT_RESULT_NOT_REUSABLE")
	}
	return nil
}

func allTestsPassed(runs []ExecutionRun) bool {
	for _, run := range runs {
		if !run.TestEventsValid || run.TestSummaryPassed != 1 || run.TestSummaryFailed != 0 || run.TestSummaryErrored != 0 || run.TestSummarySkipped != 0 {
			return false
		}
	}
	return len(runs) == 2
}

func validArtifactManifest(files []string, digests map[string]string, manifest string) bool {
	if len(files) != 13 || len(digests) != len(files) || !sort.StringsAreSorted(files) || !validDigest(manifest) {
		return false
	}
	for index, file := range files {
		if file == "" || strings.HasPrefix(file, "/") || strings.Contains(file, "..") || (index > 0 && files[index-1] == file) || !validDigest(digests[file]) {
			return false
		}
	}
	return artifactManifestDigest(files, digests) == manifest
}

func artifactManifestDigest(files []string, digests map[string]string) string {
	var builder strings.Builder
	for _, file := range files {
		builder.WriteString(file)
		builder.WriteByte('\t')
		builder.WriteString(digests[file])
		builder.WriteByte('\n')
	}
	return DigestBytes([]byte(builder.String()))
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameReuseDigests(left, right ReuseAccounting) bool {
	leftDigests, rightDigests := reuseDigests(left), reuseDigests(right)
	for index := range leftDigests {
		if leftDigests[index] != rightDigests[index] {
			return false
		}
	}
	return true
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
