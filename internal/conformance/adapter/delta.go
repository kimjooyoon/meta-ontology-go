package adapter

import (
	"fmt"
	"path"
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

func normalizeFacts(input []Fact) ([]Fact, error) {
	facts := append([]Fact(nil), input...)
	for _, fact := range facts {
		if strings.TrimSpace(fact.SubjectID) == "" || strings.TrimSpace(fact.Predicate) == "" ||
			strings.TrimSpace(fact.ObjectID) == "" || strings.TrimSpace(fact.Class) == "" {
			return nil, fmt.Errorf("facts require subject, predicate, object, and class")
		}
		if err := validateSourceURI(fact.SourceURI); err != nil {
			return nil, err
		}
		if err := validateRange(fact.Start, fact.End); err != nil {
			return nil, fmt.Errorf("fact %q: %w", fact.SubjectID, err)
		}
	}
	sort.Slice(facts, func(i, j int) bool {
		left, right := facts[i], facts[j]
		if left.SubjectID != right.SubjectID {
			return left.SubjectID < right.SubjectID
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.ObjectID != right.ObjectID {
			return left.ObjectID < right.ObjectID
		}
		return left.Class < right.Class
	})
	for i := 1; i < len(facts); i++ {
		if factsEqual(facts[i], facts[i-1]) {
			return nil, fmt.Errorf("duplicate fact %q", facts[i].SubjectID)
		}
	}
	return emptyFactsIfNil(facts), nil
}

func factsEqual(left, right Fact) bool {
	return left.SubjectID == right.SubjectID && left.Predicate == right.Predicate &&
		left.ObjectID == right.ObjectID && left.Class == right.Class && left.SourceURI == right.SourceURI &&
		left.Start == right.Start && left.End == right.End
}

func validateRange(start, end int) error {
	if start < 0 || end < 0 || end < start {
		return fmt.Errorf("invalid range %d:%d", start, end)
	}
	return nil
}

func validateSourceURI(sourceURI string) error {
	if sourceURI == "" {
		return nil
	}
	if strings.ContainsAny(sourceURI, "\\\x00") || path.IsAbs(sourceURI) {
		return fmt.Errorf("source_uri must be a safe relative path")
	}
	clean := path.Clean(sourceURI)
	if clean != sourceURI || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("source_uri must not escape the fixture root")
	}
	return nil
}

func hasDuplicateStrings(values []string) bool {
	for i := 1; i < len(values); i++ {
		if values[i] == values[i-1] {
			return true
		}
	}
	return false
}

func emptyStringsIfNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func emptyRegionsIfNil(values []Region) []Region {
	if values == nil {
		return []Region{}
	}
	return values
}

func emptySlotsIfNil(values []Slot) []Slot {
	if values == nil {
		return []Slot{}
	}
	return values
}

func emptyImportsIfNil(values []Import) []Import {
	if values == nil {
		return []Import{}
	}
	return values
}

func emptyMappingsIfNil(values []Mapping) []Mapping {
	if values == nil {
		return []Mapping{}
	}
	return values
}

func emptyFactsIfNil(values []Fact) []Fact {
	if values == nil {
		return []Fact{}
	}
	return values
}
