package verify

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	ReportSchema          = "gooo/self-improvement-termination-judge/v2"
	inputSchema           = "gooo/self-improvement-termination-input/v2"
	receiptSchema         = "gooo/self-improvement-termination-receipt/v2"
	metaprogram           = "internal/meta/selfimprovementtermination"
	producer              = "selfimprovementtermination.Evaluate"
	consumer              = "self-improvement-cycle"
	metaOperation         = "prove-self-improvement-termination"
	proofChoice           = "TERMINATION"
	traceStage            = "META_RUN"
	claimStage            = "CLAIM"
	interventionStage     = "INTERVENTION"
	sourcePath            = "examples/self-improvement-termination/main.gooo"
	sourceProgramSchema   = "termination-case/v2"
	interventionSchema    = "termination-intervention/v1"
	maxTraceSteps         = 64
	indicatorTotal        = 2
	decisionFixedPoint    = "FIXED_POINT"
	decisionInProgress    = "IN_PROGRESS"
	decisionCycle         = "CYCLE"
	decisionDivergence    = "DIVERGENCE_POSSIBLE"
	decisionFailClosed    = "FAIL_CLOSED"
	resolutionExact       = "EXACT"
	resolutionLower       = "LOWER_RESOLUTION"
	receiptBound          = "BOUND"
	receiptFailClosed     = "FAIL_CLOSED"
	claimOpen             = "OPEN"
	claimDischarged       = "DISCHARGED"
	claimRefuted          = "REFUTED"
	upstreamNoChange      = "NO_CHANGE"
	upstreamChanged       = "CHANGED"
	reasonNoChange        = "NO_CHANGE_FIXED_POINT_OBSERVED"
	reasonStateChanged    = "METAPROGRAM_STATE_CHANGED"
	reasonCycle           = "REPEATED_STATE_CYCLE_OBSERVED"
	reasonInProgress      = "TRACE_ENDED_BEFORE_TERMINATION"
	reasonDivergence      = "STRICTLY_GROWING_BOUNDARY_NO_FIXED_POINT"
	reasonDecisionUnknown = "FEEDBACK_COVERAGE_DECISION_UNKNOWN"
)

type wireInput struct {
	Schema           string             `json:"schema"`
	Repository       string             `json:"repository"`
	Subject          string             `json:"subject"`
	Producer         string             `json:"producer"`
	Consumer         string             `json:"consumer"`
	MetaOperation    string             `json:"meta_operation"`
	ProofChoice      string             `json:"proof_choice"`
	Stage            string             `json:"stage"`
	Source           wireSource         `json:"source"`
	UpstreamDecision string             `json:"upstream_decision"`
	MaxSteps         int                `json:"max_steps"`
	Trace            []wireObservation  `json:"trace"`
	Interventions    []wireIntervention `json:"interventions"`
}

type wireSource struct {
	Path              string `json:"path"`
	SourceDigest      string `json:"source_digest"`
	SemanticDigest    string `json:"semantic_digest"`
	CaseID            string `json:"case_id"`
	CaseProgramDigest string `json:"case_program_digest"`
}

type wireObservation struct {
	Stage       string `json:"stage"`
	Step        int    `json:"step"`
	BeforeState string `json:"before_state"`
	AfterState  string `json:"after_state"`
	BeforeRank  int    `json:"before_rank"`
	AfterRank   int    `json:"after_rank"`
	Decision    string `json:"decision"`
	Reason      string `json:"reason"`
}

type wireIntervention struct {
	ID                   string `json:"id"`
	Schema               string `json:"schema"`
	Stage                string `json:"stage"`
	Step                 int    `json:"step"`
	Reason               string `json:"reason"`
	SourceBeforeDigest   string `json:"source_before_digest"`
	SourceAfterDigest    string `json:"source_after_digest"`
	SemanticBeforeDigest string `json:"semantic_before_digest"`
	SemanticAfterDigest  string `json:"semantic_after_digest"`
	SourceChanged        bool   `json:"source_changed"`
	SemanticChanged      bool   `json:"semantic_changed"`
}

type wireAuthority struct {
	ReadOnly            bool `json:"read_only"`
	RepositoryWrites    int  `json:"repository_writes"`
	MutationAuthorized  bool `json:"mutation_authorized"`
	PromotionAuthorized bool `json:"promotion_authorized"`
}

type wireOutcome struct {
	ObservedSteps     int    `json:"observed_steps"`
	MaxSteps          int    `json:"max_steps"`
	StateCount        int    `json:"state_count"`
	RepeatedStates    int    `json:"repeated_states"`
	DetectedPeriod    int    `json:"detected_period"`
	FinalState        string `json:"final_state"`
	TerminationProven bool   `json:"termination_proven"`
	ClaimState        string `json:"claim_state"`
}

type wireConformance struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type wireIndicator struct {
	ID            string `json:"id"`
	Route         string `json:"route"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          int    `json:"step"`
	Reason        string `json:"reason"`
	Value         string `json:"value"`
	Limit         string `json:"limit"`
	Satisfied     bool   `json:"satisfied"`
}

type wireClaimTransition struct {
	Stage  string `json:"stage"`
	Step   int    `json:"step"`
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type wireReceipt struct {
	Schema           string                `json:"schema"`
	Metaprogram      string                `json:"metaprogram"`
	Repository       string                `json:"repository"`
	Subject          string                `json:"subject"`
	Producer         string                `json:"producer"`
	Consumer         string                `json:"consumer"`
	MetaOperation    string                `json:"meta_operation"`
	ProofChoice      string                `json:"proof_choice"`
	Stage            string                `json:"stage"`
	Status           string                `json:"status"`
	Resolution       string                `json:"resolution"`
	Decision         string                `json:"decision"`
	Reason           string                `json:"reason"`
	Source           wireSource            `json:"source"`
	UpstreamDecision string                `json:"upstream_decision"`
	InputDigest      string                `json:"input_digest"`
	TraceDigest      string                `json:"trace_digest"`
	Observations     []wireObservation     `json:"observations"`
	Interventions    []wireIntervention    `json:"interventions"`
	ClaimTransitions []wireClaimTransition `json:"claim_transitions"`
	Outcome          wireOutcome           `json:"outcome"`
	Conformance      wireConformance       `json:"conformance"`
	Authority        wireAuthority         `json:"authority"`
	Indicators       []wireIndicator       `json:"indicators"`
	ReceiptDigest    string                `json:"receipt_digest"`
	ReplayDigest     string                `json:"replay_digest"`
}

type Report struct {
	Schema             string                `json:"schema"`
	Independent        bool                  `json:"independent"`
	Repository         string                `json:"repository"`
	Subject            string                `json:"subject"`
	Stage              string                `json:"stage"`
	CaseID             string                `json:"case_id"`
	SourcePath         string                `json:"source_path"`
	SourceDigest       string                `json:"source_digest"`
	SemanticDigest     string                `json:"semantic_digest"`
	Decision           string                `json:"decision"`
	Resolution         string                `json:"resolution"`
	Status             string                `json:"status"`
	Reason             string                `json:"reason"`
	ClaimState         string                `json:"claim_state"`
	TerminationProven  bool                  `json:"termination_proven"`
	Conformance        wireConformance       `json:"conformance"`
	ClaimTransitions   []wireClaimTransition `json:"claim_transitions"`
	ReceiptDigest      string                `json:"receipt_digest"`
	VerificationDigest string                `json:"verification_digest"`
}

type sourceCase struct {
	ID               string
	Program          string
	ProgramDigest    string
	UpstreamDecision string
	MaxSteps         int
	Trace            []wireObservation
}

type sourceModel struct {
	Path           string
	SourceDigest   string
	SemanticDigest string
	Raw            []byte
	Cases          []sourceCase
}

type classification struct {
	decision, resolution, status, reason, finalState, claimState string
	period, stateCount, repeatedStates                           int
	terminationProven                                            bool
}

func Verify(root, requestedSourcePath, repository, caseID string, receiptRaw []byte) (Report, error) {
	if requestedSourcePath != sourcePath {
		return Report{}, fmt.Errorf("independent judge: source path is outside the fixed contract")
	}
	model, err := loadSource(root, requestedSourcePath)
	if err != nil {
		return Report{}, err
	}
	input, err := buildInput(model, repository, caseID)
	if err != nil {
		return Report{}, err
	}
	if err := validateInput(input); err != nil {
		return Report{}, err
	}
	class := classify(input)
	expected := receiptForValidation(input, class)
	actual, err := decodeReceipt(receiptRaw)
	if err != nil {
		return Report{}, err
	}
	expected = seal(expected)
	if !reflect.DeepEqual(actual, expected) {
		return Report{}, fmt.Errorf("independent judge: receipt does not match independently recomputed source outcome")
	}
	report := Report{
		Schema: ReportSchema, Independent: true, Repository: repository, Subject: consumer,
		Stage: traceStage, CaseID: caseID, SourcePath: model.Path,
		SourceDigest: model.SourceDigest, SemanticDigest: model.SemanticDigest,
		Decision: class.decision, Resolution: class.resolution, Status: class.status,
		Reason: class.reason, ClaimState: class.claimState,
		TerminationProven: class.terminationProven, Conformance: expected.Conformance,
		ClaimTransitions: expected.ClaimTransitions, ReceiptDigest: actual.ReceiptDigest,
	}
	report.VerificationDigest = digestJSON(struct {
		Decision, Resolution, Status, Reason, ClaimState, ReceiptDigest string
	}{report.Decision, report.Resolution, report.Status, report.Reason, report.ClaimState, report.ReceiptDigest})
	return report, nil
}

func loadSource(root, requestedPath string) (sourceModel, error) {
	canonicalPath := filepath.ToSlash(filepath.Clean(requestedPath))
	if canonicalPath != sourcePath || filepath.IsAbs(requestedPath) || strings.HasPrefix(canonicalPath, "../") {
		return sourceModel{}, fmt.Errorf("independent judge: source path is outside the repository contract")
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(canonicalPath)))
	if err != nil {
		return sourceModel{}, err
	}
	semanticDigest, ir, err := parseAndLower(canonicalPath, raw)
	if err != nil {
		return sourceModel{}, err
	}
	model := sourceModel{Path: canonicalPath, SourceDigest: digestBytes(raw), SemanticDigest: semanticDigest, Raw: append([]byte(nil), raw...)}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || !strings.HasPrefix(node.ValueProgram, sourceProgramSchema+"|") {
			continue
		}
		item, err := parseSourceCase(node.ValueProgram)
		if err != nil {
			return sourceModel{}, fmt.Errorf("independent judge: source activity %q: %w", node.Name, err)
		}
		item.Program = node.ValueProgram
		item.ProgramDigest = digestBytes([]byte(node.ValueProgram))
		model.Cases = append(model.Cases, item)
	}
	sort.Slice(model.Cases, func(left, right int) bool { return model.Cases[left].ID < model.Cases[right].ID })
	if len(model.Cases) == 0 {
		return sourceModel{}, fmt.Errorf("independent judge: source declares no executable cases")
	}
	for index := 1; index < len(model.Cases); index++ {
		if model.Cases[index-1].ID == model.Cases[index].ID {
			return sourceModel{}, fmt.Errorf("independent judge: source case %q is duplicated", model.Cases[index].ID)
		}
	}
	return model, nil
}

func parseAndLower(filename string, raw []byte) (string, semantic.IR, error) {
	file, diagnostics := syntax.ParseFile(filename, string(raw))
	if err := diagnostics.Error(); err != nil || file == nil {
		if err == nil {
			err = fmt.Errorf("parser returned no source tree")
		}
		return "", semantic.IR{}, fmt.Errorf("independent judge Gooo source: %w", err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", semantic.IR{}, fmt.Errorf("independent judge Gooo semantic lowering: %w", err)
	}
	return "sha256:" + ir.StableHash(), ir, nil
}

func parseSourceCase(program string) (sourceCase, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 5 || parts[0] != sourceProgramSchema {
		return sourceCase{}, fmt.Errorf("value program schema is not %q", sourceProgramSchema)
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || value == "" {
			return sourceCase{}, fmt.Errorf("malformed value program field %q", part)
		}
		if _, exists := values[key]; exists {
			return sourceCase{}, fmt.Errorf("duplicate value program field %q", key)
		}
		values[key] = value
	}
	for _, key := range []string{"id", "max_steps", "upstream", "trace"} {
		if values[key] == "" {
			return sourceCase{}, fmt.Errorf("value program field %q is missing", key)
		}
	}
	maxSteps, err := strconv.Atoi(values["max_steps"])
	if err != nil || maxSteps < 1 || maxSteps > maxTraceSteps {
		return sourceCase{}, fmt.Errorf("value program max_steps is invalid")
	}
	trace, err := parseTrace(values["trace"])
	if err != nil {
		return sourceCase{}, err
	}
	if len(trace) == 0 || len(trace) > maxSteps {
		return sourceCase{}, fmt.Errorf("value program trace length is outside max_steps")
	}
	return sourceCase{ID: values["id"], UpstreamDecision: values["upstream"], MaxSteps: maxSteps, Trace: trace}, nil
}

func parseTrace(raw string) ([]wireObservation, error) {
	items := strings.Split(raw, ";")
	trace := make([]wireObservation, 0, len(items))
	for _, item := range items {
		fields := strings.Split(item, ",")
		if len(fields) != 7 {
			return nil, fmt.Errorf("value program trace observation has %d fields, want 7", len(fields))
		}
		step, stepErr := strconv.Atoi(fields[0])
		beforeRank, beforeErr := strconv.Atoi(fields[3])
		afterRank, afterErr := strconv.Atoi(fields[4])
		if stepErr != nil || beforeErr != nil || afterErr != nil {
			return nil, fmt.Errorf("value program trace has non-integer coordinate")
		}
		trace = append(trace, wireObservation{Stage: traceStage, Step: step,
			BeforeState: stateDigest(fields[1]), AfterState: stateDigest(fields[2]),
			BeforeRank: beforeRank, AfterRank: afterRank, Decision: fields[5], Reason: fields[6]})
	}
	return trace, nil
}

func buildInput(model sourceModel, repository, caseID string) (wireInput, error) {
	for _, item := range model.Cases {
		if item.ID != caseID {
			continue
		}
		interventions, err := buildInterventions(model.Raw, item.Program, model.SemanticDigest)
		if err != nil {
			return wireInput{}, err
		}
		return wireInput{
			Schema: inputSchema, Repository: repository, Subject: consumer,
			Producer: producer, Consumer: consumer, MetaOperation: metaOperation,
			ProofChoice: proofChoice, Stage: traceStage,
			Source: wireSource{Path: model.Path, SourceDigest: model.SourceDigest, SemanticDigest: model.SemanticDigest,
				CaseID: item.ID, CaseProgramDigest: item.ProgramDigest},
			UpstreamDecision: item.UpstreamDecision, MaxSteps: item.MaxSteps,
			Trace: append([]wireObservation(nil), item.Trace...), Interventions: interventions,
		}, nil
	}
	return wireInput{}, fmt.Errorf("independent judge: source case %q is not declared", caseID)
}

func buildInterventions(raw []byte, program, semanticBefore string) ([]wireIntervention, error) {
	semanticSource := strings.Replace(string(raw), program, program+"|intervention=semantic-trace", 1)
	if semanticSource == string(raw) {
		return nil, fmt.Errorf("independent judge: semantic intervention did not find selected computes value")
	}
	semanticAfter, _, err := parseAndLower(sourcePath, []byte(semanticSource))
	if err != nil {
		return nil, err
	}
	commentSource := append([]byte("// nonsemantic comment intervention\n"), raw...)
	commentAfter, _, err := parseAndLower(sourcePath, commentSource)
	if err != nil {
		return nil, err
	}
	return []wireIntervention{
		{ID: "semantic-trace", Schema: interventionSchema, Stage: interventionStage, Step: 1,
			Reason:             "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST",
			SourceBeforeDigest: digestBytes(raw), SourceAfterDigest: digestBytes([]byte(semanticSource)),
			SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter,
			SourceChanged: true, SemanticChanged: semanticBefore != semanticAfter},
		{ID: "nonsemantic-comment", Schema: interventionSchema, Stage: interventionStage, Step: 2,
			Reason:             "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST",
			SourceBeforeDigest: digestBytes(raw), SourceAfterDigest: digestBytes(commentSource),
			SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: commentAfter,
			SourceChanged: true, SemanticChanged: semanticBefore != commentAfter},
	}, nil
}

func validateInput(input wireInput) error {
	if input.Schema != inputSchema || input.Repository == "" || input.Subject != consumer ||
		input.Producer != producer || input.Consumer != consumer || input.MetaOperation != metaOperation ||
		input.ProofChoice != proofChoice || input.Stage != traceStage || input.MaxSteps < 1 ||
		input.MaxSteps > maxTraceSteps || input.Source.Path != sourcePath || input.Source.CaseID == "" ||
		!validDigest(input.Source.SourceDigest) || !validDigest(input.Source.SemanticDigest) ||
		!validDigest(input.Source.CaseProgramDigest) || input.UpstreamDecision == "" {
		return fmt.Errorf("independent judge: source-bound input identity is invalid")
	}
	if len(input.Trace) == 0 || len(input.Trace) > input.MaxSteps {
		return fmt.Errorf("independent judge: trace length is outside budget")
	}
	for index, observation := range input.Trace {
		if observation.Stage != traceStage || observation.Step != index+1 || observation.BeforeRank < 0 ||
			observation.AfterRank < 0 || !validDigest(observation.BeforeState) || !validDigest(observation.AfterState) {
			return fmt.Errorf("independent judge: malformed step %d", index+1)
		}
		if index > 0 && input.Trace[index-1].AfterState != observation.BeforeState {
			return fmt.Errorf("independent judge: broken state chain at step %d", index+1)
		}
		if observation.BeforeState == observation.AfterState {
			if observation.BeforeRank != observation.AfterRank || observation.Decision != upstreamNoChange || observation.Reason != reasonNoChange {
				return fmt.Errorf("independent judge: step %d claims an unbound no-change", index+1)
			}
		} else if observation.Decision != upstreamChanged || observation.Reason != reasonStateChanged {
			return fmt.Errorf("independent judge: step %d claims an unbound change", index+1)
		}
	}
	knownUpstream := input.UpstreamDecision == upstreamNoChange || input.UpstreamDecision == upstreamChanged
	if knownUpstream && input.UpstreamDecision != input.Trace[len(input.Trace)-1].Decision {
		return fmt.Errorf("independent judge: upstream decision disagrees with final observed step")
	}
	if len(input.Interventions) != indicatorTotal {
		return fmt.Errorf("independent judge: intervention denominator is not fixed at %d", indicatorTotal)
	}
	for index, intervention := range input.Interventions {
		if intervention.Schema != interventionSchema || intervention.Stage != interventionStage || intervention.Step != index+1 ||
			!validDigest(intervention.SourceBeforeDigest) || !validDigest(intervention.SourceAfterDigest) ||
			!validDigest(intervention.SemanticBeforeDigest) || !validDigest(intervention.SemanticAfterDigest) ||
			!intervention.SourceChanged || intervention.SourceBeforeDigest != input.Source.SourceDigest ||
			intervention.SourceAfterDigest == intervention.SourceBeforeDigest || intervention.SemanticBeforeDigest != input.Source.SemanticDigest {
			return fmt.Errorf("independent judge: intervention %d is not source-bound", index+1)
		}
		switch intervention.ID {
		case "semantic-trace":
			if intervention.SemanticAfterDigest == intervention.SemanticBeforeDigest || !intervention.SemanticChanged || intervention.Reason != "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST" {
				return fmt.Errorf("independent judge: semantic intervention is not semantic")
			}
		case "nonsemantic-comment":
			if intervention.SemanticAfterDigest != intervention.SemanticBeforeDigest || intervention.SemanticChanged || intervention.Reason != "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST" {
				return fmt.Errorf("independent judge: comment intervention changed semantic meaning")
			}
		default:
			return fmt.Errorf("independent judge: intervention %d is unknown", index+1)
		}
	}
	return nil
}

func classify(input wireInput) classification {
	states := []string{input.Trace[0].BeforeState}
	for _, observation := range input.Trace {
		if observation.AfterState != states[len(states)-1] {
			states = append(states, observation.AfterState)
		}
	}
	cycleStart, period := repeatedState(states)
	repeatedStates := 0
	if cycleStart >= 0 {
		repeatedStates = 1
	}
	final := input.Trace[len(input.Trace)-1]
	if input.UpstreamDecision != upstreamNoChange && input.UpstreamDecision != upstreamChanged {
		return classification{decisionFailClosed, resolutionLower, receiptFailClosed, reasonDecisionUnknown, final.AfterState, claimOpen, period, len(states), repeatedStates, false}
	}
	if cycleStart >= 0 {
		return classification{decisionCycle, resolutionExact, receiptBound, reasonCycle, final.AfterState, claimRefuted, period, len(states), repeatedStates, false}
	}
	if final.Decision == upstreamNoChange {
		return classification{decisionFixedPoint, resolutionExact, receiptBound, reasonNoChange, final.AfterState, claimDischarged, 0, len(states), repeatedStates, true}
	}
	diverging := len(input.Trace) == input.MaxSteps
	for _, observation := range input.Trace {
		diverging = diverging && observation.Decision == upstreamChanged && observation.AfterRank > observation.BeforeRank
	}
	if diverging {
		return classification{decisionDivergence, resolutionLower, receiptBound, reasonDivergence, final.AfterState, claimOpen, 0, len(states), repeatedStates, false}
	}
	return classification{decisionInProgress, resolutionLower, receiptBound, reasonInProgress, final.AfterState, claimOpen, 0, len(states), repeatedStates, false}
}

func repeatedState(states []string) (int, int) {
	for left := 0; left < len(states); left++ {
		for right := left + 1; right < len(states); right++ {
			if states[left] == states[right] {
				return left, right - left
			}
		}
	}
	return -1, 0
}

func receiptForValidation(input wireInput, class classification) wireReceipt {
	receipt := wireReceipt{
		Schema: receiptSchema, Metaprogram: metaprogram, Repository: input.Repository,
		Subject: input.Subject, Producer: input.Producer, Consumer: input.Consumer,
		MetaOperation: input.MetaOperation, ProofChoice: input.ProofChoice, Stage: input.Stage,
		Status: class.status, Resolution: class.resolution, Decision: class.decision,
		Reason: class.reason, Source: input.Source, UpstreamDecision: input.UpstreamDecision,
		InputDigest: digestJSON(input), TraceDigest: digestJSON(input.Trace),
		Observations: append([]wireObservation(nil), input.Trace...), Interventions: append([]wireIntervention(nil), input.Interventions...),
		ClaimTransitions: transitions(class, len(input.Trace)), Outcome: wireOutcome{
			ObservedSteps: len(input.Trace), MaxSteps: input.MaxSteps, StateCount: class.stateCount,
			RepeatedStates: class.repeatedStates, DetectedPeriod: class.period, FinalState: class.finalState,
			TerminationProven: class.terminationProven, ClaimState: class.claimState,
		}, Conformance: wireConformance{Satisfied: indicatorTotal, Total: indicatorTotal, BasisPoints: 10000},
		Authority: wireAuthority{ReadOnly: true}, Indicators: indicators(input.Interventions),
	}
	return receipt
}

func transitions(class classification, finalStep int) []wireClaimTransition {
	finalTo := claimOpen
	if class.claimState == claimDischarged {
		finalTo = claimDischarged
	} else if class.claimState == claimRefuted {
		finalTo = claimRefuted
	}
	return []wireClaimTransition{
		{Stage: claimStage, Step: 0, From: claimOpen, To: claimOpen, Reason: "TRACE_BOUND"},
		{Stage: claimStage, Step: finalStep, From: claimOpen, To: finalTo, Reason: class.reason},
	}
}

func indicators(interventions []wireIntervention) []wireIndicator {
	return []wireIndicator{
		indicator(interventions[0], "gooo.termination.semantic-trace-intervention.v1", "SEMANTIC_TRACE_INTERVENTION", "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST"),
		indicator(interventions[1], "gooo.termination.nonsemantic-comment-intervention.v1", "NONSEMANTIC_COMMENT_INTERVENTION", "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST"),
	}
}

func indicator(intervention wireIntervention, id, route, reason string) wireIndicator {
	return wireIndicator{
		ID: id, Route: route, Producer: producer, Consumer: consumer,
		MetaOperation: metaOperation, ProofChoice: proofChoice, Stage: interventionStage,
		Step: intervention.Step, Reason: reason,
		Value: fmt.Sprintf("source_changed=%t;semantic_changed=%t", intervention.SourceChanged, intervention.SemanticChanged),
		Limit: "source_changed=true;semantic_changed=true|false", Satisfied: true,
	}
}

func decodeReceipt(raw []byte) (wireReceipt, error) {
	var receipt wireReceipt
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return wireReceipt{}, fmt.Errorf("independent judge: receipt JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return wireReceipt{}, fmt.Errorf("independent judge: receipt has trailing JSON")
		}
		return wireReceipt{}, fmt.Errorf("independent judge: receipt trailing data: %w", err)
	}
	return receipt, nil
}

func seal(receipt wireReceipt) wireReceipt {
	receipt.ReceiptDigest = ""
	receipt.ReplayDigest = ""
	receipt.ReceiptDigest = digestJSON(receipt)
	receipt.ReplayDigest = digestJSON(struct {
		InputDigest, TraceDigest, ReceiptDigest string
	}{receipt.InputDigest, receipt.TraceDigest, receipt.ReceiptDigest})
	return receipt
}

func digestJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(payload)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func stateDigest(label string) string {
	return digestBytes([]byte("self-improvement-state/v1:" + label))
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
