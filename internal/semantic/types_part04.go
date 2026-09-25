package semantic

import (
	"sort"
)

func normalizeAliases(raw []string, name string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(raw))
	aliases := make([]string, 0, len(raw))
	for _, value := range raw {
		alias, err := normalizeName(value)
		if err != nil {
			return nil, err
		}
		if alias == name {
			continue
		}
		if _, exists := seen[alias]; exists {
			continue
		}
		seen[alias] = struct{}{}
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	return aliases, nil
}
