package lsp

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const SelfImprovementCycleResultProvenanceSchema = "gooo.lsp.self-improvement-cycle-result.v1"

const (
	SelfImprovementCycleResultProvenanceVisible = "VISIBLE"
	SelfImprovementCycleResultProvenanceUnknown = "UNKNOWN"
)

// SelfImprovementCycleResultProvenanceObservation is an editor-facing
// projection of a cycle result. It keeps the reverse observation visible but
// never turns it into an execution, merge, or deployment authorization.
type SelfImprovementCycleResultProvenanceObservation struct {
	Schema             string `json:"schema"`
	URI                string `json:"uri"`
	Version            int    `json:"version"`
	DeclarationSymbol  string `json:"declaration_symbol"`
	OriginDigest       string `json:"origin_digest"`
	EnvironmentDigest  string `json:"environment_digest"`
	CycleResultDigest  string `json:"cycle_result_digest"`
	CycleDigest        string `json:"cycle_digest"`
	OutcomeDigest      string `json:"outcome_digest"`
	Decision           string `json:"decision"`
	Reason             string `json:"reason"`
	EvidenceCount      int    `json:"evidence_count"`
	NonAuthorizing     bool   `json:"non_authorizing"`
	ObservationDigest  string `json:"observation_digest"`
}

// ObserveSelfImprovementCycleResultProvenance binds a cycle result to the
// declaration context and origin chain without making the editor surface
// executable or trusted.
func ObserveSelfImprovementCycleResultProvenance(
	context DeclarationCompletionContext,
	origin provenance.OriginChainObservation,
	result valueexecution.SelfImprovementCycleResultObservation,
) SelfImprovementCycleResultProvenanceObservation {
	originEncoded, _ := json.Marshal(origin)
	observation := SelfImprovementCycleResultProvenanceObservation{
		Schema:            SelfImprovementCycleResultProvenanceSchema,
		URI:               context.DocumentURI,
		Version:           context.DocumentVersion,
		DeclarationSymbol: context.DeclarationSymbol,
		OriginDigest:      cache.HashBytes(originEncoded).String(),
		EnvironmentDigest: context.EnvironmentDigest,
		CycleResultDigest: result.Digest,
		CycleDigest:       result.CycleDigest,
		OutcomeDigest:     result.OutcomeDigest,
		Decision:          SelfImprovementCycleResultProvenanceUnknown,
		Reason:            "CYCLE_RESULT_UNKNOWN",
		EvidenceCount:     result.EvidenceCount,
		NonAuthorizing:    true,
	}
	switch {
	case observation.URI == "":
		observation.Reason = "MISSING_DOCUMENT_URI"
	case observation.Version < 0:
		observation.Reason = "MISSING_DOCUMENT_VERSION"
	case observation.DeclarationSymbol == "":
		observation.Reason = "MISSING_DECLARATION_SYMBOL"
	case observation.EnvironmentDigest == "":
		observation.Reason = "MISSING_ENVIRONMENT_DIGEST"
	case reflect.DeepEqual(origin, provenance.OriginChainObservation{}):
		observation.Reason = "ORIGIN_OBSERVATION_MISSING"
	case !result.NonAuthorizing || !result.ReverseObservation:
		observation.Reason = "CYCLE_RESULT_AUTHORITY_INVALID"
	case result.Status == valueexecution.SelfImprovementCycleResultStatusUnknown:
		observation.Reason = "CYCLE_RESULT_UNKNOWN"
	case !cache.Digest(result.Digest).Known():
		observation.Reason = "CYCLE_RESULT_DIGEST_MISSING"
	default:
		observation.Decision = SelfImprovementCycleResultProvenanceVisible
		observation.Reason = "CYCLE_RESULT_STATUS_BOUND"
	}
	return finalizeSelfImprovementCycleResultProvenance(observation)
}

// ValidateSelfImprovementCycleResultProvenance verifies the immutable editor
// envelope and its non-authorizing boundary.
func ValidateSelfImprovementCycleResultProvenance(value SelfImprovementCycleResultProvenanceObservation) error {
	if value.Schema != SelfImprovementCycleResultProvenanceSchema || value.URI == "" || value.Version < 0 || value.DeclarationSymbol == "" || value.EnvironmentDigest == "" || !value.NonAuthorizing || value.Reason == "" {
		return errors.New("self-improvement cycle result provenance identity is invalid")
	}
	if value.Decision != SelfImprovementCycleResultProvenanceVisible && value.Decision != SelfImprovementCycleResultProvenanceUnknown {
		return errors.New("self-improvement cycle result provenance decision is invalid")
	}
	if value.Decision == SelfImprovementCycleResultProvenanceVisible && !cache.Digest(value.CycleResultDigest).Known() {
		return errors.New("visible self-improvement cycle result provenance is incomplete")
	}
	if !cache.Digest(value.ObservationDigest).Known() || value.ObservationDigest != selfImprovementCycleResultProvenanceDigest(value) {
		return errors.New("self-improvement cycle result provenance digest is invalid")
	}
	return nil
}

func finalizeSelfImprovementCycleResultProvenance(value SelfImprovementCycleResultProvenanceObservation) SelfImprovementCycleResultProvenanceObservation {
	value.ObservationDigest = selfImprovementCycleResultProvenanceDigest(value)
	return value
}

func selfImprovementCycleResultProvenanceDigest(value SelfImprovementCycleResultProvenanceObservation) string {
	value.ObservationDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}
