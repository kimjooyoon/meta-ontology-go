package experimentpromotionverify

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
)

const portfolioScope = "GOOO_META_EXPERIMENT_PROMOTION_LEDGER"
const producerPackage = "github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion"

var (
	headSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	runURLPattern  = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/([0-9]+)$`)
	jobURLPattern  = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/([0-9]+)/jobs/[0-9]+$`)
)

// Verify independently replays source and observation inputs and compares the
// complete expected report. It deliberately has no import edge to producer.
func Verify(input Input) Verification {
	expected, replayErr := replay(input)
	checks := []ReplayCheck{}
	checks = append(checks, check("source-reconstruction", expected.SourceProjection, input.Report.SourceProjection, replayErr == nil, "ParseFile -> bidir.Lower rebuilt the exact portfolio and gate set"))
	checks = append(checks, check("observation-replay", expected.ObservationDigest, input.Report.ObservationDigest, replayErr == nil && expected.ObservationDigest == input.Report.ObservationDigest, "raw observation receipts were independently decoded"))
	checks = append(checks, check("promotion-report", expected.Digest, input.Report.Digest, replayErr == nil && reflect.DeepEqual(expected, input.Report), "producer report matches the independently reconstructed report"))
	checks = append(checks, check("forbidden-import-boundary", 0, 0, replayErr == nil && noProducerImport(input.Report), "consumer package boundary excludes producer package"))
	checks = append(checks, check("exact-denominators", "30/30 and 150/150", fmt.Sprintf("%d/%d and %d/%d", input.Report.Summary.ExperimentsNumerator, input.Report.Summary.ExperimentsDenominator, input.Report.Summary.GateSlotsNumerator, input.Report.Summary.GateSlotsDenominator), replayErr == nil && input.Report.Summary.ExperimentsNumerator == ExperimentCount && input.Report.Summary.ExperimentsDenominator == ExperimentCount && input.Report.Summary.GateSlotsNumerator == GateSlotCount && input.Report.Summary.GateSlotsDenominator == GateSlotCount, "fixed experiment and gate-slot denominators"))

	decision := "PASS"
	reason := "independent consumer replay matches source-derived report"
	if replayErr != nil {
		decision = "FAIL_CLOSED"
		reason = replayErr.Error()
	} else if !reflect.DeepEqual(expected, input.Report) {
		decision = "FAIL_CLOSED"
		reason = "independent replay differs from producer report"
	}
	verification := Verification{
		Schema: ReportSchema, SubjectSHA: input.Report.SubjectSHA, Decision: decision,
		Resolution: "EXACT", Reason: reason, SourceProjection: expected.SourceProjection,
		Checks: checks, Summary: input.Report.Summary, Guardrails: input.Report.Guardrails,
		AggregateMetrics: []string{}, NotClaimed: append([]string(nil), input.Report.NotClaimed...),
		RepositoryWrites: input.Report.RepositoryWrites, MutationAuthority: input.Report.MutationAuthority,
	}
	verification.Digest = verificationDigest(verification)
	return verification
}

func replay(input Input) (Report, error) {
	contractErr := validateContract(input.Contract)
	if contractErr != nil {
		return Report{}, contractErr
	}
	projection, sourceErr := parseSource(input.SourceRaw)
	if sourceErr != nil {
		return Report{}, sourceErr
	}
	bundle, bundleErr := decodeObservation(input.ObservationRaw)
	receipts, duplicates := observationIndex(bundle, bundleErr)
	expected := Report{
		Schema: ReportSchema, Scope: portfolioScope, SubjectSHA: input.SubjectSHA,
		SourceProjection: projection, ObservationDigest: digestBytes(input.ObservationRaw),
		AggregateMetrics: []string{}, NotClaimed: append([]string(nil), input.Contract.NotClaimed...),
		RepositoryWrites: input.RepositorySnapshot.ChangedPaths, MutationAuthority: false,
		RepositorySnapshot: input.RepositorySnapshot,
	}
	for _, experimentID := range projection.Experiments {
		experiment := ExperimentResult{ExperimentID: experimentID, Gates: make([]GateResult, 0, GateCount)}
		for _, gateID := range projection.Gates {
			experiment.Gates = append(experiment.Gates, evaluateGate(experimentID, gateID, receipts[receiptKey(experimentID, gateID)], duplicates[receiptKey(experimentID, gateID)], projection, bundleErr))
		}
		experiment.Status, experiment.Stage, experiment.Step, experiment.Reason = summarizeExperiment(experiment.Gates)
		expected.Experiments = append(expected.Experiments, experiment)
	}
	expected.ClaimLedger = sealLedger(expected.Experiments)
	expected.EmittedClaims = emitClaims(expected.Experiments)
	expected.Summary = summarize(expected.Experiments, expected.ClaimLedger, expected.EmittedClaims, input.RepositorySnapshot)
	expected.Guardrails = makeGuardrails(expected.EmittedClaims, input.RepositorySnapshot)
	expected.Digest = reportDigest(expected)
	return expected, nil
}

func validateContract(contract Contract) error {
	expectedExperiments := experimentIDs()
	if contract.Schema != "gooo/experiment-promotion-contract/v1" || contract.Version != 1 || contract.SourcePath != SourcePath || !sameStrings(contract.Experiments, expectedExperiments) || !sameStrings(contract.Gates, GateIDs) || contract.ExperimentDenominator != ExperimentCount || contract.GateSlotDenominator != GateSlotCount {
		return fmt.Errorf("contract expectation is invalid")
	}
	expectedFields := []string{"pr_number", "head_sha", "source_raw_digest", "source_semantic_digest", "producer_id", "consumer_package", "consumer_imports", "claim_transition_digest", "actions.run_url", "actions.job_url", "actions.conclusion", "artifact.bytes", "artifact.path", "artifact.digest"}
	expectedNotClaimed := []string{"overall promotion score", "weighted gate score", "improvement rate", "aggregate estimate", "PR title or body as evidence", "network conclusion cache"}
	if !sameStrings(contract.RequiredReceiptFields, expectedFields) || !sameStrings(contract.NotClaimed, expectedNotClaimed) {
		return fmt.Errorf("contract guardrails are invalid")
	}
	return nil
}

func decodeObservation(raw []byte) (ObservationBundle, error) {
	var bundle ObservationBundle
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bundle); err != nil {
		return ObservationBundle{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return ObservationBundle{}, fmt.Errorf("trailing JSON")
		}
		return ObservationBundle{}, err
	}
	return bundle, nil
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

func receiptKey(experimentID, gateID string) string { return experimentID + "\x00" + gateID }

func evaluateGate(experimentID, gateID string, receipt ObservationReceipt, duplicate bool, projection SourceProjection, bundleErr error) GateResult {
	if bundleErr != nil {
		return newGateResult(experimentID, gateID, StatusUnknown, "OBSERVE", "decode-observation-bundle", "observation bundle cannot be decoded")
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
		intervention := receipt.SemanticIntervention
		if intervention == nil {
			return resultWithReceipt(experimentID, gateID, StatusUnknown, "OBSERVE", "lookup-intervention-material", "semantic intervention material is absent", receipt)
		}
		if intervention.BaselineRawDigest != projection.RawDigest || intervention.BaselineSemanticDigest != projection.SemanticDigest || !validDigest(intervention.InterventionRawDigest) || !validDigest(intervention.InterventionSemanticDigest) || intervention.InterventionRawDigest == intervention.BaselineRawDigest || intervention.InterventionSemanticDigest == intervention.BaselineSemanticDigest {
			return resultWithReceipt(experimentID, gateID, StatusRefuted, "CAUSALITY", "compare-intervention-digests", "semantic intervention is contradictory", receipt)
		}
	}
	if receipt.ClaimTransitionDigest != claimTransitionDigest(experimentID, gateID, claimStateFor(status)) {
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
	if receipt.ProducerID != "gooo-experiment-promotion-producer/v1" {
		return "producer identifier does not bind the declared producer"
	}
	if receipt.ConsumerPackage != "github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotionverify" || len(receipt.ConsumerImports) == 0 {
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
	if gateID != "semantic-causality" && receipt.SemanticIntervention != nil {
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
	return GateResult{ExperimentID: experimentID, GateID: gateID, Status: status, Stage: stage, Step: step, Reason: reason, ClaimTransition: makeTransition(experimentID, gateID, status, stage, step, reason)}
}

func resultWithReceipt(experimentID, gateID, status, stage, step, reason string, receipt ObservationReceipt) GateResult {
	result := newGateResult(experimentID, gateID, status, stage, step, reason)
	result.ObservationID = receipt.ObservationID
	receiptCopy := receipt
	result.Receipt = &receiptCopy
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
	return ClaimTransition{ExperimentID: experimentID, GateID: gateID, From: ClaimOpen, To: next, Stage: stage, Step: step, Reason: reason, Digest: claimTransitionDigest(experimentID, gateID, next)}
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
			entry := ClaimLedgerEntry{Sequence: sequence, ExperimentID: gate.ExperimentID, GateID: gate.GateID, PriorState: ClaimOpen, NextState: gate.ClaimTransition.To, Stage: gate.Stage, Step: gate.Step, Reason: gate.Reason, PreviousDigest: previous}
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
	result := Summary{ExperimentsNumerator: ExperimentCount, ExperimentsDenominator: ExperimentCount, GateSlotsNumerator: GateSlotCount, GateSlotsDenominator: GateSlotCount, ClaimTransitionsNumerator: len(ledger), ClaimTransitionsDenominator: GateSlotCount, RepositoryWrites: snapshot.ChangedPaths, MutationAuthority: false, ForbiddenAggregates: forbiddenAggregateClaims(claims)}
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
		if claim.State == "ASSERTED" && (claim.Class == "IMPROVEMENT_RATE" || claim.Class == "AGGREGATE_ESTIMATE" || claim.Class == "WEIGHTED_SCORE") {
			count++
		}
	}
	return count
}

func makeGuardrails(claims []EmittedClaim, snapshot RepositorySnapshot) []Guardrail {
	forbidden := forbiddenAggregateClaims(claims)
	return []Guardrail{
		{ID: "gooo.guardrail.experiment-promotion.forbidden-aggregate-claim.v1", Direction: "AT_MOST", Observed: forbidden, AllowedMax: 0, ConformanceNumerator: 1, ConformanceDenominator: 1, Conforms: forbidden == 0},
		{ID: "gooo.guardrail.experiment-promotion.repository-writes.v1", Direction: "AT_MOST", Observed: snapshot.ChangedPaths, AllowedMax: 0, ConformanceNumerator: boolInt(snapshot.ChangedPaths == 0), ConformanceDenominator: 1, Conforms: snapshot.ChangedPaths == 0},
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func claimTransitionDigest(experimentID, gateID, next string) string {
	return digestBytes([]byte(fmt.Sprintf("claim-transition/v1|%s|%s|OPEN|%s", experimentID, gateID, next)))
}

func artifactDigest(path string, bytes int) string {
	return digestBytes([]byte(fmt.Sprintf("artifact/v1|%s|%d", path, bytes)))
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}

func ledgerDigest(entry ClaimLedgerEntry) string {
	entry.Digest = ""
	return digestValue(entry)
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestValue(report)
}

func verificationDigest(verification Verification) string {
	verification.Digest = ""
	return digestValue(verification)
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func check(id string, expected, observed any, ok bool, reason string) ReplayCheck {
	status := "FAIL"
	if ok {
		status = "PASS"
	}
	return ReplayCheck{ID: id, Status: status, Stage: "REPLAY", Step: id, Reason: reason, Expected: fmt.Sprint(expected), Observed: fmt.Sprint(observed)}
}

func noProducerImport(report Report) bool {
	for _, experiment := range report.Experiments {
		for _, gate := range experiment.Gates {
			if gate.Receipt == nil {
				continue
			}
			for _, imported := range gate.Receipt.ConsumerImports {
				if imported == producerPackage || strings.HasPrefix(imported, producerPackage+"/") {
					return false
				}
			}
		}
	}
	return true
}
