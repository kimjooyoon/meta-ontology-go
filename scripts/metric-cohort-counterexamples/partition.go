package main

import (
	"slices"
	"strings"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func repartitionPlan(plan *metricstrategy.Plan) error {
	slices.SortFunc(plan.Bindings, func(a, b metricstrategy.Binding) int {
		return strings.Compare(a.IndicatorID, b.IndicatorID)
	})
	for index := range plan.Candidates {
		candidate := &plan.Candidates[index]
		var ids []string
		operations := map[string]bool{}
		for _, binding := range plan.Bindings {
			if binding.Family == candidate.ProofChoice {
				ids = append(ids, binding.IndicatorID)
				operations[binding.MetaOperation] = true
			}
		}
		if slices.Equal(ids, candidate.IndicatorIDs) {
			continue
		}
		candidate.IndicatorIDs, candidate.IndicatorCount = ids, len(ids)
		candidate.MetaOperations = nil
		for operation := range operations {
			candidate.MetaOperations = append(candidate.MetaOperations, operation)
		}
		slices.Sort(candidate.MetaOperations)
		candidate.EvidenceDigest = ""
		digest, err := artifact.Digest(candidate)
		if err != nil {
			return err
		}
		candidate.EvidenceDigest = digest
	}
	plan.Input.IndicatorCount, plan.Digest = len(plan.Bindings), ""
	digest, err := artifact.Digest(plan)
	plan.Digest = digest
	return err
}
