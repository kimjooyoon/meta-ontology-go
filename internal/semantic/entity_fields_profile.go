package semantic

import (
	"errors"
	"fmt"
)

// EntityFieldsState is the support state supplied by an authoritative view.
// The semantic package accepts both contract states so semantic invariants
// remain closed when syntax support advances.
type EntityFieldsState string

const (
	EntityFieldsDeferred  EntityFieldsState = "DEFERRED"
	EntityFieldsSupported EntityFieldsState = "SUPPORTED"
)

// EntityFieldsProfile is the immutable profile tuple named by the V1 source
// contract. It is passed by value so validation never mutates the binding.
type EntityFieldsProfile struct {
	ID      string
	Version int
	Digest  string
}

const (
	EntityFieldsProfileID      = "gooo.entityfields.go-projection.v1"
	EntityFieldsProfileVersion = 1
	EntityFieldsProfileDigest  = "7e93032618d1250cd4ff480eb7b5d6832f79bfc6921e6b9eea104151db965ec0"
)

var (
	ErrEntityFieldsUnknownState       = errors.New("unknown EntityFields semantic support state")
	ErrEntityFieldsUnboundProfile     = errors.New("unbound EntityFields semantic profile")
	ErrEntityFieldsProfileMismatch    = errors.New("EntityFields semantic profile mismatch")
	ErrEntityFieldsProfileDigestError = errors.New("EntityFields semantic profile digest mismatch")
)

func CurrentEntityFieldsProfile() EntityFieldsProfile {
	return EntityFieldsProfile{
		ID: EntityFieldsProfileID, Version: EntityFieldsProfileVersion, Digest: EntityFieldsProfileDigest,
	}
}

func (p EntityFieldsProfile) Validate() error {
	if p.ID == "" && p.Version == 0 && p.Digest == "" {
		return ErrEntityFieldsUnboundProfile
	}
	if p.ID != EntityFieldsProfileID || p.Version != EntityFieldsProfileVersion {
		return fmt.Errorf("%w: id=%q version=%d", ErrEntityFieldsProfileMismatch, p.ID, p.Version)
	}
	if p.Digest != EntityFieldsProfileDigest {
		return fmt.Errorf("%w: %w", ErrEntityFieldsProfileMismatch, ErrEntityFieldsProfileDigestError)
	}
	return nil
}

// EntityFieldsBinding couples the support state to the exact profile tuple.
// Both DEFERRED and SUPPORTED are intentionally validated; neither state can
// skip semantic closure.
type EntityFieldsBinding struct {
	State   EntityFieldsState
	Profile EntityFieldsProfile
}

func (b EntityFieldsBinding) Validate() error {
	switch b.State {
	case EntityFieldsDeferred, EntityFieldsSupported:
		return b.Profile.Validate()
	default:
		return ErrEntityFieldsUnknownState
	}
}
