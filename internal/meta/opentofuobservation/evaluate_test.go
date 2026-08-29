package opentofuobservation

import (
	"strings"
	"testing"
)

func testContract() Contract {
	return Contract{Schema: ContractSchema, ID: "test-contract", Cells: Cells()}
}

func testCommand(name string) CommandReceipt {
	return CommandReceipt{Name: name, Command: []string{"tofu", name}, CwdRole: "fixture-run", ExitCode: 0,
		StdoutBytes: 1, StdoutDigest: DigestBytes([]byte(name)), StderrDigest: DigestBytes(nil), WallMS: 1, PeakRSSKiB: 1, Executed: true}
}

func testRun(index int, fixture, plan, rawPlan, events, rawEvents string) ExecutionRun {
	return ExecutionRun{Index: index, FixtureDigest: fixture, PlanJSONDigest: plan, PlanRawDigest: rawPlan,
		PlanCanonicalizer: "opentofu-plan-json/v1", PlanCanonicalizerDigest: DigestBytes([]byte("opentofu-plan-json/v1")), PlanVolatileFields: []string{"timestamp"}, PlanJSONBytes: 1,
		PlanSchemaValid: true, TestEventDigest: events, TestRawDigest: rawEvents, TestEventCount: 5,
		TestTypeCounts:         map[string]int{"version": 1, "test_abstract": 1, "test_file": 1, "test_run": 1, "test_summary": 1},
		TestAbstractDiscovered: 1, TestRunExecuted: 1, TestSummaryPassed: 1, TestSummaryFailed: 0,
		TestSummaryErrored: 0, TestSummarySkipped: 0, TestEventsValid: true,
		Commands: []CommandReceipt{testCommand("init"), testCommand("plan"), testCommand("show"), testCommand("test")}}
}

func testGraph(digest string) GraphObservation {
	graph := GraphObservation{Schema: "gooo-graph/v1", ProgramDigest: digest,
		GraphHash: strings.Repeat("a", 64), ActivityCount: 12, EdgeCount: 24}
	for _, cell := range fixedCells {
		activity := "activity-" + cell.ID
		input := "input-" + cell.ID
		output := "output-" + cell.ID
		graph.Nodes = append(graph.Nodes, GraphNode{ID: activity, Kind: "Activity", Name: cell.MetaOperation})
		graph.Nodes = append(graph.Nodes, GraphNode{ID: input, Kind: "Entity", Name: cell.ID + " input"})
		graph.Nodes = append(graph.Nodes, GraphNode{ID: output, Kind: "Entity", Name: cell.ID + " output"})
		graph.Relations = append(graph.Relations, GraphRelation{Status: "deterministic", Subject: activity, Predicate: "used", Object: input})
		graph.Relations = append(graph.Relations, GraphRelation{Status: "deterministic", Subject: output, Predicate: "wasGeneratedBy", Object: activity})
		graph.Bindings = append(graph.Bindings, GraphBinding{CellID: cell.ID, ActivityID: activity, InputID: input, OutputID: output, UsedEdgeCount: 1, GeneratedCount: 1})
	}
	return graph
}

func testObservation() Observation {
	digest := DigestBytes([]byte("digest"))
	fixture := DigestBytes([]byte("fixture"))
	plan := DigestBytes([]byte("plan-canonical"))
	versionJSON := map[string]any{"platform": "linux_amd64", "terraform_version": "1.12.6"}
	versionDigest, _ := DigestJSON(versionJSON)
	cellEvidence := make(map[string]string, len(fixedCells))
	cellProjections := make(map[string]string, len(fixedCells))
	for _, cell := range fixedCells {
		cellProjections[cell.ID] = "cell-projection:" + cell.ID
		cellEvidence[cell.ID] = DigestBytes([]byte(cellProjections[cell.ID]))
	}
	return Observation{Schema: ObservationSchema, ContractID: "test-contract", SubjectSHA: "0123456789012345678901234567890123456789", UserPaths: Paths(),
		Release: ReleaseObservation{ReleaseID: ExpectedReleaseID, AssetURL: ExpectedAssetURL, AssetSHA256: ExpectedAssetSHA, AssetBytes: ExpectedAssetSize, ChecksumsSHA256: ExpectedSumsSHA,
			VersionJSON: versionJSON, VersionJSONSHA: versionDigest, Version: "1.12.6", Platform: "linux_amd64", Command: CommandReceipt{Name: "tofu-version", Command: []string{"tofu", "version", "-json"}, CwdRole: "release", ExitCode: 0, StdoutBytes: 1, StdoutDigest: digest, StderrDigest: DigestBytes(nil), WallMS: 1, PeakRSSKiB: 1, Executed: true}},
		FixtureDigest: fixture, FixtureFiles: []string{"main.tf"}, FixturePhysicalLines: 1,
		Executions: []ExecutionRun{testRun(1, fixture, plan, DigestBytes([]byte("plan-raw-1")), DigestBytes([]byte("events")), DigestBytes([]byte("raw-1"))),
			testRun(2, fixture, plan, DigestBytes([]byte("plan-raw-2")), DigestBytes([]byte("events")), DigestBytes([]byte("raw-2")))},
		Reuse:     ReuseAccounting{RequestMode: "BASELINE", Requests: 1, Discovered: 1, Executed: 1, Reused: 0, Skipped: 0, PriorCandidates: 0, Invalidated: 0, Decision: "EXECUTE", Reason: "NO_PRIOR_RECEIPT", SourceDigest: digest, FixtureDigest: digest, ArgumentDigest: digest, EnvironmentDigest: digest, ReleaseDigest: digest, ToolchainDigest: digest, DependencyGraphDigest: digest, ExpectedResultDigest: digest, BaselinePhysicalCommandExecutions: 9, BaselinePhysicalTestExecutions: 2, ReusePhysicalCommandExecutions: 0, ReusePhysicalTestExecutions: 0, PriorReceiptsValid: 0, DecisionWallMS: 1, DecisionPeakRSSKiB: 1, RequiresExecution: true},
		Runtime:   RuntimeSummary{ConsumerBuildMS: 1, ConsumerBuildPeakRSS: 1, TofuInitMS: 1, TofuInitPeakRSS: 1, TofuPlanMS: 1, TofuPlanPeakRSS: 1, TofuShowMS: 1, TofuShowPeakRSS: 1, TofuTestMS: 1, TofuTestPeakRSS: 1, TofuTestExecutions: 2, TotalWallMS: 1, MaxPeakRSSKiB: 1},
		Inventory: Inventory{InputRegularFiles: 1, InputPhysicalLines: 1, OutputArtifactFiles: 1}, ObserverGoVersion: "go version go1.27.0 linux/amd64", ObserverGOVERSION: ExpectedGo, ObserverToolchainDigest: digest,
		CellEvidenceProjections: cellProjections, CellEvidenceDigests: cellEvidence, Graph: testGraph(digest), RepositoryWrites: 0, LocalTestExecutions: 0, ReleaseBinaryBuilds: 0, ReleaseBinaryBuildReason: "NOT_EXECUTED_RELEASE_BINARY_BOUNDARY", HumanReportReady: true}
}

func TestEvaluateExactObservationAndReport(t *testing.T) {
	contract := testContract()
	report, err := Evaluate(contract, testObservation())
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Summary.ClosedCells != 12 || report.Summary.ReplayMatches != 1 {
		t.Fatalf("report=%s cells=%d replay=%d", report.Decision, report.Summary.ClosedCells, report.Summary.ReplayMatches)
	}
	if err := ValidateReport(report, report.SubjectSHA, contract.ID); err != nil {
		t.Fatal(err)
	}
}

func TestReplayMismatchIsRefuted(t *testing.T) {
	observation := testObservation()
	observation.Executions[1].PlanJSONDigest = DigestBytes([]byte("different-plan"))
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionRefuted || report.Resolution != ResolutionExact {
		t.Fatalf("report=%s/%s", report.Decision, report.Resolution)
	}
}

func TestMissingExecutionIsTypedUnknown(t *testing.T) {
	observation := testObservation()
	observation.Executions = observation.Executions[:1]
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionUnknown || len(report.Unknowns) != 1 {
		t.Fatalf("report=%s unknowns=%d", report.Decision, len(report.Unknowns))
	}
	unknown := report.Unknowns[0]
	if unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.UnknownClass == "" || unknown.NextOperation == "" || unknown.BlockedBy == nil {
		t.Fatal("unknown coordinate is not six-field")
	}
}

func TestCacheMarkerCannotClaimReuse(t *testing.T) {
	observation := testObservation()
	observation.Reuse.Decision = "REUSED"
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionRefuted || report.Reason != "REUSE_FIRST_RUN_CLAIM_INVALID" {
		t.Fatalf("report=%s reason=%s", report.Decision, report.Reason)
	}
}

func TestGraphBindingMustUseRealNodesAndEdges(t *testing.T) {
	observation := testObservation()
	observation.Graph.Bindings[0].InputID = "virtual-input"
	if err := ValidateObservation(observation); err == nil {
		t.Fatal("virtual graph entity was accepted")
	}
}

func TestCellEvidenceDigestMustBeDistinctlyBound(t *testing.T) {
	observation := testObservation()
	observation.CellEvidenceDigests[fixedCells[1].ID] = observation.CellEvidenceDigests[fixedCells[0].ID]
	if err := ValidateObservation(observation); err == nil {
		t.Fatal("duplicate cell evidence was accepted")
	}
}

func TestPlanRawDigestMayDifferWhenCanonicalDigestMatches(t *testing.T) {
	observation := testObservation()
	if observation.Executions[0].PlanRawDigest == observation.Executions[1].PlanRawDigest || observation.Executions[0].PlanJSONDigest != observation.Executions[1].PlanJSONDigest {
		t.Fatal("raw/canonical plan distinction is absent")
	}
	if err := ValidateObservation(observation); err != nil {
		t.Fatal(err)
	}
}

func TestTestEventTypeCountsAreExact(t *testing.T) {
	observation := testObservation()
	observation.Executions[0].TestTypeCounts["test_run"] = 2
	if err := ValidateObservation(observation); err == nil {
		t.Fatal("test event type drift was accepted")
	}
}

func exactReuseObservation(t *testing.T) Observation {
	baseline := testObservation()
	priorReport, err := Evaluate(testContract(), baseline)
	if err != nil {
		t.Fatal(err)
	}
	files := []string{"SHA256SUMS", "baseline-observation.json", "baseline-report.json", "contract.json", "gooo-graph.json", "main.gooo", "plan-1.canonical.json", "plan-1.json", "plan-2.canonical.json", "plan-2.json", "test-1.ndjson", "test-2.ndjson", "version.json"}
	digests := make(map[string]string, len(files))
	for _, file := range files {
		digests[file] = DigestBytes([]byte(file))
	}
	prior := &PriorReceipt{Report: priorReport, ReceiptFileDigest: DigestBytes([]byte("baseline-report")), ArtifactFiles: files, ArtifactDigests: digests}
	prior.ArtifactManifestDigest = artifactManifestDigest(files, digests)
	observation := baseline
	observation.PriorReceipt = prior
	observation.Reuse.RequestMode = "REUSE"
	observation.Reuse.Requests = 2
	observation.Reuse.Discovered = 2
	observation.Reuse.Reused = 1
	observation.Reuse.PriorCandidates = 1
	observation.Reuse.Decision = "REUSED"
	observation.Reuse.Reason = "EXACT_PRIOR_RECEIPT"
	observation.Reuse.PriorReceiptsValid = 1
	observation.Reuse.ReusedArtifactFiles = append([]string(nil), files...)
	observation.Reuse.PriorReceiptDigest = priorReport.ReportDigest
	observation.Reuse.PriorReceiptFileDigest = prior.ReceiptFileDigest
	observation.Reuse.PriorArtifactManifestDigest = prior.ArtifactManifestDigest
	observation.Reuse.RequiresExecution = false
	return observation
}

func TestExactPriorReceiptIsReusedWithoutPhysicalExecution(t *testing.T) {
	observation := exactReuseObservation(t)
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Cells[10].Decision != DecisionPass || report.Reuse.Reused != 1 || report.Reuse.ReusePhysicalCommandExecutions != 0 || report.Reuse.ReusePhysicalTestExecutions != 0 {
		t.Fatalf("reuse report=%s cell=%s reused=%d", report.Decision, report.Cells[10].Decision, report.Reuse.Reused)
	}
	if err := ValidateReport(report, report.SubjectSHA, report.ContractID); err != nil {
		t.Fatal(err)
	}
}

func TestReuseDigestMismatchRequiresExecution(t *testing.T) {
	observation := exactReuseObservation(t)
	observation.Reuse.SourceDigest = DigestBytes([]byte("changed-source"))
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionRefuted || report.Reason != "REUSE_INPUT_DIGEST_MISMATCH" {
		t.Fatalf("digest mismatch report=%s reason=%s", report.Decision, report.Reason)
	}
}

func TestInvalidatedPriorCandidateRequiresExecution(t *testing.T) {
	observation := exactReuseObservation(t)
	observation.Reuse.Executed = 0
	observation.Reuse.Reused = 0
	observation.Reuse.Skipped = 1
	observation.Reuse.Invalidated = 1
	observation.Reuse.PriorReceiptsValid = 0
	observation.Reuse.Decision = "REQUIRES_EXECUTION"
	observation.Reuse.Reason = "REUSE_INPUT_DIGEST_MISMATCH"
	observation.Reuse.RequiresExecution = true
	observation.Reuse.SourceDigest = DigestBytes([]byte("changed-source"))
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionRefuted || report.Reason != "REUSE_INPUT_DIGEST_MISMATCH" || report.Reuse.Reused != 0 || report.Reuse.Invalidated != 1 || !report.Reuse.RequiresExecution {
		t.Fatalf("invalidated report=%s reason=%s reused=%d invalidated=%d requires=%t", report.Decision, report.Reason, report.Reuse.Reused, report.Reuse.Invalidated, report.Reuse.RequiresExecution)
	}
}

func TestMissingPriorReceiptIsTypedUnknown(t *testing.T) {
	observation := exactReuseObservation(t)
	observation.PriorReceipt = nil
	observation.Reuse.PriorReceiptDigest = ""
	observation.Reuse.PriorReceiptFileDigest = ""
	observation.Reuse.PriorArtifactManifestDigest = ""
	observation.Reuse.ReusedArtifactFiles = nil
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionUnknown || len(report.Unknowns) != 1 || report.Unknowns[0].BlockedBy == nil {
		t.Fatalf("missing prior report=%s unknowns=%d", report.Decision, len(report.Unknowns))
	}
	unknown := report.Unknowns[0]
	if unknown.Stage != "REUSE" || unknown.Step != "READ_PRIOR_RECEIPT" || unknown.Reason != "PRIOR_RECEIPT_MISSING" || unknown.UnknownClass != "DIRECT_MISSING" || unknown.NextOperation != "RESTORE_PRIOR_RECEIPT" || unknown.BlockedBy == nil || len(unknown.BlockedBy) != 0 {
		t.Fatalf("prior-missing coordinate=%+v", unknown)
	}
}

func TestMalformedPriorReceiptFailsClosed(t *testing.T) {
	observation := exactReuseObservation(t)
	observation.PriorReceipt.Report.Decision = DecisionUnknown
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionFailClosed || report.Reason != "PRIOR_RECEIPT_TOP_DECISION_INVALID" {
		t.Fatalf("malformed prior report=%s reason=%s", report.Decision, report.Reason)
	}
}

func TestPriorRefutedReceiptOutranksMissingReuseEvidence(t *testing.T) {
	observation := exactReuseObservation(t)
	observation.PriorReceipt.Report.Decision = DecisionRefuted
	observation.Reuse.PriorReceiptDigest = ""
	report, err := Evaluate(testContract(), observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionRefuted || report.Reason != "PRIOR_RECEIPT_NOT_REUSABLE" {
		t.Fatalf("refuted prior report=%s reason=%s", report.Decision, report.Reason)
	}
}
