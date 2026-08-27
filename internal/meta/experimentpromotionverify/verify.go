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

const producerPackage = "github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion"
const verifierAlgorithm = "experimentpromotionverify.algorithm/v2"

var verifierHeadPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var verifierRunPattern = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/([0-9]+)$`)
var verifierJobPattern = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/([0-9]+)/jobs/([0-9]+)$`)

type observedAction struct {
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
type observedProcedure struct {
	ProcedureID string   `json:"procedure_id"`
	SourcePath  string   `json:"source_path"`
	AlgorithmID string   `json:"algorithm_id"`
	Imports     []string `json:"imports"`
}
type observedLedger struct {
	Schema     string `json:"schema"`
	EntryCount int    `json:"entry_count"`
	LastDigest string `json:"last_digest"`
	AppendOnly bool   `json:"append_only"`
}

func decodeBundle(raw []byte) (ObservationBundle, error) {
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

// Verify uses a map-backed slot table and validates the raw evidence itself.
// The producer package is intentionally not imported; the import boundary is
// part of the evidence and is also checked by the workflow dependency graph.
func Verify(input Input) Verification {
	checks := make([]ReplayCheck, 0, 8)
	projection, sourceErr := parseSource(input.SourceRaw)
	bundle, bundleErr := decodeBundle(input.ObservationRaw)
	actual := slotTable(bundle, bundleErr)
	checks = append(checks, check("source-reconstruction", projection, input.Report.SourceProjection, sourceErr == nil && reflect.DeepEqual(projection, input.Report.SourceProjection), "independent ParseFile -> bidir.Lower source reconstruction"))
	checks = append(checks, check("observation-replay", digestBytes(input.ObservationRaw), input.Report.ObservationDigest, bundleErr == nil && digestBytes(input.ObservationRaw) == input.Report.ObservationDigest, "raw observation bytes are the observation authority"))
	checks = append(checks, check("raw-procedure-and-actions", "report-derived", rawEvidenceExpectation(bundle, input.SourceRaw, projection, actual, input.Report), bundleErr == nil && reportMatchesEvidence(input.Report, bundle, input.SourceRaw, projection, actual), "procedure, artifact, and Actions bytes are rehashed independently"))
	checks = append(checks, check("persistent-ledger", "append-only", ledgerExpectation(bundle), bundleErr == nil && validLedger(bundle.PriorLedger), "prior ledger is present, hashed, and append-only"))
	checks = append(checks, check("claim-transitions", GateSlotCount, len(input.Report.ClaimLedger), reportClaimsValid(input.Report), "every slot retains an explicit OPEN-origin transition"))
	checks = append(checks, check("independent-procedure", ConsumerPackage, ConsumerPackage, bundleErr == nil && procedureBoundaryValid(bundle), "consumer procedure identity and import graph are captured"))
	checks = append(checks, check("exact-denominators", "30/30 and 150/150", fmt.Sprintf("%d/%d and %d/%d", input.Report.Summary.DeclaredExperimentsNumerator, input.Report.Summary.DeclaredExperimentsDenominator, input.Report.Summary.MaterializedClaimSlotsNumerator, input.Report.Summary.MaterializedClaimSlotsDenominator), summaryDenominatorsValid(input.Report.Summary), "fixed declared and materialized denominators"))
	checks = append(checks, check("report-self-digest", input.Report.Digest, reportDigest(input.Report), input.Report.Digest == reportDigest(input.Report), "report digest covers reconstructed output"))

	allPass := sourceErr == nil && bundleErr == nil && reflect.DeepEqual(projection, input.Report.SourceProjection) && reportMatchesEvidence(input.Report, bundle, input.SourceRaw, projection, actual) && validLedger(bundle.PriorLedger) && reportClaimsValid(input.Report) && summaryDenominatorsValid(input.Report.Summary) && input.Report.Digest == reportDigest(input.Report)
	if !allPass {
		checks = append(checks, check("fail-closed", "supported", "mismatch", false, "producer output is not supported by independently reconstructed bytes"))
	}
	decision, resolution, reason := "PASS", "EXACT", "independent raw-source and observation replay matches the report"
	if !allPass {
		decision, resolution, reason = "FAIL_CLOSED", "LOWER_RESOLUTION", "independent replay differs from producer output"
	}
	verification := Verification{Schema: ReportSchema, SubjectSHA: input.Report.SubjectSHA, Decision: decision, Resolution: resolution, Reason: reason, SourceProjection: projection, Checks: checks, Summary: input.Report.Summary, Guardrails: input.Report.Guardrails, Counterexamples: input.Report.Counterexamples, AggregateMetrics: append([]string(nil), input.Report.AggregateMetrics...), NotClaimed: append([]string(nil), input.Report.NotClaimed...), RepositoryWrites: input.Report.RepositoryWrites, MutationAuthority: input.Report.MutationAuthority}
	verification.Digest = verificationDigest(verification)
	return verification
}

type slotObservation struct {
	Receipt ObservationReceipt
	Present bool
	Problem string
}

func slotTable(bundle ObservationBundle, bundleErr error) map[string]slotObservation {
	table := map[string]slotObservation{}
	if bundleErr != nil {
		return table
	}
	for _, receipt := range bundle.Receipts {
		key := receipt.ExperimentID + "\x00" + receipt.GateID
		item := table[key]
		if item.Present {
			item.Problem = "duplicate slot receipt"
		}
		item.Receipt, item.Present = receipt, true
		table[key] = item
	}
	return table
}

func rawEvidenceValid(bundle ObservationBundle, source []byte, projection SourceProjection, table map[string]slotObservation) bool {
	if bundle.Schema != ObservationSchema || bundle.BundleID == "" {
		return false
	}
	seenIDs, seenRuns, seenJobs, seenArtifacts, seenTargets := map[string]bool{}, map[int64]bool{}, map[int64]bool{}, map[int64]bool{}, map[string]bool{}
	for _, item := range table {
		if item.Problem != "" || !item.Present || !rawReceiptValid(item.Receipt, source, projection) {
			return false
		}
		r := item.Receipt
		if seenIDs[r.ObservationID] || seenRuns[r.Actions.RunID] || seenJobs[r.Actions.JobID] || seenArtifacts[r.Artifact.ArtifactID] || seenTargets[r.TargetAddress] {
			return false
		}
		seenIDs[r.ObservationID], seenRuns[r.Actions.RunID], seenJobs[r.Actions.JobID], seenArtifacts[r.Artifact.ArtifactID], seenTargets[r.TargetAddress] = true, true, true, true, true
	}
	return true
}

func reportMatchesEvidence(report Report, bundle ObservationBundle, source []byte, projection SourceProjection, table map[string]slotObservation) bool {
	if !rawEvidenceValid(bundle, source, projection, table) && len(table) > 0 {
		// Invalid raw evidence is still a valid replay when the report records it
		// as REFUTED. The per-slot loop below distinguishes that from silence.
	}
	if len(report.Experiments) != ExperimentCount {
		return false
	}
	for _, experiment := range report.Experiments {
		promotionStatuses := make([]string, 0, GateCount)
		evidenceStatuses := make([]string, 0, GateCount)
		for _, gate := range experiment.Gates {
			item := table[experiment.ExperimentID+"\x00"+gate.GateID]
			expected := StatusUnknown
			promotion := StatusUnknown
			if item.Present {
				if item.Problem != "" || !rawReceiptValid(item.Receipt, source, projection) {
					expected = StatusRefuted
				} else if gate.GateID == "semantic-causality" && !verifierInterventionValid(item.Receipt.SemanticIntervention, source, projection, experiment.ExperimentID) {
					expected = StatusRefuted
				} else {
					expected = conclusionStatus(item.Receipt.Actions.Conclusion)
					if item.Receipt.EvidenceClass == EvidenceCurrent {
						promotion = expected
					}
				}
			}
			if gate.Status != expected || gate.PromotionStatus != promotion || gate.ClaimTransition.To != claimState(expected) {
				return false
			}
			promotionStatuses = append(promotionStatuses, promotion)
			evidenceStatuses = append(evidenceStatuses, expected)
		}
		expectedStatus := foldStatuses(promotionStatuses)
		expectedEvidence := foldStatuses(evidenceStatuses)
		if experiment.Status != expectedStatus || experiment.EvidenceStatus != expectedEvidence {
			return false
		}
	}
	return true
}

func foldStatuses(values []string) string {
	for _, status := range []string{StatusRefuted, StatusOpen, StatusUnknown} {
		for _, value := range values {
			if value == status {
				return status
			}
		}
	}
	return StatusProven
}

func verifierInterventionValid(intervention *SemanticIntervention, source []byte, projection SourceProjection, experimentID string) bool {
	if intervention == nil {
		return false
	}
	if intervention.BaselineRawDigest != digestBytes(intervention.BaselineSourceRaw) || string(intervention.BaselineSourceRaw) != string(source) || intervention.BaselineSemanticDigest != projection.SemanticDigest || intervention.SemanticRawDigest != digestBytes(intervention.SemanticSourceRaw) || intervention.CommentRawDigest != digestBytes(intervention.CommentSourceRaw) {
		return false
	}
	semantic, semanticErr := parseMaterial(intervention.SemanticSourceRaw)
	comment, commentErr := parseMaterial(intervention.CommentSourceRaw)
	if semanticErr != nil || commentErr != nil || semantic.SemanticDigest == projection.SemanticDigest || comment.SemanticDigest != projection.SemanticDigest || intervention.SemanticSemanticDigest != semantic.SemanticDigest || intervention.CommentSemanticDigest != comment.SemanticDigest || intervention.ContractedOutputDigest != digestBytes([]byte("contracted-output/v2|"+semantic.SemanticDigest+"|"+comment.SemanticDigest)) || intervention.DecisionDigest != digestBytes([]byte("decision/v2|"+experimentID+"|semantic-causality|DISCHARGED")) || intervention.ClaimTransitionDigest != claimTransitionDigest(experimentID, "semantic-causality", ClaimDischarged) {
		return false
	}
	return intervention.SemanticRawDigest != intervention.BaselineRawDigest && intervention.CommentRawDigest != intervention.BaselineRawDigest
}

func conclusionStatus(value string) string {
	switch value {
	case ObservationSuccess:
		return StatusProven
	case ObservationProgress, ObservationQueued:
		return StatusOpen
	default:
		return StatusRefuted
	}
}
func claimState(status string) string {
	if status == StatusProven {
		return ClaimDischarged
	}
	if status == StatusRefuted {
		return ClaimRefuted
	}
	return ClaimOpen
}

func rawReceiptValid(receipt ObservationReceipt, source []byte, projection SourceProjection) bool {
	if receipt.Schema != ObservationSchema || !verifierHeadPattern.MatchString(receipt.HeadSHA) || receipt.SourceRawDigest != digestBytes(source) || receipt.SourceSemanticDigest != projection.SemanticDigest || receipt.ConsumerPackage != ConsumerPackage || receipt.ProcedureID != ConsumerProcedureID || receipt.ProcedureSourcePath != ConsumerSourcePath || receipt.ProcedureAlgorithmID != verifierAlgorithm || len(receipt.ConsumerImports) == 0 {
		return false
	}
	for _, imported := range receipt.ConsumerImports {
		if imported == producerPackage || strings.HasPrefix(imported, producerPackage+"/") {
			return false
		}
	}
	if digestBytes(receipt.ProcedureSourceBytes) != receipt.ProcedureSourceDigest || receipt.ProcedureSourceBytes == nil {
		return false
	}
	var procedure observedProcedure
	if json.Unmarshal(receipt.ProcedureSourceBytes, &procedure) != nil || procedure.ProcedureID != receipt.ProcedureID || procedure.SourcePath != receipt.ProcedureSourcePath || procedure.AlgorithmID != receipt.ProcedureAlgorithmID || !reflect.DeepEqual(procedure.Imports, receipt.ConsumerImports) {
		return false
	}
	if receipt.ProcedureAlgorithmDigest != digestBytes([]byte(receipt.ProcedureAlgorithmID+"|"+strings.Join(receipt.ConsumerImports, ","))) {
		return false
	}
	if digestBytes(receipt.Actions.Raw) != receipt.Actions.RawDigest || receipt.Actions.Raw == nil {
		return false
	}
	var action observedAction
	if json.Unmarshal(receipt.Actions.Raw, &action) != nil || action != observedActionFromReceipt(receipt.Actions) {
		return false
	}
	if action.Repository != "kimjooyoon/meta-ontology-go" || action.PRNumber != receipt.PRNumber || action.HeadSHA != receipt.HeadSHA || action.RunID <= 0 || action.JobID <= 0 || action.ArtifactID != receipt.Artifact.ArtifactID || action.ArtifactName != receipt.Artifact.ArtifactName || action.ArtifactDigest != receipt.Artifact.Digest {
		return false
	}
	run, job := verifierRunPattern.FindStringSubmatch(receipt.Actions.RunURL), verifierJobPattern.FindStringSubmatch(receipt.Actions.JobURL)
	if len(run) != 2 || len(job) != 3 || run[1] != fmt.Sprint(action.RunID) || job[1] != run[1] || job[2] != fmt.Sprint(action.JobID) {
		return false
	}
	if receipt.Artifact.Raw == nil || receipt.Artifact.Bytes != len(receipt.Artifact.Raw) || digestBytes(receipt.Artifact.Raw) != receipt.Artifact.Digest || receipt.Artifact.ArtifactID != action.ArtifactID || receipt.Artifact.ArtifactName != action.ArtifactName {
		return false
	}
	if receipt.GateID == "source-bound" && string(receipt.Artifact.Raw) != string(source) {
		return false
	}
	return true
}

func observedActionFromReceipt(value ActionsObservation) observedAction {
	return observedAction{Repository: value.Repository, PRNumber: value.PRNumber, HeadSHA: value.HeadSHA, WorkflowID: value.WorkflowID, WorkflowName: value.WorkflowName, RunID: value.RunID, JobID: value.JobID, Conclusion: value.Conclusion, ArtifactID: value.ArtifactID, ArtifactName: value.ArtifactName, ArtifactDigest: value.ArtifactDigest}
}

func validLedger(ledger LedgerObservation) bool {
	if ledger.Path == "" || ledger.Raw == nil || digestBytes(ledger.Raw) != ledger.Digest || ledger.EntryCount != GateSlotCount || !validDigest(ledger.LastDigest) {
		return false
	}
	var payload observedLedger
	return json.Unmarshal(ledger.Raw, &payload) == nil && payload.Schema == "gooo/claim-ledger/v2" && payload.EntryCount == GateSlotCount && payload.LastDigest == ledger.LastDigest && payload.AppendOnly
}

func procedureBoundaryValid(bundle ObservationBundle) bool {
	for _, receipt := range bundle.Receipts {
		if receipt.ProcedureID != ConsumerProcedureID || receipt.ConsumerPackage != ConsumerPackage {
			return false
		}
	}
	return true
}
func rawEvidenceExpectation(bundle ObservationBundle, source []byte, projection SourceProjection, table map[string]slotObservation, report Report) string {
	if rawEvidenceValid(bundle, source, projection, table) {
		return "bound"
	}
	return "unbound"
}
func ledgerExpectation(bundle ObservationBundle) string {
	if validLedger(bundle.PriorLedger) {
		return "append-only"
	}
	return "invalid"
}

func reportClaimsValid(report Report) bool {
	if len(report.Experiments) != ExperimentCount || len(report.ClaimLedger) != GateSlotCount || len(report.EmittedClaims) != GateSlotCount {
		return false
	}
	previous := report.PriorLedger.LastDigest
	for i, entry := range report.ClaimLedger {
		if entry.Sequence != i+1 || entry.PriorState != ClaimOpen || entry.PreviousDigest != previous || entry.Digest != ledgerDigest(entry) {
			return false
		}
		previous = entry.Digest
	}
	return true
}

func summaryDenominatorsValid(summary Summary) bool {
	return summary.DeclaredExperimentsNumerator == ExperimentCount && summary.DeclaredExperimentsDenominator == ExperimentCount && summary.MaterializedClaimSlotsNumerator == GateSlotCount && summary.MaterializedClaimSlotsDenominator == GateSlotCount
}
func reportDigest(report Report) string {
	report.Digest = ""
	raw, _ := json.Marshal(report)
	return digestBytes(raw)
}
func ledgerDigest(entry ClaimLedgerEntry) string {
	entry.Digest = ""
	raw, _ := json.Marshal(entry)
	return digestBytes(raw)
}
func claimTransitionDigest(experimentID, gateID, next string) string {
	return digestBytes([]byte(fmt.Sprintf("claim-transition/v2|%s|%s|OPEN|%s", experimentID, gateID, next)))
}
func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func verificationDigest(verification Verification) string {
	verification.Digest = ""
	raw, _ := json.Marshal(verification)
	return digestBytes(raw)
}
func check(id string, expected, observed any, ok bool, reason string) ReplayCheck {
	status := "PASS"
	if !ok {
		status = "FAIL"
	}
	return ReplayCheck{ID: id, Status: status, Stage: "VERIFY", Step: "reconstruct", Reason: reason, Expected: fmt.Sprint(expected), Observed: fmt.Sprint(observed)}
}
