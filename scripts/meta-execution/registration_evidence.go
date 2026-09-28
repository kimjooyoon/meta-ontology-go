package main

import (
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func finishRegistrationMaterialization(plan generation.Plan, action generation.Action,
	candidate syntaxregistration.Candidate, materialized operationMaterialization) (operationMaterialization, *operationError) {
	if candidate.RequiredArtifacts != syntaxregistration.RequiredArtifacts ||
		len(candidate.Artifacts) != syntaxregistration.RequiredArtifacts ||
		candidate.Required != candidate.Emitted || candidate.Emitted != len(candidate.Members) ||
		candidate.Emitted == 0 || candidate.RepositoryWrites != 0 || candidate.ApplyAuthorized ||
		candidate.PromotionAllowed || candidate.Admission != "UNASSESSED" {
		return materialized, newOperationError("ARTIFACT", "observe-native-obligations",
			"REGISTRATION_OBLIGATION_REFUTED", "KNOWN_CONTRADICTION", "preserve-registration-counterexample")
	}
	materialized.OperationID = string(sourcepolicy.OperationRegisterSyntax)
	materialized.ContractDigest = candidate.ContractDigest
	for _, id := range action.RequiredIndicatorIDs {
		switch id {
		case "registration.artifact-completeness/v1", "registration.execution-identity/v1",
			"registration.native-conformance/v1", "registration.replay/v1":
		default:
			return materialized, newOperationError("ARTIFACT", "bind-indicator",
				"REGISTRATION_INDICATOR_UNBOUND", "KNOWN_CONTRADICTION", "restore-native-indicator-binding")
		}
		observation := generation.IndicatorObservation{Schema: generation.IndicatorObservationSchema,
			IndicatorID: id, Subject: action.Subject, HeadSHA: plan.HeadSHA,
			OperationID: materialized.OperationID, ValueKind: "integer", ActualValue: 1,
			ExpectedPredicate: "equal", ExpectedBound: 1, TransformedSubject: action.Subject}
		raw, err := json.Marshal(observation)
		if err != nil {
			return materialized, registrationNativeFailure(err, "encode-indicator")
		}
		materialized.Indicators = append(materialized.Indicators, generation.IndicatorReceipt{
			ID: id, Verdict: generation.IndicatorVerdictPass, ProofChoice: action.ProofChoice,
			EvidenceDigest: strings.TrimPrefix(digestBytes(raw), "sha256:"), Observation: &observation})
	}
	evidence := struct {
		Schema        string                       `json:"schema"`
		RequestDigest string                       `json:"request_digest"`
		Candidate     syntaxregistration.Candidate `json:"candidate"`
		Process       operationReplayEvidence      `json:"process"`
	}{Schema: "gooo/native-registration-instance/v1", RequestDigest: candidate.RequestDigest,
		Candidate: candidate, Process: operationReplayEvidenceFrom(materialized.Executor, materialized.Evaluator, materialized.Verifier)}
	raw, err := json.Marshal(evidence)
	if err != nil {
		return materialized, registrationNativeFailure(err, "encode-native-instance")
	}
	materialized.Canonical, materialized.InstanceDigest = raw, digestBytes(raw)
	return materialized, nil
}
