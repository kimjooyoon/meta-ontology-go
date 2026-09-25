package closure

import (
	"fmt"
	"regexp"
	"strings"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var shaPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func NormalizeDigest(value string) (string, error) {
	value = strings.TrimSpace(value)
	if regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(value) {
		value = "sha256:" + value
	}
	if !digestPattern.MatchString(value) {
		return "", fmt.Errorf("invalid sha256 digest %q", value)
	}
	return value, nil
}

func normalizeInput(in Input) (Input, error) {
	digest, err := NormalizeDigest(in.Artifact.Digest)
	if err != nil {
		return in, err
	}
	in.Artifact.Digest = digest
	return in, validateIdentity(in)
}
