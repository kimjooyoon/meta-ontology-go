package artifactcoverage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/actionability"
)

func Load(actionabilityPath, observationPath string) (actionability.Report, ObservationDocument, error) {
	action, err := decodeDocument[actionability.Report](actionabilityPath)
	if err != nil {
		return actionability.Report{}, ObservationDocument{}, err
	}
	if err := verifyActionabilityDigest(action); err != nil {
		return actionability.Report{}, ObservationDocument{}, err
	}
	observations, err := decodeDocument[ObservationDocument](observationPath)
	if err != nil {
		return actionability.Report{}, ObservationDocument{}, err
	}
	normalizeObservations(&observations)
	return action, observations, nil
}

func decodeDocument[T any](path string) (T, error) {
	var document T
	data, err := os.ReadFile(path)
	if err != nil {
		return document, err
	}
	if err := json.Unmarshal(data, &document); err != nil {
		return document, err
	}
	return document, nil
}

func verifyActionabilityDigest(report actionability.Report) error {
	digest := report.ReportDigest
	report.ReportDigest = ""
	if !validDigest(digest) || digestJSON(report) != digest {
		return fmt.Errorf("actionability report digest is invalid")
	}
	return nil
}
