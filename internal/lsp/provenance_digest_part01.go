package lsp

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

// knownLSPProvenanceDigest accepts the two representations used at the LSP
// boundary: cache digests are bare lowercase hex, while document cache keys
// carry their explicit sha256 scheme prefix. Both represent the same digest;
// neither representation grants authorization.
func knownLSPProvenanceDigest(value string) bool {
	const scheme = "sha256:"
	if strings.HasPrefix(value, scheme) {
		value = strings.TrimPrefix(value, scheme)
	}
	return cache.Digest(value).Known()
}
