package publictestreuse

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

const ExecutionToolProfile = toolchaincli.GeneratedTestGoProfile

// ExecutionToolInputs keeps the receipt's named type as well as its JSON shape.
// Canonical digests include the qualified type name; an alias would change it.
type ExecutionToolInputs toolchaincli.GeneratedTestGoInputs

func (inputs ExecutionToolInputs) Validate() error {
	return toolchaincli.GeneratedTestGoInputs(inputs).Validate()
}

func (inputs ExecutionToolInputs) Digest() string {
	digest, err := cache.DigestOf(inputs)
	if err != nil {
		return ""
	}
	return digest.String()
}

var ErrUnobservedToolInputs = errors.New("historical receipt has no selected Go command observations")
