package opentofuobservation

import "testing"

func testContract() Contract {
	return Contract{Schema: ContractSchema, ID: "test-contract", Cells: Cells()}
}

func testCommand(name string) CommandReceipt {
	return CommandReceipt{Name: name, Command: []string{"tofu", name}, CwdRole: "fixture-run", ExitCode: 0,
		StdoutBytes: 1, StdoutDigest: DigestBytes([]byte(name)), StderrDigest: DigestBytes(nil), WallMS: 1, PeakRSSKiB: 1, Executed: true}
}

func testObservation() Observation {
	commands := []CommandReceipt{testCommand("init"), testCommand("plan"), testCommand("show"), testCommand("test")}
	run := ExecutionRun{Index: 1, FixtureDigest: DigestBytes([]byte("fixture")), PlanJSONDigest: DigestBytes([]byte("plan")), PlanJSONBytes: 1, PlanSchemaValid: true, TestEventDigest: DigestBytes([]byte("events")), TestRawDigest: DigestBytes([]byte("raw")), TestEventCount: 1, TestEventsValid: true, Commands: commands}
	run2 := run
	run2.Index = 2
	bindings := make([]GraphBinding, 0, len(fixedCells))
	for _, cell := range fixedCells { bindings = append(bindings, GraphBinding{CellID: cell.ID, ActivityID: cell.MetaOperation, InputID: cell.ID + "-input", OutputID: cell.ID + "-output", UsedEdgeCount: 1, GeneratedCount: 1}) }
	digest := DigestBytes([]byte("digest"))
	return Observation{Schema: ObservationSchema, ContractID: "test-contract", SubjectSHA: "0123456789012345678901234567890123456789", UserPaths: Paths(),
		Release: ReleaseObservation{ReleaseID: ExpectedReleaseID, AssetURL: ExpectedAssetURL, AssetSHA256: ExpectedAssetSHA, AssetBytes: ExpectedAssetSize, ChecksumsSHA256: ExpectedSumsSHA, Version: "1.12.6", Platform: "linux_amd64"},
		FixtureDigest: DigestBytes([]byte("fixture")), FixtureFiles: []string{"main.tf"}, FixturePhysicalLines: 1, Executions: []ExecutionRun{run, run2},
		Reuse: ReuseAccounting{Discovered: 1, Executed: 1, Reused: 0, Skipped: 0, PriorCandidates: 0, Invalidated: 0, Decision: "NOT_REUSED_FIRST_RUN", Reason: "NO_PRIOR_RECEIPT", SourceDigest: digest, FixtureDigest: digest, ArgumentDigest: digest, EnvironmentDigest: digest, ReleaseDigest: digest, ToolchainDigest: digest, DependencyGraphDigest: digest, ExpectedResultDigest: digest},
		Runtime: RuntimeSummary{ConsumerBuildMS: 1, ConsumerBuildPeakRSS: 1, TofuInitMS: 1, TofuInitPeakRSS: 1, TofuPlanMS: 1, TofuPlanPeakRSS: 1, TofuShowMS: 1, TofuShowPeakRSS: 1, TofuTestMS: 1, TofuTestPeakRSS: 1, TotalWallMS: 1, MaxPeakRSSKiB: 1},
		Inventory: Inventory{InputRegularFiles: 1, InputPhysicalLines: 1, OutputArtifactFiles: 1}, ObserverGoVersion: "go version go1.27.0 linux/amd64", ObserverGOVERSION: ExpectedGo, ObserverToolchainDigest: digest,
		Graph: GraphObservation{Schema: "gooo/opentofu-observation-graph/v1", ProgramDigest: digest, GraphHash: digest, ActivityCount: 12, EdgeCount: 24, Bindings: bindings}, RepositoryWrites: 0, LocalTestExecutions: 0, HumanReportReady: true}
}

func TestEvaluateExactObservationAndReport(t *testing.T) {
	contract := testContract()
	report, err := Evaluate(contract, testObservation())
	if err != nil { t.Fatal(err) }
	if report.Decision != DecisionPass || report.Summary.ClosedCells != 12 || report.Summary.ReplayMatches != 1 { t.Fatalf("report=%s cells=%d replay=%d", report.Decision, report.Summary.ClosedCells, report.Summary.ReplayMatches) }
	if err := ValidateReport(report, report.SubjectSHA, contract.ID); err != nil { t.Fatal(err) }
}

func TestReplayMismatchIsRefuted(t *testing.T) {
	observation := testObservation()
	observation.Executions[1].PlanJSONDigest = DigestBytes([]byte("different-plan"))
	report, err := Evaluate(testContract(), observation)
	if err != nil { t.Fatal(err) }
	if report.Decision != DecisionRefuted || report.Resolution != ResolutionExact { t.Fatalf("report=%s/%s", report.Decision, report.Resolution) }
}

func TestMissingExecutionIsTypedUnknown(t *testing.T) {
	observation := testObservation()
	observation.Executions = observation.Executions[:1]
	report, err := Evaluate(testContract(), observation)
	if err != nil { t.Fatal(err) }
	if report.Decision != DecisionUnknown || len(report.Unknowns) != 1 { t.Fatalf("report=%s unknowns=%d", report.Decision, len(report.Unknowns)) }
	unknown := report.Unknowns[0]
	if unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.UnknownClass == "" || unknown.NextOperation == "" || unknown.BlockedBy == nil { t.Fatal("unknown coordinate is not six-field") }
}

func TestCacheMarkerCannotClaimReuse(t *testing.T) {
	observation := testObservation()
	observation.Reuse.Decision = "REUSED"
	report, err := Evaluate(testContract(), observation)
	if err != nil { t.Fatal(err) }
	if report.Decision != DecisionRefuted || report.Reason != "REUSE_FIRST_RUN_CLAIM_INVALID" { t.Fatalf("report=%s reason=%s", report.Decision, report.Reason) }
}
