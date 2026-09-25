package opentofuobservation

import "errors"

import "maps"

var unknownReasonCatalog = map[string]Unknown{
	"ASSET_CHECKSUM_EVIDENCE_UNAVAILABLE":   {Stage: "FOUNDATION", Step: "VERIFY_ASSET_CHECKSUM", Reason: "ASSET_CHECKSUM_EVIDENCE_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_RELEASE_CHECKSUM", BlockedBy: []string{}},
	"RELEASE_IDENTITY_EVIDENCE_UNAVAILABLE": {Stage: "FOUNDATION", Step: "VERIFY_RELEASE_IDENTITY", Reason: "RELEASE_IDENTITY_EVIDENCE_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_RELEASE_IDENTITY", BlockedBy: []string{}},
	"CLI_VERSION_JSON_UNAVAILABLE":          {Stage: "FOUNDATION", Step: "OBSERVE_CLI_VERSION", Reason: "CLI_VERSION_JSON_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "REEXECUTE_VERSION_COMMAND", BlockedBy: []string{}},
	"CLI_VERSION_COMMAND_RECEIPT_MISSING":   {Stage: "FOUNDATION", Step: "READ_VERSION_RECEIPT", Reason: "CLI_VERSION_COMMAND_RECEIPT_MISSING", UnknownClass: "DIRECT_MISSING", NextOperation: "REEXECUTE_VERSION_COMMAND", BlockedBy: []string{}},
	"OBSERVER_GO_TOOLCHAIN_UNAVAILABLE":     {Stage: "FOUNDATION", Step: "READ_OBSERVER_TOOLCHAIN", Reason: "OBSERVER_GO_TOOLCHAIN_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_OBSERVER_TOOLCHAIN", BlockedBy: []string{}},
	"FIXTURE_INPUT_UNAVAILABLE":             {Stage: "FOUNDATION", Step: "READ_FIXTURE_INPUT", Reason: "FIXTURE_INPUT_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_OPENTOFU_FIXTURE", BlockedBy: []string{}},
	"EXECUTION_RECEIPT_MISSING":             {Stage: "COHERENCE", Step: "READ_EXECUTION_RECEIPT", Reason: "EXECUTION_RECEIPT_MISSING", UnknownClass: "DIRECT_MISSING", NextOperation: "REEXECUTE_OPENTOFU_COMMAND", BlockedBy: []string{}},
	"EXECUTION_DIGEST_EVIDENCE_UNAVAILABLE": {Stage: "COHERENCE", Step: "READ_EXECUTION_DIGEST", Reason: "EXECUTION_DIGEST_EVIDENCE_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_EXECUTION_RECEIPT", BlockedBy: []string{}},
	"OPENTOFU_JSON_EVIDENCE_INCOMPLETE":     {Stage: "COHERENCE", Step: "VALIDATE_OPENTOFU_JSON", Reason: "OPENTOFU_JSON_EVIDENCE_INCOMPLETE", UnknownClass: "DIRECT_MISSING", NextOperation: "RECAPTURE_OPENTOFU_JSON", BlockedBy: []string{}},
	"COMMAND_RUNTIME_RECEIPT_MISSING":       {Stage: "COHERENCE", Step: "READ_COMMAND_RUNTIME", Reason: "COMMAND_RUNTIME_RECEIPT_MISSING", UnknownClass: "DIRECT_MISSING", NextOperation: "RECAPTURE_COMMAND_RECEIPTS", BlockedBy: []string{}},
	"REUSE_ELIGIBILITY_EVIDENCE_MISSING":    {Stage: "REGRESSION", Step: "VALIDATE_REUSE_ELIGIBILITY", Reason: "REUSE_ELIGIBILITY_EVIDENCE_MISSING", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_REUSE_DIGESTS", BlockedBy: []string{}},
	"PRIOR_RECEIPT_MISSING":                 {Stage: "REUSE", Step: "READ_PRIOR_RECEIPT", Reason: "PRIOR_RECEIPT_MISSING", UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_PRIOR_RECEIPT", BlockedBy: []string{}},
	"RUNTIME_OBSERVATION_MISSING":           {Stage: "COHERENCE", Step: "READ_RUNTIME_OBSERVATION", Reason: "RUNTIME_OBSERVATION_MISSING", UnknownClass: "DIRECT_MISSING", NextOperation: "RECAPTURE_RUNTIME_OBSERVATION", BlockedBy: []string{}},
}

func Evaluate(contract Contract, observation Observation) (Report, error) {
	if err := ValidateContract(contract); err != nil {
		return Report{}, err
	}
	report := baseReport(contract, observation)
	if err := ValidateObservation(observation); err != nil {
		var typed *ValidationError
		if errors.As(err, &typed) && typed.Decision == DecisionRefuted {
			report.GraphValidation = typed.GraphDiagnostic
			return failRefuted(report, typed.Reason)
		}
		if errors.As(err, &typed) && typed.Decision == DecisionFailClosed {
			report.GraphValidation = typed.GraphDiagnostic
			return failClosed(report, typed.Reason)
		}
		if errors.As(err, &typed) && typed.Decision == DecisionUnknown {
			context, ok := unknownReasonContext(typed.Reason)
			if !ok {
				return failClosed(report, "UNKNOWN_CAUSE_UNCATALOGED")
			}
			return failUnknown(report, context)
		}
		return failClosed(report, "UNKNOWN_CAUSE_UNCATALOGED")
	}
	report.Cells = evaluateCells(observation)
	report.Summary = summarize(report.Cells, observation)
	report.Decision, report.Resolution, report.Reason = decide(report.Cells)
	report.Counterexamples = FixedCounterexamples()
	sealed, err := sealedReportDigest(report)
	if err != nil {
		return Report{}, err
	}
	report.ReportDigest = sealed
	return report, nil
}

func baseReport(contract Contract, observation Observation) Report {
	return Report{Schema: ReportSchema, ContractID: contract.ID, SubjectSHA: observation.SubjectSHA,
		MetaOperation: MetaOperation, UserPaths: observation.UserPaths, Release: observation.Release,
		FixtureDigest: observation.FixtureDigest, FixtureFiles: observation.FixtureFiles,
		FixturePhysicalLines: observation.FixturePhysicalLines, Executions: observation.Executions,
		Reuse: observation.Reuse, Runtime: observation.Runtime, Inventory: observation.Inventory,
		CellEvidenceProjections: copyEvidenceDigests(observation.CellEvidenceProjections),
		CellEvidenceDigests:     copyEvidenceDigests(observation.CellEvidenceDigests),
		Graph:                   observation.Graph, RepositoryWrites: observation.RepositoryWrites,
		LocalTestExecutions: observation.LocalTestExecutions, ReleaseBinaryBuilds: observation.ReleaseBinaryBuilds,
		ReleaseBinaryBuildReason: observation.ReleaseBinaryBuildReason, ObserverGoVersion: observation.ObserverGoVersion,
		ObserverGOVERSION: observation.ObserverGOVERSION, ObserverToolchainDigest: observation.ObserverToolchainDigest,
		HumanReportReady: observation.HumanReportReady, PromotionAuthorized: false, PriorReceipt: observation.PriorReceipt}
}

func copyEvidenceDigests(source map[string]string) map[string]string {
	copy := make(map[string]string, len(source))
	maps.Copy(copy, source)
	return copy
}

func unknownReasonContext(reason string) (Unknown, bool) {
	context, ok := unknownReasonCatalog[reason]
	if !ok {
		return Unknown{}, false
	}
	context.BlockedBy = append([]string{}, context.BlockedBy...)
	return context, true
}

func failUnknown(report Report, context Unknown) (Report, error) {
	report.Decision, report.Resolution, report.Reason = DecisionUnknown, ResolutionLower, context.Reason
	report.Unknowns = []Unknown{context}
	report.Counterexamples = FixedCounterexamples()
	sealed, err := sealedReportDigest(report)
	if err != nil {
		return Report{}, err
	}
	report.ReportDigest = sealed
	return report, nil
}

func failRefuted(report Report, reason string) (Report, error) {
	report.Decision, report.Resolution, report.Reason = DecisionRefuted, ResolutionExact, reason
	report.Counterexamples = FixedCounterexamples()
	sealed, err := sealedReportDigest(report)
	if err != nil {
		return Report{}, err
	}
	report.ReportDigest = sealed
	return report, nil
}

func failClosed(report Report, reason string) (Report, error) {
	report.Decision, report.Resolution, report.Reason = DecisionFailClosed, ResolutionLower, reason
	report.Unknowns = []Unknown{{Stage: "OBSERVATION", Step: "VALIDATE_INPUT", Reason: reason, UnknownClass: "MALFORMED_EVIDENCE", NextOperation: "RECAPTURE_OPENTOFU_OBSERVATION", BlockedBy: []string{}}}
	report.Counterexamples = FixedCounterexamples()
	sealed, err := sealedReportDigest(report)
	if err != nil {
		return Report{}, err
	}
	report.ReportDigest = sealed
	return report, nil
}

func decide(cells []CellResult) (string, string, string) {
	for _, cell := range cells {
		if cell.Decision == DecisionRefuted {
			return DecisionRefuted, ResolutionExact, "KNOWN_CONTRADICTION"
		}
	}
	for _, cell := range cells {
		if cell.Decision == DecisionUnknown {
			return DecisionUnknown, ResolutionLower, "OBSERVATION_EVIDENCE_UNAVAILABLE"
		}
	}
	return DecisionPass, ResolutionExact, "OPENTOFU_RELEASED_CLI_OBSERVED"
}

func evaluateCells(observation Observation) []CellResult {
	planEqual := observation.Executions[0].PlanJSONDigest == observation.Executions[1].PlanJSONDigest
	testEqual := observation.Executions[0].TestEventDigest == observation.Executions[1].TestEventDigest
	values := []cellValue{
		{1, 1, true, "RELEASE_PIN_EXACT"}, {1, 1, true, "ASSET_CHECKSUM_EXACT"},
		{1, 1, true, "CLI_VERSION_JSON_EXACT"}, {1, 1, true, "FIXTURE_DIGEST_EXACT"},
		{boolInt(observation.Executions[0].PlanSchemaValid), 1, observation.Executions[0].PlanSchemaValid, "PLAN_JSON_SCHEMA_OBSERVED"},
		{boolInt(observation.Executions[0].TestEventsValid), 1, observation.Executions[0].TestEventsValid, "TEST_EVENT_INVENTORY_OBSERVED"},
		{runtimeReady(observation), 1, runtimeReady(observation) == 1, "COMMAND_RUNTIME_OBSERVED"},
		{positiveRSS(observation), 1, positiveRSS(observation) == 1, "PEAK_RSS_OBSERVED"},
		{boolInt(planEqual), 1, planEqual, "PLAN_REPLAY_EQUAL"},
		{boolInt(testEqual), 1, testEqual, "TEST_REPLAY_EQUAL"},
		{boolInt(reuseReady(observation)), 1, reuseReady(observation), "FIRST_RUN_NOT_REUSED"},
		{boolInt(observation.HumanReportReady), 1, observation.HumanReportReady, "HUMAN_REPORT_READY"},
	}
	result := make([]CellResult, 0, len(fixedCells))
	for index, spec := range fixedCells {
		value := values[index]
		decision, state := DecisionPass, "CLOSED"
		if !value.ok {
			decision, state = DecisionRefuted, "REFUTED"
		}
		result = append(result, CellResult{ID: spec.ID, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
			Indicator: spec.Indicator, Producer: "opentofuobservation.Evaluate", Consumer: "opentofu-released-cli-verifier",
			Decision: decision, State: state, Observed: value.observed, Expected: value.expected, Reason: value.reason,
			EvidenceDigest: observation.CellEvidenceDigests[spec.ID]})
	}
	return result
}

type cellValue struct {
	observed, expected int
	ok                 bool
	reason             string
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func runtimeReady(observation Observation) int {
	if err := validateRuntime(observation); err != nil {
		return 0
	}
	return 1
}

func positiveRSS(observation Observation) int {
	return boolInt(observation.Runtime.MaxPeakRSSKiB > 0)
}

func reuseReady(observation Observation) bool {
	return validateReuse(observation) == nil
}

func summarize(cells []CellResult, observation Observation) Summary {
	summary := Summary{CellsTotal: len(cells), ThreePaths: len(observation.UserPaths), Executions: len(observation.Executions), RepositoryWrites: observation.RepositoryWrites, LocalTests: observation.LocalTestExecutions}
	for _, cell := range cells {
		if cell.State == "CLOSED" {
			summary.ClosedCells++
			switch cell.ProofChoice {
			case "FOUNDATION":
				summary.FoundationClosed++
			case "COHERENCE":
				summary.CoherenceClosed++
			case "REGRESSION":
				summary.RegressionClosed++
			}
		}
		if cell.Decision == DecisionUnknown {
			summary.UnknownCells++
		}
		if cell.Decision == DecisionRefuted {
			summary.RefutedCells++
		}
	}
	summary.OpenCells = summary.CellsTotal - summary.ClosedCells - summary.UnknownCells - summary.RefutedCells
	if len(observation.Executions) == 2 && observation.Executions[0].PlanJSONDigest == observation.Executions[1].PlanJSONDigest && observation.Executions[0].TestEventDigest == observation.Executions[1].TestEventDigest {
		summary.ReplayMatches = 1
	}
	return summary
}
