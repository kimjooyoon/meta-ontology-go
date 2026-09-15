package verify

import (
	"fmt"
	"regexp"
	"strings"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var shaPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func normalizeInput(in Input) (Input, error) {
	value := strings.TrimSpace(in.Artifact.Digest)
	if regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(value) {
		value = "sha256:" + value
	}
	if !digestPattern.MatchString(value) {
		return in, fmt.Errorf("invalid artifact digest")
	}
	in.Artifact.Digest = value
	return in, nil
}
