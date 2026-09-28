package analysisprovenance

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

// Digest returns the stable identity of one semantic-analysis provenance tuple.
// It intentionally lives below generation and LSP so both layers can bind the
// same identity without introducing a package cycle.
func Digest(sourceDigest, profileDigest, toolchainDigest, contractDigest string) string {
	return cache.HashBytes([]byte(strings.Join([]string{sourceDigest, profileDigest, toolchainDigest, contractDigest}, "\x00"))).String()
}
