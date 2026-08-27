package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	receiptSchema = "gooo/claim-lifecycle-calculus/v2"
	metaOperation = "preserve-claim-lifecycle"
	producerID    = "claimlifecyclecalculus.Evaluate"
	consumerID    = "claim-lifecycle-calculus-judge"
	caseTotal     = 6
)

type SourceCase struct {
	CaseID            string `json:"case_id"`
	ClaimID           string `json:"claim_id"`
	PriorState        string `json:"prior_state"`
	EvidenceKind      string `json:"evidence_kind"`
	EvidenceID        string `json:"evidence_id"`
	DependencyClaimID string `json:"dependency_claim_id"`
	ObservedStage     string `json:"observed_stage"`
	ObservedStep      string `json:"observed_step"`
	ObservedReason    string `json:"observed_reason"`
	ExpectedStatus    string `json:"expected_status"`
	ExpectedDecision  string `json:"expected_decision"`
	Provenance        string `json:"provenance"`
}

type EntityBinding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type ActivityBinding struct {
	Name               string     `json:"name"`
	Inputs             []string   `json:"inputs"`
	Output             string     `json:"output"`
	ValueProgram       string     `json:"value_program"`
	SemanticNodeDigest string     `json:"semantic_node_digest"`
	Case               SourceCase `json:"case"`
}

type SourceRelation struct {
	Package          string            `json:"package"`
	Namespace        string            `json:"namespace"`
	SemanticIRDigest string            `json:"semantic_ir_digest"`
	Entities         []EntityBinding   `json:"entities"`
	Activities       []ActivityBinding `json:"activities"`
	Digest           string            `json:"digest"`
}

type Claim struct {
	ID                string     `json:"id"`
	CaseID            string     `json:"case_id"`
	Statement         string     `json:"statement"`
	DeclarationDigest string     `json:"declaration_digest"`
	Status            string     `json:"status"`
	Coordinate        Coordinate `json:"coordinate"`
	Reason            string     `json:"reason"`
	EvidenceDigest    string     `json:"evidence_digest"`
	Provenance        string     `json:"provenance"`
}

type Evidence struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	ClaimID      string `json:"claim_id"`
	SourceCaseID string `json:"source_case_id"`
	Digest       string `json:"digest"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type CauseReceipt struct {
	Sequence           int        `json:"sequence"`
	ClaimID            string     `json:"claim_id"`
	Kind               string     `json:"kind"`
	EvidenceIDs        []string   `json:"evidence_ids,omitempty"`
	DependencyClaimIDs []string   `json:"dependency_claim_ids,omitempty"`
	Coordinate         Coordinate `json:"coordinate"`
	Reason             string     `json:"reason"`
	Provenance         string     `json:"provenance"`
	Digest             string     `json:"digest"`
}

type Transition struct {
	Sequence           int    `json:"sequence"`
	ClaimID            string `json:"claim_id"`
	DeclarationDigest  string `json:"declaration_digest"`
	Event              string `json:"event"`
	Before             string `json:"before"`
	After              string `json:"after"`
	EvidenceDigest     string `json:"evidence_digest,omitempty"`
	CauseReceiptDigest string `json:"cause_receipt_digest"`
	PreviousDigest     string `json:"previous_digest,omitempty"`
	Digest             string `json:"digest"`
}

type CaseResult struct {
	ID               string     `json:"id"`
	ClaimID          string     `json:"claim_id"`
	ExpectedStatus   string     `json:"expected_status"`
	ObservedStatus   string     `json:"observed_status"`
	ExpectedDecision string     `json:"expected_decision"`
	Decision         string     `json:"decision"`
	Conformance      string     `json:"conformance"`
	ObservedEvent    string     `json:"observed_event"`
	CauseKind        string     `json:"cause_kind"`
	Coordinate       Coordinate `json:"coordinate"`
	Reason           string     `json:"reason"`
	Provenance       string     `json:"provenance"`
}

type Metric struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Satisfied     bool   `json:"satisfied"`
}

type CaseCounts struct {
	Pass       int `json:"pass"`
	FailClosed int `json:"fail_closed"`
	Unknown    int `json:"unknown"`
}

type Summary struct {
	ClaimsTotal            int        `json:"claims_total"`
	TransitionsTotal       int        `json:"transitions_total"`
	CasesTotal             int        `json:"cases_total"`
	OpenClaims             int        `json:"open_claims"`
	DischargedClaims       int        `json:"discharged_claims"`
	RefutedClaims          int        `json:"refuted_claims"`
	ContradictingClosures  int        `json:"contradicting_closures"`
	DirectUnknownCases     int        `json:"direct_unknown_cases"`
	DependencyBlockedCases int        `json:"dependency_blocked_cases"`
	CaseDecisionCounts     CaseCounts `json:"case_decision_counts"`
}

type Effects struct {
	RepositoryWrites     int    `json:"repository_writes"`
	MutationAuthority    bool   `json:"mutation_authority"`
	BeforeSnapshotDigest string `json:"before_snapshot_digest"`
	AfterSnapshotDigest  string `json:"after_snapshot_digest"`
	RuntimeVersion       string `json:"runtime_version"`
}

type Decision struct {
	Value      string `json:"value"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
}

type SubjectCounts struct {
	Open              int `json:"open"`
	Discharged        int `json:"discharged"`
	Refuted           int `json:"refuted"`
	DirectUnknown     int `json:"direct_unknown"`
	DependencyBlocked int `json:"dependency_blocked"`
	FailClosed        int `json:"fail_closed"`
}

type Receipt struct {
	Schema                string         `json:"schema"`
	Scope                 string         `json:"scope"`
	HeadSHA               string         `json:"head_sha"`
	GoVersion             string         `json:"go_version"`
	SourcePath            string         `json:"source_path"`
	RawSourceDigest       string         `json:"raw_source_digest"`
	Producer              string         `json:"producer"`
	Consumer              string         `json:"consumer"`
	MetaOperation         string         `json:"meta_operation"`
	SourceRelation        SourceRelation `json:"source_relation"`
	Claims                []Claim        `json:"claims"`
	Evidence              []Evidence     `json:"evidence"`
	Transitions           []Transition   `json:"transitions"`
	CauseReceipts         []CauseReceipt `json:"cause_receipts"`
	Cases                 []CaseResult   `json:"cases"`
	Metrics               []Metric       `json:"metrics"`
	Summary               Summary        `json:"summary"`
	Effects               Effects        `json:"effects"`
	ConformanceDecision   Decision       `json:"conformance_decision"`
	SubjectCounts         SubjectCounts  `json:"subject_counts"`
	SubjectResolution     Decision       `json:"subject_resolution"`
	SemanticReceiptDigest string         `json:"semantic_receipt_digest"`
	ReceiptDigest         string         `json:"receipt_digest"`
}

type observation struct {
	Event          string
	After          string
	Decision       string
	CauseKind      string
	EvidenceDigest string
}

func main() {
	sourcePath := flag.String("source", "examples/claim-lifecycle-calculus/main.gooo", "Gooo source")
	sourceLabel := flag.String("source-label", "", "stable source label recorded in the receipt")
	headSHA := flag.String("head-sha", "", "exact producer head SHA")
	outputPath := flag.String("output", "", "receipt output path")
	flag.Parse()
	if *outputPath == "" {
		fail("-output is required")
	}
	if *sourceLabel == "" {
		*sourceLabel = *sourcePath
	}
	raw, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	relation, err := inspectSource(*sourcePath, raw)
	if err != nil {
		fail(err.Error())
	}
	before := repositorySnapshotDigest()
	receipt := buildReceipt(*sourceLabel, *headSHA, raw, relation, before, "")
	after := repositorySnapshotDigest()
	receipt.Effects.AfterSnapshotDigest = after
	receipt.Effects.RepositoryWrites = boolInt(before != after)
	receipt.Metrics = buildMetrics(receipt)
	receipt.SemanticReceiptDigest = digestWithoutSemanticReceipt(receipt)
	receipt.ReceiptDigest = digestWithoutReceipt(receipt)
	encoded, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	encoded = append(encoded, '\n')
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, encoded, 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("claim lifecycle claims=%d transitions=%d cases=%d pass=%d fail_closed=%d unknown=%d conformance=%s subject=%s repository_writes=%d mutation_authority=%t\n", receipt.Summary.ClaimsTotal, receipt.Summary.TransitionsTotal, receipt.Summary.CasesTotal, receipt.Summary.CaseDecisionCounts.Pass, receipt.Summary.CaseDecisionCounts.FailClosed, receipt.Summary.CaseDecisionCounts.Unknown, receipt.ConformanceDecision.Value, receipt.SubjectResolution.Value, receipt.Effects.RepositoryWrites, receipt.Effects.MutationAuthority)
}

func inspectSource(path string, raw []byte) (SourceRelation, error) {
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return SourceRelation{}, diagnostics.Error()
	}
	if file.Package == nil || file.Namespace == nil || file.Package.Name != "claimlifecycle" || file.Namespace.Name != "claimlifecycle" {
		return SourceRelation{}, fmt.Errorf("Gooo package relation is not claimlifecycle")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceRelation{}, fmt.Errorf("lower Gooo source: %w", err)
	}
	activityNodes := make(map[string]string)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind.String() == "Activity" {
			activityNodes[node.Name] = node.ValueProgram + "\x00" + digestString(node.SemanticCanonical())
		}
	}
	relation := SourceRelation{Package: file.Package.Name, Namespace: file.Namespace.Name, SemanticIRDigest: ir.StableHash()}
	for _, declaration := range file.Declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			relation.Entities = append(relation.Entities, EntityBinding{Name: value.Name, ID: value.ID})
		case *syntax.ActivityDecl:
			binding, err := activityFromSource(value, activityNodes)
			if err != nil {
				return SourceRelation{}, err
			}
			relation.Activities = append(relation.Activities, binding)
		}
	}
	if len(relation.Entities) != caseTotal || len(relation.Activities) != caseTotal {
		return SourceRelation{}, fmt.Errorf("Gooo source must expose %d entities and %d source cases", caseTotal, caseTotal)
	}
	seenCases := make(map[string]bool, caseTotal)
	seenClaims := make(map[string]bool, caseTotal)
	for _, activity := range relation.Activities {
		if seenCases[activity.Case.CaseID] || seenClaims[activity.Case.ClaimID] {
			return SourceRelation{}, fmt.Errorf("source case or claim identity is duplicated")
		}
		seenCases[activity.Case.CaseID] = true
		seenClaims[activity.Case.ClaimID] = true
	}
	for _, activity := range relation.Activities {
		if dependency := activity.Case.DependencyClaimID; dependency != "" && !seenClaims[dependency] {
			return SourceRelation{}, fmt.Errorf("dependency %q is not a source claim", dependency)
		}
	}
	relation.Digest = digestWithoutRelation(relation)
	return relation, nil
}

func activityFromSource(activity *syntax.ActivityDecl, nodes map[string]string) (ActivityBinding, error) {
	if activity.ValueProgram == "" {
		return ActivityBinding{}, fmt.Errorf("activity %q has no structured claim-case value program", activity.Name)
	}
	parts, ok := nodes[activity.Name]
	if !ok {
		return ActivityBinding{}, fmt.Errorf("lowered semantic IR has no activity %q", activity.Name)
	}
	programAndDigest := strings.SplitN(parts, "\x00", 2)
	if len(programAndDigest) != 2 || programAndDigest[0] != activity.ValueProgram {
		return ActivityBinding{}, fmt.Errorf("activity %q value program did not survive lowering", activity.Name)
	}
	caseInfo, err := parseSourceCase(activity.ValueProgram)
	if err != nil {
		return ActivityBinding{}, fmt.Errorf("activity %q: %w", activity.Name, err)
	}
	return ActivityBinding{Name: activity.Name, Inputs: names(activity.Inputs), Output: activity.Output, ValueProgram: activity.ValueProgram, SemanticNodeDigest: programAndDigest[1], Case: caseInfo}, nil
}

func parseSourceCase(program string) (SourceCase, error) {
	const prefix = "claim-case/v1;"
	if !strings.HasPrefix(program, prefix) {
		return SourceCase{}, fmt.Errorf("value program is not claim-case/v1")
	}
	values := make(map[string]string)
	for _, item := range strings.Split(strings.TrimPrefix(program, prefix), ";") {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			return SourceCase{}, fmt.Errorf("malformed source case field %q", item)
		}
		if _, exists := values[key]; exists {
			return SourceCase{}, fmt.Errorf("source case field %q is duplicated", key)
		}
		values[key] = value
	}
	read := func(key string) (string, error) {
		value, ok := values[key]
		if !ok {
			return "", fmt.Errorf("missing source case field %q", key)
		}
		return value, nil
	}
	caseInfo := SourceCase{}
	fields := []struct {
		key string
		dst *string
	}{
		{"case_id", &caseInfo.CaseID}, {"claim_id", &caseInfo.ClaimID}, {"prior_state", &caseInfo.PriorState},
		{"evidence_kind", &caseInfo.EvidenceKind}, {"evidence_id", &caseInfo.EvidenceID}, {"dependency_claim_id", &caseInfo.DependencyClaimID},
		{"observed_stage", &caseInfo.ObservedStage}, {"observed_step", &caseInfo.ObservedStep}, {"observed_reason", &caseInfo.ObservedReason},
		{"expected_status", &caseInfo.ExpectedStatus}, {"expected_decision", &caseInfo.ExpectedDecision}, {"provenance", &caseInfo.Provenance},
	}
	for _, field := range fields {
		value, err := read(field.key)
		if err != nil {
			return SourceCase{}, err
		}
		*field.dst = value
	}
	if caseInfo.CaseID == "" || caseInfo.ClaimID == "" || caseInfo.PriorState == "" || caseInfo.EvidenceKind == "" || caseInfo.ObservedStage == "" || caseInfo.ObservedStep == "" || caseInfo.ObservedReason == "" || caseInfo.ExpectedStatus == "" || caseInfo.ExpectedDecision == "" || caseInfo.Provenance == "" {
		return SourceCase{}, fmt.Errorf("source case has an empty required observation")
	}
	return caseInfo, nil
}

func buildReceipt(sourcePath, headSHA string, raw []byte, relation SourceRelation, before, after string) Receipt {
	claims := make([]Claim, 0, len(relation.Activities))
	for _, activity := range relation.Activities {
		statement := statementOf(activity)
		claims = append(claims, Claim{ID: activity.Case.ClaimID, CaseID: activity.Case.CaseID, Statement: statement, DeclarationDigest: digestJSON(struct {
			ID, Statement string
		}{activity.Case.ClaimID, statement}), Status: "OPEN", Provenance: provenanceOf(activity, relation)})
	}
	evidence := buildEvidence(relation)
	evidenceByID := make(map[string]Evidence, len(evidence))
	for _, item := range evidence {
		evidenceByID[item.ID] = item
	}
	transitions, causes, cases := buildLedger(relation, claims, evidenceByID)
	for _, item := range cases {
		for claimIndex := range claims {
			if claims[claimIndex].ID != item.ClaimID {
				continue
			}
			claims[claimIndex].Status = item.ObservedStatus
			claims[claimIndex].Coordinate = item.Coordinate
			claims[claimIndex].Reason = item.Reason
			for _, transition := range transitions {
				if transition.ClaimID == item.ClaimID && transition.Event != "CLAIM_OPENED" {
					claims[claimIndex].EvidenceDigest = transition.EvidenceDigest
					break
				}
			}
			break
		}
	}
	summary := summarize(claims, transitions, cases)
	receipt := Receipt{
		Schema: receiptSchema, Scope: "CLAIM_LIFECYCLE_ONLY", HeadSHA: headSHA, GoVersion: runtime.Version(),
		SourcePath: sourcePath, RawSourceDigest: digestBytes(raw), Producer: producerID, Consumer: consumerID,
		MetaOperation: metaOperation, SourceRelation: relation, Claims: claims, Evidence: evidence,
		Transitions: transitions, CauseReceipts: causes, Cases: cases, Summary: summary,
		Effects: Effects{RepositoryWrites: boolInt(before != after), MutationAuthority: false, BeforeSnapshotDigest: before, AfterSnapshotDigest: after, RuntimeVersion: runtime.Version()},
	}
	receipt.ConformanceDecision = conformanceDecision(cases)
	receipt.SubjectCounts = subjectCounts(claims, cases)
	receipt.SubjectResolution = subjectResolution(receipt.SubjectCounts)
	receipt.Metrics = buildMetrics(receipt)
	receipt.SemanticReceiptDigest = digestWithoutSemanticReceipt(receipt)
	receipt.ReceiptDigest = digestWithoutReceipt(receipt)
	return receipt
}

func buildEvidence(relation SourceRelation) []Evidence {
	result := make([]Evidence, 0)
	seen := make(map[string]bool)
	for _, activity := range relation.Activities {
		caseInfo := activity.Case
		if caseInfo.EvidenceID == "" || (caseInfo.EvidenceKind != "SUPPORTING" && caseInfo.EvidenceKind != "CONTRADICTING") || seen[caseInfo.EvidenceID] {
			continue
		}
		seen[caseInfo.EvidenceID] = true
		result = append(result, Evidence{ID: caseInfo.EvidenceID, Kind: caseInfo.EvidenceKind, ClaimID: caseInfo.ClaimID, SourceCaseID: caseInfo.CaseID, Digest: digestJSON(struct {
			ID, Kind, ClaimID, SourceCaseID, RelationDigest string
		}{caseInfo.EvidenceID, caseInfo.EvidenceKind, caseInfo.ClaimID, caseInfo.CaseID, relation.Digest})})
	}
	return result
}

func buildLedger(relation SourceRelation, claims []Claim, evidence map[string]Evidence) ([]Transition, []CauseReceipt, []CaseResult) {
	transitions := make([]Transition, 0, len(claims)*2)
	causes := make([]CauseReceipt, 0, len(claims)*2)
	cases := make([]CaseResult, 0, len(claims))
	appendEvent := func(claim Claim, event, before, after, evidenceDigest string, cause CauseReceipt) {
		cause.Sequence = len(transitions) + 1
		cause.ClaimID = claim.ID
		cause.Digest = digestWithoutCause(cause)
		causes = append(causes, cause)
		transition := Transition{Sequence: len(transitions) + 1, ClaimID: claim.ID, DeclarationDigest: claim.DeclarationDigest, Event: event, Before: before, After: after, EvidenceDigest: evidenceDigest, CauseReceiptDigest: cause.Digest}
		if len(transitions) > 0 {
			transition.PreviousDigest = transitions[len(transitions)-1].Digest
		}
		transition.Digest = digestWithoutTransition(transition)
		transitions = append(transitions, transition)
	}
	for index, claim := range claims {
		activity := relation.Activities[index]
		appendEvent(claim, "CLAIM_OPENED", "UNRECORDED", "OPEN", "", CauseReceipt{Kind: "DECLARATION", Coordinate: Coordinate{Stage: "CLAIM", Step: "declare", Reason: "CLAIM_OPENED"}, Reason: "source activity " + activity.Name + " declared claim " + claim.ID, Provenance: provenanceOf(activity, relation)})
	}
	for index, activity := range relation.Activities {
		claim := claims[index]
		observed := observeCase(activity.Case, evidence)
		cause := CauseReceipt{Kind: observed.CauseKind, Coordinate: Coordinate{Stage: activity.Case.ObservedStage, Step: activity.Case.ObservedStep, Reason: activity.Case.ObservedReason}, Reason: activity.Case.ObservedReason, Provenance: provenanceOf(activity, relation)}
		if activity.Case.EvidenceID != "" {
			if item, ok := evidence[activity.Case.EvidenceID]; ok {
				cause.EvidenceIDs = []string{item.ID}
			}
		}
		if activity.Case.DependencyClaimID != "" {
			cause.DependencyClaimIDs = []string{activity.Case.DependencyClaimID}
		}
		appendEvent(claim, observed.Event, activity.Case.PriorState, observed.After, observed.EvidenceDigest, cause)
		cases = append(cases, CaseResult{ID: activity.Case.CaseID, ClaimID: claim.ID, ExpectedStatus: activity.Case.ExpectedStatus, ObservedStatus: observed.After, ExpectedDecision: activity.Case.ExpectedDecision, Decision: observed.Decision, Conformance: caseConformance(activity.Case, observed), ObservedEvent: observed.Event, CauseKind: observed.CauseKind, Coordinate: cause.Coordinate, Reason: activity.Case.ObservedReason, Provenance: provenanceOf(activity, relation)})
	}
	return transitions, causes, cases
}

func observeCase(caseInfo SourceCase, evidence map[string]Evidence) observation {
	result := observation{After: caseInfo.PriorState, Decision: "FAIL_CLOSED", CauseKind: "INVALID_OBSERVATION"}
	if caseInfo.PriorState != "OPEN" {
		result.Event = "EVIDENCE_INVALID"
		return result
	}
	switch caseInfo.EvidenceKind {
	case "SUPPORTING":
		result.Event, result.After, result.Decision, result.CauseKind = "EVIDENCE_ACCEPTED", "DISCHARGED", "PASS", "SUPPORTING_EVIDENCE"
	case "CONTRADICTING":
		result.Event, result.After, result.Decision, result.CauseKind = "EVIDENCE_CONTRADICTED", "REFUTED", "PASS", "CONTRADICTING_EVIDENCE"
	case "UNAVAILABLE":
		result.Event, result.After, result.Decision, result.CauseKind = "EVIDENCE_UNAVAILABLE", "OPEN", "UNKNOWN", "DIRECT_UNKNOWN"
	case "DEPENDENCY_BLOCKED":
		result.Event, result.After, result.Decision, result.CauseKind = "EVIDENCE_DEPENDENCY_BLOCKED", "OPEN", "UNKNOWN", "DEPENDENCY_BLOCKED"
	case "AMBIGUOUS":
		result.Event, result.After, result.Decision, result.CauseKind = "EVIDENCE_AMBIGUOUS", "OPEN", "FAIL_CLOSED", "AMBIGUOUS_EVIDENCE"
	default:
		result.Event = "EVIDENCE_INVALID"
		return result
	}
	if (caseInfo.EvidenceKind == "SUPPORTING" || caseInfo.EvidenceKind == "CONTRADICTING") && caseInfo.EvidenceID != "" {
		if item, ok := evidence[caseInfo.EvidenceID]; ok && item.Kind == caseInfo.EvidenceKind {
			result.EvidenceDigest = item.Digest
			return result
		}
	}
	if caseInfo.EvidenceKind == "SUPPORTING" || caseInfo.EvidenceKind == "CONTRADICTING" {
		result.Event, result.After, result.Decision, result.CauseKind = "EVIDENCE_INVALID", "OPEN", "FAIL_CLOSED", "INVALID_OBSERVATION"
	}
	if caseInfo.EvidenceKind == "DEPENDENCY_BLOCKED" && caseInfo.DependencyClaimID == "" {
		result.Event, result.Decision, result.CauseKind = "EVIDENCE_INVALID", "FAIL_CLOSED", "INVALID_OBSERVATION"
	}
	return result
}

func caseConformance(source SourceCase, observed observation) string {
	if source.ExpectedStatus == observed.After && source.ExpectedDecision == observed.Decision {
		return "PASS"
	}
	return "FAIL_CLOSED"
}

func summarize(claims []Claim, transitions []Transition, cases []CaseResult) Summary {
	summary := Summary{ClaimsTotal: len(claims), TransitionsTotal: len(transitions), CasesTotal: len(cases)}
	for _, claim := range claims {
		switch claim.Status {
		case "OPEN":
			summary.OpenClaims++
		case "DISCHARGED":
			summary.DischargedClaims++
		case "REFUTED":
			summary.RefutedClaims++
		}
	}
	for _, item := range cases {
		switch item.Decision {
		case "PASS":
			summary.CaseDecisionCounts.Pass++
		case "FAIL_CLOSED":
			summary.CaseDecisionCounts.FailClosed++
		case "UNKNOWN":
			summary.CaseDecisionCounts.Unknown++
		}
		switch item.CauseKind {
		case "DIRECT_UNKNOWN":
			summary.DirectUnknownCases++
		case "DEPENDENCY_BLOCKED":
			summary.DependencyBlockedCases++
		}
		if item.CauseKind == "CONTRADICTING_EVIDENCE" && item.ObservedStatus == "REFUTED" {
			summary.ContradictingClosures++
		}
	}
	return summary
}

func conformanceDecision(cases []CaseResult) Decision {
	for _, item := range cases {
		if item.Conformance != "PASS" {
			return Decision{Value: "FAIL_CLOSED", Resolution: "SOURCE_CASE_MISMATCH", Reason: "observed lifecycle did not match source case " + item.ID}
		}
	}
	return Decision{Value: "PASS", Resolution: "SOURCE_CASES_RECONSTRUCTED", Reason: "receipt decisions and transitions were derived from lowered Gooo cases"}
}

func subjectCounts(claims []Claim, cases []CaseResult) SubjectCounts {
	counts := SubjectCounts{}
	for _, claim := range claims {
		switch claim.Status {
		case "OPEN":
			counts.Open++
		case "DISCHARGED":
			counts.Discharged++
		case "REFUTED":
			counts.Refuted++
		}
	}
	for _, item := range cases {
		if item.CauseKind == "DIRECT_UNKNOWN" {
			counts.DirectUnknown++
		}
		if item.CauseKind == "DEPENDENCY_BLOCKED" {
			counts.DependencyBlocked++
		}
		if item.Decision == "FAIL_CLOSED" {
			counts.FailClosed++
		}
	}
	return counts
}

func subjectResolution(counts SubjectCounts) Decision {
	if counts.UnknownTotal() > 0 && counts.FailClosed > 0 {
		return Decision{Value: "UNKNOWN_AND_FAIL_CLOSED_PRESENT", Resolution: "SUBJECT_OUTCOME", Reason: "open claims include direct or dependency-blocked uncertainty and an ambiguous case"}
	}
	if counts.UnknownTotal() > 0 {
		return Decision{Value: "UNKNOWN_PRESENT", Resolution: "SUBJECT_OUTCOME", Reason: "open claims retain direct or dependency-blocked uncertainty"}
	}
	if counts.FailClosed > 0 {
		return Decision{Value: "FAIL_CLOSED_PRESENT", Resolution: "SUBJECT_OUTCOME", Reason: "at least one subject case refuses closure"}
	}
	return Decision{Value: "CLOSED", Resolution: "SUBJECT_OUTCOME", Reason: "all subject cases closed"}
}

func (counts SubjectCounts) UnknownTotal() int {
	return counts.DirectUnknown + counts.DependencyBlocked
}

func buildMetrics(receipt Receipt) []Metric {
	metric := func(id, class, proof string, numerator, denominator int) Metric {
		return Metric{ID: id, Class: class, ProofChoice: proof, Numerator: numerator, Denominator: denominator, Producer: producerID, Consumer: consumerID, MetaOperation: metaOperation, Satisfied: denominator == 0 || numerator == denominator}
	}
	closures := receipt.Summary.DischargedClaims + receipt.Summary.RefutedClaims
	closureDenominator := 0
	contradictionDenominator := 0
	unknownDenominator := 0
	ambiguousDenominator := 0
	for _, activity := range receipt.SourceRelation.Activities {
		switch activity.Case.EvidenceKind {
		case "SUPPORTING", "CONTRADICTING":
			closureDenominator++
		case "UNAVAILABLE", "DEPENDENCY_BLOCKED":
			unknownDenominator++
		case "AMBIGUOUS":
			ambiguousDenominator++
		}
		if activity.Case.EvidenceKind == "CONTRADICTING" {
			contradictionDenominator++
		}
	}
	return []Metric{
		metric("source-cases-reconstructed.v1", "DRIVER", "FOUNDATION", len(receipt.SourceRelation.Activities), caseTotal),
		metric("producer-import-boundary.v1", "GUARDRAIL", "FOUNDATION", 0, 0),
		metric("claim-identity-preserved.v1", "DRIVER", "FOUNDATION", len(receipt.Claims), len(receipt.SourceRelation.Activities)),
		metric("persistent-claim-ledger.v1", "DRIVER", "COHERENCE", persistentClaimCount(receipt), len(receipt.Claims)),
		metric("append-only-transition-ledger.v1", "DRIVER", "FOUNDATION", len(receipt.Transitions), len(receipt.Claims)*2),
		metric("terminal-evidence-closures.v1", "OUTCOME", "COHERENCE", closures, closureDenominator),
		metric("contradiction-to-refutation.v1", "OUTCOME", "COHERENCE", receipt.Summary.ContradictingClosures, contradictionDenominator),
		metric("unknown-cause-partition.v1", "GUARDRAIL", "REGRESSION", receipt.Summary.DirectUnknownCases+receipt.Summary.DependencyBlockedCases, unknownDenominator),
		metric("ambiguous-evidence-fail-closed.v1", "GUARDRAIL", "REGRESSION", receipt.Summary.CaseDecisionCounts.FailClosed, ambiguousDenominator),
		metric("read-only-effect-boundary.v1", "GUARDRAIL", "REGRESSION", boolInt(receipt.Effects.RepositoryWrites == 0 && !receipt.Effects.MutationAuthority && receipt.Effects.BeforeSnapshotDigest == receipt.Effects.AfterSnapshotDigest), 1),
	}
}

func persistentClaimCount(receipt Receipt) int {
	opened := make(map[string]bool)
	resolved := make(map[string]bool)
	for _, transition := range receipt.Transitions {
		if transition.Event == "CLAIM_OPENED" {
			opened[transition.ClaimID] = true
		} else {
			resolved[transition.ClaimID] = true
		}
	}
	count := 0
	for _, claim := range receipt.Claims {
		if opened[claim.ID] && resolved[claim.ID] {
			count++
		}
	}
	return count
}

func names(refs []syntax.NameRef) []string {
	result := make([]string, len(refs))
	for index, ref := range refs {
		result[index] = ref.Name
	}
	return result
}

func statementOf(activity ActivityBinding) string {
	return activity.Name + "(" + strings.Join(activity.Inputs, ",") + ")->" + activity.Output + " computes " + activity.ValueProgram
}

func provenanceOf(activity ActivityBinding, relation SourceRelation) string {
	return "gooo:activity/" + activity.Name + ";ir=" + relation.SemanticIRDigest + ";node=" + activity.SemanticNodeDigest + ";case=" + activity.Case.CaseID
}

func repositorySnapshotDigest() string {
	output, err := exec.Command("git", "status", "--porcelain=v1", "--untracked-files=all").Output()
	if err != nil {
		fail("repository snapshot: " + err.Error())
	}
	return digestBytes(output)
}

func digestString(value string) string { return digestBytes([]byte(value)) }

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(encoded)
}

func digestWithoutRelation(value SourceRelation) string {
	value.Digest = ""
	return digestJSON(value)
}

func digestWithoutCause(value CauseReceipt) string {
	value.Digest = ""
	return digestJSON(value)
}

func digestWithoutTransition(value Transition) string {
	value.Digest = ""
	return digestJSON(value)
}

func digestWithoutReceipt(value Receipt) string {
	value.ReceiptDigest = ""
	return digestJSON(value)
}

func digestWithoutSemanticReceipt(value Receipt) string {
	value.SourcePath = ""
	value.RawSourceDigest = ""
	value.SemanticReceiptDigest = ""
	value.ReceiptDigest = ""
	value.Effects.BeforeSnapshotDigest = ""
	value.Effects.AfterSnapshotDigest = ""
	return digestJSON(value)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
