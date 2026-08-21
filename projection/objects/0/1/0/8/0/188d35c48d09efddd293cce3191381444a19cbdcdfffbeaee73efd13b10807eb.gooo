//go:build detector_bridge

package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func literalProductionMutationExpectations() map[string]productionVector {
	return decodeLiteralProductionVectors(literalProducerMutationGZIP)
}

type productionResultMutation struct {
	name     string
	mutate   func(*production.Result)
	truth    productionVector
	observed productionVector
}

func literalResultMutations(base production.Result, input production.Input, authority production.AuthorityContext) []productionResultMutation {
	observed := literalProducerResultObservedExpectations()
	truth := literalProductionCorpusExpectations()["positive-no-delta"]
	return []productionResultMutation{
		{name: "result-wrong-decision", mutate: func(result *production.Result) {
			result.Status = production.StatusFailClosed
			result.Reasons = []production.Reason{{Code: production.ReasonDigestMismatch, Detail: "producer-only mutation"}}
		}, truth: truth, observed: observed["result-wrong-decision"]},
		{name: "result-wrong-reason", mutate: func(result *production.Result) {
			result.Reasons = []production.Reason{{Code: production.ReasonDigestMismatch, Detail: "producer-only mutation"}}
		}, truth: truth, observed: observed["result-wrong-reason"]},
		{name: "result-missing-accepted-surface", mutate: func(result *production.Result) {
			result.AcceptedSurfaceIDs = nil
		}, truth: truth, observed: observed["result-missing-accepted-surface"]},
		{name: "result-extra-accepted-surface", mutate: func(result *production.Result) {
			result.AcceptedSurfaceIDs = []semantic.ID{bridgeID("urn:gooo:surface:billing/pay-order"), bridgeID("urn:gooo:surface:unexpected")}
		}, truth: truth, observed: observed["result-extra-accepted-surface"]},
		{name: "result-count-drift", mutate: func(result *production.Result) {
			result.Observation.InferenceRecords.Value++
		}, truth: truth, observed: observed["result-count-drift"]},
		{name: "result-resource-drift", mutate: func(result *production.Result) {
			result.Observation.CPU.Value++
		}, truth: truth, observed: observed["result-resource-drift"]},
		{name: "result-input-digest-drift", mutate: func(result *production.Result) {
			result.InputDigest = bridgeHash("producer-only-input-digest")
		}, truth: truth, observed: observed["result-input-digest-drift"]},
		{name: "result-result-digest-drift", mutate: func(result *production.Result) {
			result.Digest = bridgeHash("producer-only-result-digest")
		}, truth: truth, observed: observed["result-result-digest-drift"]},
	}
}

type productionInputMutationCase struct {
	name   string
	mutate func(*production.Input)
}
type productionInputMutation struct {
	name            string
	input           production.Input
	authorityBefore production.AuthorityContext
	want            productionVector
}

func literalInputMutations(base production.Input, authority production.AuthorityContext) []productionInputMutation {
	wants := literalProductionMutationExpectations()
	mutations := make([]productionInputMutation, 0)
	for _, mutation := range literalInputMutationCases() {
		input := cloneProductionInput(base)
		mutation.mutate(&input)
		mutations = append(mutations, productionInputMutation{name: mutation.name, input: input, authorityBefore: authority, want: wants[mutation.name]})
	}
	return mutations
}
