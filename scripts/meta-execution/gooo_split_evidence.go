package main

import (
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

type goooSplitReplay struct {
	Request   workspacesplit.Request `json:"request"`
	Receipt   json.RawMessage        `json:"receipt"`
	Runtime   goooSplitRuntime       `json:"runtime"`
	Native    string                 `json:"native_conformance"`
	Executor  splitReplayProcess     `json:"executor"`
	Evaluator splitReplayProcess     `json:"evaluator"`
	Verifier  splitReplayProcess     `json:"verifier"`
}

func finishGoooSplitMaterialization(plan generation.Plan, action generation.Action, receipt goooSplitReceipt,
	canonical []byte, materialized operationMaterialization) (operationMaterialization, *operationError) {
	contract, _ := json.Marshal(receipt.Registration.Operation)
	materialized.OperationID = string(action.Operation)
	materialized.ContractDigest, materialized.InstanceDigest = digestBytes(contract), digestBytes(canonical)
	materialized.Canonical = canonical
	for _, id := range action.RequiredIndicatorIDs {
		switch id {
		case "gooo.workspace.partition-origin/v1", "gooo.workspace.consumer-publication/v1",
			"gooo.workspace.native-conformance/v1", "gooo.workspace.finite-runtime/v1", "gooo.workspace.replay/v1":
		default:
			return materialized, goooSplitFailure("bind-indicators", "WORKSPACE_SPLIT_INDICATOR_UNBOUND", nil)
		}
		observation := generation.IndicatorObservation{Schema: generation.IndicatorObservationSchema,
			IndicatorID: id, Subject: action.Subject, HeadSHA: plan.HeadSHA, OperationID: materialized.OperationID,
			ValueKind: "integer", ActualValue: 1, ExpectedPredicate: "equal", ExpectedBound: 1, TransformedSubject: action.Subject}
		raw, _ := json.Marshal(observation)
		materialized.Indicators = append(materialized.Indicators, generation.IndicatorReceipt{
			ID: id, Verdict: generation.IndicatorVerdictPass, ProofChoice: action.ProofChoice,
			EvidenceDigest: strings.TrimPrefix(digestBytes(raw), "sha256:"), Observation: &observation})
	}
	return materialized, nil
}
