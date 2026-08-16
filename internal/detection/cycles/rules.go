package cycles

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

type relationRule struct {
	subject Kind
	object  Kind
}

var relationRules = map[Relation]relationRule{
	Used:              {subject: Activity, object: Entity},
	WasGeneratedBy:    {subject: Entity, object: Activity},
	WasDerivedFrom:    {subject: Entity, object: Entity},
	WasAssociatedWith: {subject: Activity, object: Agent},
	WasInformedBy:     {subject: Activity, object: Activity},
	WasAttributedTo:   {subject: Entity, object: Agent},
	ActedOnBehalfOf:   {subject: Agent, object: Agent},
	SpecializationOf:  {subject: Entity, object: Entity},
	AlternateOf:       {subject: Entity, object: Entity},
}

func expectedKinds(predicate Relation) (Kind, Kind, bool) {
	rule, ok := relationRules[predicate]
	if !ok {
		return "", "", false
	}
	return rule.subject, rule.object, true
}

// ExpectedKinds returns the required subject and object roles for predicate.
// The boolean is false when predicate is not in the supported PROV vocabulary.
func ExpectedKinds(predicate Relation) (Kind, Kind, bool) {
	return expectedKinds(predicate)
}

// KnownRelation reports whether predicate has a direction rule.
func KnownRelation(predicate Relation) bool {
	_, _, ok := expectedKinds(predicate)
	return ok
}

func canonicalID(raw ID) (ID, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("stable ID is empty")
	}
	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("stable ID contains whitespace")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("stable ID is not a URI: %v", err)
	}
	if parsed.Scheme == "" {
		return "", fmt.Errorf("stable ID must have a URI scheme")
	}
	if strings.Contains(value, "://") && parsed.Host == "" {
		return "", fmt.Errorf("stable ID has an empty URI authority")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Host != "" {
		parsed.Host = strings.ToLower(parsed.Host)
	}
	canonical := parsed.String()
	if canonical == "" {
		return "", fmt.Errorf("stable ID is empty after normalization")
	}
	return canonical, nil
}

func normalizedName(raw string) string {
	return strings.Join(strings.Fields(raw), " ")
}

func normalizedNamespace(raw string) string {
	return strings.TrimSpace(raw)
}

func relationName(edge Edge) Relation {
	return edge.predicate()
}

func directionMessage(edge Edge, subject, object Node, known bool) string {
	predicate := relationName(edge)
	if !known {
		return fmt.Sprintf("unknown relation %q", predicate)
	}
	expectedSubject, expectedObject, _ := expectedKinds(predicate)
	return fmt.Sprintf("%s expects %s -> %s, got %s -> %s", predicate,
		expectedSubject, expectedObject, subject.Kind, object.Kind)
}
