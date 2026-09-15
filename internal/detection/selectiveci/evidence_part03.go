package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func pathRequirement(raw PathRequirement) (pathclosure.Requirement, error) {
	pathID, err := semantic.ParseIdentity(raw.PathID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	start, err := semantic.ParseIdentity(raw.StartID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	end, err := semantic.ParseIdentity(raw.EndID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	records := make([]semantic.ID, len(raw.RecordIDs))
	kinds := make([]semantic.InferenceKind, len(raw.ExpectedKinds))
	for i := range raw.RecordIDs {
		records[i], err = semantic.ParseIdentity(raw.RecordIDs[i])
		if err != nil {
			return pathclosure.Requirement{}, err
		}
		kinds[i] = semantic.InferenceKind(raw.ExpectedKinds[i])
	}
	return pathclosure.Requirement{PathID: pathID, RecordIDs: records, ExpectedKinds: kinds, StartID: start, EndID: end}, nil
}
