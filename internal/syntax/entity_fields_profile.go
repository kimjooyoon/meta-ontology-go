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
)

var ErrEntityFieldsProfileMismatch = errors.New("syntax: EntityFields profile mismatch")

func (p EntityFieldsProfile) Validate() error {
	want := EntityFieldsProfile{
		ID: EntityFieldsProfileID, Version: EntityFieldsProfileVersion, Digest: EntityFieldsProfileDigest,
	}
	if p != want {
		return ErrEntityFieldsProfileMismatch
	}
	return nil
}
