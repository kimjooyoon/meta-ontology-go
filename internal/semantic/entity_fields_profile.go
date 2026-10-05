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
	EntityFieldsProfileID        = "gooo.entityfields.go-projection.v1"
	EntityFieldsProfileVersion   = 1
	EntityFieldsProfileDigest    = "7e93032618d1250cd4ff480eb7b5d6832f79bfc6921e6b9eea104151db965ec0"
	EntityFieldsV2ProfileID      = "gooo.entityfields.go-projection.v2"
	EntityFieldsV2ProfileVersion = 2
	EntityFieldsV2ProfileDigest  = "d890716e88e65947968cee26e2254c4df2133c8ea420a5cb487afb2fa080ef4c"
	EntityFieldsV3ProfileID      = "gooo.entityfields.go-projection.v3"
	EntityFieldsV3ProfileVersion = 3
	EntityFieldsV3ProfileDigest  = "a41e79696fe46d18b0d2d74c0a687b9898397e71e340a9bcd432534fd6df476b"
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
	switch p.ID {
	case EntityFieldsProfileID:
		if p.Version == EntityFieldsProfileVersion && p.Digest == EntityFieldsProfileDigest {
			return nil
		}
	case EntityFieldsV2ProfileID:
		if p.Version == EntityFieldsV2ProfileVersion && p.Digest == EntityFieldsV2ProfileDigest {
			return nil
		}
	case EntityFieldsV3ProfileID:
		if p.Version == EntityFieldsV3ProfileVersion && p.Digest == EntityFieldsV3ProfileDigest {
			return nil
		}
	}
	if p.Digest != EntityFieldsProfileDigest && p.Digest != EntityFieldsV2ProfileDigest && p.Digest != EntityFieldsV3ProfileDigest {
		return fmt.Errorf("%w: %w", ErrEntityFieldsProfileMismatch, ErrEntityFieldsProfileDigestError)
	}
	return fmt.Errorf("%w: id=%q version=%d", ErrEntityFieldsProfileMismatch, p.ID, p.Version)
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
