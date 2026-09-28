package lsp

import (
	"encoding/json"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

const DeclarationCompletionObservationSchema = "gooo.lsp.declaration-completion-observation.v1"

type DeclarationCompletionStatus string

const DeclarationCompletionStatusComplete DeclarationCompletionStatus = "COMPLETE"
const DeclarationCompletionStatusUnknown DeclarationCompletionStatus = "UNKNOWN"

type DeclarationCompletionContext struct {
	DocumentURI       string `json:"document_uri"`
	DocumentVersion   int    `json:"document_version"`
	DeclarationSymbol string `json:"declaration_symbol"`
	EnvironmentDigest string `json:"environment_digest"`
}

type DeclarationCompletionObservation struct {
	Schema            string                      `json:"schema"`
	DocumentURI       string                      `json:"document_uri"`
	DocumentVersion   int                         `json:"document_version"`
	DeclarationSymbol string                      `json:"declaration_symbol"`
	Label             string                      `json:"label"`
	Detail            string                      `json:"detail"`
	OriginDigest      string                      `json:"origin_digest"`
	EnvironmentDigest string                      `json:"environment_digest"`
	Status            DeclarationCompletionStatus `json:"status"`
	Reason            string                      `json:"reason"`
	NonAuthorizing    bool                        `json:"non_authorizing"`
	Digest            string                      `json:"digest"`
}

// ObserveDeclarationCompletion binds an LSP completion candidate to the
// declaration and provenance context without making it executable or trusted.
func ObserveDeclarationCompletion(
	context DeclarationCompletionContext,
	label string,
	detail string,
	origin provenance.OriginChainObservation,
) DeclarationCompletionObservation {
	originEncoded, _ := json.Marshal(origin)
	observation := DeclarationCompletionObservation{
		Schema:            DeclarationCompletionObservationSchema,
		DocumentURI:       context.DocumentURI,
		DocumentVersion:   context.DocumentVersion,
		DeclarationSymbol: context.DeclarationSymbol,
		Label:             label,
		Detail:            detail,
		OriginDigest:      cache.HashBytes(originEncoded).String(),
		EnvironmentDigest: context.EnvironmentDigest,
		Status:            DeclarationCompletionStatusUnknown,
		Reason:            "ORIGIN_OBSERVATION_MISSING",
		NonAuthorizing:    true,
	}
	switch {
	case context.DocumentURI == "" || context.DeclarationSymbol == "":
		observation.Reason = "DECLARATION_CONTEXT_MISSING"
	case context.EnvironmentDigest == "":
		observation.Reason = "ENVIRONMENT_DIGEST_MISSING"
	case reflect.DeepEqual(origin, provenance.OriginChainObservation{}):
		observation.Reason = "ORIGIN_OBSERVATION_MISSING"
	default:
		observation.Status = DeclarationCompletionStatusComplete
		observation.Reason = "DECLARATION_ORIGIN_BOUND"
	}
	observation.Digest = declarationCompletionObservationDigest(observation)
	return observation
}

func declarationCompletionObservationDigest(observation DeclarationCompletionObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
