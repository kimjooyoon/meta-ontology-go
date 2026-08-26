package adapter

import (
	"fmt"
	"sort"
	"strings"
)

func normalizeDelta(input *Delta) (*Delta, error) {
	if input == nil {
		return nil, nil
	}
	result := *input
	var err error
	if result.Added, err = normalizeFacts(result.Added); err != nil {
		return nil, fmt.Errorf("added facts: %w", err)
	}
	if result.Removed, err = normalizeFacts(result.Removed); err != nil {
		return nil, fmt.Errorf("removed facts: %w", err)
	}
	if result.Candidates, err = normalizeFacts(result.Candidates); err != nil {
		return nil, fmt.Errorf("candidate facts: %w", err)
	}
	result.Locality = append([]string(nil), result.Locality...)
	sort.Strings(result.Locality)
	if hasDuplicateStrings(result.Locality) {
		return nil, fmt.Errorf("duplicate locality id")
	}
	result.Locality = emptyStringsIfNil(result.Locality)
	result.Conflicts = append([]Conflict(nil), result.Conflicts...)
	for _, conflict := range result.Conflicts {
		if strings.TrimSpace(conflict.Code) == "" {
			return nil, fmt.Errorf("conflicts require code")
		}
	}
	sort.Slice(result.Conflicts, func(i, j int) bool {
		if result.Conflicts[i].Code != result.Conflicts[j].Code {
			return result.Conflicts[i].Code < result.Conflicts[j].Code
		}
		if result.Conflicts[i].SemanticID != result.Conflicts[j].SemanticID {
			return result.Conflicts[i].SemanticID < result.Conflicts[j].SemanticID
		}
		return result.Conflicts[i].Detail < result.Conflicts[j].Detail
	})
	return &result, nil
}
