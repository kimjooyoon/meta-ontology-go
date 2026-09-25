package languagesemantic

import (
	"encoding/json"
	"os"
)

var knownLaws = map[string]struct{}{
	"PRESENTATION_INVARIANCE": {},
	"CANDIDATE_NON_AUTHORITY": {},
	"DETERMINISTIC_AUTHORITY": {},
}

func LoadRegistry(path string) (Registry, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, nil, err
	}
	var registry Registry
	if err := json.Unmarshal(raw, &registry); err != nil {
		return Registry{}, nil, err
	}
	if err := registry.Validate(); err != nil {
		return Registry{}, nil, err
	}
	return registry, raw, nil
}
