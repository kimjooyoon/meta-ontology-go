package lsp

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const DeclarationCompletionGuardObservationSchema = "gooo.lsp.declaration-completion-guard.v1"

type DeclarationCompletionGuardStatus string

const DeclarationCompletionGuardStatusMatched DeclarationCompletionGuardStatus = "MATCHED"
const DeclarationCompletionGuardStatusStale DeclarationCompletionGuardStatus = "STALE"
const DeclarationCompletionGuardStatusUnknown DeclarationCompletionGuardStatus = "UNKNOWN"

type DeclarationCompletionGuardObservation struct {
	Schema         string                           `json:"schema"`
	ProposedDigest string                           `json:"proposed_digest"`
	CurrentDigest  string                           `json:"current_digest"`
	Status         DeclarationCompletionGuardStatus `json:"status"`
	Reason         string                           `json:"reason"`
	NonAuthorizing bool                             `json:"non_authorizing"`
	Digest         string                           `json:"digest"`
}

// ObserveDeclarationCompletionGuard checks freshness before a client uses a
// completion candidate. MATCHED is a consistency result, not authorization.
func ObserveDeclarationCompletionGuard(
	proposed DeclarationCompletionContext,
	current DeclarationCompletionContext,
) DeclarationCompletionGuardObservation {
	proposedDigest := declarationCompletionContextDigest(proposed)
	currentDigest := declarationCompletionContextDigest(current)
	observation := DeclarationCompletionGuardObservation{
		Schema:         DeclarationCompletionGuardObservationSchema,
		ProposedDigest: proposedDigest,
		CurrentDigest:  currentDigest,
		Status:         DeclarationCompletionGuardStatusUnknown,
		Reason:         "COMPLETION_CONTEXT_UNKNOWN",
		NonAuthorizing: true,
	}
	switch {
	case proposed.DocumentURI == "" || current.DocumentURI == "":
		observation.Reason = "DOCUMENT_URI_MISSING"
	case proposed.DeclarationSymbol == "" || current.DeclarationSymbol == "":
		observation.Reason = "DECLARATION_SYMBOL_MISSING"
	case proposed.EnvironmentDigest == "" || current.EnvironmentDigest == "":
		observation.Reason = "ENVIRONMENT_DIGEST_MISSING"
	case proposed.DocumentURI != current.DocumentURI ||
		proposed.DocumentVersion != current.DocumentVersion ||
		proposed.DeclarationSymbol != current.DeclarationSymbol ||
		proposed.EnvironmentDigest != current.EnvironmentDigest:
		observation.Status = DeclarationCompletionGuardStatusStale
		observation.Reason = "COMPLETION_CONTEXT_STALE"
	default:
		observation.Status = DeclarationCompletionGuardStatusMatched
		observation.Reason = "COMPLETION_CONTEXT_MATCHED"
	}
	observation.Digest = declarationCompletionGuardObservationDigest(observation)
	return observation
}

func declarationCompletionContextDigest(context DeclarationCompletionContext) string {
	encoded, _ := json.Marshal(context)
	return cache.HashBytes(encoded).String()
}

func declarationCompletionGuardObservationDigest(observation DeclarationCompletionGuardObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
