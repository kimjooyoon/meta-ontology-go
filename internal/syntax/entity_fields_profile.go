package syntax

import "errors"

type EntityFieldsProfile struct {
	ID      string
	Version int
	Digest  string
}

const (
	EntityFieldsProfileID      = "gooo.entityfields.go-projection.v1"
	EntityFieldsProfileVersion = 1
	EntityFieldsProfileDigest  = "7e93032618d1250cd4ff480eb7b5d6832f79bfc6921e6b9eea104151" +
		"db965ec0"
	EntityFieldsV2ProfileID      = "gooo.entityfields.go-projection.v2"
	EntityFieldsV2ProfileVersion = 2
	// V2 digest is SHA256 of the canonical lines documented in EntityFieldsV2Support.
	EntityFieldsV2ProfileDigest = "d890716e88e65947968cee26e2254c4df2133c8ea420a5cb487afb2fa080ef4c"
)

var ErrEntityFieldsProfileMismatch = errors.New("syntax: EntityFields profile mismatch")

func (p EntityFieldsProfile) Validate() error {
	switch p.ID {
	case EntityFieldsProfileID:
		if p.Version == EntityFieldsProfileVersion && p.Digest == EntityFieldsProfileDigest {
			return nil
		}
	case EntityFieldsV2ProfileID:
		if p.Version == EntityFieldsV2ProfileVersion && p.Digest == EntityFieldsV2ProfileDigest {
			return nil
		}
	}
	return ErrEntityFieldsProfileMismatch
}
