package evidencequorumconsumer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

type sourceExecutionReceipt struct {
	Schema         string                      `json:"schema"`
	Decision       string                      `json:"decision"`
	Reason         string                      `json:"reason"`
	Resolution     string                      `json:"resolution"`
	Filename       string                      `json:"filename"`
	SourceDigest   string                      `json:"source_digest"`
	SemanticDigest string                      `json:"semantic_digest,omitempty"`
	Entry          sourceExecutionEntry        `json:"entry"`
	Events         []sourceExecutionEvent      `json:"events"`
	Diagnostics    []sourceExecutionDiagnostic `json:"diagnostics"`
	Effects        sourceExecutionEffects      `json:"effects"`
	Digest         string                      `json:"digest"`
}

type sourceExecutionEntry struct {
	Package   string                   `json:"package"`
	Namespace string                   `json:"namespace"`
	Activity  string                   `json:"activity"`
	Inputs    []sourceExecutionBinding `json:"inputs"`
	Output    sourceExecutionBinding   `json:"output"`
}
type sourceExecutionBinding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
type sourceExecutionEvent struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Subject  string `json:"subject"`
}
type sourceExecutionDiagnostic struct {
	Stage   string `json:"stage"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
type sourceExecutionEffects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

func Evaluate(input Input) Report {
	sourceRawDigest := evidencequorumwire.DigestBytes(input.Source)
	sourceSemanticDigest, sourceErr := evidencequorumpolicy.SemanticDigest(input.SourcePath, input.Source)
	report := Report{
		Schema:               ReportSchema,
		Scope:                Scope,
		HeadSHA:              input.HeadSHA,
		SourcePath:           input.SourcePath,
		SourceEntry:          input.Policy.SourceEntry,
		SourceRawDigest:      sourceRawDigest,
		SourceSemanticDigest: sourceSemanticDigest,
		PolicySemanticDigest: input.Policy.SemanticDigest,
		NotClaimed: []string{
			"confidence-weighted voting or confidence averaging",
			"full Byzantine consensus",
			"full compiler semantic correctness",
			"identity or trust of a self-reported origin label",
			"repository mutation or side effects",
		},
	}
	if sourceErr != nil {
		report.Reason = "SOURCE_RECONSTRUCTION_FAILED"
	} else {
		for _, item := range input.Cases {
			report.Cases = append(report.Cases, evaluateCase(input, item, sourceRawDigest, sourceSemanticDigest))
		}
	}
	report.Summary = summarize(report.Cases, input.Policy)
	report.Decision = DecisionClosed
	report.SubjectDecision = DecisionClosed
	report.Resolution = ResolutionInvariant
	report.Reason = "EVIDENCE_QUORUM_CASE_CONFORMANCE_FAILED"
	if sourceErr == nil && len(input.Cases) == input.Policy.CaseDenominator && report.Summary.CasesSatisfied == report.Summary.CasesTotal {
		report.Decision = DecisionPass
		report.Reason = "EVIDENCE_QUORUM_CASES_CONFORM"
	}
	if len(report.Cases) > 0 {
		report.SubjectDecision = report.Cases[0].SubjectDecision
		report.Resolution = report.Cases[0].Resolution
	}
	report.ReceiptDigests, report.Summary.RawReceiptsTotal, report.Summary.CurrentEvidenceTotal,
		report.Summary.SyntheticEvidenceTotal, report.Summary.DistinctProvenanceGroups,
		report.Summary.CollapsedReplicas = inventory(input.Cases)
	report.Summary.SourceReconstructionTotal = 1
	report.Summary.ProducerPackageImportTotal = 1
	if sourceErr == nil {
		report.Summary.SourceReconstructionCount = 1
	}
	report.Digest = reportDigest(report)
	return report
}

func evaluateCase(input Input, item CaseInput, sourceRawDigest, sourceSemanticDigest string) CaseResult {
	result := CaseResult{ID: item.ID, Status: "NOT_SATISFIED", ConformanceDecision: DecisionClosed,
		SubjectState: StatusOpen, SubjectDecision: DecisionClosed, ObservationState: ObservationExact,
		Resolution: ResolutionInvariant, Stage: "QUORUM_DECISION", Step: "receipt-integrity"}
	var evidence []classifiedEvidence
	invalid := false
	for _, raw := range item.Receipts {
		receipt, err := DecodeReceipt(raw)
		if err != nil || !evidencequorumwire.Verify(receipt) {
			invalid = true
			continue
		}
		role, value, predicate, ok := classify(receipt, input.Policy)
		if !ok || !validReceipt(receipt, input, sourceRawDigest, sourceSemanticDigest) {
			invalid = true
			continue
		}
		if value == "CONTRADICTS" && predicate != input.Policy.ContradictionPredicate {
			invalid = true
		}
		lineage := receipt.ExecutableDigest + "|" + receipt.DependencyDigest
		evidence = append(evidence, classifiedEvidence{Receipt: receipt, Digest: receipt.Digest, Role: role,
			Value: value, Predicate: predicate, Provenance: provenance(receipt, lineage)})
	}
	result.RawReceipts = len(item.Receipts)
	for _, item := range evidence {
		if item.Receipt.EvidenceClass == evidencequorumwire.CurrentEvidence {
			result.CurrentEvidence++
		} else {
			result.SyntheticEvidence++
		}
	}
	result.Groups = groupResults(evidence)
	result.IndependentGroups = currentGroupCount(evidence)
	result.CollapsedReplicas = len(evidence) - len(result.Groups)
	result.ConflictGroups = validConflictCount(evidence, input.Policy)
	result.Claims = claimResults(input.Policy, evidence, result)
	if invalid {
		return finishCase(result, StatusOpen, DecisionClosed, ObservationExact, ResolutionLower,
			"EVIDENCE_RECEIPT_INVALID", "receipt-integrity", input.Policy)
	}
	if hasUnknown(evidence) {
		return finishCase(result, StatusOpen, DecisionUnknown, ObservationUnknown, ResolutionLower,
			"QUORUM_EVIDENCE_UNKNOWN", "UNKNOWN", input.Policy)
	}
	if result.ConflictGroups > 0 {
		return finishCase(result, StatusRefuted, DecisionClosed, ObservationExact, ResolutionInvariant,
			"QUORUM_VALID_CONTRADICTION", "valid-contradiction", input.Policy)
	}
	if result.IndependentGroups < input.Policy.Threshold || !hasRequiredRoles(evidence, input.Policy.RequiredRoles) ||
		!hasRequiredPredicates(evidence, input.Policy.RequiredPredicates) {
		return finishCase(result, StatusOpen, DecisionClosed, ObservationExact, ResolutionLower,
			"QUORUM_INSUFFICIENT_INDEPENDENT_GROUPS", "minimum-independent-groups", input.Policy)
	}
	return finishCase(result, StatusDischarged, DecisionPass, ObservationExact, ResolutionExact,
		"QUORUM_CLAIM_DISCHARGED", "independent-provenance-quorum", input.Policy)
}

func classify(receipt evidencequorumwire.Receipt, policy evidencequorumpolicy.Policy) (string, string, string, bool) {
	switch receipt.Channel {
	case "gooo-source-execution":
		return "producer", "SUPPORTS", "SOURCE_ACTIVITY_EXECUTED", receipt.Predicate == "SOURCE_ACTIVITY_EXECUTED"
	case "raw-source-semantic-reconstructor":
		return "consumer", "SUPPORTS", "RAW_SOURCE_SEMANTIC_RECONSTRUCTED", receipt.Predicate == "RAW_SOURCE_SEMANTIC_RECONSTRUCTED"
	case "generated-artifact-observer":
		return "meta-operation", "SUPPORTS", "GENERATED_ARTIFACT_OBSERVED", receipt.Predicate == "GENERATED_ARTIFACT_OBSERVED"
	case "synthetic-duplicate":
		return "producer", "SUPPORTS", "DUPLICATE_REPLICA", receipt.Predicate == "DUPLICATE_REPLICA"
	case "synthetic-valid-conflict":
		return "meta-operation", "CONTRADICTS", receipt.Predicate, receipt.Predicate == policy.ContradictionPredicate
	case "synthetic-invalid-conflict":
		return "meta-operation", "CONTRADICTS", receipt.Predicate, true
	case "synthetic-unknown":
		return "producer", "UNKNOWN", "UNKNOWN_OBSERVATION", receipt.Predicate == "UNKNOWN_OBSERVATION"
	default:
		return "", "", "", false
	}
}

func validReceipt(receipt evidencequorumwire.Receipt, input Input, sourceRawDigest, sourceSemanticDigest string) bool {
	if receipt.Schema != evidencequorumwire.Schema || receipt.HeadSHA != input.HeadSHA || receipt.SourcePath != input.SourcePath ||
		receipt.SubjectRawDigest != sourceRawDigest || receipt.SubjectSemanticDigest != sourceSemanticDigest ||
		receipt.PolicySemanticDigest != input.Policy.SemanticDigest || !validDigest(receipt.ExecutableDigest) ||
		!validDigest(receipt.DependencyDigest) || !validDigest(receipt.ObservationDigest) ||
		receipt.DependencyDigest != evidencequorumwire.DependencyDigest(receipt.DependencyPaths) ||
		receipt.Producer != input.Policy.Claim.Producer || receipt.Consumer != input.Policy.Claim.Consumer ||
		receipt.MetaOperation != input.Policy.Claim.MetaOperation || receipt.ProofChoice != input.Policy.Claim.ProofChoice ||
		receipt.RepositoryWrites != 0 || receipt.MutationAuthority || receipt.ObservationDigest != evidencequorumwire.ObservationDigest(receipt) {
		return false
	}
	for index := 1; index < len(receipt.DependencyPaths); index++ {
		if receipt.DependencyPaths[index-1] >= receipt.DependencyPaths[index] {
			return false
		}
	}
	if receipt.EvidenceClass != evidencequorumwire.CurrentEvidence && receipt.EvidenceClass != evidencequorumwire.SyntheticCounterexample {
		return false
	}
	if receipt.Channel == "gooo-source-execution" {
		return validSourceExecution(receipt, input, sourceRawDigest, sourceSemanticDigest)
	}
	return true
}

func validSourceExecution(receipt evidencequorumwire.Receipt, input Input, sourceRawDigest, sourceSemanticDigest string) bool {
	if receipt.SourceExecutionReceiptDigest == "" || len(receipt.SourceExecutionReceipt) == 0 {
		return false
	}
	var nested sourceExecutionReceipt
	if err := json.Unmarshal(receipt.SourceExecutionReceipt, &nested); err != nil {
		return false
	}
	if nested.Digest != receipt.SourceExecutionReceiptDigest || !validDigest(nested.Digest) ||
		nested.Schema != "gooo/source-execution-receipt/v1" || nested.Decision != DecisionPass ||
		nested.Reason != "SOURCE_ACTIVITY_EXECUTED" || nested.Resolution != ResolutionExact ||
		nested.Filename != input.SourcePath || nested.SourceDigest != sourceRawDigest || nested.SemanticDigest != sourceSemanticDigest ||
		nested.Entry.Activity != input.Policy.SourceEntry || len(nested.Events) != 4 || len(nested.Diagnostics) != 0 ||
		nested.Effects.RepositoryWrites != 0 || nested.Effects.MutationAuthority {
		return false
	}
	digest := nested.Digest
	nested.Digest = ""
	return digest == evidencequorumwire.DigestJSON(nested)
}

func provenance(receipt evidencequorumwire.Receipt, lineage string) Provenance {
	return Provenance{OriginGroup: "provenance:" + strings.TrimPrefix(evidencequorumwire.DigestBytes([]byte(lineage)), "sha256:"),
		LineageKey: lineage, ExecutableDigest: receipt.ExecutableDigest, DependencyPaths: append([]string(nil), receipt.DependencyPaths...),
		DependencyDigest: receipt.DependencyDigest, SubjectRawDigest: receipt.SubjectRawDigest,
		SubjectSemanticDigest: receipt.SubjectSemanticDigest, ObservationDigest: receipt.ObservationDigest}
}

func groupResults(evidence []classifiedEvidence) []GroupResult {
	groups := map[string]*GroupResult{}
	for _, item := range evidence {
		group := groups[item.Provenance.OriginGroup]
		if group == nil {
			group = &GroupResult{OriginGroup: item.Provenance.OriginGroup, Provenance: item.Provenance, Independent: true}
			groups[group.OriginGroup] = group
		}
		group.EvidenceIDs = append(group.EvidenceIDs, item.Digest)
		group.Roles = appendUnique(group.Roles, item.Role)
		group.Values = appendUnique(group.Values, item.Value)
		group.EvidenceClasses = appendUnique(group.EvidenceClasses, item.Receipt.EvidenceClass)
	}
	result := make([]GroupResult, 0, len(groups))
	for _, group := range groups {
		sort.Strings(group.EvidenceIDs)
		sort.Strings(group.Roles)
		sort.Strings(group.Values)
		sort.Strings(group.EvidenceClasses)
		result = append(result, *group)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].OriginGroup < result[j].OriginGroup })
	return result
}

func currentGroupCount(evidence []classifiedEvidence) int {
	groups := map[string]bool{}
	for _, item := range evidence {
		if item.Receipt.EvidenceClass == evidencequorumwire.CurrentEvidence && item.Value == "SUPPORTS" {
			groups[item.Provenance.OriginGroup] = true
		}
	}
	return len(groups)
}

func validConflictCount(evidence []classifiedEvidence, policy evidencequorumpolicy.Policy) int {
	count := 0
	for _, item := range evidence {
		if item.Value == "CONTRADICTS" && item.Predicate == policy.ContradictionPredicate {
			count++
		}
	}
	return count
}

func hasUnknown(evidence []classifiedEvidence) bool {
	for _, item := range evidence {
		if item.Value == "UNKNOWN" {
			return true
		}
	}
	return false
}

func hasRequiredRoles(evidence []classifiedEvidence, required []string) bool {
	seen := map[string]bool{}
	for _, item := range evidence {
		if item.Receipt.EvidenceClass == evidencequorumwire.CurrentEvidence && item.Value == "SUPPORTS" {
			seen[item.Role] = true
		}
	}
	for _, role := range required {
		if !seen[role] {
			return false
		}
	}
	return true
}

func hasRequiredPredicates(evidence []classifiedEvidence, required []string) bool {
	seen := map[string]bool{}
	for _, item := range evidence {
		if item.Receipt.EvidenceClass == evidencequorumwire.CurrentEvidence && item.Value == "SUPPORTS" {
			seen[item.Predicate] = true
		}
	}
	for _, predicate := range required {
		if !seen[predicate] {
			return false
		}
	}
	return true
}

func claimResults(policy evidencequorumpolicy.Policy, evidence []classifiedEvidence, result CaseResult) []ClaimResult {
	digests := make([]string, 0, len(evidence))
	provenanceValues := make([]Provenance, 0)
	seen := map[string]bool{}
	for _, item := range evidence {
		digests = append(digests, item.Digest)
		if !seen[item.Provenance.OriginGroup] {
			seen[item.Provenance.OriginGroup] = true
			provenanceValues = append(provenanceValues, item.Provenance)
		}
	}
	sort.Strings(digests)
	state, decision, observation, resolution, reason, stage, step := result.SubjectState, result.SubjectDecision, result.ObservationState, result.Resolution, result.Reason, result.Stage, result.Step
	transition := ClaimTransition{From: policy.PriorClaimState, To: state, PreviousDigest: previousClaimDigest(policy.Claim.ID),
		EvidenceDigests: digests, Provenance: provenanceValues, Stage: stage, Step: step, Reason: reason}
	return []ClaimResult{{ID: policy.Claim.ID, Producer: policy.Claim.Producer, Consumer: policy.Claim.Consumer,
		MetaOperation: policy.Claim.MetaOperation, ProofChoice: policy.Claim.ProofChoice, State: state,
		SubjectDecision: decision, ObservationState: observation, Resolution: resolution, Reason: reason,
		Stage: stage, Step: step, EvidenceDigests: digests, Transitions: []ClaimTransition{transition}}}
}

func finishCase(result CaseResult, state, decision, observation, resolution, reason, step string, policy evidencequorumpolicy.Policy) CaseResult {
	result.SubjectState, result.SubjectDecision, result.ObservationState = state, decision, observation
	result.Resolution, result.Reason, result.Step = resolution, reason, step
	if observation == ObservationUnknown {
		result.Stage = "UNKNOWN"
	}
	result.Claims = claimResults(policy, nil, result)
	if len(result.Claims) == 1 {
		// Rebuild the transition with the evidence digests captured before the
		// final state was selected.
		result.Claims[0].EvidenceDigests = claimEvidence(result.Groups)
		result.Claims[0].Transitions[0].EvidenceDigests = append([]string(nil), result.Claims[0].EvidenceDigests...)
		result.Claims[0].Transitions[0].Provenance = groupProvenance(result.Groups)
	}
	if expectedCase(result.ID, policy) == (resultExpectation{State: state, Decision: decision, Observation: observation, Resolution: resolution, Reason: reason}) {
		result.Status = "SATISFIED"
		result.ConformanceDecision = DecisionPass
	}
	return result
}

type resultExpectation struct{ State, Decision, Observation, Resolution, Reason string }

func expectedCase(id string, _ evidencequorumpolicy.Policy) resultExpectation {
	switch id {
	case "current-quorum", "synthetic-duplicate":
		return resultExpectation{StatusDischarged, DecisionPass, ObservationExact, ResolutionExact, "QUORUM_CLAIM_DISCHARGED"}
	case "synthetic-valid-conflict":
		return resultExpectation{StatusRefuted, DecisionClosed, ObservationExact, ResolutionInvariant, "QUORUM_VALID_CONTRADICTION"}
	case "synthetic-invalid-conflict", "insufficient-current":
		return resultExpectation{StatusOpen, DecisionClosed, ObservationExact, ResolutionLower, map[string]string{"synthetic-invalid-conflict": "EVIDENCE_RECEIPT_INVALID", "insufficient-current": "QUORUM_INSUFFICIENT_INDEPENDENT_GROUPS"}[id]}
	case "synthetic-unknown":
		return resultExpectation{StatusOpen, DecisionUnknown, ObservationUnknown, ResolutionLower, "QUORUM_EVIDENCE_UNKNOWN"}
	default:
		return resultExpectation{}
	}
}

func claimEvidence(groups []GroupResult) []string {
	var result []string
	for _, group := range groups {
		result = append(result, group.EvidenceIDs...)
	}
	sort.Strings(result)
	return result
}
func groupProvenance(groups []GroupResult) []Provenance {
	result := make([]Provenance, 0, len(groups))
	for _, group := range groups {
		result = append(result, group.Provenance)
	}
	return result
}

func summarize(cases []CaseResult, policy evidencequorumpolicy.Policy) Summary {
	summary := Summary{CasesTotal: len(cases), ClaimsTotal: len(cases), MinimumIndependentGroups: policy.Threshold}
	for _, item := range cases {
		if item.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		if item.SubjectState == StatusDischarged {
			summary.DischargedClaims++
			summary.QuorumSatisfiedCases++
		}
		if item.SubjectState == StatusOpen {
			summary.OpenClaims++
		}
		if item.SubjectState == StatusRefuted {
			summary.RefutedClaims++
			summary.ConflictCases++
		}
		if item.Resolution == ResolutionLower {
			summary.LowerResolutionCases++
		}
		if item.ObservationState == ObservationUnknown {
			summary.UnknownObservationCases++
		}
	}
	return summary
}

func inventory(cases []CaseInput) (digests []string, raw, current, synthetic, groups, collapsed int) {
	seenReceipts := map[string]evidencequorumwire.Receipt{}
	seenGroups := map[string]bool{}
	for _, item := range cases {
		raw += len(item.Receipts)
		for _, bytes := range item.Receipts {
			receipt, err := DecodeReceipt(bytes)
			if err != nil || !evidencequorumwire.Verify(receipt) {
				continue
			}
			seenReceipts[receipt.Digest] = receipt
			lineage := receipt.ExecutableDigest + "|" + receipt.DependencyDigest
			seenGroups[lineage] = true
		}
	}
	for digest, receipt := range seenReceipts {
		digests = append(digests, digest)
		if receipt.EvidenceClass == evidencequorumwire.CurrentEvidence {
			current++
		}
		if receipt.EvidenceClass == evidencequorumwire.SyntheticCounterexample {
			synthetic++
		}
	}
	sort.Strings(digests)
	groups = len(seenGroups)
	collapsed = len(seenReceipts) - groups
	return
}

func validDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && len(value) == len("sha256:")+64
}
func appendUnique(values []string, value string) []string {
	for _, old := range values {
		if old == value {
			return values
		}
	}
	return append(values, value)
}
func reportError(message string) error { return fmt.Errorf("evidence quorum: %s", message) }
