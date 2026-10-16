package main

import (
	"fmt"
	"os"
	"path/filepath"

	conceptoperation "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func readObservationInputs(directory string) (*conceptoperation.ObservationInputs, error) {
	document, documentErr := os.ReadFile(filepath.Join(directory, "operation-observations.json"))
	policy, policyErr := os.ReadFile(filepath.Join(directory, "observation-policy.gooo"))
	if os.IsNotExist(documentErr) && os.IsNotExist(policyErr) {
		return nil, nil
	}
	if documentErr != nil {
		return nil, fmt.Errorf("read operation observations: %w", documentErr)
	}
	if policyErr != nil {
		return nil, fmt.Errorf("read observation policy: %w", policyErr)
	}
	return &conceptoperation.ObservationInputs{Document: document, PolicySource: policy}, nil
}
