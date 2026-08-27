package effects

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion/broker"
)

var ErrCapabilityTokenRequired = errors.New("capability broker token required")

// Request is the only effect API. The type system requires a broker.Token;
// callers cannot construct a valid token because its authority is private to
// broker.IssueToken.
func Request(token broker.Token, kind, target string) error {
	if !token.Valid() {
		return ErrCapabilityTokenRequired
	}
	capability := token.Capability()
	if capability.Kind != kind || capability.Target != target {
		return ErrCapabilityTokenRequired
	}
	return nil
}
