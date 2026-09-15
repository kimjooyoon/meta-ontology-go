package pathclosure

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func normalizeRequirement(raw Requirement) (Requirement, error) {
	pathID, err := semantic.ParseIdentity(raw.PathID.String())
	if err != nil {
		return Requirement{}, fmt.Errorf("path ID: %w", err)
	}
	start, err := semantic.ParseIdentity(raw.StartID.String())
	if err != nil {
		return Requirement{}, fmt.Errorf("start ID: %w", err)
	}
	end, err := semantic.ParseIdentity(raw.EndID.String())
	if err != nil {
		return Requirement{}, fmt.Errorf("end ID: %w", err)
	}
	if len(raw.RecordIDs) == 0 || len(raw.RecordIDs) != len(raw.ExpectedKinds) {
		return Requirement{}, errors.New("record and kind sequences must be non-empty and equal")
	}
	records := make([]semantic.ID, len(raw.RecordIDs))
	seen := make(map[semantic.ID]struct{}, len(records))
	for i, rawID := range raw.RecordIDs {
		recordID, parseErr := semantic.ParseIdentity(rawID.String())
		if parseErr != nil {
			return Requirement{}, fmt.Errorf("record ID: %w", parseErr)
		}
		if _, exists := seen[recordID]; exists {
			return Requirement{}, fmt.Errorf("duplicate record ID %s", recordID)
		}
		seen[recordID] = struct{}{}
		records[i] = recordID
		if !raw.ExpectedKinds[i].Valid() {
			return Requirement{}, fmt.Errorf("unknown inference kind %q", raw.ExpectedKinds[i])
		}
	}
	return Requirement{
		PathID: pathID, RecordIDs: records,
		ExpectedKinds: append([]semantic.InferenceKind(nil), raw.ExpectedKinds...),
		StartID:       start, EndID: end,
	}, nil
}
