package adapter

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

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
