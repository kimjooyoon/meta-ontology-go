package nonmonotonicrefutationoracle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	oracleSchema     = "gooo/meta-nonmonotonic-refutation-oracle/v2"
	producerSchema   = "gooo/meta-nonmonotonic-refutation-producer/v2"
	sourceSchema     = "gooo/meta-nonmonotonic-refutation-source/v1"
	producerID       = "producer://nonmonotonic-refutation"
	consumerID       = "consumer://nonmonotonic-refutation-oracle"
	metaOperation    = "meta://revise-claim-by-evidence"
	statusOpen       = "OPEN"
	statusDischarged = "DISCHARGED"
	statusRefuted    = "REFUTED"
	reopenPolicy     = "REOPEN_IF_NEWER_ADMISSIBLE"
)

// Judge independently parses and lowers source, reconstructs its own model,
// then compares the producer's model with that reconstruction before replay.
func Judge(producerBytes, source []byte) (Report, error) {
	input, err := decode(producerBytes)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		Schema: oracleSchema, ProducerSchema: input.Schema,
		ProducerReceiptDigest: digestBytes(producerBytes), SourcePath: input.SourcePath,
		SourceDigest: input.SourceDigest, SourceSemanticDigest: input.SourceSemanticDigest,
		SourceModelDigest: input.SourceModelDigest, Producer: input.Producer, Consumer: input.Consumer,
		MetaOperation: input.MetaOperation, ProofChoice: input.ProofChoice,
		Effects:               effects{RepositoryWrites: input.Effects.RepositoryWrites, MutationAuthority: input.Effects.MutationAuthority},
		MetaValue:             "current knowledge is revisable while historical states remain inspectable",
		FalsifiablePrediction: "changing an observed value or revision policy changes the adjudication, while comment-only edits preserve semantic digest and verdict",
	}
	model, reason := reconstructSource(source)
	if reason == "" {
		reason = validateInput(input, source, model)
	}
	if reason != "" {
		return finishFailure(report, reason), nil
	}
	cases, transitions, metrics, reason := replay(model)
	if reason != "" {
		report.Cases, report.Transitions, report.Metrics = cases, transitions, metrics
		return finishFailure(report, reason), nil
	}
	report.Cases, report.Transitions, report.Metrics = cases, transitions, metrics
	report.SubjectResolution = subjectResolution(cases)
	report.Conformance = Conformance{Decision: "PASS", Resolution: "EXACT", Reason: "SOURCE_RECONSTRUCTION_AND_REPLAY"}
	return finish(report, "PASS", "EXACT", "NONMONOTONIC_REFUTATION_OBSERVED"), nil
}

func decode(data []byte) (producerInput, error) {
	var input producerInput
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, fmt.Errorf("decode producer receipt: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return input, fmt.Errorf("producer receipt has trailing JSON")
	}
	return input, nil
}

func reconstructSource(source []byte) (sourceModel, string) {
	file, diagnostics := syntax.ParseFile("source.gooo", string(source))
	if diagnostics.HasErrors() {
		return sourceModel{}, "GOOO_PARSE_REJECTED"
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, "GOOO_LOWER_REJECTED"
	}
	model := sourceModel{SemanticDigest: "sha256:" + semantic.StableHash([]byte(ir.SemanticCanonical()))}
	entities := make(map[string]string)
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			continue
		}
		entities[entity.Name] = entity.ID
		if strings.HasPrefix(entity.ID, "gooo://nonmonotonic-refutation/claim/") {
			model.Contract.Claims = append(model.Contract.Claims, sourceClaim{ID: entity.ID})
		}
	}
	if len(model.Contract.Claims) != 3 {
		return sourceModel{}, "SOURCE_CLAIM_DENOMINATOR_MISMATCH"
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "meta.observe:v1;") {
			continue
		}
		if len(activity.Inputs) < 3 || activity.Output == "" {
			return sourceModel{}, "SOURCE_OBSERVATION_ENDPOINT_MISMATCH"
		}
		observationID, ok := entities[activity.Output]
		if !ok || !strings.HasPrefix(observationID, "gooo://nonmonotonic-refutation/observation/") {
			return sourceModel{}, "SOURCE_OBSERVATION_ID_MISMATCH"
		}
		fields, ok := parseObservationProgram(activity.ValueProgram)
		if !ok {
			return sourceModel{}, "SOURCE_OBSERVATION_FIELDS_INVALID"
		}
		claimID := "gooo://nonmonotonic-refutation/claim/" + fields["claim"]
		if entities[activity.Inputs[0].Name] != claimID || entities[activity.Inputs[1].Name] != "gooo://nonmonotonic-refutation/predicate/"+fields["predicate"] || entities[activity.Inputs[2].Name] != "gooo://nonmonotonic-refutation/value/one" || fields["expected"] != "1" {
			return sourceModel{}, "SOURCE_SUBJECT_RESOLUTION_MISMATCH"
		}
		sequence := len(model.Contract.Observations) + 1
		model.Contract.Observations = append(model.Contract.Observations, sourceObservation{
			ID: observationID, Activity: activity.Name, ClaimID: claimID, Sequence: sequence,
			Predicate: fields["predicate"], ExpectedValue: fields["expected"], ObservedValue: fields["observed"],
			Provenance: fields["provenance"], EvidenceDigest: fields["evidence_digest"], PriorState: fields["prior"],
			RevisionPolicy: fields["revision_policy"], Producer: fields["producer"], Consumer: fields["consumer"],
			MetaOperation: fields["meta_operation"], ProofChoice: fields["proof_choice"],
			Coordinate: coordinate{Stage: fields["stage"], Step: fields["step"]},
		})
	}
	if len(model.Contract.Observations) != 6 {
		return sourceModel{}, "SOURCE_OBSERVATION_DENOMINATOR_MISMATCH"
	}
	model.Contract.Schema = sourceSchema
	model.Contract.FixedClaimTotal = len(model.Contract.Claims)
	model.Contract.FixedObservationTotal = len(model.Contract.Observations)
	model.Contract.FixedTransitionTotal = len(model.Contract.Observations)
	if err := completeClaims(&model.Contract); err != nil {
		return sourceModel{}, "SOURCE_CLAIM_BINDING_MISMATCH"
	}
	return model, ""
}

func parseObservationProgram(program string) (map[string]string, bool) {
	fields := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		if part == "meta.observe:v1" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || value == "" {
			return nil, false
		}
		if !knownObservationField(key) {
			return nil, false
		}
		if _, exists := fields[key]; exists {
			return nil, false
		}
		fields[key] = value
	}
	for _, key := range []string{"claim", "predicate", "expected", "observed", "provenance", "evidence_digest", "prior", "revision_policy", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if fields[key] == "" {
			return nil, false
		}
	}
	if len(fields["evidence_digest"]) != len("sha256:")+64 || !strings.HasPrefix(fields["evidence_digest"], "sha256:") {
		return nil, false
	}
	return fields, true
}

func knownObservationField(key string) bool {
	for _, known := range []string{"claim", "predicate", "expected", "observed", "provenance", "evidence_digest", "prior", "revision_policy", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if key == known {
			return true
		}
	}
	return false
}

func completeClaims(contract *sourceContract) error {
	for index := range contract.Claims {
		claim := &contract.Claims[index]
		for _, observation := range contract.Observations {
			if observation.ClaimID != claim.ID {
				continue
			}
			if claim.InitialStatus == "" {
				claim.InitialStatus = observation.PriorState
				claim.Predicate = observation.Predicate
				claim.ExpectedValue = observation.ExpectedValue
				claim.RevisionPolicy = observation.RevisionPolicy
			}
			if claim.Predicate != observation.Predicate || claim.ExpectedValue != observation.ExpectedValue || claim.RevisionPolicy != observation.RevisionPolicy {
				return fmt.Errorf("claim source changed its predicate, expected value, or revision policy")
			}
		}
		if claim.InitialStatus == "" {
			return fmt.Errorf("claim has no observation")
		}
	}
	return nil
}

func validateInput(input producerInput, source []byte, model sourceModel) string {
	if input.Schema != producerSchema || input.ReceiptDigest == "" {
		return "PRODUCER_SCHEMA_OR_RECEIPT_MISSING"
	}
	claimed := input.ReceiptDigest
	input.ReceiptDigest = ""
	if claimed != digestJSON(input) {
		return "PRODUCER_RECEIPT_DIGEST_MISMATCH"
	}
	if input.SourceDigest == "" || input.SourceDigest != digestBytes(source) || !strings.HasSuffix(input.SourcePath, ".gooo") {
		return "SOURCE_BINDING_MISMATCH"
	}
	if input.SourceSemanticDigest != model.SemanticDigest || input.SourceModelDigest != digestJSON(model.Contract) || input.Contract.Schema != sourceSchema || input.Contract.FixedClaimTotal != 3 || input.Contract.FixedObservationTotal != 6 || input.Contract.FixedTransitionTotal != 6 {
		return "SOURCE_RECONSTRUCTION_MISMATCH"
	}
	if !sameSourceContract(input.Contract, model.Contract) {
		return "PRODUCER_SOURCE_MODEL_MISMATCH"
	}
	if input.Producer != producerID || input.Consumer != consumerID || input.MetaOperation != metaOperation || input.Effects.RepositoryWrites != 0 || input.Effects.MutationAuthority {
		return "PRODUCER_PROVENANCE_OR_EFFECTS_MISMATCH"
	}
	return ""
}

func sameSourceContract(left, right sourceContract) bool {
	return digestJSON(left) == digestJSON(right)
}

func replay(model sourceModel) ([]CaseResult, []Transition, Metrics, string) {
	metrics := Metrics{FixedClaimTotal: model.Contract.FixedClaimTotal, FixedObservationTotal: model.Contract.FixedObservationTotal, FixedTransitionTotal: model.Contract.FixedTransitionTotal, InScopeClaimTotal: len(model.Contract.Claims)}
	cases := make([]CaseResult, len(model.Contract.Claims))
	status := make(map[string]string, len(model.Contract.Claims))
	caseIndex := make(map[string]int, len(model.Contract.Claims))
	for index, claim := range model.Contract.Claims {
		cases[index] = CaseResult{ID: strings.TrimPrefix(claim.ID, "gooo://nonmonotonic-refutation/claim/"), ClaimID: claim.ID, InitialStatus: claim.InitialStatus, StatusHistory: []string{claim.InitialStatus}, RevisionPolicy: claim.RevisionPolicy, Producer: producerID, Consumer: consumerID, MetaOperation: metaOperation}
		status[claim.ID] = claim.InitialStatus
		caseIndex[claim.ID] = index
	}
	transitions := make([]Transition, 0, len(model.Contract.Observations))
	previousDigest := ""
	for index, observation := range model.Contract.Observations {
		caseNumber, ok := caseIndex[observation.ClaimID]
		if !ok || observation.Sequence != index+1 {
			return cases, transitions, metrics, "SOURCE_OBSERVATION_ORDER_MISMATCH"
		}
		before := status[observation.ClaimID]
		if observation.PriorState != before {
			return cases, transitions, metrics, "PRIOR_STATE_DOES_NOT_MATCH_LEDGER"
		}
		kind := classify(observation)
		if kind == "" {
			return cases, transitions, metrics, "INDEPENDENT_CLASSIFIER_REJECTED_PREDICATE"
		}
		after, accepted, reason := revise(before, kind, observation.RevisionPolicy)
		coordinate := coordinate{Stage: observation.Coordinate.Stage, Step: observation.Coordinate.Step, Reason: reason}
		transition := Transition{Sequence: index + 1, CaseID: cases[caseNumber].ID, ClaimID: observation.ClaimID, Before: before, After: after, Accepted: accepted, EvidenceID: observation.ID, EvidenceKind: kind, EvidenceBasis: evidenceBasis(observation), EvidenceDigest: observation.EvidenceDigest, EvidenceProvenance: observation.Provenance, ProofChoice: observation.ProofChoice, Coordinate: coordinate, PreviousDigest: previousDigest}
		transition.TransitionDigest = transitionDigest(transition)
		previousDigest = transition.TransitionDigest
		transitions = append(transitions, transition)
		metrics.TransitionTotal++
		if !accepted {
			finalizePartialCases(cases, status, &metrics)
			return cases, transitions, metrics, reason
		}
		status[observation.ClaimID] = after
		cases[caseNumber].StatusHistory = append(cases[caseNumber].StatusHistory, after)
		switch before + "->" + after {
		case statusOpen + "->" + statusDischarged:
			metrics.OpenToDischargedTotal++
		case statusDischarged + "->" + statusRefuted:
			metrics.DischargedToRefutedTotal++
			metrics.NonMonotonicRevisionTotal++
			cases[caseNumber].RefutationObserved = true
		case statusRefuted + "->" + statusDischarged:
			metrics.RefutedToDischargedTotal++
		}
	}
	for index := range cases {
		cases[index].CurrentStatus = status[cases[index].ClaimID]
		cases[index].HistoryRetained = len(cases[index].StatusHistory) >= 1
		if cases[index].CurrentStatus == statusDischarged {
			metrics.CurrentDischargedTotal++
		}
		if cases[index].CurrentStatus == statusRefuted {
			metrics.CurrentRefutedTotal++
		}
		metrics.RetainedStateTotal += len(cases[index].StatusHistory)
	}
	if metrics.FixedClaimTotal != 3 || metrics.FixedObservationTotal != 6 || metrics.FixedTransitionTotal != 6 || metrics.TransitionTotal != 6 || metrics.OpenToDischargedTotal != 3 || metrics.DischargedToRefutedTotal != 2 || metrics.RefutedToDischargedTotal != 1 || metrics.CurrentDischargedTotal != 2 || metrics.CurrentRefutedTotal != 1 || metrics.RetainedStateTotal != 9 {
		return cases, transitions, metrics, "FIXED_TRANSITION_COUNT_MISMATCH"
	}
	metrics.CurrentDischargeBasisPoints = metrics.CurrentDischargedTotal * 10000 / metrics.FixedClaimTotal
	return cases, transitions, metrics, ""
}

func finalizePartialCases(cases []CaseResult, status map[string]string, metrics *Metrics) {
	for index := range cases {
		cases[index].CurrentStatus = status[cases[index].ClaimID]
		cases[index].HistoryRetained = len(cases[index].StatusHistory) >= 1
		metrics.RetainedStateTotal += len(cases[index].StatusHistory)
	}
}

func classify(observation sourceObservation) string {
	if observation.Predicate != "equality" {
		return ""
	}
	if observation.ObservedValue == observation.ExpectedValue {
		return "SUPPORT"
	}
	return "CONTRADICTING"
}

func revise(before, kind, policy string) (string, bool, string) {
	switch {
	case before == statusOpen && kind == "SUPPORT":
		return statusDischarged, true, "EQUALITY_MATCH_ACCEPTED"
	case before == statusDischarged && kind == "CONTRADICTING":
		return statusRefuted, true, "NEW_OBSERVATION_CONTRADICTS_DISCHARGED"
	case before == statusRefuted && kind == "SUPPORT" && policy == reopenPolicy:
		return statusDischarged, true, "REVISION_POLICY_REOPEN_ACCEPTS_MATCH"
	case before == statusRefuted && kind == "SUPPORT":
		return statusRefuted, false, "REVISION_POLICY_UNKNOWN_OR_FORBIDS_REOPEN"
	default:
		return before, false, "TRANSITION_NOT_ADMISSIBLE"
	}
}

func evidenceBasis(observation sourceObservation) string {
	return fmt.Sprintf("predicate=%s expected=%s observed=%s provenance=%s digest=%s", observation.Predicate, observation.ExpectedValue, observation.ObservedValue, observation.Provenance, observation.EvidenceDigest)
}

func subjectResolution(cases []CaseResult) SubjectResolution {
	for _, result := range cases {
		if result.CurrentStatus == statusRefuted {
			return SubjectResolution{Decision: "OBSERVED", Resolution: "PARTIAL", Reason: "CURRENT_KNOWLEDGE_INCLUDES_REFUTED_SUBJECT"}
		}
	}
	return SubjectResolution{Decision: "OBSERVED", Resolution: "EXACT", Reason: "ALL_SUBJECTS_DISCHARGED"}
}

func finishFailure(report Report, reason string) Report {
	report.Conformance = Conformance{Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Reason: reason}
	report.SubjectResolution = SubjectResolution{Decision: "UNRESOLVED", Resolution: "FAIL_CLOSED", Reason: reason}
	return finish(report, "FAIL_CLOSED", "LOWER_RESOLUTION", reason)
}

func finish(report Report, decision, resolution, reason string) Report {
	report.Decision, report.Resolution, report.Reason = decision, resolution, reason
	report.ReportDigest = reportDigest(report)
	return report
}
