package analysisprovenance

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

// SubjectDigest binds the stable source name to its observed bytes. It is an
// identity for provenance correlation, not an authorization or trust claim.
func SubjectDigest(sourcePath, sourceDigest string) string {
	return cache.HashBytes([]byte(strings.Join([]string{sourcePath, sourceDigest}, "\x00"))).String()
}
