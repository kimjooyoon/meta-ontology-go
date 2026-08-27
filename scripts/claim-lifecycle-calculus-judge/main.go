package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
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
	receiptPath := flag.String("receipt", "", "producer receipt")
	sourcePath := flag.String("source", "examples/claim-lifecycle-calculus/main.gooo", "Gooo source")
	flag.Parse()
	if *receiptPath == "" {
		fail("-receipt is required")
	}
	receiptRaw, err := os.ReadFile(*receiptPath)
	if err != nil {
		fail(err.Error())
	}
	sourceRaw, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	var receipt Receipt
	decoder := json.NewDecoder(strings.NewReader(string(receiptRaw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		fail(err.Error())
	}
	if err := validate(receipt, *sourcePath, sourceRaw); err != nil {
		fail(err.Error())
	}
	fmt.Printf("judge decision=%s claims=%d/%d transitions=%d/%d pass=%d fail_closed=%d unknown=%d direct_unknown=%d dependency_blocked=%d\n", receipt.Decision.Value, receipt.Summary.ClaimsTotal, claimTotal, receipt.Summary.TransitionsTotal, claimTotal*2, receipt.Summary.CaseDecisionCounts.Pass, receipt.Summary.CaseDecisionCounts.FailClosed, receipt.Summary.CaseDecisionCounts.Unknown, receipt.Summary.DirectUnknownCases, receipt.Summary.DependencyBlockedCases)
}

func validate(receipt Receipt, sourcePath string, sourceRaw []byte) error {
	if receipt.Schema != receiptSchema || receipt.Scope != "CLAIM_LIFECYCLE_ONLY" || receipt.GoVersion != goVersion || receipt.SourcePath != sourcePath || receipt.Producer != producerID || receipt.Consumer != consumerID || receipt.MetaOperation != metaOperation {
		return fmt.Errorf("receipt header or producer boundary changed")
	}
	if receipt.SourceDigest != digestBytes(sourceRaw) || !validDigest(receipt.SourceDigest) {
		return fmt.Errorf("receipt is not bound to the supplied Gooo source")
	}
	relation, err := inspectSource(sourcePath, sourceRaw)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(receipt.SourceRelation, relation) {
		return fmt.Errorf("Gooo activity relation changed")
	}
	expectedClaims := claimsFor(relation)
	if !reflect.DeepEqual(receipt.Claims, expectedClaims) {
		return fmt.Errorf("claim identity or declaration status changed")
	}
	expectedEvidence := []Evidence{
		{ID: "evidence:supporting", Kind: "SUPPORTING", Digest: digestString("supporting evidence for claim:supporting-discharge")},
		{ID: "evidence:contradicting", Kind: "CONTRADICTING", Digest: digestString("contradicting evidence for claim:contradicting-refutation")},
	}
	if !reflect.DeepEqual(receipt.Evidence, expectedEvidence) {
		return fmt.Errorf("evidence identity changed")
	}
	if err := validateLedger(receipt); err != nil {
		return err
	}
	expectedCases := expectedCaseResults(receipt)
	if !reflect.DeepEqual(receipt.Cases, expectedCases) {
		return fmt.Errorf("case classification changed")
	}
	expectedSummary := summarize(receipt.Claims, receipt.Transitions, receipt.Cases)
	if receipt.Summary != expectedSummary {
		return fmt.Errorf("summary denominator or case partition changed")
	}
	expectedMetrics := buildMetrics(receipt)
	if !reflect.DeepEqual(receipt.Metrics, expectedMetrics) {
		return fmt.Errorf("metric numerator, denominator, or provenance changed")
	}
	if receipt.Effects != (Effects{RepositoryWrites: 0, MutationAuthority: false}) {
		return fmt.Errorf("read-only effect boundary changed")
	}
	if receipt.Decision != (Decision{Value: "PASS", Resolution: "LIFECYCLE_CASES_EXACT", Reason: "CLAIMS_PRESERVED_WITH_EXPLICIT_OPEN_DISCHARGED_REFUTED_TRANSITIONS"}) {
		return fmt.Errorf("receipt decision changed")
	}
	if receipt.ReceiptDigest != digestWithoutReceipt(receipt) || !validDigest(receipt.ReceiptDigest) {
		return fmt.Errorf("receipt digest changed")
	}
	return nil
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
	relation.Digest = digestWithoutRelation(relation)
	return relation, nil
}

func names(refs []syntax.NameRef) []string {
	result := make([]string, len(refs))
	for index, ref := range refs {
		result[index] = ref.Name
	}
	return result
}

func claimsFor(relation SourceRelation) []Claim {
	claims := make([]Claim, 0, claimTotal)
	for _, activity := range relation.Activities {
		statement := statementOf(activity)
		claims = append(claims, Claim{ID: activity.ClaimID, Statement: statement, DeclarationDigest: digestJSON(struct {
			ID, Statement string
		}{activity.ClaimID, statement}), Status: "OPEN"})
	}
	claims[1].Status = "DISCHARGED"
	claims[2].Status = "REFUTED"
	return claims
}

func validateLedger(receipt Receipt) error {
	if len(receipt.Claims) != claimTotal || len(receipt.Transitions) != claimTotal*2 || len(receipt.CauseReceipts) != claimTotal*2 {
		return fmt.Errorf("claim, transition, or cause denominator changed")
	}
	previous := ""
	for index, transition := range receipt.Transitions {
		if transition.Sequence != index+1 || transition.ClaimID != receipt.Claims[claimIndex(index)].ID || transition.DeclarationDigest != receipt.Claims[claimIndex(index)].DeclarationDigest || transition.PreviousDigest != previous || transition.Digest != digestWithoutTransition(transition) || !validDigest(transition.Digest) {
			return fmt.Errorf("transition %d is not append-only or claim-bound", index+1)
		}
		cause := receipt.CauseReceipts[index]
		if cause.Sequence != index+1 || cause.ClaimID != transition.ClaimID || cause.Digest != digestWithoutCause(cause) || transition.CauseReceiptDigest != cause.Digest || !validDigest(cause.Digest) {
			return fmt.Errorf("cause receipt %d is not bound to its transition", index+1)
		}
		if err := validateEvent(index, transition, cause, receipt); err != nil {
			return err
		}
		previous = transition.Digest
	}
	return nil
}

func validateEvent(index int, transition Transition, cause CauseReceipt, receipt Receipt) error {
	if index < claimTotal {
		if transition.Event != "CLAIM_OPENED" || transition.Before != "UNRECORDED" || transition.After != "OPEN" || transition.EvidenceDigest != "" || cause.Kind != "DECLARATION" || cause.Coordinate != (Coordinate{Stage: "CLAIM", Step: "declare", Reason: "CLAIM_OPENED"}) || cause.Reason != "Gooo activity declaration created a durable claim slot" {
			return fmt.Errorf("claim registration %d changed", index+1)
		}
		return nil
	}
	resolutionIndex := index - claimTotal
	expected := []struct {
		event, after, kind, reason, coordinateStage, coordinateStep, coordinateReason string
		evidenceDigest                                                                string
		dependencies                                                                  []string
	}{
		{"EVIDENCE_UNAVAILABLE", "OPEN", "DIRECT_UNKNOWN", "no evidence was supplied; the claim remains", "EVIDENCE", "observe", "EVIDENCE_UNAVAILABLE", "", nil},
		{"EVIDENCE_ACCEPTED", "DISCHARGED", "SUPPORTING_EVIDENCE", "supporting evidence closes the claim as discharged", "EVIDENCE", "classify", "SUPPORTING_EVIDENCE", receipt.Evidence[0].Digest, nil},
		{"EVIDENCE_CONTRADICTED", "REFUTED", "CONTRADICTING_EVIDENCE", "contradicting evidence closes the claim as refuted", "EVIDENCE", "classify", "CONTRADICTING_EVIDENCE", receipt.Evidence[1].Digest, nil},
		{"EVIDENCE_UNAVAILABLE", "OPEN", "DIRECT_UNKNOWN", "direct evidence is unavailable; the claim remains open", "EVIDENCE", "observe", "EVIDENCE_UNAVAILABLE", "", nil},
		{"EVIDENCE_DEPENDENCY_BLOCKED", "OPEN", "DEPENDENCY_BLOCKED", "dependent resolution is blocked by an open upstream claim", "RESOLVE", "dependency", "UPSTREAM_CLAIM_OPEN", "", []string{receipt.Claims[3].ID}},
		{"EVIDENCE_AMBIGUOUS", "OPEN", "AMBIGUOUS_EVIDENCE", "ambiguous evidence cannot close a claim", "EVIDENCE", "classify", "AMBIGUOUS_EVIDENCE", "", nil},
	}[resolutionIndex]
	if transition.Event != expected.event || transition.Before != "OPEN" || transition.After != expected.after || transition.EvidenceDigest != expected.evidenceDigest || cause.Kind != expected.kind || cause.Reason != expected.reason || cause.Coordinate != (Coordinate{Stage: expected.coordinateStage, Step: expected.coordinateStep, Reason: expected.coordinateReason}) || !reflect.DeepEqual(cause.DependencyClaimIDs, expected.dependencies) {
		return fmt.Errorf("resolution event %d changed", resolutionIndex+1)
	}
	if expected.evidenceDigest != "" && !reflect.DeepEqual(cause.EvidenceIDs, []string{receipt.Evidence[resolutionIndex-1].ID}) {
		return fmt.Errorf("resolution evidence cause %d changed", resolutionIndex+1)
	}
	return nil
}

func expectedCaseResults(receipt Receipt) []CaseResult {
	resolution := receipt.Transitions[claimTotal:]
	causes := receipt.CauseReceipts[claimTotal:]
	return []CaseResult{
		{ID: "declared-claim-without-evidence", ClaimID: resolution[0].ClaimID, ExpectedStatus: "OPEN", ObservedStatus: resolution[0].After, Decision: "UNKNOWN", ObservedEvent: resolution[0].Event, CauseKind: causes[0].Kind, Coordinate: causes[0].Coordinate, Reason: "direct evidence is absent"},
		{ID: "supporting-evidence-closes", ClaimID: resolution[1].ClaimID, ExpectedStatus: "DISCHARGED", ObservedStatus: resolution[1].After, Decision: "PASS", ObservedEvent: resolution[1].Event, CauseKind: causes[1].Kind, Coordinate: causes[1].Coordinate, Reason: "supporting evidence discharges the preserved claim"},
		{ID: "conflicting-evidence-closes-refuted", ClaimID: resolution[2].ClaimID, ExpectedStatus: "REFUTED", ObservedStatus: resolution[2].After, Decision: "PASS", ObservedEvent: resolution[2].Event, CauseKind: causes[2].Kind, Coordinate: causes[2].Coordinate, Reason: "conflicting evidence refutes without deleting the claim"},
		{ID: "direct-missing-evidence", ClaimID: resolution[3].ClaimID, ExpectedStatus: "OPEN", ObservedStatus: resolution[3].After, Decision: "UNKNOWN", ObservedEvent: resolution[3].Event, CauseKind: causes[3].Kind, Coordinate: causes[3].Coordinate, Reason: "direct evidence is unavailable"},
		{ID: "dependency-missing-evidence", ClaimID: resolution[4].ClaimID, ExpectedStatus: "OPEN", ObservedStatus: resolution[4].After, Decision: "UNKNOWN", ObservedEvent: resolution[4].Event, CauseKind: causes[4].Kind, Coordinate: causes[4].Coordinate, Reason: "dependency is blocked by a direct unknown"},
		{ID: "ambiguous-evidence-refuses-closure", ClaimID: resolution[5].ClaimID, ExpectedStatus: "OPEN", ObservedStatus: resolution[5].After, Decision: "FAIL_CLOSED", ObservedEvent: resolution[5].Event, CauseKind: causes[5].Kind, Coordinate: causes[5].Coordinate, Reason: "ambiguous evidence fails closed"},
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

func claimIndex(index int) int {
	if index >= claimTotal {
		return index - claimTotal
	}
	return index
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

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
