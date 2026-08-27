package experimentpromotion

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	producerID      = "gooo-experiment-promotion-producer/v1"
	consumerPackage = "github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotionverify"
	producerPackage = "github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion"
)

var (
	headSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	runURLPattern  = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/([0-9]+)$`)
	jobURLPattern  = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/([0-9]+)/jobs/[0-9]+$`)
)

// Evaluate derives every state from the source projection and observations.
// The contract is checked as an expectation after source reconstruction; it
// is never used as the source of the portfolio or gate members.
func Evaluate(input Input) Report {
	contract := input.Contract
	report := Report{
		Schema:             ReportSchema,
		Scope:              PortfolioScope,
		SubjectSHA:         input.SubjectSHA,
		ObservationDigest:  DigestBytes(input.ObservationRaw),
		AggregateMetrics:   []string{},
		NotClaimed:         append([]string(nil), contract.NotClaimed...),
		RepositoryWrites:   input.RepositorySnapshot.ChangedPaths,
		MutationAuthority:  false,
		RepositorySnapshot: input.RepositorySnapshot,
	}

	projection, sourceErr := parseSource(input.SourceRaw)
	report.SourceProjection = projection
	ids, gates := experimentIDs(), append([]string(nil), GateIDs...)
	if sourceErr == nil {
		ids, gates = projection.Experiments, projection.Gates
	}
	bundle, bundleErr := DecodeObservation(input.ObservationRaw)
	receipts, duplicates := observationIndex(bundle, bundleErr)

	for _, experimentID := range ids {
		experiment := ExperimentResult{ExperimentID: experimentID, Gates: make([]GateResult, 0, len(gates))}
		for _, gateID := range gates {
			result := evaluateGate(experimentID, gateID, receipts[receiptKey(experimentID, gateID)], duplicates[receiptKey(experimentID, gateID)], projection, sourceErr, bundleErr)
			experiment.Gates = append(experiment.Gates, result)
		}
		experiment.Status, experiment.Stage, experiment.Step, experiment.Reason = summarizeExperiment(experiment.Gates)
		report.Experiments = append(report.Experiments, experiment)
	}

	report.ClaimLedger = sealLedger(report.Experiments)
	report.EmittedClaims = emitClaims(report.Experiments)
	report.Summary = summarize(report.Experiments, report.ClaimLedger, report.EmittedClaims, input.RepositorySnapshot)
	report.Guardrails = makeGuardrails(report.EmittedClaims, input.RepositorySnapshot)
	report.Digest = reportDigest(report)
	return report
}

func observationIndex(bundle ObservationBundle, bundleErr error) (map[string]ObservationReceipt, map[string]bool) {
	receipts := make(map[string]ObservationReceipt)
	duplicates := make(map[string]bool)
	if bundleErr != nil || bundle.Schema != ObservationSchema || bundle.BundleID == "" {
		return receipts, duplicates
	}
	for _, receipt := range bundle.Receipts {
		key := receiptKey(receipt.ExperimentID, receipt.GateID)
		if _, exists := receipts[key]; exists {
			duplicates[key] = true
			continue
		}
		receipts[key] = receipt
	}
	return receipts, duplicates
}

func receiptKey(experimentID, gateID string) string {
	return experimentID + "\x00" + gateID
}

func evaluateGate(experimentID, gateID string, receipt ObservationReceipt, duplicate bool, projection SourceProjection, sourceErr, bundleErr error) GateResult {
	if bundleErr != nil {
		return newGateResult(experimentID, gateID, StatusUnknown, "OBSERVE", "decode-observation-bundle", "observation bundle cannot be decoded")
	}
	if sourceErr != nil {
		return newGateResult(experimentID, gateID, StatusUnknown, "LOWER_RESOLUTION", "reconstruct-source", "source reconstruction is unavailable")
	}
	if duplicate {
		return newGateResult(experimentID, gateID, StatusRefuted, "VERIFY", "reject-duplicate-receipt", "more than one receipt binds the same experiment and gate")
	}
	if receipt.ObservationID == "" {
		return newGateResult(experimentID, gateID, StatusUnknown, "OBSERVE", "lookup-receipt", "observation receipt is absent")
	}
	if missingReceiptField(receipt) {
		return resultWithReceipt(experimentID, gateID, StatusUnknown, "OBSERVE", "validate-observation-receipt", "a required observation receipt field is absent", receipt)
	}

	if reason := validateReceipt(receipt, experimentID, gateID, projection); reason != "" {
		return resultWithReceipt(experimentID, gateID, StatusRefuted, "VERIFY", "validate-observation-receipt", reason, receipt)
	}
	status, stage, step, reason := classifyConclusion(receipt.Actions.Conclusion)
	if gateID == "semantic-causality" {
		if receipt.SemanticIntervention == nil {
			return resultWithReceipt(experimentID, gateID, StatusUnknown, "OBSERVE", "lookup-intervention-material", "semantic intervention material is absent", receipt)
		}
		intervention := receipt.SemanticIntervention
		if intervention.BaselineRawDigest != projection.RawDigest || intervention.BaselineSemanticDigest != projection.SemanticDigest || !validDigest(intervention.InterventionRawDigest) || !validDigest(intervention.InterventionSemanticDigest) || intervention.InterventionRawDigest == intervention.BaselineRawDigest || intervention.InterventionSemanticDigest == intervention.BaselineSemanticDigest {
			return resultWithReceipt(experimentID, gateID, StatusRefuted, "CAUSALITY", "compare-intervention-digests", "semantic intervention is contradictory", receipt)
		}
	}
	next := claimStateFor(status)
	if receipt.ClaimTransitionDigest != claimTransitionDigest(experimentID, gateID, next) {
		return resultWithReceipt(experimentID, gateID, StatusRefuted, "CLAIM", "verify-claim-transition-digest", "claim transition digest does not bind the derived transition", receipt)
	}
	return resultWithReceipt(experimentID, gateID, status, stage, step, reason, receipt)
}

func validateReceipt(receipt ObservationReceipt, experimentID, gateID string, projection SourceProjection) string {
	if receipt.Schema != ObservationSchema {
		return "observation receipt schema is invalid"
	}
	if receipt.ExperimentID != experimentID || receipt.GateID != gateID {
		return "observation receipt slot binding is invalid"
	}
	if receipt.PRNumber <= 0 {
		return "pull request number is missing"
	}
	if !headSHAPattern.MatchString(receipt.HeadSHA) {
		return "exact head SHA is malformed"
	}
	if receipt.SourceRawDigest != projection.RawDigest {
		return "source raw digest does not bind the reconstructed source"
	}
	if receipt.SourceSemanticDigest != projection.SemanticDigest {
		return "source semantic digest does not bind the reconstructed source"
	}
	if receipt.ProducerID != producerID {
		return "producer identifier does not bind the declared producer"
	}
	if receipt.ConsumerPackage != consumerPackage || len(receipt.ConsumerImports) == 0 {
		return "consumer package or import boundary is missing"
	}
	for _, imported := range receipt.ConsumerImports {
		if imported == producerPackage || strings.HasPrefix(imported, producerPackage+"/") {
			return "consumer import boundary includes the producer package"
		}
	}
	if !actionsURLsBind(receipt.Actions.RunURL, receipt.Actions.JobURL) {
		return "exact Actions run or job URL is malformed"
	}
	if !isActionConclusion(receipt.Actions.Conclusion) {
		return "Actions conclusion is missing or unsupported"
	}
	if receipt.Artifact.Bytes <= 0 || receipt.Artifact.Path == "" {
		return "artifact byte count or path is missing"
	}
	if !validDigest(receipt.Artifact.Digest) || receipt.Artifact.Digest != artifactDigest(receipt.Artifact.Path, receipt.Artifact.Bytes) {
		return "artifact digest is malformed or unbound"
	}
	if gateID := receipt.GateID; gateID != "semantic-causality" && receipt.SemanticIntervention != nil {
		return "semantic intervention is attached to the wrong gate"
	}
	return ""
}

func missingReceiptField(receipt ObservationReceipt) bool {
	return receipt.PRNumber <= 0 || receipt.HeadSHA == "" || receipt.SourceRawDigest == "" || receipt.SourceSemanticDigest == "" || receipt.ProducerID == "" || receipt.ConsumerPackage == "" || len(receipt.ConsumerImports) == 0 || receipt.ClaimTransitionDigest == "" || receipt.Actions.RunURL == "" || receipt.Actions.JobURL == "" || receipt.Actions.Conclusion == "" || receipt.Artifact.Bytes <= 0 || receipt.Artifact.Path == "" || receipt.Artifact.Digest == ""
}

func actionsURLsBind(runURL, jobURL string) bool {
	run := runURLPattern.FindStringSubmatch(runURL)
	job := jobURLPattern.FindStringSubmatch(jobURL)
	return len(run) == 2 && len(job) == 2 && run[1] == job[1]
}

func isActionConclusion(value string) bool {
	switch value {
	case ObservationSuccess, ObservationProgress, ObservationQueued, ObservationFailure, ObservationCanceled:
		return true
	default:
		return false
	}
}

func classifyConclusion(value string) (string, string, string, string) {
	switch value {
	case ObservationSuccess:
		return StatusProven, "OBSERVE", "bind-actions-observation", "required observation fields are bound"
	case ObservationProgress, ObservationQueued:
		return StatusOpen, "OBSERVE", "await-actions-conclusion", "Actions conclusion is not final"
	default:
		return StatusRefuted, "OBSERVE", "classify-actions-conclusion", "Actions conclusion is not successful"
	}
}

func newGateResult(experimentID, gateID, status, stage, step, reason string) GateResult {
	transition := makeTransition(experimentID, gateID, status, stage, step, reason)
	return GateResult{ExperimentID: experimentID, GateID: gateID, Status: status, Stage: stage, Step: step, Reason: reason, ClaimTransition: transition}
}

func resultWithReceipt(experimentID, gateID, status, stage, step, reason string, receipt ObservationReceipt) GateResult {
	result := newGateResult(experimentID, gateID, status, stage, step, reason)
	result.ObservationID = receipt.ObservationID
	result.Receipt = &receipt
	return result
}

func claimStateFor(status string) string {
	if status == StatusProven {
		return ClaimDischarged
	}
	if status == StatusRefuted {
		return ClaimRefuted
	}
	return ClaimOpen
}

func makeTransition(experimentID, gateID, status, stage, step, reason string) ClaimTransition {
	next := claimStateFor(status)
	return ClaimTransition{
		ExperimentID: experimentID,
		GateID:       gateID,
		From:         ClaimOpen,
		To:           next,
		Stage:        stage,
		Step:         step,
		Reason:       reason,
		Digest:       claimTransitionDigest(experimentID, gateID, next),
	}
}

func summarizeExperiment(gates []GateResult) (string, string, string, string) {
	for _, gate := range gates {
		if gate.Status == StatusRefuted {
			return StatusRefuted, gate.Stage, gate.Step, gate.Reason
		}
	}
	for _, gate := range gates {
		if gate.Status == StatusOpen {
			return StatusOpen, gate.Stage, gate.Step, gate.Reason
		}
	}
	for _, gate := range gates {
		if gate.Status == StatusUnknown {
			return StatusUnknown, gate.Stage, gate.Step, gate.Reason
		}
	}
	return StatusProven, "DERIVE", "all-gates-proven", "all five gates are proven from observations"
}

func sealLedger(experiments []ExperimentResult) []ClaimLedgerEntry {
	ledger := make([]ClaimLedgerEntry, 0, GateSlotCount)
	previous := ""
	sequence := 0
	for _, experiment := range experiments {
		for _, gate := range experiment.Gates {
			sequence++
			entry := ClaimLedgerEntry{
				Sequence: sequence, ExperimentID: gate.ExperimentID, GateID: gate.GateID,
				PriorState: ClaimOpen, NextState: gate.ClaimTransition.To,
				Stage: gate.Stage, Step: gate.Step, Reason: gate.Reason,
				PreviousDigest: previous,
			}
			entry.Digest = ledgerDigest(entry)
			ledger = append(ledger, entry)
			previous = entry.Digest
		}
	}
	return ledger
}

func emitClaims(experiments []ExperimentResult) []EmittedClaim {
	claims := make([]EmittedClaim, 0, GateSlotCount)
	for _, experiment := range experiments {
		for _, gate := range experiment.Gates {
			claims = append(claims, EmittedClaim{ExperimentID: gate.ExperimentID, GateID: gate.GateID, Class: "PROMOTION_GATE", State: gate.ClaimTransition.To})
		}
	}
	return claims
}

func summarize(experiments []ExperimentResult, ledger []ClaimLedgerEntry, claims []EmittedClaim, snapshot RepositorySnapshot) Summary {
	result := Summary{
		ExperimentsNumerator: ExperimentCount, ExperimentsDenominator: ExperimentCount,
		GateSlotsNumerator: GateSlotCount, GateSlotsDenominator: GateSlotCount,
		ClaimTransitionsNumerator: len(ledger), ClaimTransitionsDenominator: GateSlotCount,
		RepositoryWrites: snapshot.ChangedPaths, MutationAuthority: false,
		ForbiddenAggregates: forbiddenAggregateClaims(claims),
	}
	for _, experiment := range experiments {
		addState(&result.ExperimentStates, experiment.Status)
		for _, gate := range experiment.Gates {
			addState(&result.GateStates, gate.Status)
		}
	}
	return result
}

func addState(counts *StateCounts, status string) {
	switch status {
	case StatusProven:
		counts.Proven++
	case StatusOpen:
		counts.Open++
	case StatusUnknown:
		counts.Unknown++
	case StatusRefuted:
		counts.Refuted++
	}
}

func forbiddenAggregateClaims(claims []EmittedClaim) int {
	count := 0
	for _, claim := range claims {
		if claim.State != "ASSERTED" {
			continue
		}
		switch claim.Class {
		case "IMPROVEMENT_RATE", "AGGREGATE_ESTIMATE", "WEIGHTED_SCORE":
			count++
		}
	}
	return count
}

func makeGuardrails(claims []EmittedClaim, snapshot RepositorySnapshot) []Guardrail {
	return []Guardrail{
		{
			ID: "gooo.guardrail.experiment-promotion.forbidden-aggregate-claim.v1", Direction: "AT_MOST",
			Observed: forbiddenAggregateClaims(claims), AllowedMax: 0,
			ConformanceNumerator: 1, ConformanceDenominator: 1,
			Conforms: forbiddenAggregateClaims(claims) == 0,
		},
		{
			ID: "gooo.guardrail.experiment-promotion.repository-writes.v1", Direction: "AT_MOST",
			Observed: snapshot.ChangedPaths, AllowedMax: 0,
			ConformanceNumerator: boolInt(snapshot.ChangedPaths == 0), ConformanceDenominator: 1,
			Conforms: snapshot.ChangedPaths == 0,
		},
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// FormatSummary is intentionally exact-count-only. It is used by CI and
// cannot produce a score, percentage, or improvement estimate.
func FormatSummary(report Report) string {
	return fmt.Sprintf("experiments=%d/%d gate_slots=%d/%d states=PROVEN:%d,OPEN:%d,UNKNOWN:%d,REFUTED:%d gate_states=PROVEN:%d,OPEN:%d,UNKNOWN:%d,REFUTED:%d", report.Summary.ExperimentsNumerator, report.Summary.ExperimentsDenominator, report.Summary.GateSlotsNumerator, report.Summary.GateSlotsDenominator, report.Summary.ExperimentStates.Proven, report.Summary.ExperimentStates.Open, report.Summary.ExperimentStates.Unknown, report.Summary.ExperimentStates.Refuted, report.Summary.GateStates.Proven, report.Summary.GateStates.Open, report.Summary.GateStates.Unknown, report.Summary.GateStates.Refuted)
}
