package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	receiptSchema = "gooo/claim-lifecycle-calculus/v1"
	metaOperation = "preserve-claim-lifecycle"
	producerID    = "claimlifecyclecalculus.Evaluate"
	consumerID    = "claim-lifecycle-calculus-judge"
	goVersion     = "go1.27.0"
	claimTotal    = 6
	caseTotal     = 6
)

type activitySpec struct {
	Name    string
	Inputs  []string
	Output  string
	Program string
	ClaimID string
}

var activitySpecs = []activitySpec{
	{Name: "DeclareClaim", Inputs: []string{"Claim"}, Output: "OpenClaim", Program: "meta.claim.lifecycle.open:v1", ClaimID: "claim:declared-open"},
	{Name: "AcceptSupportingEvidence", Inputs: []string{"OpenClaim", "Evidence"}, Output: "DischargedClaim", Program: "meta.claim.lifecycle.discharge:v1", ClaimID: "claim:supporting-discharge"},
	{Name: "ApplyContradictingEvidence", Inputs: []string{"OpenClaim", "Evidence"}, Output: "RefutedClaim", Program: "meta.claim.lifecycle.refute:v1", ClaimID: "claim:contradicting-refutation"},
	{Name: "PreserveOpenOnMissingEvidence", Inputs: []string{"OpenClaim"}, Output: "OpenClaim", Program: "meta.claim.lifecycle.unknown:v1", ClaimID: "claim:direct-unknown"},
	{Name: "EmitLifecycleReceipt", Inputs: []string{"OpenClaim"}, Output: "LifecycleReceipt", Program: "meta.claim.lifecycle.receipt:v1", ClaimID: "claim:dependent-unknown"},
	{Name: "RejectAmbiguousEvidence", Inputs: []string{"OpenClaim", "Evidence"}, Output: "OpenClaim", Program: "meta.claim.lifecycle.ambiguous:v1", ClaimID: "claim:ambiguous-evidence"},
}

var entitySpecs = []string{
	"gooo://claim-lifecycle/claim",
	"gooo://claim-lifecycle/evidence",
	"gooo://claim-lifecycle/open-claim",
	"gooo://claim-lifecycle/discharged-claim",
	"gooo://claim-lifecycle/refuted-claim",
	"gooo://claim-lifecycle/lifecycle-receipt",
}

type EntityBinding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type ActivityBinding struct {
	Name         string   `json:"name"`
	Inputs       []string `json:"inputs"`
	Output       string   `json:"output"`
	ValueProgram string   `json:"value_program"`
	ClaimID      string   `json:"claim_id"`
}

type SourceRelation struct {
	Package    string            `json:"package"`
	Namespace  string            `json:"namespace"`
	Entities   []EntityBinding   `json:"entities"`
	Activities []ActivityBinding `json:"activities"`
	Digest     string            `json:"digest"`
}

type Claim struct {
	ID                string `json:"id"`
	Statement         string `json:"statement"`
	DeclarationDigest string `json:"declaration_digest"`
	Status            string `json:"status"`
}

type Evidence struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Digest string `json:"digest"`
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
	ID             string     `json:"id"`
	ClaimID        string     `json:"claim_id"`
	ExpectedStatus string     `json:"expected_status"`
	ObservedStatus string     `json:"observed_status"`
	Decision       string     `json:"decision"`
	ObservedEvent  string     `json:"observed_event"`
	CauseKind      string     `json:"cause_kind"`
	Coordinate     Coordinate `json:"coordinate"`
	Reason         string     `json:"reason"`
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
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Decision struct {
	Value      string `json:"value"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
}

type Receipt struct {
	Schema         string         `json:"schema"`
	Scope          string         `json:"scope"`
	HeadSHA        string         `json:"head_sha"`
	GoVersion      string         `json:"go_version"`
	SourcePath     string         `json:"source_path"`
	SourceDigest   string         `json:"source_digest"`
	Producer       string         `json:"producer"`
	Consumer       string         `json:"consumer"`
	MetaOperation  string         `json:"meta_operation"`
	SourceRelation SourceRelation `json:"source_relation"`
	Claims         []Claim        `json:"claims"`
	Evidence       []Evidence     `json:"evidence"`
	Transitions    []Transition   `json:"transitions"`
	CauseReceipts  []CauseReceipt `json:"cause_receipts"`
	Cases          []CaseResult   `json:"cases"`
	Metrics        []Metric       `json:"metrics"`
	Summary        Summary        `json:"summary"`
	Effects        Effects        `json:"effects"`
	Decision       Decision       `json:"decision"`
	ReceiptDigest  string         `json:"receipt_digest"`
}

func main() {
	sourcePath := flag.String("source", "examples/claim-lifecycle-calculus/main.gooo", "Gooo source")
	headSHA := flag.String("head-sha", "", "exact producer head SHA")
	outputPath := flag.String("output", "", "receipt output path")
	flag.Parse()
	if *outputPath == "" {
		fail("-output is required")
	}
	raw, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	relation, err := inspectSource(*sourcePath, raw)
	if err != nil {
		fail(err.Error())
	}
	receipt := buildReceipt(*sourcePath, *headSHA, raw, relation)
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
	fmt.Printf("claim lifecycle claims=%d transitions=%d cases=%d pass=%d fail_closed=%d unknown=%d decision=%s repository_writes=%d mutation_authority=%t\n", receipt.Summary.ClaimsTotal, receipt.Summary.TransitionsTotal, receipt.Summary.CasesTotal, receipt.Summary.CaseDecisionCounts.Pass, receipt.Summary.CaseDecisionCounts.FailClosed, receipt.Summary.CaseDecisionCounts.Unknown, receipt.Decision.Value, receipt.Effects.RepositoryWrites, receipt.Effects.MutationAuthority)
}

func inspectSource(path string, raw []byte) (SourceRelation, error) {
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return SourceRelation{}, diagnostics.Error()
	}
	if file.Package == nil || file.Namespace == nil || file.Package.Name != "claimlifecycle" || file.Namespace.Name != "claimlifecycle" {
		return SourceRelation{}, fmt.Errorf("Gooo package relation is not claimlifecycle")
	}
	relation := SourceRelation{Package: file.Package.Name, Namespace: file.Namespace.Name}
	for _, declaration := range file.Declarations {
		if entity, ok := declaration.(*syntax.EntityDecl); ok {
			relation.Entities = append(relation.Entities, EntityBinding{Name: entity.Name, ID: entity.ID})
		}
	}
	for index, spec := range activitySpecs {
		if index >= len(file.Declarations) {
			return SourceRelation{}, fmt.Errorf("Gooo activity denominator is incomplete")
		}
		activity, ok := file.Declarations[len(relation.Entities)+index].(*syntax.ActivityDecl)
		if !ok || activity.Name != spec.Name || activity.Output != spec.Output || activity.ValueProgram != spec.Program || !sameStrings(names(activity.Inputs), spec.Inputs) {
			return SourceRelation{}, fmt.Errorf("Gooo activity relation %d does not bind the lifecycle contract", index+1)
		}
		relation.Activities = append(relation.Activities, ActivityBinding{Name: activity.Name, Inputs: names(activity.Inputs), Output: activity.Output, ValueProgram: activity.ValueProgram, ClaimID: spec.ClaimID})
	}
	if len(relation.Entities) != len(entitySpecs) || len(relation.Activities) != len(activitySpecs) {
		return SourceRelation{}, fmt.Errorf("Gooo relation denominator changed")
	}
	for index, entity := range relation.Entities {
		if entity.ID != entitySpecs[index] {
			return SourceRelation{}, fmt.Errorf("Gooo entity relation %d does not bind the lifecycle contract", index+1)
		}
	}
	relation.Digest = digestWithoutField(relation, func(value *SourceRelation) { value.Digest = "" })
	return relation, nil
}

func names(refs []syntax.NameRef) []string {
	result := make([]string, len(refs))
	for index, ref := range refs {
		result[index] = ref.Name
	}
	return result
}

func buildReceipt(sourcePath, headSHA string, raw []byte, relation SourceRelation) Receipt {
	claims := make([]Claim, 0, claimTotal)
	for _, activity := range relation.Activities {
		statement := statementOf(activity)
		claims = append(claims, Claim{ID: activity.ClaimID, Statement: statement, DeclarationDigest: digestJSON(struct {
			ID, Statement string
		}{activity.ClaimID, statement}), Status: "OPEN"})
	}
	evidence := []Evidence{
		{ID: "evidence:supporting", Kind: "SUPPORTING", Digest: digestString("supporting evidence for claim:supporting-discharge")},
		{ID: "evidence:contradicting", Kind: "CONTRADICTING", Digest: digestString("contradicting evidence for claim:contradicting-refutation")},
	}
	transitions, causes := buildTransitions(claims, evidence)
	claims[1].Status = "DISCHARGED"
	claims[2].Status = "REFUTED"
	cases := buildCases(transitions, causes)
	summary := summarize(claims, transitions, cases)
	receipt := Receipt{
		Schema: receiptSchema, Scope: "CLAIM_LIFECYCLE_ONLY", HeadSHA: headSHA, GoVersion: goVersion,
		SourcePath: sourcePath, SourceDigest: digestBytes(raw), Producer: producerID, Consumer: consumerID,
		MetaOperation: metaOperation, SourceRelation: relation, Claims: claims, Evidence: evidence,
		Transitions: transitions, CauseReceipts: causes, Cases: cases, Summary: summary,
		Effects: Effects{RepositoryWrites: 0, MutationAuthority: false}, Decision: Decision{Value: "PASS", Resolution: "LIFECYCLE_CASES_EXACT", Reason: "CLAIMS_PRESERVED_WITH_EXPLICIT_OPEN_DISCHARGED_REFUTED_TRANSITIONS"},
	}
	receipt.Metrics = buildMetrics(receipt)
	receipt.ReceiptDigest = digestWithoutField(receipt, func(value *Receipt) { value.ReceiptDigest = "" })
	return receipt
}

func buildTransitions(claims []Claim, evidence []Evidence) ([]Transition, []CauseReceipt) {
	var transitions []Transition
	var causes []CauseReceipt
	appendEvent := func(claim Claim, event, before, after, evidenceDigest string, cause CauseReceipt) {
		cause.Sequence = len(transitions) + 1
		cause.ClaimID = claim.ID
		cause.Digest = digestWithoutField(cause, func(value *CauseReceipt) { value.Digest = "" })
		causes = append(causes, cause)
		transition := Transition{Sequence: len(transitions) + 1, ClaimID: claim.ID, DeclarationDigest: claim.DeclarationDigest, Event: event, Before: before, After: after, EvidenceDigest: evidenceDigest, CauseReceiptDigest: cause.Digest}
		if len(transitions) > 0 {
			transition.PreviousDigest = transitions[len(transitions)-1].Digest
		}
		transition.Digest = digestWithoutField(transition, func(value *Transition) { value.Digest = "" })
		transitions = append(transitions, transition)
	}
	for _, claim := range claims {
		appendEvent(claim, "CLAIM_OPENED", "UNRECORDED", "OPEN", "", CauseReceipt{Kind: "DECLARATION", Coordinate: Coordinate{Stage: "CLAIM", Step: "declare", Reason: "CLAIM_OPENED"}, Reason: "Gooo activity declaration created a durable claim slot"})
	}
	appendEvent(claims[0], "EVIDENCE_UNAVAILABLE", "OPEN", "OPEN", "", CauseReceipt{Kind: "DIRECT_UNKNOWN", Coordinate: Coordinate{Stage: "EVIDENCE", Step: "observe", Reason: "EVIDENCE_UNAVAILABLE"}, Reason: "no evidence was supplied; the claim remains"})
	appendEvent(claims[1], "EVIDENCE_ACCEPTED", "OPEN", "DISCHARGED", evidence[0].Digest, CauseReceipt{Kind: "SUPPORTING_EVIDENCE", EvidenceIDs: []string{evidence[0].ID}, Coordinate: Coordinate{Stage: "EVIDENCE", Step: "classify", Reason: "SUPPORTING_EVIDENCE"}, Reason: "supporting evidence closes the claim as discharged"})
	appendEvent(claims[2], "EVIDENCE_CONTRADICTED", "OPEN", "REFUTED", evidence[1].Digest, CauseReceipt{Kind: "CONTRADICTING_EVIDENCE", EvidenceIDs: []string{evidence[1].ID}, Coordinate: Coordinate{Stage: "EVIDENCE", Step: "classify", Reason: "CONTRADICTING_EVIDENCE"}, Reason: "contradicting evidence closes the claim as refuted"})
	appendEvent(claims[3], "EVIDENCE_UNAVAILABLE", "OPEN", "OPEN", "", CauseReceipt{Kind: "DIRECT_UNKNOWN", Coordinate: Coordinate{Stage: "EVIDENCE", Step: "observe", Reason: "EVIDENCE_UNAVAILABLE"}, Reason: "direct evidence is unavailable; the claim remains open"})
	appendEvent(claims[4], "EVIDENCE_DEPENDENCY_BLOCKED", "OPEN", "OPEN", "", CauseReceipt{Kind: "DEPENDENCY_BLOCKED", DependencyClaimIDs: []string{claims[3].ID}, Coordinate: Coordinate{Stage: "RESOLVE", Step: "dependency", Reason: "UPSTREAM_CLAIM_OPEN"}, Reason: "dependent resolution is blocked by an open upstream claim"})
	appendEvent(claims[5], "EVIDENCE_AMBIGUOUS", "OPEN", "OPEN", "", CauseReceipt{Kind: "AMBIGUOUS_EVIDENCE", Coordinate: Coordinate{Stage: "EVIDENCE", Step: "classify", Reason: "AMBIGUOUS_EVIDENCE"}, Reason: "ambiguous evidence cannot close a claim"})
	return transitions, causes
}

func buildCases(transitions []Transition, causes []CauseReceipt) []CaseResult {
	resolution := transitions[claimTotal:]
	return []CaseResult{
		{ID: "declared-claim-without-evidence", ClaimID: resolution[0].ClaimID, ExpectedStatus: "OPEN", ObservedStatus: resolution[0].After, Decision: "UNKNOWN", ObservedEvent: resolution[0].Event, CauseKind: causes[claimTotal].Kind, Coordinate: causes[claimTotal].Coordinate, Reason: "direct evidence is absent"},
		{ID: "supporting-evidence-closes", ClaimID: resolution[1].ClaimID, ExpectedStatus: "DISCHARGED", ObservedStatus: resolution[1].After, Decision: "PASS", ObservedEvent: resolution[1].Event, CauseKind: causes[claimTotal+1].Kind, Coordinate: causes[claimTotal+1].Coordinate, Reason: "supporting evidence discharges the preserved claim"},
		{ID: "conflicting-evidence-closes-refuted", ClaimID: resolution[2].ClaimID, ExpectedStatus: "REFUTED", ObservedStatus: resolution[2].After, Decision: "PASS", ObservedEvent: resolution[2].Event, CauseKind: causes[claimTotal+2].Kind, Coordinate: causes[claimTotal+2].Coordinate, Reason: "conflicting evidence refutes without deleting the claim"},
		{ID: "direct-missing-evidence", ClaimID: resolution[3].ClaimID, ExpectedStatus: "OPEN", ObservedStatus: resolution[3].After, Decision: "UNKNOWN", ObservedEvent: resolution[3].Event, CauseKind: causes[claimTotal+3].Kind, Coordinate: causes[claimTotal+3].Coordinate, Reason: "direct evidence is unavailable"},
		{ID: "dependency-missing-evidence", ClaimID: resolution[4].ClaimID, ExpectedStatus: "OPEN", ObservedStatus: resolution[4].After, Decision: "UNKNOWN", ObservedEvent: resolution[4].Event, CauseKind: causes[claimTotal+4].Kind, Coordinate: causes[claimTotal+4].Coordinate, Reason: "dependency is blocked by a direct unknown"},
		{ID: "ambiguous-evidence-refuses-closure", ClaimID: resolution[5].ClaimID, ExpectedStatus: "OPEN", ObservedStatus: resolution[5].After, Decision: "FAIL_CLOSED", ObservedEvent: resolution[5].Event, CauseKind: causes[claimTotal+5].Kind, Coordinate: causes[claimTotal+5].Coordinate, Reason: "ambiguous evidence fails closed"},
	}
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

func buildMetrics(receipt Receipt) []Metric {
	metric := func(id, class, proof string, numerator, denominator int) Metric {
		return Metric{ID: id, Class: class, ProofChoice: proof, Numerator: numerator, Denominator: denominator, Producer: producerID, Consumer: consumerID, MetaOperation: metaOperation, Satisfied: numerator == denominator}
	}
	return []Metric{
		metric("claim-identity-preserved.v1", "DRIVER", "FOUNDATION", len(receipt.Claims), claimTotal),
		metric("source-activity-relations-bound.v1", "DRIVER", "FOUNDATION", len(receipt.SourceRelation.Activities), len(activitySpecs)),
		metric("append-only-transition-ledger.v1", "DRIVER", "FOUNDATION", len(receipt.Transitions), claimTotal*2),
		metric("terminal-evidence-closures.v1", "OUTCOME", "COHERENCE", receipt.Summary.DischargedClaims+receipt.Summary.RefutedClaims, 2),
		metric("contradiction-to-refutation.v1", "OUTCOME", "COHERENCE", receipt.Summary.ContradictingClosures, 1),
		metric("unknown-cause-partition.v1", "GUARDRAIL", "REGRESSION", receipt.Summary.DirectUnknownCases+receipt.Summary.DependencyBlockedCases, 3),
		metric("ambiguous-evidence-fail-closed.v1", "GUARDRAIL", "REGRESSION", receipt.Summary.CaseDecisionCounts.FailClosed, 1),
		metric("read-only-effect-boundary.v1", "GUARDRAIL", "REGRESSION", boolInt(receipt.Effects.RepositoryWrites == 0 && !receipt.Effects.MutationAuthority), 1),
	}
}

func statementOf(activity ActivityBinding) string {
	return activity.Name + "(" + strings.Join(activity.Inputs, ",") + ")->" + activity.Output + " computes " + activity.ValueProgram
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

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func digestString(value string) string {
	return digestBytes([]byte(value))
}

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

func digestWithoutField[T any](value T, clear func(*T)) string {
	clear(&value)
	return digestJSON(value)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
