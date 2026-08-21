package metricstrategy

import (
	"fmt"
	"sort"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func buildCandidates(bindings []Binding) ([]Candidate, error) {
	candidates := make([]Candidate, 0, 3)
	for _, choice := range proofChoices() {
		subset := make([]Binding, 0)
		ids := make([]string, 0)
		operations := make(map[string]bool)
		unsatisfied := 0
		for _, binding := range bindings {
			if binding.Family != choice {
				continue
			}
			subset, ids = append(subset, binding), append(ids, binding.IndicatorID)
			operations[binding.MetaOperation] = true
			if binding.Status != "SATISFIED" {
				unsatisfied++
			}
		}
		if len(subset) == 0 {
			return nil, fmt.Errorf("metric strategy choice %s has no indicators", choice)
		}
		operationList := make([]string, 0, len(operations))
		for operation := range operations {
			operationList = append(operationList, operation)
		}
		sort.Strings(operationList)
		digest, err := artifact.Digest(subset)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, Candidate{ProofChoice: choice, IndicatorIDs: ids, MetaOperations: operationList, IndicatorCount: len(subset), UnsatisfiedCount: unsatisfied, Admissible: unsatisfied == 0, EvidenceDigest: digest})
	}
	return candidates, nil
}

