package publictestreuse

import (
	"errors"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const ExecutionToolProfile = "gooo/generated-test-native-go-command/v1"

// ExecutionToolInputs describes the selected cmd/go and its frozen child
// environment. It does not establish transitive SDK, module or effect closure.
type ExecutionToolInputs struct {
	Profile           string `json:"profile"`
	Path              string `json:"path"`
	ExecutableDigest  string `json:"executable_digest"`
	Version           string `json:"version"`
	Platform          string `json:"platform"`
	EnvironmentDigest string `json:"environment_digest"`
}

var ErrUnobservedToolInputs = errors.New("historical receipt has no selected Go command observations")

func (inputs ExecutionToolInputs) Validate() error {
	if inputs.Profile != ExecutionToolProfile || inputs.Path == "" || len(inputs.Path) > 4096 ||
		strings.ContainsRune(inputs.Path, 0) || inputs.Version != "go1.27.2" ||
		inputs.Platform == "" || len(inputs.Platform) > 64 ||
		!cache.Digest(inputs.ExecutableDigest).Known() || !cache.Digest(inputs.EnvironmentDigest).Known() {
		return errors.New("selected Go command observations are incomplete")
	}
	return nil
}

func (inputs ExecutionToolInputs) Digest() string {
	digest, err := cache.DigestOf(inputs)
	if err != nil {
		return ""
	}
	return digest.String()
}
