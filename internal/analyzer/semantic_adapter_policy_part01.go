package analyzer

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

// CurrentSemanticAdapterPolicy is a versioned, opt-in mapping contract.
const CurrentSemanticAdapterPolicy = "analyzer-semantic-adapter/v1"

var ErrSemanticAdapter = errors.New("semantic adapter rejected input")

// AdapterErrorCode identifies a fail-closed adapter rejection.
type AdapterErrorCode string

const (
	AdapterInvalidPolicy       AdapterErrorCode = "invalid-policy"
	AdapterUnknownRelation     AdapterErrorCode = "unknown-relation"
	AdapterUnknownEndpoint     AdapterErrorCode = "unknown-endpoint"
	AdapterEndpointKind        AdapterErrorCode = "endpoint-kind"
	AdapterEvidenceConfig      AdapterErrorCode = "evidence-config"
	AdapterSourceConfig        AdapterErrorCode = "source-config"
	AdapterAnalysisDiagnostics AdapterErrorCode = "analysis-diagnostics"
	AdapterSlotConfig          AdapterErrorCode = "slot-config"
	AdapterLocalityConfig      AdapterErrorCode = "locality-config"
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
