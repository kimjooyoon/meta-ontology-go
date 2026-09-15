package main

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"strings"
)

func provenanceErrorCode(err error) string {
	var conflict *provenance.ConflictError
	var chain *provenance.ChainError
	var claim *provenance.ClaimError
	var freshness *provenance.FreshnessError
	var corruption *provenance.CorruptionError
	switch {
	case errors.As(err, &conflict):
		return "provenance.conflict"
	case errors.As(err, &chain):
		return "provenance.chain-gap"
	case errors.As(err, &claim):
		return "provenance.claim-not-verified"
	case errors.As(err, &freshness):
		return "provenance.stale-source"
	case errors.As(err, &corruption):
		return "provenance.corruption"
	case strings.Contains(err.Error(), "decode evidence"), strings.Contains(err.Error(), "evidence is incomplete"):
		return "evidence.malformed"
	case strings.Contains(err.Error(), "different from authoritative"):
		return "evidence.binding"
	case strings.Contains(err.Error(), "read source"):
		return "source.read"
	case strings.Contains(err.Error(), "parse source"), strings.Contains(err.Error(), "source diagnostics"):
		return "source.parse"
	case strings.Contains(err.Error(), "lower source"):
		return "source.semantic"
	case strings.Contains(err.Error(), "read evidence"):
		return "evidence.read"
	case strings.Contains(err.Error(), "canonical provenance reread"):
		return "provenance.reread"
	default:
		return "provenance.rejected"
	}
}
