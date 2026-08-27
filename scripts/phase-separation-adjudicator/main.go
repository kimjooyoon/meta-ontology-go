package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	schema                 = "gooo/meta.phase-separation-witness/v2"
	decisionPass           = "PASS"
	decisionUnknown        = "UNKNOWN"
	resolutionExact        = "EXACT"
	resolutionLower        = "LOWER_RESOLUTION"
	reasonExact            = "PHASE_SEPARATION_WITNESS_EXACT"
	reasonUnknownContract  = "UNKNOWN_SOURCE_CONTRACT"
	reasonUnknownSyntax    = "UNKNOWN_SOURCE_SYNTAX"
	toolchain              = "go1.27.0"
	producerID             = "source-authority"
	consumerID             = "independent-adjudicator"
	metaOperationID        = "preserve-explicit-claim"
	proofChoiceID          = "boundary-receipt"
	stateOpen              = "OPEN"
	stateDischarged        = "DISCHARGED"
	stateRefuted           = "REFUTED"
	payloadClaim           = "claim"
	payloadValue           = "value"
	payloadAuthority       = "authority"
	payloadEvidence        = "evidence"
	expectedSourceCases    = 6
	expectedCleanCases     = 1
	expectedLeakageCases   = 5
	expectedClaimTransfers = 2
	expectedIndicators     = 12
	expectedViews          = 3
	expectedProofs         = 3
	zeroDigest             = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
)

var expectedLeakReasons = map[string]string{
	payloadValue:     "VALUE_CROSSES_PHASE",
	payloadAuthority: "AUTHORITY_CROSSES_PHASE",
	payloadEvidence:  "EVIDENCE_CROSSES_PHASE",
	"phase-skip":     "PHASE_EDGE_SKIPS",
	"phase-reverse":  "PHASE_EDGE_REVERSES",
}

var expectedLiteralClass = map[string]string{"source": "declared", "expansion": "expanded", "execution": "observed"}

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type sourceRecord struct {
	ActivityName     string
	ActivityID       string
	CaseKey          string
	TransferID       string
	ValueID          string
	FromValueID      string
	ToValueID        string
	LiteralClass     string
	FromLiteralClass string
	ToLiteralClass   string
	FromPhase        string
	ToPhase          string
	PayloadClass     string
	ClaimDigest      string
	TargetDigest     string
	Provenance       string
	ClaimStateFrom   string
	ClaimStateTo     string
	Stage            string
	Step             string
	DeclaredReason   string
	Program          string
}

type parsedFile struct {
	Filename     string
	Source       []byte
	File         *syntax.File
	IR           semantic.IR
	SemanticHash string
	Activities   []sourceRecord
	EntityIDs    map[string]string
}

type derivedRecord struct {
	sourceRecord
	EvidenceDigest string
	PreviousDigest string
}

type caseResult struct {
	Name            string   `json:"name"`
	Class           string   `json:"class"`
	Outcome         string   `json:"outcome"`
	Reason          string   `json:"reason"`
	Stage           string   `json:"stage"`
	Step            string   `json:"step"`
	ClaimState      string   `json:"claim_state"`
	TransferCount   int      `json:"transfer_count"`
	TransferIDs     []string `json:"transfer_ids"`
	ValueIDs        []string `json:"value_ids"`
	PayloadClasses  []string `json:"payload_classes"`
	EvidenceDigests []string `json:"evidence_digests"`
	Provenances     []string `json:"provenances"`
	PreviousDigests []string `json:"previous_digests"`
	Passed          bool     `json:"passed"`
}

type claimTransition struct {
	ID             string `json:"id"`
	FromPhase      string `json:"from_phase"`
	ToPhase        string `json:"to_phase"`
	FromClaim      string `json:"from_claim"`
	ToClaim        string `json:"to_claim"`
	FromState      string `json:"from_state"`
	ToState        string `json:"to_state"`
	ClaimDigest    string `json:"claim_digest"`
	TargetDigest   string `json:"target_digest"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	PreviousDigest string `json:"previous_digest"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	Preserved      bool   `json:"preserved"`
}

type indicator struct {
	ID            string `json:"id"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Satisfied     bool   `json:"satisfied"`
}

type view struct {
	Audience      string `json:"audience"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Satisfied     int    `json:"satisfied"`
	Total         int    `json:"total"`
	BasisPoints   int    `json:"basis_points"`
}

type proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	Passed         bool   `json:"passed"`
}

type intervention struct {
	Kind                    string `json:"kind"`
	BaseSemanticDigest      string `json:"base_semantic_digest"`
	VariantSemanticDigest   string `json:"variant_semantic_digest"`
	BaseDecision            string `json:"base_decision"`
	VariantDecision         string `json:"variant_decision"`
	BaseTransitionDigest    string `json:"base_transition_digest"`
	VariantTransitionDigest string `json:"variant_transition_digest"`
	Changed                 bool   `json:"changed"`
	Preserved               bool   `json:"preserved"`
	Passed                  bool   `json:"passed"`
	Numerator               int    `json:"numerator"`
	Denominator             int    `json:"denominator"`
}

type summary struct {
	SourceCasesProcessed         int `json:"source_cases_processed"`
	SourceCasesTotal             int `json:"source_cases_total"`
	CleanCasesPassed             int `json:"clean_cases_passed"`
	CleanCasesTotal              int `json:"clean_cases_total"`
	LeakageRejections            int `json:"leakage_rejections"`
	LeakageRejectionsTotal       int `json:"leakage_rejections_total"`
	ClaimTransitionsPreserved    int `json:"claim_transitions_preserved"`
	ClaimTransitionsTotal        int `json:"claim_transitions_total"`
	ExplicitClaimTransfers       int `json:"explicit_claim_transfers"`
	ExplicitClaimTransfersTotal  int `json:"explicit_claim_transfers_total"`
	IndicatorsSatisfied          int `json:"indicators_satisfied"`
	IndicatorsTotal              int `json:"indicators_total"`
	SemanticCausality            int `json:"semantic_causality"`
	SemanticCausalityTotal       int `json:"semantic_causality_total"`
	NonsemanticPreservation      int `json:"nonsemantic_preservation"`
	NonsemanticPreservationTotal int `json:"nonsemantic_preservation_total"`
	UnknownCases                 int `json:"unknown_cases"`
	RepositoryWrites             int `json:"repository_writes"`
}

type authority struct {
	Execution bool `json:"execution"`
	Mutation  bool `json:"mutation"`
	Promotion bool `json:"promotion"`
}

type ciSnapshot struct {
	RepositoryWrites   int    `json:"repository_writes"`
	MutationAuthority  bool   `json:"mutation_authority"`
	PromotionAuthority bool   `json:"promotion_authority"`
	ExecutionAuthority bool   `json:"execution_authority"`
	Permissions        string `json:"permissions"`
	BeforeStatusDigest string `json:"before_status_digest"`
	AfterStatusDigest  string `json:"after_status_digest"`
}

type unknownResult struct {
	Decision       string     `json:"decision"`
	Coordinate     coordinate `json:"coordinate"`
	ClaimState     string     `json:"claim_state"`
	EvidenceDigest string     `json:"evidence_digest"`
	Provenance     string     `json:"provenance"`
	PreviousDigest string     `json:"previous_digest"`
}

type receipt struct {
	Schema                  string            `json:"schema"`
	Decision                string            `json:"decision"`
	Reason                  string            `json:"reason"`
	Resolution              string            `json:"resolution"`
	HeadSHA                 string            `json:"head_sha"`
	Toolchain               string            `json:"toolchain"`
	SourcePath              string            `json:"source_path"`
	SourceDigest            string            `json:"source_digest"`
	LeakSourcePath          string            `json:"leak_source_path"`
	LeakSourceDigest        string            `json:"leak_source_digest"`
	UnknownSourcePath       string            `json:"unknown_source_path"`
	UnknownSourceDigest     string            `json:"unknown_source_digest"`
	Producer                string            `json:"producer"`
	Consumer                string            `json:"consumer"`
	MetaOperation           string            `json:"meta_operation"`
	ProofChoice             string            `json:"proof_choice"`
	Cases                   []caseResult      `json:"cases"`
	Transitions             []claimTransition `json:"claim_transitions"`
	Indicators              []indicator       `json:"indicators"`
	Views                   []view            `json:"views"`
	Proofs                  []proof           `json:"proofs"`
	SemanticIntervention    intervention      `json:"semantic_intervention"`
	NonsemanticIntervention intervention      `json:"nonsemantic_intervention"`
	Summary                 summary           `json:"summary"`
	Authority               authority         `json:"authority"`
	CISnapshot              ciSnapshot        `json:"ci_snapshot"`
	Unknown                 unknownResult     `json:"unknown"`
	Coordinate              coordinate        `json:"coordinate"`
	Digest                  string            `json:"digest"`
}

type evaluation struct {
	Decision         string
	Cases            []caseResult
	Transitions      []claimTransition
	Summary          summary
	TransitionDigest string
}

func main() {
	source := flag.String("source", "examples/phase-separation-witness/main.gooo", "clean Gooo source")
	leaks := flag.String("leaks", "examples/phase-separation-witness/leaks.gooo", "leakage Gooo source")
	unknown := flag.String("unknown", "examples/phase-separation-witness/unknown.gooo", "UNKNOWN Gooo source")
	receiptPath := flag.String("receipt", "", "producer receipt")
	snapshotPath := flag.String("ci-snapshot", "", "independent CI observation")
	expectedHead := flag.String("expected-head", "", "exact source commit")
	flag.Parse()
	if err := run(*source, *leaks, *unknown, *receiptPath, *snapshotPath, *expectedHead); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("phase separation independent adjudicator: PASS")
}

func run(sourcePath, leaksPath, unknownPath, receiptPath, snapshotPath, expectedHead string) error {
	if receiptPath == "" || snapshotPath == "" || expectedHead == "" {
		return fmt.Errorf("receipt, ci-snapshot, and expected-head are required")
	}
	mainFile, err := readSource(sourcePath)
	if err != nil {
		return err
	}
	leakFile, err := readSource(leaksPath)
	if err != nil {
		return err
	}
	unknownFile, err := readSource(unknownPath)
	if err != nil {
		return err
	}
	var snapshot ciSnapshot
	snapshotBytes, err := os.ReadFile(snapshotPath)
	if err != nil {
		return fmt.Errorf("read CI snapshot: %w", err)
	}
	if err := json.Unmarshal(snapshotBytes, &snapshot); err != nil {
		return fmt.Errorf("decode CI snapshot: %w", err)
	}
	data, err := os.ReadFile(receiptPath)
	if err != nil {
		return fmt.Errorf("read receipt: %w", err)
	}
	var got receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&got); err != nil {
		return fmt.Errorf("decode receipt: %w", err)
	}
	if got.Digest != digestReceipt(got) {
		return fmt.Errorf("receipt digest mismatch")
	}
	if got.HeadSHA != expectedHead || got.Schema != schema || got.Toolchain != toolchain {
		return fmt.Errorf("receipt identity is not exact")
	}
	base := evaluateCorpus(mainFile.Activities, leakFile.Activities)
	want := reconstructReceipt(sourcePath, mainFile, leaksPath, leakFile, unknownPath, unknownFile, expectedHead, snapshot, base)
	if got.SourceDigest != digestBytes(mainFile.Source) || got.LeakSourceDigest != digestBytes(leakFile.Source) || got.UnknownSourceDigest != digestBytes(unknownFile.Source) {
		return fmt.Errorf("receipt source digests do not match checked-out source")
	}
	if got.SourcePath != sourcePath || got.LeakSourcePath != leaksPath || got.UnknownSourcePath != unknownPath {
		return fmt.Errorf("receipt source paths do not match checked-out source")
	}
	if got.Producer != producerID || got.Consumer != consumerID || got.MetaOperation != metaOperationID || got.ProofChoice != proofChoiceID {
		return fmt.Errorf("receipt authority identity is not exact")
	}
	if !reflect.DeepEqual(got.Cases, want.Cases) || !reflect.DeepEqual(got.Transitions, want.Transitions) || !reflect.DeepEqual(got.Summary, want.Summary) || !reflect.DeepEqual(got.Indicators, want.Indicators) || !reflect.DeepEqual(got.Views, want.Views) || !reflect.DeepEqual(got.Proofs, want.Proofs) || !reflect.DeepEqual(got.SemanticIntervention, want.SemanticIntervention) || !reflect.DeepEqual(got.NonsemanticIntervention, want.NonsemanticIntervention) || !reflect.DeepEqual(got.Unknown, want.Unknown) {
		return fmt.Errorf("receipt is not a source-derived reconstruction")
	}
	if got.Authority != want.Authority || got.CISnapshot != snapshot {
		return fmt.Errorf("receipt CI authority observation is not exact")
	}
	if want.Decision != decisionPass || got.Decision != want.Decision || got.Reason != want.Reason || got.Resolution != want.Resolution || got.Coordinate != want.Coordinate {
		return fmt.Errorf("source-derived decision is not exact")
	}
	return nil
}

func readSource(filename string) (parsedFile, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return parsedFile{}, fmt.Errorf("read source %s: %w", filename, err)
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return parsedFile{}, fmt.Errorf("%s: syntax.ParseFile: %v", filename, diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return parsedFile{}, fmt.Errorf("%s: bidir.Lower: %w", filename, err)
	}
	parsed := parsedFile{Filename: filename, Source: append([]byte(nil), source...), File: file, IR: ir, SemanticHash: ir.StableHash(), EntityIDs: make(map[string]string)}
	for _, declaration := range file.Declarations {
		if entity, ok := declaration.(*syntax.EntityDecl); ok {
			parsed.EntityIDs[entity.Name] = entity.ID
		}
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if len(activity.Inputs) != 1 || activity.Output == "" {
			return parsedFile{}, fmt.Errorf("%s: activity %q is not one-input/one-output", filename, activity.Name)
		}
		fromID, fromOK := parsed.EntityIDs[activity.Inputs[0].Name]
		toID, toOK := parsed.EntityIDs[activity.Output]
		if !fromOK || !toOK {
			return parsedFile{}, fmt.Errorf("%s: activity %q endpoint is not an entity", filename, activity.Name)
		}
		node, ok := activityNode(ir, activity.Name)
		if !ok || node.ValueProgram != activity.ValueProgram {
			return parsedFile{}, fmt.Errorf("%s: activity %q was not retained by bidir.Lower", filename, activity.Name)
		}
		record, err := decodeActivity(filename, activity, fromID, toID, node.ID.String())
		if err != nil {
			return parsedFile{}, err
		}
		parsed.Activities = append(parsed.Activities, record)
	}
	if len(parsed.Activities) == 0 {
		return parsedFile{}, fmt.Errorf("%s: no phase activities", filename)
	}
	return parsed, nil
}

func activityNode(ir semantic.IR, name string) (semantic.Node, bool) {
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity && node.Name == name {
			return node, true
		}
	}
	return semantic.Node{}, false
}

func decodeActivity(filename string, activity *syntax.ActivityDecl, fromID, toID, activityID string) (sourceRecord, error) {
	fields, err := parseComputes(activity.ValueProgram)
	if err != nil {
		return sourceRecord{}, fmt.Errorf("%s: activity %q: %w", filename, activity.Name, err)
	}
	record := sourceRecord{ActivityName: activity.Name, ActivityID: activityID, Program: activity.ValueProgram, CaseKey: fields["case"], TransferID: fields["transfer_id"], ValueID: fields["value_id"], FromValueID: fields["from_value_id"], ToValueID: fields["to_value_id"], LiteralClass: fields["literal_class"], FromLiteralClass: fields["from_literal_class"], ToLiteralClass: fields["to_literal_class"], FromPhase: fields["from_phase"], ToPhase: fields["to_phase"], PayloadClass: fields["payload_class"], ClaimDigest: fields["claim_digest"], TargetDigest: fields["target_digest"], Provenance: fields["provenance"], ClaimStateFrom: fields["claim_state_from"], ClaimStateTo: fields["claim_state_to"], Stage: fields["stage"], Step: fields["step"], DeclaredReason: fields["reason"]}
	if record.CaseKey == "" || record.TransferID == "" || record.ValueID != record.FromValueID || record.FromValueID != fromID || record.ToValueID != toID {
		return sourceRecord{}, fmt.Errorf("%s: activity %q: value IDs do not bind to endpoints", filename, activity.Name)
	}
	fromPhase, err := phaseOfID(fromID)
	if err != nil {
		return sourceRecord{}, fmt.Errorf("%s: activity %q: %w", filename, activity.Name, err)
	}
	toPhase, err := phaseOfID(toID)
	if err != nil {
		return sourceRecord{}, fmt.Errorf("%s: activity %q: %w", filename, activity.Name, err)
	}
	if record.FromPhase != fromPhase || record.ToPhase != toPhase || record.LiteralClass != record.FromLiteralClass || expectedLiteralClass[fromPhase] != record.FromLiteralClass || expectedLiteralClass[toPhase] != record.ToLiteralClass {
		return sourceRecord{}, fmt.Errorf("%s: activity %q: phase-local payload disagrees with endpoints", filename, activity.Name)
	}
	if record.PayloadClass != payloadClaim && record.PayloadClass != payloadValue && record.PayloadClass != payloadAuthority && record.PayloadClass != payloadEvidence {
		return sourceRecord{}, fmt.Errorf("%s: activity %q: unknown payload class", filename, activity.Name)
	}
	if record.Provenance == "" || record.Stage == "" || record.Step == "" || record.DeclaredReason == "" || record.ClaimStateFrom != stateOpen || (record.ClaimStateTo != stateOpen && record.ClaimStateTo != stateDischarged && record.ClaimStateTo != stateRefuted) {
		return sourceRecord{}, fmt.Errorf("%s: activity %q: incomplete evidence or lifecycle", filename, activity.Name)
	}
	if record.PayloadClass == payloadClaim {
		if !isDigest(record.ClaimDigest) || !isDigest(record.TargetDigest) {
			return sourceRecord{}, fmt.Errorf("%s: activity %q: claim digest is missing", filename, activity.Name)
		}
	} else if record.ClaimDigest != "none" || record.TargetDigest != "none" {
		return sourceRecord{}, fmt.Errorf("%s: activity %q: non-claim carries claim digest", filename, activity.Name)
	}
	return record, nil
}

func parseComputes(program string) (map[string]string, error) {
	fields := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		key, value, ok := strings.Cut(part, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" || fields[key] != "" {
			return nil, fmt.Errorf("invalid computes field")
		}
		fields[key] = value
	}
	for _, key := range []string{"case", "transfer_id", "value_id", "from_value_id", "to_value_id", "literal_class", "from_literal_class", "to_literal_class", "from_phase", "to_phase", "payload_class", "claim_digest", "target_digest", "provenance", "claim_state_from", "claim_state_to", "stage", "step", "reason"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing computes field %q", key)
		}
	}
	return fields, nil
}

func phaseOfID(id string) (string, error) {
	_, path, ok := strings.Cut(id, "://")
	if !ok {
		return "", fmt.Errorf("phase value ID %q has no URI path", id)
	}
	parts := strings.Split(path, "/")
	if len(parts) != 2 || expectedLiteralClass[parts[0]] == "" || parts[1] == "" {
		return "", fmt.Errorf("phase value ID %q is not phase-local", id)
	}
	return parts[0], nil
}

func evaluateCorpus(mainRecords, leakRecords []sourceRecord) evaluation {
	all := append(append([]sourceRecord(nil), mainRecords...), leakRecords...)
	annotated := annotate(all)
	byCase := make(map[string][]derivedRecord)
	for _, record := range annotated {
		byCase[record.CaseKey] = append(byCase[record.CaseKey], record)
	}
	result := evaluation{Decision: decisionUnknown}
	result.Summary.SourceCasesProcessed, result.Summary.SourceCasesTotal = len(byCase), len(byCase)
	result.Summary.CleanCasesTotal = boolToInt(len(byCase["clean"]) > 0)
	result.Summary.LeakageRejectionsTotal = len(leakRecords)
	result.Summary.ClaimTransitionsTotal = len(byCase["clean"])
	result.Summary.ExplicitClaimTransfersTotal = len(byCase["clean"])
	clean, transitions := evaluateClean(byCase["clean"])
	result.Cases = append(result.Cases, clean)
	result.Transitions = transitions
	result.Summary.CleanCasesPassed = boolToInt(clean.Passed)
	result.Summary.ClaimTransitionsPreserved = countPreserved(transitions)
	result.Summary.ExplicitClaimTransfers = countPreserved(transitions)
	for _, name := range []string{"value-leak", "authority-leak", "evidence-leak", "phase-skip", "phase-reverse"} {
		caseResult := evaluateLeak(name, byCase[name])
		result.Cases = append(result.Cases, caseResult)
		result.Summary.LeakageRejections += boolToInt(caseResult.Passed)
	}
	result.TransitionDigest = digestValue(result.Transitions)
	if result.Summary.CleanCasesPassed == 1 && result.Summary.LeakageRejections == result.Summary.LeakageRejectionsTotal && result.Summary.ClaimTransitionsPreserved == result.Summary.ClaimTransitionsTotal && result.Summary.ExplicitClaimTransfers == result.Summary.ExplicitClaimTransfersTotal {
		result.Decision = decisionPass
	}
	return result
}

func annotate(records []sourceRecord) []derivedRecord {
	result := make([]derivedRecord, 0, len(records))
	previous := zeroDigest
	for _, record := range records {
		evidence := evidenceDigest(record)
		result = append(result, derivedRecord{sourceRecord: record, EvidenceDigest: evidence, PreviousDigest: previous})
		previous = evidence
	}
	return result
}

func evaluateClean(records []derivedRecord) (caseResult, []claimTransition) {
	result := caseFromRecords("clean", "CLEAN", records)
	transitions := make([]claimTransition, 0, len(records))
	valid := len(records) == expectedClaimTransfers
	for index, record := range records {
		transition, reason := deriveTransition(record)
		if index == 0 && record.ClaimStateTo != stateOpen {
			valid = false
		}
		if index == len(records)-1 && record.ClaimStateTo != stateDischarged {
			valid = false
		}
		if reason != "" {
			valid = false
		}
		transitions = append(transitions, transition)
	}
	result.Stage, result.Step = "EXECUTION", "ADJUDICATE"
	if valid {
		result.Outcome, result.ClaimState, result.Reason, result.Passed = stateDischarged, stateDischarged, "EXPLICIT_CLAIM_DISCHARGED", true
	} else {
		result.Outcome, result.ClaimState, result.Reason = stateRefuted, stateRefuted, "CLAIM_TRANSFER_NOT_PRESERVED"
	}
	return result, transitions
}

func evaluateLeak(name string, records []derivedRecord) caseResult {
	result := caseFromRecords(name, "LEAKAGE", records)
	result.Stage, result.Step = "EXECUTION", "ADJUDICATE"
	if len(records) != 1 {
		result.Outcome, result.ClaimState, result.Reason = stateRefuted, stateRefuted, reasonUnknownContract
		return result
	}
	result.Outcome, result.ClaimState, result.Reason = stateRefuted, stateRefuted, deriveLeakReason(records[0])
	result.Passed = isKnownLeakReason(result.Reason)
	return result
}

func caseFromRecords(name, class string, records []derivedRecord) caseResult {
	result := caseResult{Name: name, Class: class, TransferCount: len(records)}
	for _, record := range records {
		result.TransferIDs = append(result.TransferIDs, record.TransferID)
		result.ValueIDs = append(result.ValueIDs, record.ValueID, record.ToValueID)
		result.PayloadClasses = append(result.PayloadClasses, record.PayloadClass)
		result.EvidenceDigests = append(result.EvidenceDigests, record.EvidenceDigest)
		result.Provenances = append(result.Provenances, record.Provenance)
		result.PreviousDigests = append(result.PreviousDigests, record.PreviousDigest)
	}
	return result
}

func deriveTransition(record derivedRecord) (claimTransition, string) {
	transition := claimTransition{ID: record.TransferID, FromPhase: record.FromPhase, ToPhase: record.ToPhase, FromClaim: record.FromValueID, ToClaim: record.ToValueID, FromState: record.ClaimStateFrom, ToState: record.ClaimStateTo, ClaimDigest: claimDigest(record.sourceRecord), TargetDigest: targetDigest(record.sourceRecord), EvidenceDigest: record.EvidenceDigest, Provenance: record.Provenance, PreviousDigest: record.PreviousDigest, MetaOperation: metaOperationID, ProofChoice: proofChoiceID}
	if record.PayloadClass != payloadClaim {
		return transition, leakReason(record.PayloadClass)
	}
	if !allowedClaimEdge(record.FromPhase, record.ToPhase) {
		return transition, edgeReason(record.FromPhase, record.ToPhase)
	}
	if record.ClaimDigest != transition.ClaimDigest || record.TargetDigest != transition.TargetDigest {
		return transition, "CLAIM_DIGEST_MISMATCH"
	}
	if record.Provenance == "" || record.ClaimStateFrom != stateOpen {
		return transition, "CLAIM_PROVENANCE_MISMATCH"
	}
	transition.Preserved = true
	return transition, ""
}

func deriveLeakReason(record derivedRecord) string {
	if record.PayloadClass != payloadClaim {
		return leakReason(record.PayloadClass)
	}
	return edgeReason(record.FromPhase, record.ToPhase)
}

func allowedClaimEdge(from, to string) bool {
	return (from == "source" && to == "expansion") || (from == "expansion" && to == "execution")
}

func edgeReason(from, to string) string {
	if from == "source" && to == "execution" {
		return expectedLeakReasons["phase-skip"]
	}
	if from == "execution" && to == "expansion" {
		return expectedLeakReasons["phase-reverse"]
	}
	return "PHASE_EDGE_INVALID"
}

func leakReason(payload string) string { return expectedLeakReasons[payload] }

func isKnownLeakReason(reason string) bool {
	for _, expected := range expectedLeakReasons {
		if reason == expected {
			return true
		}
	}
	return false
}

func reconstructReceipt(sourcePath string, mainFile parsedFile, leaksPath string, leakFile parsedFile, unknownPath string, unknownFile parsedFile, head string, snapshot ciSnapshot, base evaluation) receipt {
	result := receipt{Schema: schema, Decision: decisionUnknown, Reason: reasonUnknownContract, Resolution: resolutionLower, HeadSHA: head, Toolchain: toolchain, SourcePath: sourcePath, SourceDigest: digestBytes(mainFile.Source), LeakSourcePath: leaksPath, LeakSourceDigest: digestBytes(leakFile.Source), UnknownSourcePath: unknownPath, UnknownSourceDigest: digestBytes(unknownFile.Source), Producer: producerID, Consumer: consumerID, MetaOperation: metaOperationID, ProofChoice: proofChoiceID, Cases: base.Cases, Transitions: base.Transitions, Summary: base.Summary, CISnapshot: snapshot, Authority: authority{Execution: snapshot.ExecutionAuthority, Mutation: snapshot.MutationAuthority, Promotion: snapshot.PromotionAuthority}, Coordinate: coordinate{Stage: "SOURCE", Step: "DECODE", Reason: reasonUnknownContract}}
	result.Indicators, result.Views, result.Proofs = buildEvidence(base, result)
	result.SemanticIntervention = semanticIntervention(mainFile, leakFile, base)
	result.NonsemanticIntervention = nonsemanticIntervention(mainFile, leakFile, base)
	result.Summary.SemanticCausality, result.Summary.SemanticCausalityTotal = result.SemanticIntervention.Numerator, result.SemanticIntervention.Denominator
	result.Summary.NonsemanticPreservation, result.Summary.NonsemanticPreservationTotal = result.NonsemanticIntervention.Numerator, result.NonsemanticIntervention.Denominator
	result.Indicators, result.Views, result.Proofs = buildEvidence(base, result)
	result.Summary.IndicatorsTotal, result.Summary.IndicatorsSatisfied = len(result.Indicators), countIndicators(result.Indicators)
	result.Unknown = deriveUnknown(unknownFile.Activities)
	result.Summary.UnknownCases = boolToInt(result.Unknown.Decision == decisionUnknown)
	result.Summary.RepositoryWrites = snapshot.RepositoryWrites
	if exact(result) {
		result.Decision, result.Reason, result.Resolution, result.Coordinate = decisionPass, reasonExact, resolutionExact, coordinate{Stage: "EXECUTION", Step: "ADJUDICATE", Reason: reasonExact}
	}
	return result
}

func buildEvidence(base evaluation, report receipt) ([]indicator, []view, []proof) {
	evidenceOK := true
	for _, result := range base.Cases {
		evidenceOK = evidenceOK && len(result.EvidenceDigests) == result.TransferCount && len(result.Provenances) == result.TransferCount && len(result.PreviousDigests) == result.TransferCount
	}
	claimLifecycleOK := true
	for _, transition := range base.Transitions {
		claimLifecycleOK = claimLifecycleOK && transition.FromState == stateOpen && (transition.ToState == stateOpen || transition.ToState == stateDischarged) && transition.Preserved
	}
	values := []bool{report.Producer == producerID, report.Consumer == consumerID, base.Summary.SourceCasesProcessed == expectedSourceCases && base.Summary.SourceCasesTotal == expectedSourceCases, base.Summary.CleanCasesPassed == base.Summary.CleanCasesTotal && base.Summary.CleanCasesTotal == expectedCleanCases, base.Summary.LeakageRejections == base.Summary.LeakageRejectionsTotal && base.Summary.LeakageRejectionsTotal == expectedLeakageCases, base.Summary.ClaimTransitionsPreserved == base.Summary.ClaimTransitionsTotal && base.Summary.ClaimTransitionsTotal == expectedClaimTransfers, evidenceOK, claimLifecycleOK, report.SemanticIntervention.Passed, report.NonsemanticIntervention.Passed, report.CISnapshot.RepositoryWrites == 0 && !report.CISnapshot.MutationAuthority, !report.CISnapshot.PromotionAuthority && !report.CISnapshot.ExecutionAuthority}
	indicators := make([]indicator, 0, len(values))
	for index, satisfied := range values {
		indicators = append(indicators, indicator{ID: fmt.Sprintf("PHASE-%02d", index+1), MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Numerator: boolToInt(satisfied), Denominator: 1, Satisfied: satisfied})
	}
	views := []view{{Audience: "PRODUCER", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: base.Summary.SourceCasesProcessed, Total: base.Summary.SourceCasesTotal}, {Audience: "CONSUMER", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: base.Summary.CleanCasesPassed + base.Summary.LeakageRejections, Total: base.Summary.CleanCasesTotal + base.Summary.LeakageRejectionsTotal}, {Audience: "GOVERNOR", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: countIndicators(indicators), Total: len(indicators)}}
	for index := range views {
		views[index].BasisPoints = ratioBasisPoints(views[index].Satisfied, views[index].Total)
	}
	proofs := []proof{{Choice: "BOUNDARY", Claim: "phase-local values reject non-claim payload authority", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(base.Cases), Provenance: report.SourcePath, Passed: base.Summary.LeakageRejections == base.Summary.LeakageRejectionsTotal}, {Choice: "TRANSITION", Claim: "only digest-bound adjacent claims discharge", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(base.Transitions), Provenance: report.SourcePath, Passed: base.Summary.ClaimTransitionsPreserved == base.Summary.ClaimTransitionsTotal}, {Choice: "AUTHORITY", Claim: "CI observes zero repository writes and mutation authority", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(report.CISnapshot), Provenance: report.CISnapshot.Permissions, Passed: report.CISnapshot.RepositoryWrites == 0 && !report.CISnapshot.MutationAuthority && !report.CISnapshot.PromotionAuthority}}
	return indicators, views, proofs
}

func deriveUnknown(records []sourceRecord) unknownResult {
	if len(records) != 1 {
		return unknownResult{Decision: decisionUnknown, Coordinate: coordinate{Stage: "SOURCE", Step: "DECODE", Reason: reasonUnknownContract}, ClaimState: stateOpen, PreviousDigest: zeroDigest}
	}
	record := records[0]
	return unknownResult{Decision: decisionUnknown, Coordinate: coordinate{Stage: record.Stage, Step: record.Step, Reason: record.DeclaredReason}, ClaimState: stateOpen, EvidenceDigest: evidenceDigest(record), Provenance: record.Provenance, PreviousDigest: zeroDigest}
}

func semanticIntervention(mainFile, leakFile parsedFile, base evaluation) intervention {
	result := intervention{Kind: "SEMANTIC", Denominator: 1}
	variantSource := bytes.Replace(mainFile.Source, []byte("payload_class=claim"), []byte("payload_class=value"), 1)
	variantSource = bytes.Replace(variantSource, []byte("claim_digest=sha256:ee3e8ebaa490d076fa230325fe30ef061c629c7cd2e5d5ef41df5e52a06548c3"), []byte("claim_digest=none"), 1)
	variantSource = bytes.Replace(variantSource, []byte("target_digest=sha256:f71af2266668baa89342e7740fe22b77b58f5aa612fe716b7fe9257380fb34fa"), []byte("target_digest=none"), 1)
	variant, err := readSourceBytes(mainFile.Filename, variantSource)
	if err != nil {
		return result
	}
	variantEvaluation := evaluateCorpus(variant.Activities, leakFile.Activities)
	result.BaseSemanticDigest, result.VariantSemanticDigest, result.BaseDecision, result.VariantDecision = mainFile.SemanticHash, variant.SemanticHash, base.Decision, variantEvaluation.Decision
	result.BaseTransitionDigest, result.VariantTransitionDigest = base.TransitionDigest, variantEvaluation.TransitionDigest
	result.Changed = result.BaseSemanticDigest != result.VariantSemanticDigest && (result.BaseDecision != result.VariantDecision || result.BaseTransitionDigest != result.VariantTransitionDigest)
	result.Numerator, result.Passed = boolToInt(result.Changed), result.Changed
	return result
}

func nonsemanticIntervention(mainFile, leakFile parsedFile, base evaluation) intervention {
	result := intervention{Kind: "NONSEMANTIC", Denominator: 1}
	variantSource := append(append([]byte(nil), mainFile.Source...), []byte("\n// nonsemantic intervention\n")...)
	variant, err := readSourceBytes(mainFile.Filename, variantSource)
	if err != nil {
		return result
	}
	variantEvaluation := evaluateCorpus(variant.Activities, leakFile.Activities)
	result.BaseSemanticDigest, result.VariantSemanticDigest, result.BaseDecision, result.VariantDecision = mainFile.SemanticHash, variant.SemanticHash, base.Decision, variantEvaluation.Decision
	result.BaseTransitionDigest, result.VariantTransitionDigest = base.TransitionDigest, variantEvaluation.TransitionDigest
	result.Preserved = result.BaseSemanticDigest == result.VariantSemanticDigest && result.BaseDecision == result.VariantDecision && result.BaseTransitionDigest == result.VariantTransitionDigest
	result.Numerator, result.Passed = boolToInt(result.Preserved), result.Preserved
	return result
}

func readSourceBytes(filename string, source []byte) (parsedFile, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return parsedFile{}, fmt.Errorf("%s: syntax.ParseFile: %v", filename, diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return parsedFile{}, err
	}
	parsed := parsedFile{Filename: filename, Source: append([]byte(nil), source...), File: file, IR: ir, SemanticHash: ir.StableHash(), EntityIDs: make(map[string]string)}
	for _, declaration := range file.Declarations {
		if entity, ok := declaration.(*syntax.EntityDecl); ok {
			parsed.EntityIDs[entity.Name] = entity.ID
		}
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		fromID := parsed.EntityIDs[activity.Inputs[0].Name]
		toID := parsed.EntityIDs[activity.Output]
		node, ok := activityNode(ir, activity.Name)
		if !ok {
			return parsedFile{}, fmt.Errorf("%s: missing activity node", filename)
		}
		record, err := decodeActivity(filename, activity, fromID, toID, node.ID.String())
		if err != nil {
			return parsedFile{}, err
		}
		parsed.Activities = append(parsed.Activities, record)
	}
	return parsed, nil
}

func exact(report receipt) bool {
	return report.Summary.SourceCasesProcessed == expectedSourceCases && report.Summary.SourceCasesTotal == expectedSourceCases && report.Summary.CleanCasesPassed == expectedCleanCases && report.Summary.CleanCasesTotal == expectedCleanCases && report.Summary.LeakageRejections == expectedLeakageCases && report.Summary.LeakageRejectionsTotal == expectedLeakageCases && report.Summary.ClaimTransitionsPreserved == expectedClaimTransfers && report.Summary.ClaimTransitionsTotal == expectedClaimTransfers && report.Summary.ExplicitClaimTransfers == expectedClaimTransfers && report.Summary.ExplicitClaimTransfersTotal == expectedClaimTransfers && report.Summary.IndicatorsSatisfied == expectedIndicators && report.Summary.IndicatorsTotal == expectedIndicators && report.Summary.SemanticCausality == 1 && report.Summary.SemanticCausalityTotal == 1 && report.Summary.NonsemanticPreservation == 1 && report.Summary.NonsemanticPreservationTotal == 1 && report.Summary.UnknownCases == 1 && report.Summary.RepositoryWrites == 0 && report.Authority == (authority{}) && report.Unknown.Decision == decisionUnknown && report.Unknown.ClaimState == stateOpen && len(report.Cases) == expectedCleanCases+expectedLeakageCases && len(report.Transitions) == expectedClaimTransfers && len(report.Indicators) == expectedIndicators && len(report.Views) == expectedViews && len(report.Proofs) == expectedProofs && allIndicatorsPassed(report.Indicators) && allProofsPassed(report.Proofs)
}

func allIndicatorsPassed(values []indicator) bool {
	for _, value := range values {
		if !value.Satisfied || value.Numerator != value.Denominator {
			return false
		}
	}
	return true
}

func allProofsPassed(values []proof) bool {
	for _, value := range values {
		if !value.Passed {
			return false
		}
	}
	return true
}

func claimDigest(record sourceRecord) string {
	return digestString(strings.Join([]string{"claim", record.CaseKey, record.TransferID, record.FromValueID, record.ToValueID, record.FromPhase, record.ToPhase, record.FromLiteralClass, record.ToLiteralClass, record.Provenance}, "|"))
}

func targetDigest(record sourceRecord) string {
	return digestString(strings.Join([]string{"target", record.ToValueID, record.ToLiteralClass}, "|"))
}

func evidenceDigest(record sourceRecord) string {
	return digestString("source-value-program|" + record.Program)
}

func digestBytes(value []byte) string { return digestString(string(value)) }

func digestValue(value interface{}) string {
	encoded, _ := json.Marshal(value)
	return digestString(string(encoded))
}

func digestReceipt(value receipt) string {
	value.Digest = ""
	encoded, _ := json.Marshal(value)
	return digestString(string(encoded))
}

func digestString(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func countPreserved(values []claimTransition) int {
	count := 0
	for _, value := range values {
		count += boolToInt(value.Preserved)
	}
	return count
}

func countIndicators(values []indicator) int {
	count := 0
	for _, value := range values {
		count += boolToInt(value.Satisfied)
	}
	return count
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func ratioBasisPoints(numerator, denominator int) int {
	if denominator == 0 {
		return 0
	}
	return numerator * 10000 / denominator
}

func isDigest(value string) bool {
	return len(value) == len("sha256:")+64 && strings.HasPrefix(value, "sha256:")
}
