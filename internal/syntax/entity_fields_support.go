package syntax

import "errors"

type EntityFieldsState string

const (
	EntityFieldsDeferred  EntityFieldsState = "DEFERRED"
	EntityFieldsSupported EntityFieldsState = "SUPPORTED"
)

type EntityFieldsSupport struct {
	State   EntityFieldsState
	Profile EntityFieldsProfile
}

var (
	ErrEntityFieldsUnknownState       = errors.New("syntax: unknown entity fields support state")
	ErrEntityFieldsSupportUnavailable = errors.New("syntax: EntityFields SUPPORTED state is not implemented")
)

func (s EntityFieldsState) Valid() bool {
	return s == EntityFieldsDeferred || s == EntityFieldsSupported
}

func (s EntityFieldsSupport) Validate() error {
	switch s.State {
	case EntityFieldsDeferred:
		return s.Profile.Validate()
	case EntityFieldsSupported:
		return s.Profile.Validate()
	default:
		return ErrEntityFieldsUnknownState
	}
}

func CurrentEntityFieldsSupport() EntityFieldsSupport {
	return EntityFieldsSupport{
		State: EntityFieldsDeferred,
		Profile: EntityFieldsProfile{
			ID: EntityFieldsProfileID, Version: EntityFieldsProfileVersion, Digest: EntityFieldsProfileDigest,
		},
	}
}

// EntityFieldsV1Support returns the explicitly profile-bound V1 capability.
// The ordinary parser remains deferred until a caller opts into this exact
// contract, so older consumers cannot accidentally accept a partial shape.
func EntityFieldsV1Support() EntityFieldsSupport {
	support := CurrentEntityFieldsSupport()
	support.State = EntityFieldsSupported
	return support
}
