package toolchainconformance

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func inspectConcept(raw []byte) (conceptCounts, error) {
	artifact := conceptArtifact{}
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return conceptCounts{}, fmt.Errorf("decode concept artifact: %w", err)
	}
	if artifact.Decision != DecisionPass || !validDigest(artifact.ArtifactDigest) ||
		!validDigest(artifact.CatalogDigest) {
		return conceptCounts{}, fmt.Errorf("concept artifact is not exact")
	}
	matches := make([]conceptItem, 0, 1)
	for _, item := range artifact.Report.Concepts {
		if item.ID == ExpectedConceptID {
			matches = append(matches, item)
		}
	}
	if len(matches) != 1 {
		return conceptCounts{}, fmt.Errorf("conformance concept is not unique")
	}
	item := matches[0]
	useCases := make([]string, len(item.UseCases))
	for index := range item.UseCases {
		useCases[index] = item.UseCases[index].ID
	}
	if item.MetaOperation != ExpectedMetaOperation || item.Stage != "OPERATING" ||
		item.NoveltyClaim || !reflect.DeepEqual(item.CodeBindings, expectedCodeBindings) ||
		!reflect.DeepEqual(item.MetricBindings, metricIDs()) ||
		!reflect.DeepEqual(useCases, expectedUseCases) {
		return conceptCounts{}, fmt.Errorf("conformance concept binding drift")
	}
	return conceptCounts{
		ArtifactDigest:  artifact.ArtifactDigest,
		CatalogDigest:   artifact.CatalogDigest,
		ConceptBindings: 1,
		CodeBindings:    len(item.CodeBindings),
		MetricBindings:  len(item.MetricBindings),
		UseCaseBindings: len(item.UseCases),
	}, nil
}
