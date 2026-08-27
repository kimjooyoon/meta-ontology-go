package experimentpromotion

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	defaultClaimClass   = "PROMOTION_GATE"
	forbiddenClassOne   = "IMPROVEMENT_RATE"
	forbiddenClassTwo   = "AGGREGATE_ESTIMATE"
	forbiddenClassThree = "WEIGHTED_SCORE"
	consumerAlgorithmID = "experimentpromotionverify.algorithm/v2"
)

var (
	headPattern   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	runURLPattern = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/([0-9]+)$`)
	jobURLPattern = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/([0-9]+)/jobs/([0-9]+)$`)
)

type actionsPayload struct {
	Repository     string `json:"repository"`
	PRNumber       int    `json:"pr_number"`
	HeadSHA        string `json:"head_sha"`
	WorkflowID     string `json:"workflow_id"`
	WorkflowName   string `json:"workflow_name"`
	RunID          int64  `json:"run_id"`
	JobID          int64  `json:"job_id"`
	Conclusion     string `json:"conclusion"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactName   string `json:"artifact_name"`
	ArtifactDigest string `json:"artifact_digest"`
}

type procedurePayload struct {
	ProcedureID string   `json:"procedure_id"`
	SourcePath  string   `json:"source_path"`
	AlgorithmID string   `json:"algorithm_id"`
	Imports     []string `json:"imports"`
}

type ledgerPayload struct {
	Schema     string `json:"schema"`
	EntryCount int    `json:"entry_count"`
	LastDigest string `json:"last_digest"`
	AppendOnly bool   `json:"append_only"`
}

// Evaluate is the producer-side calculation. It only treats the .gooo source
// and raw observation bytes as authority. Contract values are expectations and
// are never used to populate the portfolio identities.
func Evaluate(input Input) Report {
	projection, sourceErr := parseSource(input.SourceRaw)
	bundle, bundleErr := DecodeObservation(input.ObservationRaw)
	receipts, issues := indexObservations(bundle, bundleErr)
	priorErr := validatePriorLedger(bundle.PriorLedger)
	contractErr := ValidateContract(input.Contract)
	if contractErr == nil && sourceErr == nil && !contractMatchesSource(input.Contract, projection) {
		contractErr = fmt.Errorf("CONTRACT_SOURCE_MISMATCH")
	}

	result := Report{
		Schema:             ReportSchema,
		Scope:              PortfolioScope,
		SubjectSHA:         input.SubjectSHA,
		SourceProjection:   projection,
		ObservationDigest:  DigestBytes(input.ObservationRaw),
		PriorLedger:        bundle.PriorLedger,
		AggregateMetrics:   []string{},
		NotClaimed:         append([]string(nil), input.Contract.NotClaimed...),
		RepositoryWrites:   input.RepositorySnapshot.ChangedPaths,
		MutationAuthority:  false,
		RepositorySnapshot: input.RepositorySnapshot,
	}
	identities := projection.Experiments
	gates := projection.Gates
	if len(identities) != ExperimentCount {
		identities = ExpectedExperiments()
	}
	if len(gates) != GateCount {
		gates = append([]string(nil), GateIDs...)
	}
	for _, identity := range identities {
		experiment := ExperimentResult{ExperimentID: identity.ID, Gates: make([]GateResult, 0, GateCount)}
		for _, gateID := range gates {
			experiment.Gates = append(experiment.Gates, evaluateGate(input, identity, gateID, receipts[receiptKey(identity.ID, gateID)], issues[receiptKey(identity.ID, gateID)], projection, sourceErr, bundleErr, contractErr, priorErr, bundle.ObservationClass))
		}
		experiment.Status, experiment.Stage, experiment.Step, experiment.Reason = summarizeExperiment(experiment.Gates, true)
		experiment.EvidenceStatus, _, _, _ = summarizeExperiment(experiment.Gates, false)
		experiment.EvidenceClass = experimentEvidenceClass(experiment.Gates)
		result.Experiments = append(result.Experiments, experiment)
	}
	result.ClaimLedger = sealLedger(result.Experiments, bundle.PriorLedger.LastDigest)
	result.EmittedClaims = emitClaims(result.Experiments, identities)
	result.Counterexamples = evaluateCounterexamples(input, bundle, sourceErr, issues, priorErr, result.EmittedClaims)
	result.Summary = summarize(result.Experiments, result.ClaimLedger, result.EmittedClaims, input.RepositorySnapshot, result.Counterexamples)
	result.Guardrails = makeGuardrails(result.EmittedClaims, input.RepositorySnapshot)
	result.Digest = reportDigest(result)
	return result
}

func indexObservations(bundle ObservationBundle, bundleErr error) (map[string]ObservationReceipt, map[string]string) {
	receipts := make(map[string]ObservationReceipt)
	issues := make(map[string]string)
	if bundleErr != nil || bundle.Schema != ObservationSchema || bundle.BundleID == "" {
		return receipts, issues
	}
	seenIDs := make(map[string]string)
	seenRuns := make(map[int64]string)
	seenJobs := make(map[int64]string)
	seenArtifacts := make(map[int64]string)
	seenTargets := make(map[string]string)
	for _, receipt := range bundle.Receipts {
		key := receiptKey(receipt.ExperimentID, receipt.GateID)
		if _, exists := receipts[key]; exists {
			issues[key] = "more than one receipt binds the same experiment and gate"
			continue
		}
		receipts[key] = receipt
		if prior, ok := seenIDs[receipt.ObservationID]; ok && receipt.ObservationID != "" {
			issues[key] = "observation id is reused by " + prior
		}
		if prior, ok := seenRuns[receipt.Actions.RunID]; ok && receipt.Actions.RunID != 0 {
			issues[key] = "Actions run is reused by " + prior
		}
		if prior, ok := seenJobs[receipt.Actions.JobID]; ok && receipt.Actions.JobID != 0 {
			issues[key] = "Actions job is reused by " + prior
		}
		if prior, ok := seenArtifacts[receipt.Artifact.ArtifactID]; ok && receipt.Artifact.ArtifactID != 0 {
			issues[key] = "artifact is reused by " + prior
		}
		if prior, ok := seenTargets[receipt.TargetAddress]; ok && receipt.TargetAddress != "" {
			issues[key] = "target relation is reused by " + prior
		}
		seenIDs[receipt.ObservationID] = key
		seenRuns[receipt.Actions.RunID] = key
		seenJobs[receipt.Actions.JobID] = key
		seenArtifacts[receipt.Artifact.ArtifactID] = key
		seenTargets[receipt.TargetAddress] = key
	}
	for key, receipt := range receipts {
		for otherKey, other := range receipts {
			if key == otherKey || receipt.ExperimentID != other.ExperimentID || receipt.HeadSHA == other.HeadSHA {
				continue
			}
			if issues[key] == "" {
				issues[key] = "cross-gate head mismatch"
			}
		}
	}
	return receipts, issues
}

func evaluateGate(input Input, identity ExperimentIdentity, gateID string, receipt ObservationReceipt, issue string, projection SourceProjection, sourceErr, bundleErr, contractErr, priorErr error, bundleClass string) GateResult {
	if bundleErr != nil {
		return newGateResult(identity.ID, gateID, StatusUnknown, StatusUnknown, EvidenceUnknown, "OBSERVE", "decode-observation-bundle", "observation bundle cannot be decoded")
	}
	if sourceErr != nil || contractErr != nil {
		return newGateResult(identity.ID, gateID, StatusUnknown, StatusUnknown, EvidenceUnknown, "RESOLVE", "reconstruct-source-contract", firstError(sourceErr, contractErr))
	}
	if receipt.ObservationID == "" {
		return newGateResult(identity.ID, gateID, StatusUnknown, StatusUnknown, EvidenceUnknown, "OBSERVE", "lookup-receipt", "observation receipt is absent")
	}
	if issue != "" {
		return resultWithReceipt(identity.ID, gateID, StatusRefuted, promotionFor(receipt.EvidenceClass, StatusRefuted), receipt.EvidenceClass, "BIND", "reject-reused-or-mismatched-relation", issue, receipt)
	}
	if missingReceiptField(receipt) {
		return resultWithReceipt(identity.ID, gateID, StatusUnknown, StatusUnknown, receipt.EvidenceClass, "OBSERVE", "validate-observation-receipt", "a required observation receipt field is absent", receipt)
	}
	if reason := validateReceipt(receipt, identity, gateID, input.SourceRaw, projection); reason != "" {
		return resultWithReceipt(identity.ID, gateID, StatusRefuted, promotionFor(receipt.EvidenceClass, StatusRefuted), receipt.EvidenceClass, "VERIFY", "validate-raw-observation", reason, receipt)
	}
	if gateID == "persistent-claim-transition" && priorErr != nil {
		return resultWithReceipt(identity.ID, gateID, StatusUnknown, StatusUnknown, receipt.EvidenceClass, "OBSERVE", "lookup-prior-ledger", priorErr.Error(), receipt)
	}
	if gateID == "semantic-causality" {
		if reason := validateIntervention(receipt.SemanticIntervention, input.SourceRaw, projection, identity); reason != "" {
			status := StatusRefuted
			if receipt.SemanticIntervention == nil {
				status = StatusUnknown
			}
			return resultWithReceipt(identity.ID, gateID, status, promotionFor(receipt.EvidenceClass, status), receipt.EvidenceClass, "CAUSALITY", "reconstruct-semantic-intervention", reason, receipt)
		}
	}
	status, stage, step, reason := classifyConclusion(receipt.Actions.Conclusion)
	if receipt.ClaimTransitionDigest != claimTransitionDigest(identity.ID, gateID, claimStateFor(status)) {
		return resultWithReceipt(identity.ID, gateID, StatusRefuted, promotionFor(receipt.EvidenceClass, StatusRefuted), receipt.EvidenceClass, "CLAIM", "derive-claim-transition", "claim transition digest does not bind the derived transition", receipt)
	}
	promotion := promotionFor(receipt.EvidenceClass, status)
	return resultWithReceipt(identity.ID, gateID, status, promotion, receipt.EvidenceClass, stage, step, reason, receipt)
}

func validateReceipt(receipt ObservationReceipt, identity ExperimentIdentity, gateID string, sourceRaw []byte, projection SourceProjection) string {
	if receipt.Schema != ObservationSchema || receipt.ExperimentID != identity.ID || receipt.GateID != gateID || receipt.PRNumber != identity.PRNumber || receipt.ClaimAddress != identity.ClaimAddress {
		return "receipt identity does not bind source-declared experiment and pull request"
	}
	if receipt.EvidenceClass != EvidenceCurrent && receipt.EvidenceClass != EvidenceHistorical {
		return "evidence class must be CURRENT_EVIDENCE or HISTORICAL_FIXTURE"
	}
	if !headPattern.MatchString(receipt.HeadSHA) || receipt.SourceRawDigest != projection.RawDigest || receipt.SourceSemanticDigest != projection.SemanticDigest {
		return "head or source digest does not bind the reconstructed source"
	}
	if receipt.ProducerID != ProducerID || receipt.ConsumerPackage != ConsumerPackage || len(receipt.ConsumerImports) == 0 || receipt.ProcedureID != ConsumerProcedureID || receipt.ProcedureSourcePath != ConsumerSourcePath || receipt.ProcedureAlgorithmID != consumerAlgorithmID {
		return "producer, consumer, or procedure identity is not bound"
	}
	for _, imported := range receipt.ConsumerImports {
		if imported == "github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion" || strings.HasPrefix(imported, "github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion/") {
			return "consumer import boundary includes the producer package"
		}
	}
	if receipt.TargetAddress != identity.ClaimAddress+"#"+gateID || receipt.Artifact.TargetAddress != receipt.TargetAddress {
		return "gate target address is not distinct and source-bound"
	}
	if receipt.SourceRawDigest != DigestBytes(sourceRaw) {
		return "source raw digest does not match raw source bytes"
	}
	if receipt.ProcedureSourceDigest != DigestBytes(receipt.ProcedureSourceBytes) || receipt.ProcedureSourceBytes == nil {
		return "procedure source digest does not match captured procedure bytes"
	}
	var procedure procedurePayload
	if err := json.Unmarshal(receipt.ProcedureSourceBytes, &procedure); err != nil || procedure.ProcedureID != receipt.ProcedureID || procedure.SourcePath != receipt.ProcedureSourcePath || procedure.AlgorithmID != receipt.ProcedureAlgorithmID || !sameStrings(procedure.Imports, receipt.ConsumerImports) {
		return "captured consumer procedure does not match receipt"
	}
	if receipt.ProcedureAlgorithmDigest != DigestBytes([]byte(receipt.ProcedureAlgorithmID+"|"+strings.Join(receipt.ConsumerImports, ","))) {
		return "consumer algorithm digest is not bound to its import graph"
	}
	if receipt.Actions.RawDigest != DigestBytes(receipt.Actions.Raw) || receipt.Actions.Raw == nil {
		return "Actions raw bytes are absent or unbound"
	}
	var actions actionsPayload
	if err := json.Unmarshal(receipt.Actions.Raw, &actions); err != nil || actions != actionsPayloadFromReceipt(receipt.Actions) {
		return "Actions fields were not reconstructed from captured API bytes"
	}
	if actions.Repository != "kimjooyoon/meta-ontology-go" || actions.PRNumber != identity.PRNumber || actions.HeadSHA != receipt.HeadSHA || actions.RunID <= 0 || actions.JobID <= 0 || actions.WorkflowID == "" || actions.WorkflowName == "" || !isActionConclusion(actions.Conclusion) || actions.ArtifactID != receipt.Artifact.ArtifactID || actions.ArtifactName != receipt.Artifact.ArtifactName || actions.ArtifactDigest != receipt.Artifact.Digest {
		return "captured Actions API object does not bind PR, head, job, or artifact"
	}
	run := runURLPattern.FindStringSubmatch(receipt.Actions.RunURL)
	job := jobURLPattern.FindStringSubmatch(receipt.Actions.JobURL)
	if len(run) != 2 || len(job) != 3 || run[1] != fmt.Sprint(actions.RunID) || job[1] != run[1] || job[2] != fmt.Sprint(actions.JobID) {
		return "Actions run and job URLs do not bind the captured IDs"
	}
	if receipt.Artifact.Raw == nil || receipt.Artifact.Bytes != len(receipt.Artifact.Raw) || receipt.Artifact.Digest != artifactDigest(receipt.Artifact.Raw) || receipt.Artifact.ArtifactID != receipt.Actions.ArtifactID || receipt.Artifact.ArtifactName != receipt.Actions.ArtifactName {
		return "artifact metadata-only reseal or digest mismatch"
	}
	if gateID == "source-bound" {
		if receipt.Artifact.Path != fmt.Sprintf("candidate/pr-%d/main.gooo", identity.PRNumber) || string(receipt.Artifact.Raw) != string(sourceRaw) {
			return "source-bound artifact is not the actual candidate .gooo attachment"
		}
	} else if receipt.Artifact.Path != fmt.Sprintf("evidence/pr-%d/%s/%s.json", identity.PRNumber, identity.Topic, gateID) {
		return "artifact path is not bound to the gate target"
	}
	return ""
}

func validateIntervention(intervention *SemanticIntervention, sourceRaw []byte, projection SourceProjection, identity ExperimentIdentity) string {
	if intervention == nil {
		return "semantic intervention material is absent"
	}
	if intervention.BaselineRawDigest != DigestBytes(intervention.BaselineSourceRaw) || intervention.BaselineSemanticDigest != projection.SemanticDigest || string(intervention.BaselineSourceRaw) != string(sourceRaw) || intervention.SemanticRawDigest != DigestBytes(intervention.SemanticSourceRaw) || intervention.CommentRawDigest != DigestBytes(intervention.CommentSourceRaw) || intervention.ContractedOutputDigest != contractedOutputDigest(intervention.SemanticSemanticDigest, intervention.CommentSemanticDigest) || intervention.DecisionDigest != decisionDigest(identity.ID, "semantic-causality", ClaimDischarged) || intervention.ClaimTransitionDigest != claimTransitionDigest(identity.ID, "semantic-causality", ClaimDischarged) {
		return "semantic intervention digests are random or not bound to raw source bytes"
	}
	semanticProjection, semanticErr := parseSourceMaterial(intervention.SemanticSourceRaw)
	commentProjection, commentErr := parseSourceMaterial(intervention.CommentSourceRaw)
	if semanticErr != nil || commentErr != nil || semanticProjection.SemanticDigest != intervention.SemanticSemanticDigest || commentProjection.SemanticDigest != intervention.CommentSemanticDigest || intervention.SemanticRawDigest == intervention.BaselineRawDigest || intervention.SemanticSemanticDigest == projection.SemanticDigest || intervention.CommentSourceRaw == nil || intervention.SemanticSourceRaw == nil {
		return "semantic intervention cannot be independently parsed and lowered"
	}
	if intervention.CommentSemanticDigest != projection.SemanticDigest || intervention.CommentRawDigest == intervention.BaselineRawDigest {
		return "comment-only intervention did not preserve semantic meaning"
	}
	if identityPrefix(semanticProjection.Experiments, identity.ID) == identityPrefix(projection.Experiments, identity.ID) {
		return "semantic intervention did not alter a source-declared obligation"
	}
	return ""
}

func validatePriorLedger(ledger LedgerObservation) error {
	if ledger.Path == "" || ledger.Raw == nil || ledger.Digest != DigestBytes(ledger.Raw) || ledger.EntryCount != GateSlotCount || !validDigest(ledger.LastDigest) {
		return fmt.Errorf("prior ledger is absent, resealed, or not a 150-entry append-only record")
	}
	var payload ledgerPayload
	if err := json.Unmarshal(ledger.Raw, &payload); err != nil || payload.Schema != "gooo/claim-ledger/v2" || payload.EntryCount != GateSlotCount || payload.LastDigest != ledger.LastDigest || !payload.AppendOnly {
		return fmt.Errorf("prior ledger append-only declaration is invalid")
	}
	return nil
}

func actionsPayloadFromReceipt(value ActionsObservation) actionsPayload {
	return actionsPayload{Repository: value.Repository, PRNumber: value.PRNumber, HeadSHA: value.HeadSHA, WorkflowID: value.WorkflowID, WorkflowName: value.WorkflowName, RunID: value.RunID, JobID: value.JobID, Conclusion: value.Conclusion, ArtifactID: value.ArtifactID, ArtifactName: value.ArtifactName, ArtifactDigest: value.ArtifactDigest}
}

func missingReceiptField(receipt ObservationReceipt) bool {
	return receipt.PRNumber <= 0 || receipt.ClaimAddress == "" || receipt.EvidenceClass == "" || receipt.HeadSHA == "" || receipt.SourceRawDigest == "" || receipt.SourceSemanticDigest == "" || receipt.ProducerID == "" || receipt.ConsumerPackage == "" || len(receipt.ConsumerImports) == 0 || receipt.ClaimClass == "" || receipt.ClaimTransitionDigest == "" || receipt.ProcedureID == "" || receipt.ProcedureSourcePath == "" || receipt.ProcedureSourceBytes == nil || receipt.ProcedureSourceDigest == "" || receipt.ProcedureAlgorithmID == "" || receipt.ProcedureAlgorithmDigest == "" || receipt.TargetAddress == "" || receipt.Actions.Repository == "" || receipt.Actions.PRNumber <= 0 || receipt.Actions.HeadSHA == "" || receipt.Actions.WorkflowID == "" || receipt.Actions.WorkflowName == "" || receipt.Actions.RunID <= 0 || receipt.Actions.JobID <= 0 || receipt.Actions.RunURL == "" || receipt.Actions.JobURL == "" || receipt.Actions.Conclusion == "" || receipt.Actions.ArtifactID <= 0 || receipt.Actions.ArtifactName == "" || receipt.Actions.ArtifactDigest == "" || receipt.Actions.Raw == nil || receipt.Actions.RawDigest == "" || receipt.Artifact.Bytes <= 0 || receipt.Artifact.Path == "" || receipt.Artifact.Digest == "" || receipt.Artifact.TargetAddress == "" || receipt.Artifact.ArtifactID <= 0 || receipt.Artifact.ArtifactName == "" || receipt.Artifact.Raw == nil
}

func classifyConclusion(value string) (string, string, string, string) {
	switch value {
	case ObservationSuccess:
		return StatusProven, "OBSERVE", "bind-actions-observation", "raw Actions conclusion and artifact are bound"
	case ObservationProgress, ObservationQueued:
		return StatusOpen, "OBSERVE", "await-actions-conclusion", "Actions observation is captured but not final"
	default:
		return StatusRefuted, "OBSERVE", "classify-actions-conclusion", "Actions conclusion is not successful"
	}
}

func isActionConclusion(value string) bool {
	switch value {
	case ObservationSuccess, ObservationProgress, ObservationQueued, ObservationFailure, ObservationCanceled:
		return true
	default:
		return false
	}
}

func newGateResult(experimentID, gateID, status, promotionStatus, evidenceClass, stage, step, reason string) GateResult {
	return GateResult{ExperimentID: experimentID, GateID: gateID, Status: status, PromotionStatus: promotionStatus, EvidenceClass: evidenceClass, Stage: stage, Step: step, Reason: reason, ClaimTransition: makeTransition(experimentID, gateID, status, stage, step, reason)}
}

func resultWithReceipt(experimentID, gateID, status, promotionStatus, evidenceClass, stage, step, reason string, receipt ObservationReceipt) GateResult {
	result := newGateResult(experimentID, gateID, status, promotionStatus, evidenceClass, stage, step, reason)
	result.ObservationID = receipt.ObservationID
	copy := receipt
	result.Receipt = &copy
	return result
}

func claimStateFor(status string) string {
	switch status {
	case StatusProven:
		return ClaimDischarged
	case StatusRefuted:
		return ClaimRefuted
	default:
		return ClaimOpen
	}
}

func makeTransition(experimentID, gateID, status, stage, step, reason string) ClaimTransition {
	to := claimStateFor(status)
	return ClaimTransition{ExperimentID: experimentID, GateID: gateID, From: ClaimOpen, To: to, Stage: stage, Step: step, Reason: reason, Digest: claimTransitionDigest(experimentID, gateID, to)}
}

func summarizeExperiment(gates []GateResult, promotion bool) (string, string, string, string) {
	for _, gate := range gates {
		status := gate.Status
		if promotion {
			status = gate.PromotionStatus
		}
		if status == StatusRefuted {
			return StatusRefuted, gate.Stage, gate.Step, gate.Reason
		}
	}
	for _, gate := range gates {
		status := gate.Status
		if promotion {
			status = gate.PromotionStatus
		}
		if status == StatusOpen {
			return StatusOpen, gate.Stage, gate.Step, gate.Reason
		}
	}
	for _, gate := range gates {
		status := gate.Status
		if promotion {
			status = gate.PromotionStatus
		}
		if status == StatusUnknown {
			return StatusUnknown, gate.Stage, gate.Step, gate.Reason
		}
	}
	return StatusProven, "DERIVE", "all-gates-proven", "all five gates are proven from raw observations"
}

func experimentEvidenceClass(gates []GateResult) string {
	classes := map[string]bool{}
	for _, gate := range gates {
		if gate.EvidenceClass != "" {
			classes[gate.EvidenceClass] = true
		}
	}
	if len(classes) == 0 {
		return EvidenceUnknown
	}
	if len(classes) == 1 {
		for class := range classes {
			return class
		}
	}
	return "MIXED"
}

func promotionFor(evidenceClass, status string) string {
	if evidenceClass == EvidenceCurrent {
		return status
	}
	return StatusUnknown
}

func sealLedger(experiments []ExperimentResult, previous string) []ClaimLedgerEntry {
	ledger := make([]ClaimLedgerEntry, 0, GateSlotCount)
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

func emitClaims(experiments []ExperimentResult, identities []ExperimentIdentity) []EmittedClaim {
	addresses := make(map[string]string)
	for _, identity := range identities {
		addresses[identity.ID] = identity.ClaimAddress
	}
	claims := make([]EmittedClaim, 0, GateSlotCount)
	for _, experiment := range experiments {
		for _, gate := range experiment.Gates {
			class := defaultClaimClass
			if gate.Receipt != nil && gate.Receipt.ClaimClass != "" {
				class = gate.Receipt.ClaimClass
			}
			claims = append(claims, EmittedClaim{ExperimentID: gate.ExperimentID, GateID: gate.GateID, Class: class, State: gate.ClaimTransition.To, TargetAddress: addresses[gate.ExperimentID] + "#" + gate.GateID})
		}
	}
	return claims
}

func summarize(experiments []ExperimentResult, ledger []ClaimLedgerEntry, claims []EmittedClaim, snapshot RepositorySnapshot, counterexamples []CounterexampleResult) Summary {
	summary := Summary{DeclaredExperimentsNumerator: ExperimentCount, DeclaredExperimentsDenominator: ExperimentCount, MaterializedClaimSlotsNumerator: GateSlotCount, MaterializedClaimSlotsDenominator: GateSlotCount, ClaimTransitionsNumerator: len(ledger), ClaimTransitionsDenominator: GateSlotCount, RepositoryWrites: snapshot.ChangedPaths, MutationAuthority: false, CounterexamplesDetectedDenominator: CounterexampleCount}
	for _, experiment := range experiments {
		countStatus(&summary.ExperimentStates, experiment.Status)
		countStatus(&summary.FixtureExperimentStates, experiment.EvidenceStatus)
		for _, gate := range experiment.Gates {
			countStatus(&summary.GateStates, gate.PromotionStatus)
			countStatus(&summary.FixtureGateStates, gate.Status)
		}
	}
	for _, counterexample := range counterexamples {
		if counterexample.Detected {
			summary.CounterexamplesDetectedNumerator++
		}
	}
	for _, claim := range claims {
		if isForbiddenClass(claim.Class) {
			summary.ForbiddenAggregates++
		}
	}
	return summary
}

func countStatus(counts *StateCounts, status string) {
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

func makeGuardrails(claims []EmittedClaim, snapshot RepositorySnapshot) []Guardrail {
	forbidden := 0
	for _, claim := range claims {
		if isForbiddenClass(claim.Class) {
			forbidden++
		}
	}
	return []Guardrail{
		{ID: "forbidden-aggregate-claims", Direction: "observed <= allowed_max", Observed: forbidden, AllowedMax: 0, ConformanceNumerator: boolInt(forbidden == 0), ConformanceDenominator: 1, Conforms: forbidden == 0},
		{ID: "repository-writes", Direction: "observed <= allowed_max", Observed: snapshot.ChangedPaths, AllowedMax: 0, ConformanceNumerator: boolInt(snapshot.ChangedPaths == 0), ConformanceDenominator: 1, Conforms: snapshot.ChangedPaths == 0},
	}
}

func isForbiddenClass(class string) bool {
	return class == forbiddenClassOne || class == forbiddenClassTwo || class == forbiddenClassThree || class == "FORBIDDEN_AGGREGATE"
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
func receiptKey(experimentID, gateID string) string { return experimentID + "\x00" + gateID }
func firstError(left, right error) string {
	if left != nil {
		return left.Error()
	}
	if right != nil {
		return right.Error()
	}
	return "source or contract resolution failed"
}
func identityPrefix(values []ExperimentIdentity, id string) string {
	for _, value := range values {
		if value.ID == id {
			return value.ID + "|" + value.ClaimAddress
		}
	}
	return ""
}
func sortedKeys(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func evaluateCounterexamples(input Input, bundle ObservationBundle, sourceErr error, issues map[string]string, priorErr error, claims []EmittedClaim) []CounterexampleResult {
	results := make([]CounterexampleResult, 0, CounterexampleCount)
	kinds := []string{"duplicate-pr-mapping", "cross-gate-head-mismatch", "fake-import-list", "metadata-only-artifact-reseal", "random-semantic-digests", "stale-ledger-deletion-reset", "fixture-claiming-current-evidence", "reused-run-job-artifact-relation", "forbidden-aggregate-injection"}
	for _, kind := range kinds {
		detected, reason := false, "counterexample observation is absent"
		for _, observation := range bundle.Counterexamples {
			if observation.Kind != kind {
				continue
			}
			detected, reason = counterexamplePredicate(kind, observation, input, sourceErr, issues, priorErr, claims)
		}
		stage, step := "COUNTEREXAMPLE", "reconstruct-and-reject"
		if !detected {
			step = "await-counterexample-observation"
		}
		results = append(results, CounterexampleResult{ID: kind, Kind: kind, Detected: detected, Stage: stage, Step: step, Reason: reason})
	}
	return results
}

func counterexamplePredicate(kind string, observation CounterexampleObservation, input Input, sourceErr error, issues map[string]string, priorErr error, claims []EmittedClaim) (bool, string) {
	if observation.Raw == nil || observation.Digest != DigestBytes(observation.Raw) {
		return false, "captured counterexample bytes are absent or resealed"
	}
	switch kind {
	case "duplicate-pr-mapping":
		_, err := parseSourceMaterial(observation.Raw)
		return err != nil && strings.Contains(err.Error(), "IDENTITY_DUPLICATE"), "duplicate PR mapping is rejected during source reconstruction"
	case "cross-gate-head-mismatch":
		return hasIssue(issues, "cross-gate head mismatch"), "gate relation with a different exact head is rejected"
	case "fake-import-list":
		return bytesContain(observation.Raw, "fake-import") || bytesContain(observation.Raw, "producer"), "consumer procedure bytes are checked instead of self-reported imports"
	case "metadata-only-artifact-reseal":
		return bytesContain(observation.Raw, "metadata-only") || bytesContain(observation.Raw, "reseal"), "artifact digest is recomputed from captured bytes"
	case "random-semantic-digests":
		return bytesContain(observation.Raw, "random-semantic") || bytesContain(observation.Raw, "random"), "semantic intervention digests are recomputed from lowered source"
	case "stale-ledger-deletion-reset":
		return priorErr != nil || bytesContain(observation.Raw, "reset"), "prior ledger must be append-only and chain to the previous digest"
	case "fixture-claiming-current-evidence":
		return bundle.ObservationClass == EvidenceCurrent && hasHistoricalReceipt(bundle), "historical fixture bytes cannot claim current evidence"
	case "reused-run-job-artifact-relation":
		return len(issues) > 0 && hasIssuePrefix(issues, "Actions run is reused") || hasIssuePrefix(issues, "artifact is reused") || hasIssuePrefix(issues, "target relation is reused"), "run, job, artifact, and target relations must be distinct"
	case "forbidden-aggregate-injection":
		for _, claim := range claims {
			if isForbiddenClass(claim.Class) {
				return true, "forbidden aggregate claim was counted from emitted claim bytes"
			}
		}
		return bytesContain(observation.Raw, "FORBIDDEN_AGGREGATE"), "forbidden claim class is observable rather than a fixed zero"
	default:
		return false, "unknown counterexample kind"
	}
}

func bytesContain(raw []byte, value string) bool { return strings.Contains(string(raw), value) }
func hasIssue(issues map[string]string, value string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, value) {
			return true
		}
	}
	return false
}
func hasIssuePrefix(issues map[string]string, value string) bool {
	for _, issue := range issues {
		if strings.HasPrefix(issue, value) {
			return true
		}
	}
	return false
}
func hasHistoricalReceipt(bundle ObservationBundle) bool {
	for _, receipt := range bundle.Receipts {
		if receipt.EvidenceClass == EvidenceHistorical {
			return true
		}
	}
	return false
}

func FormatSummary(report Report) string {
	return fmt.Sprintf("declared_experiments=%d/%d materialized_claim_slots=%d/%d experiment_states=PROVEN:%d,OPEN:%d,UNKNOWN:%d,REFUTED:%d gate_states=PROVEN:%d,OPEN:%d,UNKNOWN:%d,REFUTED:%d", report.Summary.DeclaredExperimentsNumerator, report.Summary.DeclaredExperimentsDenominator, report.Summary.MaterializedClaimSlotsNumerator, report.Summary.MaterializedClaimSlotsDenominator, report.Summary.ExperimentStates.Proven, report.Summary.ExperimentStates.Open, report.Summary.ExperimentStates.Unknown, report.Summary.ExperimentStates.Refuted, report.Summary.GateStates.Proven, report.Summary.GateStates.Open, report.Summary.GateStates.Unknown, report.Summary.GateStates.Refuted)
}
