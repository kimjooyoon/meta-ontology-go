package metricintervention

import (
	"encoding/json"
	"fmt"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metrictransition"
)

func Project(registry Registry, baseline metrictransition.Counts, predicted, observed metric.Delta) ([]Projection, error) {
	if err := ValidateRegistry(registry); err != nil {
		return nil, err
	}
	baseVector, err := numericVector(baseline)
	if err != nil {
		return nil, err
	}
	predictedVector, err := numericVector(predicted)
	if err != nil {
		return nil, err
	}
	observedVector, err := numericVector(observed)
	if err != nil {
		return nil, err
	}
	projections := make([]Projection, 0, len(registry.Dimensions))
	for _, dimension := range registry.Dimensions {
		prediction, predictedOK := predictedVector[dimension.ID]
		observation, observedOK := observedVector[dimension.ID]
		base, baselineOK := baseVector[dimension.ID]
		if !predictedOK || !observedOK || (dimension.Kind == "STATE" && !baselineOK) {
			return nil, fmt.Errorf("dimension %q has no vector binding", dimension.ID)
		}
		residual := observation - prediction
		projection := Projection{DimensionID: dimension.ID, Kind: dimension.Kind, Baseline: base, PredictedDelta: prediction, ObservedDelta: observation, Residual: residual, Projected: base + prediction, Status: metricStatus(residual == 0)}
		digest, err := artifact.Digest(projection)
		if err != nil {
			return nil, err
		}
		projection.EvidenceDigest = digest
		projections = append(projections, projection)
	}
	return projections, nil
}

func numericVector(value any) (map[string]int, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	vector := make(map[string]int)
	err = json.Unmarshal(data, &vector)
	return vector, err
}

func metricStatus(satisfied bool) string {
	if satisfied {
		return "SATISFIED"
	}
	return "VIOLATED"
}
