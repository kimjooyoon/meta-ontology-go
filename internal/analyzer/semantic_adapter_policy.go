package analyzer

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// CurrentSemanticAdapterPolicy is a versioned, opt-in mapping contract.
const CurrentSemanticAdapterPolicy = "analyzer-semantic-adapter/v1"

var ErrSemanticAdapter = errors.New("semantic adapter rejected input")

// AdapterErrorCode identifies a fail-closed adapter rejection.
type AdapterErrorCode string

const (
	AdapterInvalidPolicy   AdapterErrorCode = "invalid-policy"
	AdapterUnknownRelation AdapterErrorCode = "unknown-relation"
	AdapterUnknownEndpoint AdapterErrorCode = "unknown-endpoint"
	AdapterEndpointKind    AdapterErrorCode = "endpoint-kind"
	AdapterEvidenceConfig  AdapterErrorCode = "evidence-config"
	AdapterSourceConfig    AdapterErrorCode = "source-config"
)

// AdapterError carries a stable class while keeping detail local to the
// rejected observation. Errors are returned before the caller's IR changes.
type AdapterError struct {
	Code        AdapterErrorCode     `json:"code"`
	Relation    Relation             `json:"relation"`
	Identity    string               `json:"identity"`
	Detail      string               `json:"detail"`
	WriteEffect ReconcileWriteEffect `json:"write_effect"`
}

func (e AdapterError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Detail)
}

func (e AdapterError) Unwrap() error { return ErrSemanticAdapter }

func adapterError(code AdapterErrorCode, relation Relation, identity, detail string) error {
	return AdapterError{Code: code, Relation: relation, Identity: identity, Detail: detail, WriteEffect: ReconcileNoWrite}
}

// RelationMapping is an explicit typed rule. Reverse is part of the rule and
// is never inferred from the analyzer-local relation name.
type RelationMapping struct {
	Source            Relation
	Predicate         semantic.Relation
	SourceSubjectKind semantic.Kind
	SourceObjectKind  semantic.Kind
	Reverse           bool
	AllowedOrigins    []ObservationOrigin
}

// MappingPolicy contains no default mappings. Every analyzer relation must be
// registered explicitly before it can produce a semantic fact.
type MappingPolicy struct {
	Revision string
	mappings map[Relation]RelationMapping
}

func NewMappingPolicy(revision string) (MappingPolicy, error) {
	revision = strings.TrimSpace(revision)
	if revision == "" {
		return MappingPolicy{}, adapterError(AdapterInvalidPolicy, "", "", "revision is required")
	}
	return MappingPolicy{Revision: revision, mappings: make(map[Relation]RelationMapping)}, nil
}

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

func (m RelationMapping) allowsOrigin(origin ObservationOrigin) bool {
	if len(m.AllowedOrigins) == 0 {
		return true
	}
	for _, allowed := range m.AllowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}

func (p MappingPolicy) lookup(relation Relation) (RelationMapping, bool) {
	mapping, ok := p.mappings[relation]
	return mapping, ok
}

func knownAnalyzerRelation(relation Relation) bool {
	switch relation {
	case RelationInvokes, RelationUses, RelationGenerates, RelationReferences:
		return true
	default:
		return false
	}
}
