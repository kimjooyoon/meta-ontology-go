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
	oracleSchema     = "gooo/meta-nonmonotonic-refutation-oracle/v3"
	producerSchema   = "gooo/meta-nonmonotonic-refutation-producer/v3"
	sourceSchema     = "gooo/meta-nonmonotonic-refutation-source/v2"
	producerID       = "producer://nonmonotonic-refutation"
	consumerID       = "consumer://nonmonotonic-refutation-oracle"
	metaOperation    = "meta://revise-claim-by-evidence"
	statusOpen       = "OPEN"
	statusDischarged = "DISCHARGED"
	statusRefuted    = "REFUTED"
)

// Judge is an independent adjudicator. It parses and lowers raw source
// itself, reconstructs its own source model, and only then replays evidence.
func Judge(producerBytes, source []byte) (Report, error) {
	input, err := decode(producerBytes)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		Schema: oracleSchema, ProducerSchema: input.Schema,
		ProducerReceiptDigest: digestBytes(producerBytes), SourcePath: input.SourcePath,
		SourceDigest: input.SourceDigest, SourceSemanticDigest: input.SourceSemanticDigest,
		SourceBindingDigest: input.SourceBindingDigest, SourceModelDigest: input.SourceModelDigest,
		Producer: input.Producer, Consumer: input.Consumer, MetaOperation: input.MetaOperation,
		ProofChoice:           input.ProofChoice,
		Effects:               effects{RepositoryWrites: input.Effects.RepositoryWrites, MutationAuthority: input.Effects.MutationAuthority, PromotionCount: input.Effects.PromotionCount},
		MetaValue:             "evidence revises current knowledge while every prior ledger row remains inspectable",
		FalsifiablePrediction: "changing a proposition or observed value changes the relation, subject resolution, and ledger transition; comment-only edits preserve semantic meaning",
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
	report.Conformance = Conformance{Decision: "PASS", Resolution: "EXACT", Reason: "SOURCE_RECONSTRUCTION_AND_REPLAY"}
	report.SubjectResolution = subjectResolution(cases)
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
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "meta.observe:v2;") {
			continue
		}
		if len(activity.Inputs) < 5 || activity.Output == "" {
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
		if entities[activity.Inputs[0].Name] != claimID ||
			entities[activity.Inputs[1].Name] != "gooo://nonmonotonic-refutation/predicate/"+fields["predicate"] ||
			entities[activity.Inputs[2].Name] != "gooo://nonmonotonic-refutation/subject/"+fields["subject"] ||
			entities[activity.Inputs[3].Name] != "gooo://nonmonotonic-refutation/input/"+fields["input"] ||
			entities[activity.Inputs[4].Name] != "gooo://nonmonotonic-refutation/value/one" || fields["expected"] != "1" {
			return sourceModel{}, "SOURCE_SUBJECT_INPUT_RESOLUTION_MISMATCH"
		}
		model.Contract.Observations = append(model.Contract.Observations, sourceObservation{
			ID: observationID, Activity: activity.Name, ClaimID: claimID, Sequence: len(model.Contract.Observations) + 1,
			Proposition: fields["proposition"], Subject: fields["subject"], Input: fields["input"],
			Predicate: fields["predicate"], ExpectedValue: fields["expected"], ObservedValue: fields["observed"],
			Provenance: fields["provenance"], EvidenceDigest: fields["evidence_digest"], Producer: fields["producer"],
			Consumer: fields["consumer"], MetaOperation: fields["meta_operation"], ProofChoice: fields["proof_choice"],
			Coordinate: coordinate{Stage: fields["stage"], Step: fields["step"]},
		})
	}
	if len(model.Contract.Observations) != 6 {
		return sourceModel{}, "SOURCE_OBSERVATION_DENOMINATOR_MISMATCH"
	}
	model.Contract.Schema = sourceSchema
	model.Contract.FixedCaseTotal = len(model.Contract.Claims)
	model.Contract.FixedClaimTotal = len(model.Contract.Claims)
	model.Contract.FixedObservationTotal = len(model.Contract.Observations)
	model.Contract.FixedLedgerRowTotal = len(model.Contract.Observations)
	if err := completeClaims(&model.Contract); err != nil {
		return sourceModel{}, "SOURCE_CLAIM_BINDING_MISMATCH"
	}
	return model, ""
}

func parseObservationProgram(program string) (map[string]string, bool) {
	fields := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		if part == "meta.observe:v2" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || (value == "" && key != "observed") || !knownObservationField(key) {
			return nil, false
		}
		if _, exists := fields[key]; exists {
			return nil, false
		}
		fields[key] = value
	}
	for _, key := range []string{"claim", "proposition", "subject", "input", "predicate", "expected", "provenance", "evidence_digest", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if fields[key] == "" {
			return nil, false
		}
	}
	if _, ok := fields["observed"]; !ok || len(fields["evidence_digest"]) != len("sha256:")+64 || !strings.HasPrefix(fields["evidence_digest"], "sha256:") {
		return nil, false
	}
	return fields, true
}

func knownObservationField(key string) bool {
	for _, known := range []string{"claim", "proposition", "subject", "input", "predicate", "expected", "observed", "provenance", "evidence_digest", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
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
			if claim.Proposition == "" {
				claim.Proposition = observation.Proposition
				claim.Subject = observation.Subject
				claim.Input = observation.Input
				claim.Predicate = observation.Predicate
				claim.ExpectedValue = observation.ExpectedValue
			}
			if claim.Proposition != observation.Proposition || claim.Subject != observation.Subject || claim.Input != observation.Input || claim.Predicate != observation.Predicate || claim.ExpectedValue != observation.ExpectedValue {
				return fmt.Errorf("claim source changed its proposition or subject/input")
			}
		}
		if claim.Proposition == "" {
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
		return "SOURCE_RAW_DIGEST_MISMATCH"
	}
	if input.SourceBindingDigest != digestJSON(sourceBinding{RawDigest: input.SourceDigest, SemanticDigest: input.SourceSemanticDigest}) {
		return "SOURCE_BINDING_DIGEST_MISMATCH"
	}
	if input.SourceSemanticDigest != model.SemanticDigest || input.Contract.Schema != sourceSchema || input.Contract.FixedCaseTotal != 3 || input.Contract.FixedClaimTotal != 3 || input.Contract.FixedObservationTotal != 6 || input.Contract.FixedLedgerRowTotal != 6 {
		return "SOURCE_RECONSTRUCTION_MISMATCH"
	}
	if input.SourceModelDigest != digestJSON(model.Contract) || digestJSON(input.Contract) != digestJSON(model.Contract) {
		return "PRODUCER_SOURCE_MODEL_MISMATCH"
	}
	if input.Producer != producerID || input.Consumer != consumerID || input.MetaOperation != metaOperation || input.Effects.RepositoryWrites != 0 || input.Effects.MutationAuthority || input.Effects.PromotionCount != 0 {
		return "PRODUCER_PROVENANCE_OR_EFFECTS_MISMATCH"
	}
	return ""
}

func replay(model sourceModel) ([]CaseResult, []Transition, Metrics, string) {
	metrics := Metrics{FixedCaseTotal: model.Contract.FixedCaseTotal, FixedClaimTotal: model.Contract.FixedClaimTotal, FixedObservationTotal: model.Contract.FixedObservationTotal, FixedLedgerRowTotal: model.Contract.FixedLedgerRowTotal, InScopeClaimTotal: len(model.Contract.Claims)}
	cases := make([]CaseResult, len(model.Contract.Claims))
	status := make(map[string]string, len(model.Contract.Claims))
	caseIndex := make(map[string]int, len(model.Contract.Claims))
	for index, claim := range model.Contract.Claims {
		caseID := strings.TrimPrefix(claim.ID, "gooo://nonmonotonic-refutation/claim/")
		cases[index] = CaseResult{ID: caseID, ClaimID: claim.ID, Proposition: claim.Proposition, Subject: claim.Subject, Input: claim.Input, InitialStatus: statusOpen, StatusHistory: []string{statusOpen}}
		status[claim.ID] = statusOpen
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
		relation := classify(model.Contract.Claims[caseNumber], observation)
		after, accepted, reason := revise(before, relation)
		transition := Transition{Sequence: index + 1, CaseID: cases[caseNumber].ID, ClaimID: observation.ClaimID, Before: before, After: after, Accepted: accepted, EvidenceID: observation.ID, Relation: relation, EvidenceBasis: evidenceBasis(observation), EvidenceDigest: observation.EvidenceDigest, EvidenceProvenance: observation.Provenance, ProofChoice: observation.ProofChoice, Coordinate: coordinate{Stage: observation.Coordinate.Stage, Step: observation.Coordinate.Step, Reason: reason}, PreviousDigest: previousDigest}
		transition.TransitionDigest = transitionDigest(transition)
		previousDigest = transition.TransitionDigest
		transitions = append(transitions, transition)
		metrics.TransitionTotal++
		switch relation {
		case "SUPPORTS":
			metrics.SupportsTotal++
		case "CONTRADICTS":
			metrics.ContradictsTotal++
		case "INSUFFICIENT":
			metrics.InsufficientTotal++
		case "UNKNOWN":
			metrics.UnknownTotal++
		}
		if before != after {
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
		status[observation.ClaimID] = after
		cases[caseNumber].StatusHistory = append(cases[caseNumber].StatusHistory, after)
		cases[caseNumber].ObservationTotal++
	}
	for index := range cases {
		cases[index].CurrentStatus = status[cases[index].ClaimID]
		cases[index].HistoryRetained = len(cases[index].StatusHistory) == cases[index].ObservationTotal+1
		metrics.RetainedStateTotal += len(cases[index].StatusHistory)
		switch cases[index].CurrentStatus {
		case statusDischarged:
			metrics.CurrentDischargedTotal++
		case statusRefuted:
			metrics.CurrentRefutedTotal++
		case statusOpen:
			metrics.CurrentOpenTotal++
		}
	}
	if metrics.FixedCaseTotal != 3 || metrics.FixedClaimTotal != 3 || metrics.FixedObservationTotal != 6 || metrics.FixedLedgerRowTotal != 6 || metrics.TransitionTotal != 6 {
		return cases, transitions, metrics, "FIXED_SOURCE_COUNT_MISMATCH"
	}
	metrics.CurrentDischargeBasisPoints = metrics.CurrentDischargedTotal * 10000 / metrics.FixedClaimTotal
	return cases, transitions, metrics, ""
}

func classify(claim sourceClaim, observation sourceObservation) string {
	if claim.ID != observation.ClaimID || claim.Proposition != observation.Proposition || claim.Subject != observation.Subject || claim.Input != observation.Input || claim.Predicate != observation.Predicate || claim.ExpectedValue != observation.ExpectedValue {
		return "UNKNOWN"
	}
	if observation.Predicate != "equality" || observation.Subject == "" || observation.Input == "" {
		return "UNKNOWN"
	}
	propositionPrefix := "equals:" + observation.Subject + ":" + observation.Input + ":"
	if !strings.HasPrefix(observation.Proposition, propositionPrefix) {
		return "UNKNOWN"
	}
	propositionValue := strings.TrimPrefix(observation.Proposition, propositionPrefix)
	if propositionValue == "" || strings.Contains(propositionValue, ":") {
		return "UNKNOWN"
	}
	if observation.ObservedValue == "" {
		return "INSUFFICIENT"
	}
	if observation.ObservedValue == propositionValue {
		return "SUPPORTS"
	}
	return "CONTRADICTS"
}

func revise(before, relation string) (string, bool, string) {
	switch relation {
	case "SUPPORTS":
		return statusDischarged, true, "PROPOSITION_MATCHES_OBSERVATION"
	case "CONTRADICTS":
		return statusRefuted, true, "OBSERVATION_DIRECTLY_CONTRADICTS_PROPOSITION"
	case "INSUFFICIENT":
		return statusOpen, false, "INSUFFICIENT_EVIDENCE_LEAVES_CLAIM_OPEN"
	default:
		return statusOpen, false, "UNKNOWN_RELATION_LEAVES_CLAIM_OPEN"
	}
}

func evidenceBasis(observation sourceObservation) string {
	return fmt.Sprintf("proposition=%s subject=%s input=%s predicate=%s expected=%s observed=%s provenance=%s digest=%s", observation.Proposition, observation.Subject, observation.Input, observation.Predicate, observation.ExpectedValue, observation.ObservedValue, observation.Provenance, observation.EvidenceDigest)
}

func subjectResolution(cases []CaseResult) SubjectResolution {
	discharged, refuted, open := 0, 0, 0
	for _, result := range cases {
		switch result.CurrentStatus {
		case statusDischarged:
			discharged++
		case statusRefuted:
			refuted++
		default:
			open++
		}
	}
	resolution := "EXACT"
	if open > 0 {
		resolution = "LOWER_RESOLUTION"
	} else if refuted > 0 {
		resolution = "PARTIAL"
	}
	return SubjectResolution{Decision: fmt.Sprintf("DISCHARGED=%d;REFUTED=%d;OPEN=%d", discharged, refuted, open), Resolution: resolution, Reason: "CURRENT_LEDGER_DISTRIBUTION"}
}

func finishFailure(report Report, reason string) Report {
	report.Conformance = Conformance{Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Reason: reason}
	report.SubjectResolution = SubjectResolution{Decision: "UNRESOLVED", Resolution: "LOWER_RESOLUTION", Reason: reason}
	return finish(report, "FAIL_CLOSED", "LOWER_RESOLUTION", reason)
}

func finish(report Report, decision, resolution, reason string) Report {
	report.Decision, report.Resolution, report.Reason = decision, resolution, reason
	report.ReportDigest = reportDigest(report)
	return report
}
