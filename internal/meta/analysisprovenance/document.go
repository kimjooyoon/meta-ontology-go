package analysisprovenance

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

// DocumentDigest binds source bytes to the semantic lowering and its execution
// contract. It is an observation identity, not an authorization decision.
func DocumentDigest(sourceDigest, semanticDigest, profileDigest, toolchainDigest, contractDigest string) string {
	return cache.HashBytes([]byte(strings.Join([]string{
		sourceDigest, semanticDigest, profileDigest, toolchainDigest, contractDigest,
	}, "\x00"))).String()
}

// DocumentDigestWithSymbolMap binds a document's analysis inputs to the exact
// source-origin projection exposed by an editor integration.
func DocumentDigestWithSymbolMap(sourceDigest, semanticDigest, profileDigest, toolchainDigest, contractDigest, symbolMapDigest string) string {
	return DocumentDigestWithSymbolMapAndReferences(sourceDigest, semanticDigest, profileDigest, toolchainDigest, contractDigest, symbolMapDigest, "")
}

// DocumentDigestWithSymbolMapAndReferences extends the editor provenance
// identity with the deterministic reverse-observation projection of references.
// It remains an observation identity, not an authorization decision.
func DocumentDigestWithSymbolMapAndReferences(sourceDigest, semanticDigest, profileDigest, toolchainDigest, contractDigest, symbolMapDigest, referenceMapDigest string) string {
	return cache.HashBytes([]byte(strings.Join([]string{
		sourceDigest, semanticDigest, profileDigest, toolchainDigest, contractDigest, symbolMapDigest, referenceMapDigest,
	}, "\x00"))).String()
}
