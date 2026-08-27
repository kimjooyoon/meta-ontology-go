package nonmonotonicrefutationoracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	oracleSchema                     = "gooo/meta-nonmonotonic-refutation-oracle/v4"
	producerSchema                   = "gooo/meta-nonmonotonic-refutation-producer/v4"
	sourceSchema                     = "gooo/meta-nonmonotonic-refutation-source/v3"
	producerID                       = "producer://nonmonotonic-refutation"
	consumerID                       = "consumer://nonmonotonic-refutation-oracle"
	metaOperation                    = "meta://revise-claim-by-evidence"
	statusOpen                       = "OPEN"
	statusDischarged                 = "DISCHARGED"
	statusRefuted                    = "REFUTED"
	relationSupports                 = "SUPPORTS"
	relationContradicts              = "CONTRADICTS"
	relationInsufficient             = "INSUFFICIENT"
	relationUnknown                  = "UNKNOWN"
	revisionNone                     = "NONE"
	revisionSupersedes               = "SUPERSEDES"
	providerHistoricalFixture        = "HISTORICAL_FIXTURE"
	policyUnknownRetain              = "RETAIN_CURRENT"
	policyInsufficientRetain         = "RETAIN_CURRENT"
	policyOrdinarySupportRetain      = "RETAIN_REFUTED"
	policyCorrectionTargetEvidence   = "EVIDENCE_DIGEST"
	policyFoundationFirstClaimEvent  = "FIRST_CLAIM_OBSERVATION"
	policyCoherenceLaterClaimOpening = "LATER_OBSERVATION_AFTER_FIRST_SOURCE_EVENT"
	policyRegressionTargetedHistory  = "SEQUENCE_AT_LEAST_5_WITH_PRIOR_CLAIM"
	noEvidenceTarget                 = "none"
)

// Judge is an independent adjudicator. It parses and lowers raw source,
// reconstructs its own policy and fixture model, validates producer evidence,
// and only then replays the append-only claim ledger.
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
		ProofChoice: input.ProofChoice,
		Effects:     input.Effects,
		Vocabulary: Vocabulary{
			FixtureKnowledge: "HISTORICAL_FIXTURE_ONLY",
			CurrentEvidence:  "ACCEPTED_APPEND_ONLY_LEDGER_EVIDENCE",
			UnknownEvidence:  "UNKNOWN_OR_INSUFFICIENT_NOT_CURRENT_EVIDENCE",
		},
		MetaValue:             "evidence revises current knowledge while every prior ledger row remains inspectable",
		FalsifiablePrediction: "changing a proposition, observed fixture value, proof admission, or correction target changes relation, resolution, or transition; comment-only edits preserve semantic evidence",
	}
	model, reason := reconstructSource(source)
	if reason == "" {
		report.Policy = model.Contract.Policy
		reason = validateInput(input, source, model)
	}
	if reason != "" {
		return finishFailure(report, reason), nil
	}
	cases, transitions, metrics, reason := replay(model)
	report.Cases, report.Transitions, report.Metrics = cases, transitions, metrics
	if reason != "" {
		return finishFailure(report, reason), nil
	}
	report.Conformance = Conformance{Decision: "PASS", Resolution: "EXACT", Reason: "SOURCE_RECONSTRUCTION_AND_REPLAY"}
	report.SubjectResolution = subjectResolution(cases, metrics)
	return finish(report, "PASS", reportResolution(metrics), "NONMONOTONIC_REFUTATION_OBSERVED"), nil
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
		if entity.ID == "" {
			return sourceModel{}, "SOURCE_ENTITY_ID_MISSING"
		}
		entities[entity.Name] = entity.ID
		if strings.HasPrefix(entity.ID, "gooo://nonmonotonic-refutation/claim/") {
			model.Contract.Claims = append(model.Contract.Claims, sourceClaim{ID: entity.ID})
		}
	}
	if len(model.Contract.Claims) != 3 {
		return sourceModel{}, "SOURCE_CLAIM_DENOMINATOR_MISMATCH"
	}

	var policySeen bool
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "meta.revision-policy:v1;") {
			continue
		}
		if policySeen || len(activity.Inputs) != 1 || activity.Output == "" {
			return sourceModel{}, "SOURCE_POLICY_BINDING_MISMATCH"
		}
		policySeen = true
		policyID := entities[activity.Inputs[0].Name]
		outputID := entities[activity.Output]
		if !strings.HasPrefix(policyID, "gooo://nonmonotonic-refutation/policy/") || !strings.HasPrefix(outputID, "gooo://nonmonotonic-refutation/policy-binding/") {
			return sourceModel{}, "SOURCE_POLICY_ENDPOINT_MISMATCH"
		}
		fields, ok := parsePolicyProgram(activity.ValueProgram)
		if !ok || fields["policy_id"] != policyID {
			return sourceModel{}, "SOURCE_POLICY_FIELDS_INVALID"
		}
		model.Contract.Policy = revisionPolicy{
			ID: fields["policy_id"], CorrectionRelation: fields["correction_relation"], CorrectionTarget: fields["correction_target"],
			UnknownAction: fields["unknown_action"], InsufficientAction: fields["insufficient_action"],
			OrdinarySupportAfterRefuted: fields["ordinary_support_after_refuted"], FoundationRule: fields["foundation_rule"],
			CoherenceRule: fields["coherence_rule"], RegressionRule: fields["regression_rule"], FixtureClass: fields["fixture_class"], PolicyDigest: fields["policy_digest"],
		}
		if !validPolicy(model.Contract.Policy) {
			return sourceModel{}, "SOURCE_POLICY_DIGEST_OR_RULE_MISMATCH"
		}
	}
	if !policySeen {
		return sourceModel{}, "SOURCE_POLICY_MISSING"
	}

	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "meta.observe:v3;") {
			continue
		}
		if len(activity.Inputs) != 5 || activity.Output == "" {
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
		subjectID := entities[activity.Inputs[2].Name]
		inputID := entities[activity.Inputs[3].Name]
		if entities[activity.Inputs[0].Name] != claimID ||
			entities[activity.Inputs[1].Name] != "gooo://nonmonotonic-refutation/predicate/"+fields["predicate"] ||
			subjectID != "gooo://nonmonotonic-refutation/subject/"+fields["subject"] ||
			inputID != "gooo://nonmonotonic-refutation/input/"+fields["input"] ||
			entities[activity.Inputs[4].Name] != "gooo://nonmonotonic-refutation/value/one" || fields["expected"] != "1" {
			return sourceModel{}, "SOURCE_SUBJECT_INPUT_RESOLUTION_MISMATCH"
		}
		observation := sourceObservation{
			ID: observationID, Activity: activity.Name, ClaimID: claimID, Sequence: len(model.Contract.Observations) + 1,
			Proposition: fields["proposition"], Subject: fields["subject"], Input: fields["input"], Predicate: fields["predicate"], ExpectedValue: fields["expected"], ObservedValue: fields["observed"],
			ObservedMaterial: fields["observed_material"], ObservationQuality: fields["observation_quality"], ProviderClass: fields["provider_class"], Provenance: fields["provenance"],
			RevisionRelation: fields["revision_relation"], SupersedesEvidenceDigest: fields["supersedes_evidence_digest"], PolicyID: fields["policy_id"], PolicyDigest: fields["policy_digest"],
			Producer: fields["producer"], Consumer: fields["consumer"], MetaOperation: fields["meta_operation"], ProofChoice: fields["proof_choice"], Coordinate: coordinate{Stage: fields["stage"], Step: fields["step"]}, TargetAddress: subjectID + "|" + inputID,
		}
		if !validObservation(observation, model.Contract.Policy) {
			return sourceModel{}, "SOURCE_OBSERVATION_POLICY_OR_RECIPE_MISMATCH"
		}
		model.Contract.Observations = append(model.Contract.Observations, observation)
	}
	if len(model.Contract.Observations) != 6 {
		return sourceModel{}, "SOURCE_OBSERVATION_DENOMINATOR_MISMATCH"
	}
	model.Contract.Schema = sourceSchema
	model.Contract.FixedCaseTotal = len(model.Contract.Claims)
	model.Contract.FixedClaimTotal = len(model.Contract.Claims)
	model.Contract.FixedObservationTotal = len(model.Contract.Observations)
	model.Contract.FixedLedgerRowTotal = len(model.Contract.Observations)
	if !completeClaims(&model.Contract) {
		return sourceModel{}, "SOURCE_CLAIM_BINDING_MISMATCH"
	}
	for index := range model.Contract.Observations {
		model.Contract.Observations[index].EvidenceDigest = evidenceDigest(model.Contract.Observations[index])
	}
	return model, ""
}

func parsePolicyProgram(program string) (map[string]string, bool) {
	return parseFields(program, "meta.revision-policy:v1", knownPolicyField, []string{"policy_id", "correction_relation", "correction_target", "unknown_action", "insufficient_action", "ordinary_support_after_refuted", "foundation_rule", "coherence_rule", "regression_rule", "fixture_class", "policy_digest"})
}

func parseObservationProgram(program string) (map[string]string, bool) {
	return parseFields(program, "meta.observe:v3", knownObservationField, []string{"claim", "proposition", "subject", "input", "predicate", "expected", "observed", "observed_material", "observation_quality", "provider_class", "provenance", "revision_relation", "supersedes_evidence_digest", "policy_id", "policy_digest", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"})
}

func parseFields(program, marker string, known func(string) bool, required []string) (map[string]string, bool) {
	fields := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		if part == marker {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || !known(key) || (value == "" && key != "observed") {
			return nil, false
		}
		if _, exists := fields[key]; exists {
			return nil, false
		}
		fields[key] = value
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok || (fields[key] == "" && key != "observed") {
			return nil, false
		}
	}
	return fields, true
}

func knownPolicyField(key string) bool {
	for _, known := range []string{"policy_id", "correction_relation", "correction_target", "unknown_action", "insufficient_action", "ordinary_support_after_refuted", "foundation_rule", "coherence_rule", "regression_rule", "fixture_class", "policy_digest"} {
		if key == known {
			return true
		}
	}
	return false
}

func knownObservationField(key string) bool {
	for _, known := range []string{"claim", "proposition", "subject", "input", "predicate", "expected", "observed", "observed_material", "observation_quality", "provider_class", "provenance", "revision_relation", "supersedes_evidence_digest", "policy_id", "policy_digest", "producer", "consumer", "meta_operation", "proof_choice", "stage", "step"} {
		if key == known {
			return true
		}
	}
	return false
}

func validPolicy(policy revisionPolicy) bool {
	if !strings.HasPrefix(policy.ID, "gooo://nonmonotonic-refutation/policy/") || policy.CorrectionRelation != revisionSupersedes || policy.CorrectionTarget != policyCorrectionTargetEvidence || policy.UnknownAction != policyUnknownRetain || policy.InsufficientAction != policyInsufficientRetain || policy.OrdinarySupportAfterRefuted != policyOrdinarySupportRetain || policy.FoundationRule != policyFoundationFirstClaimEvent || policy.CoherenceRule != policyCoherenceLaterClaimOpening || policy.RegressionRule != policyRegressionTargetedHistory || policy.FixtureClass != providerHistoricalFixture || !validDigest(policy.PolicyDigest) {
		return false
	}
	candidate := policy
	candidate.PolicyDigest = ""
	return digestJSON(candidate) == policy.PolicyDigest
}

func validObservation(observation sourceObservation, policy revisionPolicy) bool {
	if observation.Proposition == "" || observation.Subject == "" || observation.Input == "" || observation.Predicate == "" || observation.ExpectedValue == "" || observation.ObservedMaterial == "" || observation.Provenance == "" || observation.TargetAddress == "" {
		return false
	}
	if observation.ObservationQuality != "SUFFICIENT" && observation.ObservationQuality != "UNRESOLVED" {
		return false
	}
	if observation.ProviderClass != policy.FixtureClass || observation.ProviderClass != providerHistoricalFixture || observation.PolicyID != policy.ID || observation.PolicyDigest != policy.PolicyDigest {
		return false
	}
	if observation.Producer != producerID || observation.Consumer != consumerID || observation.MetaOperation != metaOperation {
		return false
	}
	if observation.ProofChoice != "FOUNDATION" && observation.ProofChoice != "COHERENCE" && observation.ProofChoice != "REGRESSION" {
		return false
	}
	if observation.RevisionRelation != revisionNone && observation.RevisionRelation != revisionSupersedes {
		return false
	}
	if observation.RevisionRelation == revisionNone && observation.SupersedesEvidenceDigest != noEvidenceTarget {
		return false
	}
	if observation.RevisionRelation == revisionSupersedes && observation.SupersedesEvidenceDigest != noEvidenceTarget && !validDigest(observation.SupersedesEvidenceDigest) {
		return false
	}
	return observation.RevisionRelation != revisionNone || observation.SupersedesEvidenceDigest != ""
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func evidenceDigest(observation sourceObservation) string {
	return digestJSON(evidenceMaterial{
		ClaimID: observation.ClaimID, Proposition: observation.Proposition, TargetAddress: observation.TargetAddress,
		ObservedMaterial: observation.ObservedMaterial, ObservedValue: observation.ObservedValue, ObservationQuality: observation.ObservationQuality,
		ProviderClass: observation.ProviderClass, Sequence: observation.Sequence, SupersededEvidenceDigest: observation.SupersedesEvidenceDigest,
	})
}

func completeClaims(contract *sourceContract) bool {
	for index := range contract.Claims {
		claim := &contract.Claims[index]
		for _, observation := range contract.Observations {
			if observation.ClaimID != claim.ID {
				continue
			}
			if claim.Proposition == "" {
				claim.Proposition, claim.Subject, claim.Input, claim.Predicate, claim.ExpectedValue = observation.Proposition, observation.Subject, observation.Input, observation.Predicate, observation.ExpectedValue
			}
			if claim.Proposition != observation.Proposition || claim.Subject != observation.Subject || claim.Input != observation.Input || claim.Predicate != observation.Predicate || claim.ExpectedValue != observation.ExpectedValue {
				return false
			}
		}
		if claim.Proposition == "" {
			return false
		}
	}
	return true
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
	if input.SourceBindingDigest != digestJSON(sourceBinding{RawDigest: input.SourceDigest, SemanticDigest: input.SourceSemanticDigest, PolicyID: input.Contract.Policy.ID, PolicyDigest: input.Contract.Policy.PolicyDigest}) {
		return "SOURCE_BINDING_DIGEST_MISMATCH"
	}
	if input.SourceSemanticDigest != model.SemanticDigest || input.Contract.Schema != sourceSchema || input.Contract.FixedCaseTotal != 3 || input.Contract.FixedClaimTotal != 3 || input.Contract.FixedObservationTotal != 6 || input.Contract.FixedLedgerRowTotal != 6 {
		return "SOURCE_RECONSTRUCTION_MISMATCH"
	}
	if input.SourceModelDigest != digestJSON(model.Contract) || digestJSON(input.Contract) != digestJSON(model.Contract) {
		return "PRODUCER_SOURCE_MODEL_MISMATCH"
	}
	expectedRepositoryObservation := "NET_STATUS_CHANGED"
	if input.Effects.NetRepositoryStatusUnchanged {
		expectedRepositoryObservation = "NONE_OBSERVED_IN_NET_STATUS"
	}
	if input.Producer != producerID || input.Consumer != consumerID || input.MetaOperation != metaOperation || input.Effects.RepositoryWriteObservation != expectedRepositoryObservation || input.Effects.MutationAuthorityResolution != "UNKNOWN" || input.Effects.PromotionOperationsObserved != 0 {
		return "PRODUCER_PROVENANCE_OR_EFFECTS_MISMATCH"
	}
	return ""
}

func replay(model sourceModel) ([]CaseResult, []Transition, Metrics, string) {
	metrics := Metrics{FixedCaseTotal: model.Contract.FixedCaseTotal, FixedClaimTotal: model.Contract.FixedClaimTotal, FixedObservationTotal: model.Contract.FixedObservationTotal, FixedLedgerRowTotal: model.Contract.FixedLedgerRowTotal, InScopeClaimTotal: len(model.Contract.Claims)}
	cases := make([]CaseResult, len(model.Contract.Claims))
	status := make(map[string]string, len(model.Contract.Claims))
	caseIndex := make(map[string]int, len(model.Contract.Claims))
	claimObservationCount := make(map[string]int, len(model.Contract.Claims))
	currentEvidenceID := make(map[string]string, len(model.Contract.Claims))
	currentEvidenceDigest := make(map[string]string, len(model.Contract.Claims))
	for index, claim := range model.Contract.Claims {
		caseID := strings.TrimPrefix(claim.ID, "gooo://nonmonotonic-refutation/claim/")
		cases[index] = CaseResult{ID: caseID, ClaimID: claim.ID, Proposition: claim.Proposition, Subject: claim.Subject, Input: claim.Input, FixtureKnowledge: model.Contract.Policy.FixtureClass, InitialStatus: statusOpen, StatusHistory: []string{statusOpen}}
		status[claim.ID] = statusOpen
		caseIndex[claim.ID] = index
	}
	transitions := make([]Transition, 0, len(model.Contract.Observations))
	previousDigest := ""
	for index, observation := range model.Contract.Observations {
		caseNumber, ok := caseIndex[observation.ClaimID]
		if !ok || observation.Sequence != index+1 || evidenceDigest(observation) != observation.EvidenceDigest {
			return cases, transitions, metrics, "SOURCE_OBSERVATION_ORDER_OR_DIGEST_MISMATCH"
		}
		before := status[observation.ClaimID]
		claim := model.Contract.Claims[caseNumber]
		relation := classify(claim, observation)
		proofAdmitted, proofReason := admitProof(model.Contract.Policy, observation, claimObservationCount[observation.ClaimID])
		after, accepted, revisionReason := revise(model.Contract.Policy, before, relation, observation, transitions, proofAdmitted)
		transition := Transition{Sequence: index + 1, CaseID: cases[caseNumber].ID, ClaimID: observation.ClaimID, Before: before, After: after, Accepted: accepted, EvidenceID: observation.ID, Relation: relation, RevisionRelation: observation.RevisionRelation, SupersedesEvidenceDigest: observation.SupersedesEvidenceDigest, EvidenceBasis: evidenceBasis(observation), EvidenceDigest: observation.EvidenceDigest, EvidenceProvenance: observation.Provenance, ProviderClass: observation.ProviderClass, ProofChoice: observation.ProofChoice, ProofAdmitted: proofAdmitted, ProofAdmission: proofReason, Coordinate: coordinate{Stage: observation.Coordinate.Stage, Step: observation.Coordinate.Step, Reason: revisionReason}, PreviousDigest: previousDigest}
		transition.TransitionDigest = transitionDigest(transition)
		previousDigest = transition.TransitionDigest
		transitions = append(transitions, transition)
		metrics.ObservationAttemptTotal++
		metrics.TransitionTotal++
		if !accepted {
			metrics.RejectedObservationTotal++
			cases[caseNumber].RejectedObservationTotal++
		}
		if accepted && before != after {
			metrics.AcceptedStateTransitionTotal++
		}
		switch relation {
		case relationSupports:
			metrics.SupportsTotal++
		case relationContradicts:
			metrics.ContradictsTotal++
		case relationInsufficient:
			metrics.InsufficientTotal++
		case relationUnknown:
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
		if relation == relationContradicts {
			cases[caseNumber].RefutationObserved = true
		}
		if accepted {
			currentEvidenceID[observation.ClaimID] = observation.ID
			currentEvidenceDigest[observation.ClaimID] = observation.EvidenceDigest
		}
		status[observation.ClaimID] = after
		claimObservationCount[observation.ClaimID]++
		cases[caseNumber].StatusHistory = append(cases[caseNumber].StatusHistory, after)
		cases[caseNumber].ObservationTotal++
	}
	for index := range cases {
		cases[index].CurrentStatus = status[cases[index].ClaimID]
		cases[index].CurrentEvidenceID = currentEvidenceID[cases[index].ClaimID]
		cases[index].CurrentEvidenceDigest = currentEvidenceDigest[cases[index].ClaimID]
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
	if metrics.FixedCaseTotal != 3 || metrics.FixedClaimTotal != 3 || metrics.FixedObservationTotal != 6 || metrics.FixedLedgerRowTotal != 6 || metrics.ObservationAttemptTotal != 6 || metrics.TransitionTotal != 6 {
		return cases, transitions, metrics, "FIXED_SOURCE_COUNT_MISMATCH"
	}
	metrics.CurrentDischargeBasisPoints = metrics.CurrentDischargedTotal * 10000 / metrics.FixedClaimTotal
	return cases, transitions, metrics, ""
}

func classify(claim sourceClaim, observation sourceObservation) string {
	if claim.ID != observation.ClaimID || claim.Proposition != observation.Proposition || claim.Subject != observation.Subject || claim.Input != observation.Input || claim.Predicate != observation.Predicate || claim.ExpectedValue != observation.ExpectedValue || observation.ProviderClass != providerHistoricalFixture || observation.ObservationQuality != "SUFFICIENT" || evidenceDigest(observation) != observation.EvidenceDigest {
		return relationUnknown
	}
	if observation.Predicate != "equality" || observation.Subject == "" || observation.Input == "" {
		return relationUnknown
	}
	propositionPrefix := "equals:" + observation.Subject + ":" + observation.Input + ":"
	if !strings.HasPrefix(observation.Proposition, propositionPrefix) {
		return relationUnknown
	}
	propositionValue := strings.TrimPrefix(observation.Proposition, propositionPrefix)
	if propositionValue == "" || strings.Contains(propositionValue, ":") || propositionValue != observation.ExpectedValue {
		return relationUnknown
	}
	if observation.ObservedValue == "" {
		return relationInsufficient
	}
	if observation.ObservedValue == propositionValue {
		return relationSupports
	}
	return relationContradicts
}

func admitProof(policy revisionPolicy, observation sourceObservation, priorClaimObservations int) (bool, string) {
	if observation.ProviderClass != policy.FixtureClass || evidenceDigest(observation) != observation.EvidenceDigest || !validDigest(observation.EvidenceDigest) {
		return false, "EVIDENCE_DIGEST_RECOMPUTATION_REJECTED"
	}
	switch observation.ProofChoice {
	case "FOUNDATION":
		if policy.FoundationRule == policyFoundationFirstClaimEvent && priorClaimObservations == 0 {
			return true, "FOUNDATION_FIRST_CLAIM_OBSERVATION_ADMITTED"
		}
		return false, "FOUNDATION_RULE_REJECTED"
	case "COHERENCE":
		if policy.CoherenceRule == policyCoherenceLaterClaimOpening && observation.Sequence > 1 {
			return true, "COHERENCE_LATER_OBSERVATION_ADMITTED"
		}
		return false, "COHERENCE_RULE_REJECTED"
	case "REGRESSION":
		if policy.RegressionRule == policyRegressionTargetedHistory && observation.Sequence >= 5 && priorClaimObservations > 0 {
			return true, "REGRESSION_PRIOR_CLAIM_HISTORY_ADMITTED"
		}
		return false, "REGRESSION_RULE_REJECTED"
	default:
		return false, "PROOF_CHOICE_REJECTED"
	}
}

func revise(policy revisionPolicy, before, relation string, observation sourceObservation, prior []Transition, proofAdmitted bool) (string, bool, string) {
	if !proofAdmitted {
		return before, false, "PROOF_ADMISSIBILITY_REJECTED_CURRENT_STATE_RETAINED"
	}
	switch relation {
	case relationSupports:
		if before != statusRefuted {
			return statusDischarged, true, "PROPOSITION_MATCHES_OBSERVATION"
		}
		if policy.OrdinarySupportAfterRefuted != policyOrdinarySupportRetain {
			return before, false, "ORDINARY_SUPPORT_POLICY_NOT_BOUNDED"
		}
		if observation.RevisionRelation != policy.CorrectionRelation || observation.RevisionRelation != revisionSupersedes {
			return before, false, "UNTARGETED_SUPPORT_RETAINS_REFUTED"
		}
		if policy.CorrectionTarget != policyCorrectionTargetEvidence || observation.SupersedesEvidenceDigest == noEvidenceTarget || !hasExactRefutationEvidence(prior, observation.SupersedesEvidenceDigest) {
			return before, false, "CORRECTION_TARGET_NOT_FOUND_CURRENT_STATE_RETAINED"
		}
		return statusDischarged, true, "TARGETED_CORRECTION_SUPERSEDES_EXACT_REFUTATION"
	case relationContradicts:
		return statusRefuted, true, "OBSERVATION_DIRECTLY_CONTRADICTS_PROPOSITION"
	case relationInsufficient:
		if policy.InsufficientAction == policyInsufficientRetain {
			return before, false, "INSUFFICIENT_EVIDENCE_RETAINS_CURRENT_STATE"
		}
		return before, false, "INSUFFICIENT_POLICY_REJECTED"
	case relationUnknown:
		if policy.UnknownAction == policyUnknownRetain {
			return before, false, "UNKNOWN_RELATION_RETAINS_CURRENT_STATE"
		}
		return before, false, "UNKNOWN_POLICY_REJECTED"
	default:
		return before, false, "UNRECOGNIZED_RELATION_RETAINS_CURRENT_STATE"
	}
}

func hasExactRefutationEvidence(prior []Transition, target string) bool {
	for _, transition := range prior {
		if transition.Relation == relationContradicts && transition.Accepted && transition.Before == statusDischarged && transition.After == statusRefuted && transition.EvidenceDigest == target {
			return true
		}
	}
	return false
}

func evidenceBasis(observation sourceObservation) string {
	return fmt.Sprintf("fixture_knowledge=%s current_evidence_candidate=%s claim=%s proposition=%s target=%s input=%s observed_material=%s observed=%s quality=%s provenance=%s digest=%s policy=%s/%s revision=%s supersedes=%s", observation.ProviderClass, observation.ID, observation.ClaimID, observation.Proposition, observation.TargetAddress, observation.Input, observation.ObservedMaterial, observation.ObservedValue, observation.ObservationQuality, observation.Provenance, observation.EvidenceDigest, observation.PolicyID, observation.PolicyDigest, observation.RevisionRelation, observation.SupersedesEvidenceDigest)
}

func subjectResolution(cases []CaseResult, metrics Metrics) SubjectResolution {
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
	reason := "CURRENT_LEDGER_DISTRIBUTION"
	if metrics.RejectedObservationTotal > 0 || open > 0 {
		resolution = "LOWER_RESOLUTION"
		reason = "CURRENT_STATE_RETAINED_AFTER_UNRESOLVED_OBSERVATION"
	} else if refuted > 0 {
		resolution = "PARTIAL"
	}
	return SubjectResolution{Decision: fmt.Sprintf("DISCHARGED=%d;REFUTED=%d;OPEN=%d", discharged, refuted, open), Resolution: resolution, Reason: reason}
}

func reportResolution(metrics Metrics) string {
	if metrics.RejectedObservationTotal > 0 {
		return "LOWER_RESOLUTION"
	}
	return "EXACT"
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

func transitionDigest(transition Transition) string {
	transition.TransitionDigest = ""
	return digestJSON(transition)
}

func reportDigest(report Report) string {
	report.ReportDigest = ""
	return digestJSON(report)
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
