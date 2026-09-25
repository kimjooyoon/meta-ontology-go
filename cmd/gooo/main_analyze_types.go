package main

import (
	"runtime"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
)

var analyzeDeltaToolchain = runtime.Version() + "|" + runtime.GOOS + "/" + runtime.GOARCH

type analyzeDeltaOptions struct {
	authority string
	goFiles   []string
}
type analyzeDeltaOutput struct {
	analyzer.SemanticNormalizedDelta
	AuthoritySemanticDigest string                        `json:"authority_semantic_digest"`
	ObservedSemanticDigest  string                        `json:"observed_semantic_digest"`
	SemanticEqual           bool                          `json:"semantic_equal"`
	WriteEffect             analyzer.ReconcileWriteEffect `json:"write_effect"`
}
type analyzeGeneratedRegion struct{ id string }
type analyzeMarkerAlias struct{ id, name string }
