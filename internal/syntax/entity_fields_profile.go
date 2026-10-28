package syntax

import "errors"

type EntityFieldsProfile struct {
	ID      string
	Version int
	Digest  string
}

const (
	EntityFieldsV7ProfileID      = "gooo.entityfields.go-projection.v7"
	EntityFieldsV7ProfileVersion = 7
	EntityFieldsV7ProfileDigest  = "db9377f341c13e9f510bf6d8902fe87b25a2b58d9020908285601729be53d7a1"
	EntityFieldsProfileID        = "gooo.entityfields.go-projection.v1"
	EntityFieldsProfileVersion   = 1
	EntityFieldsProfileDigest    = "7e93032618d1250cd4ff480eb7b5d6832f79bfc6921e6b9eea104151" +
		"db965ec0"
	EntityFieldsV2ProfileID      = "gooo.entityfields.go-projection.v2"
	EntityFieldsV2ProfileVersion = 2
	// V2 digest is SHA256 of the canonical lines documented in EntityFieldsV2Support.
	EntityFieldsV2ProfileDigest  = "d890716e88e65947968cee26e2254c4df2133c8ea420a5cb487afb2fa080ef4c"
	EntityFieldsV3ProfileID      = "gooo.entityfields.go-projection.v3"
	EntityFieldsV3ProfileVersion = 3
	EntityFieldsV3ProfileDigest  = "a41e79696fe46d18b0d2d74c0a687b9898397e71e340a9bcd432534fd6df476b"
	EntityFieldsV4ProfileID      = "gooo.entityfields.go-projection.v4"
	EntityFieldsV4ProfileVersion = 4
	EntityFieldsV4ProfileDigest  = "6afca055a3f1abaff71810e2dcb7db4ef1b98331415e0994066d358da42839bd"
	EntityFieldsV5ProfileID      = "gooo.entityfields.go-projection.v5"
	EntityFieldsV5ProfileVersion = 5
	EntityFieldsV5ProfileDigest  = "edac62212632b1f61390c95e624d0671c718b357cd5aa42ff0eef86358d90899"
	EntityFieldsV6ProfileID      = "gooo.entityfields.go-projection.v6"
	EntityFieldsV6ProfileVersion = 6
	EntityFieldsV6ProfileDigest  = "db163a3114f83e0ffd9b7e09cc812afae124e9613318ccfd6184c606c6ce6ee5"
)

var ErrEntityFieldsProfileMismatch = errors.New("syntax: EntityFields profile mismatch")

func (p EntityFieldsProfile) Validate() error {
	if p.IdentityVersionMatches() && p.Digest == p.canonicalProfile().Digest {
		return nil
	}
	return ErrEntityFieldsProfileMismatch
}

// IdentityVersionMatches distinguishes an unknown contract from a changed
// digest while keeping profile registration in this one kernel boundary.
func (p EntityFieldsProfile) IdentityVersionMatches() bool {
	known := p.canonicalProfile()
	return known.ID != "" && p.Version == known.Version
}

// EntityFieldsSupportForProfileID resolves an explicit kernel contract ID.
// It does not infer a profile from data or accept a caller-supplied digest.
func EntityFieldsSupportForProfileID(id string) (EntityFieldsSupport, bool) {
	profile := (EntityFieldsProfile{ID: id}).canonicalProfile()
	if profile.ID == "" {
		return EntityFieldsSupport{}, false
	}
	return EntityFieldsSupport{State: EntityFieldsSupported, Profile: profile}, true
}

func (p EntityFieldsProfile) canonicalProfile() EntityFieldsProfile {
	for _, profile := range supportedEntityFieldsProfiles {
		if p.ID == profile.ID {
			return profile
		}
	}
	return EntityFieldsProfile{}
}
