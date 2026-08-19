package analyzer

import (
	"sort"
	"strings"
)

// Register adds one explicit source-to-PROV rule and rejects duplicate source
// relations so a result cannot depend on registration order.
func (p *MappingPolicy) Register(mapping RelationMapping) error {
	if p == nil {
		return adapterError(AdapterInvalidPolicy, "", "", "policy is nil")
	}
	if err := p.validateMapping(mapping); err != nil {
		return err
	}
	if p.mappings == nil {
		p.mappings = make(map[Relation]RelationMapping)
	}
	if _, exists := p.mappings[mapping.Source]; exists {
		return adapterError(AdapterInvalidPolicy, mapping.Source, "", "relation is already registered")
	}
	p.mappings[mapping.Source] = mapping
	return nil
}
func (p MappingPolicy) Validate() error {
	if strings.TrimSpace(p.Revision) == "" {
		return adapterError(AdapterInvalidPolicy, "", "", "revision is required")
	}
	keys := make([]Relation, 0, len(p.mappings))
	for relation := range p.mappings {
		keys = append(keys, relation)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, relation := range keys {
		if err := p.validateMapping(p.mappings[relation]); err != nil {
			return err
		}
	}
	return nil
}
func (p MappingPolicy) validateMapping(mapping RelationMapping) error {
	if !knownAnalyzerRelation(mapping.Source) {
		return adapterError(AdapterUnknownRelation, mapping.Source, "", "mapping source is not registered vocabulary")
	}
	if !mapping.Predicate.Valid() {
		return adapterError(AdapterInvalidPolicy, mapping.Source, "", "predicate is outside the closed semantic vocabulary")
	}
	if !mapping.SourceSubjectKind.Valid() || !mapping.SourceObjectKind.Valid() {
		return adapterError(AdapterInvalidPolicy, mapping.Source, "", "endpoint kinds are invalid")
	}
	subject, object := mapping.SourceSubjectKind, mapping.SourceObjectKind
	if mapping.Reverse {
		subject, object = object, subject
	}
	if err := mapping.Predicate.ValidateKinds(subject, object); err != nil {
		return adapterError(AdapterInvalidPolicy, mapping.Source, "", err.Error())
	}
	for _, origin := range mapping.AllowedOrigins {
		if origin != OriginSignature && origin != OriginImplementation {
			return adapterError(AdapterInvalidPolicy, mapping.Source, "", "unknown observation origin")
		}
	}
	return nil
}
