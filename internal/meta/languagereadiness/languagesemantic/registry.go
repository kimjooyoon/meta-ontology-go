package languagesemantic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	expectedSources    = 13
	expectedLaws       = 3
	expectedRejections = 2
)

var knownLaws = map[string]struct{}{
	"PRESENTATION_INVARIANCE":   {},
	"CANDIDATE_NON_AUTHORITY":   {},
	"DETERMINISTIC_AUTHORITY":   {},
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

func (registry Registry) Validate() error {
	if registry.Schema != RegistrySchema {
		return fmt.Errorf("registry schema %q is not %q", registry.Schema, RegistrySchema)
	}
	if strings.TrimSpace(registry.Version) == "" {
		return fmt.Errorf("registry version is empty")
	}
	if len(registry.Cases) != FixedTotal {
		return fmt.Errorf("registry has %d cases, want %d", len(registry.Cases), FixedTotal)
	}
	seen := make(map[string]struct{}, len(registry.Cases))
	sources, laws, rejections := 0, 0, 0
	for _, definition := range registry.Cases {
		if strings.TrimSpace(definition.ID) == "" {
			return fmt.Errorf("registry contains an empty case id")
		}
		if _, exists := seen[definition.ID]; exists {
			return fmt.Errorf("registry contains duplicate case %q", definition.ID)
		}
		seen[definition.ID] = struct{}{}
		if definition.ProofChoice != "FOUNDATION" && definition.ProofChoice != "COHERENCE" && definition.ProofChoice != "REGRESSION" {
			return fmt.Errorf("case %q has invalid proof choice %q", definition.ID, definition.ProofChoice)
		}
		if strings.TrimSpace(definition.MetaOperation) == "" {
			return fmt.Errorf("case %q has no meta operation", definition.ID)
		}
		switch definition.Kind {
		case CaseSource:
			sources++
			clean := filepath.ToSlash(filepath.Clean(definition.Path))
			if clean == "." || strings.HasPrefix(clean, "../") || filepath.Ext(clean) != ".gooo" {
				return fmt.Errorf("source case %q has invalid path %q", definition.ID, definition.Path)
			}
		case CaseLaw:
			laws++
			if _, ok := knownLaws[definition.Law]; !ok {
				return fmt.Errorf("law case %q has unknown law %q", definition.ID, definition.Law)
			}
		case CaseUpstreamRejection:
			rejections++
			if definition.UpstreamCase == "" {
				return fmt.Errorf("upstream rejection %q has no upstream case", definition.ID)
			}
		default:
			return fmt.Errorf("case %q has unknown kind %q", definition.ID, definition.Kind)
		}
	}
	if sources != expectedSources || laws != expectedLaws || rejections != expectedRejections {
		return fmt.Errorf("registry topology is sources=%d laws=%d rejections=%d, want %d/%d/%d", sources, laws, rejections, expectedSources, expectedLaws, expectedRejections)
	}
	lawNames := make([]string, 0, len(knownLaws))
	for name := range knownLaws {
		lawNames = append(lawNames, name)
	}
	sort.Strings(lawNames)
	for _, name := range lawNames {
		count := 0
		for _, definition := range registry.Cases {
			if definition.Law == name {
				count++
			}
		}
		if count != 1 {
			return fmt.Errorf("registry law %q occurs %d times, want 1", name, count)
		}
	}
	return nil
}
