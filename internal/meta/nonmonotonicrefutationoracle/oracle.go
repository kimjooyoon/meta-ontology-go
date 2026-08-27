package nonmonotonicrefutationoracle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

const (
	oracleSchema   = "gooo/meta-nonmonotonic-refutation-oracle/v1"
	producerSchema = "gooo/meta-nonmonotonic-refutation-producer/v1"
	contractSchema = "gooo/meta-nonmonotonic-refutation-contract/v1"
	producerID     = "producer://nonmonotonic-refutation"
	consumerID     = "consumer://nonmonotonic-refutation-oracle"
	metaOperation  = "meta://revise-claim-by-evidence"
)

func Judge(producerBytes, source []byte) (Report, error) {
	input, err := decode(producerBytes)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		Schema: oracleSchema, ProducerSchema: input.Schema,
		ProducerReceiptDigest: digestBytes(producerBytes), SourcePath: input.SourcePath,
		SourceDigest: input.SourceDigest, Producer: input.Producer, Consumer: input.Consumer,
		MetaOperation: input.MetaOperation, ProofChoice: input.ProofChoice,
		Effects:               effects{RepositoryWrites: 0, MutationAuthority: false},
		MetaValue:             "current knowledge is revisable while historical states remain inspectable",
		FalsifiablePrediction: "reordering, removing, or relabeling one evidence item must fail closed",
	}
	if reason := validateInput(input, source); reason != "" {
		return finish(report, "FAIL_CLOSED", "INVARIANT", reason), nil
	}
	cases, transitions, metrics, reason := replay(input)
	if reason != "" {
		return finish(report, "FAIL_CLOSED", "INVARIANT", reason), nil
	}
	report.Cases, report.Transitions, report.Metrics = cases, transitions, metrics
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

func validateInput(input producerInput, source []byte) string {
	if input.Schema != producerSchema {
		return "PRODUCER_SCHEMA_MISMATCH"
	}
	if input.ReceiptDigest == "" {
		return "PRODUCER_RECEIPT_DIGEST_MISSING"
	}
	claimed := input.ReceiptDigest
	input.ReceiptDigest = ""
	if claimed != digestJSON(input) {
		return "PRODUCER_RECEIPT_DIGEST_MISMATCH"
	}
	if !reflect.DeepEqual(input.Contract, canonicalContract()) || input.Producer != producerID ||
		input.Consumer != consumerID || input.MetaOperation != metaOperation ||
		input.Effects.RepositoryWrites != 0 || input.Effects.MutationAuthority {
		return "PRODUCER_CONTRACT_OR_EFFECTS_MISMATCH"
	}
	if input.SourceDigest == "" || input.SourceDigest != digestBytes(source) ||
		!strings.HasSuffix(input.SourcePath, ".gooo") {
		return "SOURCE_BINDING_MISMATCH"
	}
	text := string(source)
	if !strings.Contains(text, "package nonmonotonicrefutation") || !strings.Contains(text, "activity") {
		return "GOOO_SOURCE_NOT_OBSERVED"
	}
	return ""
}

func replay(input producerInput) ([]CaseResult, []Transition, Metrics, string) {
	metrics := Metrics{FixedClaimTotal: 3, InScopeClaimTotal: 3}
	cases := make([]CaseResult, 0, len(input.Contract.Cases))
	transitions := make([]Transition, 0, input.Contract.FixedTransitionTotal)
	previousDigest := ""
	sequence := 0
	for _, definition := range input.Contract.Cases {
		status := definition.InitialStatus
		result := CaseResult{ID: definition.ID, ClaimID: definition.ClaimID,
			InitialStatus: definition.InitialStatus, ExpectedFinalStatus: definition.ExpectedFinalStatus,
			StatusHistory: []string{status}, Producer: definition.Producer, Consumer: definition.Consumer,
			MetaOperation: definition.MetaOperation, ProofChoice: definition.ProofChoice}
		for _, item := range definition.Evidence {
			before := status
			after, reason := nextStatus(status, item.Kind)
			if after == "" {
				return nil, nil, Metrics{}, transitionReason(definition.ID, item.Kind, before)
			}
			if item.ClaimID != definition.ClaimID || item.Producer != producerID ||
				item.Consumer != consumerID || item.MetaOperation != metaOperation ||
				item.Coordinate.Reason != reason || item.Coordinate.Step == "" || item.Coordinate.Stage == "" {
				return nil, nil, Metrics{}, "EVIDENCE_COORDINATE_MISMATCH"
			}
			sequence++
			transition := Transition{Sequence: sequence, CaseID: definition.ID, ClaimID: definition.ClaimID,
				Before: before, After: after, EvidenceID: item.ID, EvidenceKind: item.Kind,
				Coordinate: item.Coordinate, PreviousDigest: previousDigest}
			transition.TransitionDigest = transitionDigest(transition)
			previousDigest = transition.TransitionDigest
			transitions = append(transitions, transition)
			status = after
			result.StatusHistory = append(result.StatusHistory, status)
			metrics.TransitionTotal++
			switch before + "->" + after {
			case "OPEN->DISCHARGED":
				metrics.OpenToDischargedTotal++
			case "DISCHARGED->REFUTED":
				metrics.DischargedToRefutedTotal++
				metrics.NonMonotonicRevisionTotal++
				result.RefutationObserved = true
			case "REFUTED->DISCHARGED":
				metrics.RefutedToDischargedTotal++
			}
		}
		result.CurrentStatus = status
		result.HistoryRetained = len(result.StatusHistory) == len(definition.Evidence)+1
		if !result.HistoryRetained || status != definition.ExpectedFinalStatus {
			return nil, nil, Metrics{}, "CASE_HISTORY_OR_FINAL_STATUS_MISMATCH"
		}
		if status == "DISCHARGED" {
			metrics.CurrentDischargedTotal++
		}
		if status == "REFUTED" {
			metrics.CurrentRefutedTotal++
		}
		metrics.RetainedStateTotal += len(result.StatusHistory)
		cases = append(cases, result)
	}
	metrics.CurrentDischargeBasisPoints = metrics.CurrentDischargedTotal * 10000 / metrics.FixedClaimTotal
	if metrics.TransitionTotal != input.Contract.FixedTransitionTotal || metrics.OpenToDischargedTotal != 3 ||
		metrics.DischargedToRefutedTotal != 2 || metrics.RefutedToDischargedTotal != 1 ||
		metrics.CurrentDischargedTotal != 2 || metrics.CurrentRefutedTotal != 1 || metrics.RetainedStateTotal != 9 {
		return nil, nil, Metrics{}, "FIXED_TRANSITION_COUNT_MISMATCH"
	}
	return cases, transitions, metrics, ""
}

func nextStatus(before, kind string) (string, string) {
	switch {
	case before == "OPEN" && kind == "SUPPORT":
		return "DISCHARGED", "SUPPORTING_EVIDENCE_ACCEPTED"
	case before == "DISCHARGED" && kind == "REFUTE":
		return "REFUTED", "NEW_EVIDENCE_REFUTES_DISCHARGED"
	case before == "REFUTED" && kind == "SUPPORT":
		return "DISCHARGED", "LATER_SUPPORT_REDISCHARGES_REFUTED"
	default:
		return "", ""
	}
}

func transitionReason(caseID, kind, before string) string {
	return "INVALID_TRANSITION_" + caseID + "_" + before + "_" + kind
}

func finish(report Report, decision, resolution, reason string) Report {
	report.Decision, report.Resolution, report.Reason = decision, resolution, reason
	report.ReportDigest = reportDigest(report)
	return report
}

func canonicalContract() contract {
	return contract{Schema: contractSchema, FixedClaimTotal: 3, FixedTransitionTotal: 6,
		Cases: []caseDefinition{
			caseDefinition{ID: "support-only", ClaimID: "claim://support-only", InitialStatus: "OPEN", ExpectedFinalStatus: "DISCHARGED", Producer: producerID, Consumer: consumerID, MetaOperation: metaOperation, ProofChoice: "FOUNDATION", Evidence: []evidence{ev("support-only-1", "claim://support-only", "SUPPORT", "independent reproduction supports the claim", "FOUNDATION", "EVIDENCE", "accept-support", "SUPPORTING_EVIDENCE_ACCEPTED")}},
			caseDefinition{ID: "new-counterevidence", ClaimID: "claim://new-counterevidence", InitialStatus: "OPEN", ExpectedFinalStatus: "REFUTED", Producer: producerID, Consumer: consumerID, MetaOperation: metaOperation, ProofChoice: "COHERENCE", Evidence: []evidence{ev("new-counterevidence-1", "claim://new-counterevidence", "SUPPORT", "initial reproduction supports the claim", "COHERENCE", "EVIDENCE", "accept-support", "SUPPORTING_EVIDENCE_ACCEPTED"), ev("new-counterevidence-2", "claim://new-counterevidence", "REFUTE", "new counterexample contradicts the discharged claim", "COHERENCE", "EVIDENCE", "accept-counterexample", "NEW_EVIDENCE_REFUTES_DISCHARGED")}},
			caseDefinition{ID: "re-evaluation", ClaimID: "claim://re-evaluation", InitialStatus: "OPEN", ExpectedFinalStatus: "DISCHARGED", Producer: producerID, Consumer: consumerID, MetaOperation: metaOperation, ProofChoice: "REGRESSION", Evidence: []evidence{ev("re-evaluation-1", "claim://re-evaluation", "SUPPORT", "baseline reproduction supports the claim", "REGRESSION", "EVIDENCE", "accept-support", "SUPPORTING_EVIDENCE_ACCEPTED"), ev("re-evaluation-2", "claim://re-evaluation", "REFUTE", "new counterexample contradicts the discharged claim", "REGRESSION", "EVIDENCE", "accept-counterexample", "NEW_EVIDENCE_REFUTES_DISCHARGED"), ev("re-evaluation-3", "claim://re-evaluation", "SUPPORT", "later independent support defeats the refutation", "REGRESSION", "REASSESS", "reconsider-with-support", "LATER_SUPPORT_REDISCHARGES_REFUTED")}},
		}}
}

func ev(id, claimID, kind, basis, proof, stage, step, reason string) evidence {
	return evidence{ID: id, ClaimID: claimID, Kind: kind, Basis: basis, Producer: producerID,
		Consumer: consumerID, MetaOperation: metaOperation, ProofChoice: proof,
		Coordinate: coordinate{Stage: stage, Step: step, Reason: reason}}
}
