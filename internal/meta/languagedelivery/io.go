package languagedelivery

import (
	"encoding/json"
	"os"
)

func ReadContract(path string) (Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, err
	}
	return DecodeContract(data)
}

func ReadManifest(path string) (SourceManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SourceManifest{}, err
	}
	return DecodeManifest(data)
}

func ReadEvidence(paths map[SourceName]string) (EvidenceSet, error) {
	var set EvidenceSet
	for _, source := range sourceOrder {
		data, err := os.ReadFile(paths[source])
		if err != nil {
			return EvidenceSet{}, err
		}
		switch source {
		case SourceUserJourney:
			set.UserJourney = data
		case SourceConformance:
			set.Conformance = data
		case SourceLSP:
			set.LSP = data
		case SourceRelease:
			set.Release = data
		case SourceExecution:
			set.Execution = data
		case SourceReadiness:
			set.Readiness = data
		}
	}
	return set, nil
}

func WriteReport(path string, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
