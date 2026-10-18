package publicpartialreuse

import (
	"context"
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// Selection projects the source's ordered mask onto the observed partition
// identities. The driver owns execution and its subsequent input observations.
type Selection struct {
	Partitions []Partition
	Assessment Assessment
	Origin     bodycodegen.FixedValue
}

func Select(ctx context.Context, input EvaluationInput) (Selection, error) {
	if err := input.Policy.Validate(); err != nil {
		return Selection{}, err
	}
	observation, _, _ := observe(input)
	value, assessment, err := input.Policy.Assess(ctx, observation)
	if err != nil {
		return Selection{}, err
	}
	selection := Selection{Assessment: assessment, Origin: value}
	if assessment.Decision != DecisionClosed {
		return selection, nil
	}
	if len(assessment.ExecutePartitions) != len(input.Policy.Partitions) {
		return Selection{}, errors.New("Gooo selection has no complete ordered partition mask")
	}
	for i, execute := range assessment.ExecutePartitions {
		if execute {
			selection.Partitions = append(selection.Partitions, input.Policy.Partitions[i])
		}
	}
	if len(selection.Partitions) != assessment.TestUnitsRequired ||
		len(input.Policy.Partitions)-len(selection.Partitions) != assessment.TestUnitsReusable {
		return Selection{}, errors.New("Gooo selection counts differ from its ordered mask")
	}
	return selection, nil
}
