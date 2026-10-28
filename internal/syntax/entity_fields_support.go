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

// EntityFieldsV2Support adds required single Boolean fields while retaining
// the V1 source ordering, identity, presence, and cardinality rules.
// Its digest is SHA256 of these UTF-8 lines, each ending in LF:
// id=gooo.entityfields.go-projection.v2
// types=urn:gooo:type:string,urn:gooo:type:boolean
// presence=required
// cardinality=one
// ordering=source
func EntityFieldsV2Support() EntityFieldsSupport {
	return EntityFieldsSupport{State: EntityFieldsSupported, Profile: EntityFieldsProfile{
		ID: EntityFieldsV2ProfileID, Version: EntityFieldsV2ProfileVersion, Digest: EntityFieldsV2ProfileDigest,
	}}
}

// EntityFieldsV3Support adds required single Integer fields while preserving
// the V2 Boolean and V1 string rules. Its digest binds the exact accepted types.
func EntityFieldsV3Support() EntityFieldsSupport {
	return EntityFieldsSupport{State: EntityFieldsSupported, Profile: EntityFieldsProfile{
		ID: EntityFieldsV3ProfileID, Version: EntityFieldsV3ProfileVersion, Digest: EntityFieldsV3ProfileDigest,
	}}
}

// EntityFieldsV4Support preserves V3 scalar types and adds optional single
// scalar fields. The Go projection represents an absent optional value as nil
// through a pointer to the declared scalar type.
func EntityFieldsV4Support() EntityFieldsSupport {
	return EntityFieldsSupport{State: EntityFieldsSupported, Profile: EntityFieldsProfile{
		ID: EntityFieldsV4ProfileID, Version: EntityFieldsV4ProfileVersion, Digest: EntityFieldsV4ProfileDigest,
	}}
}

// EntityFieldsV5Support represents ordered scalar collections. Required fields
// must be present; empty arrays are present values, and null is never a value.
// The profile binds a maximum of 256 elements per repeated field.
func EntityFieldsV5Support() EntityFieldsSupport {
	return EntityFieldsSupport{State: EntityFieldsSupported, Profile: EntityFieldsProfile{
		ID: EntityFieldsV5ProfileID, Version: EntityFieldsV5ProfileVersion, Digest: EntityFieldsV5ProfileDigest,
	}}
}

const RepeatedFieldLimit = 256

// EntityFieldsSupportForFile advances only for declared field capabilities.
// Source without collections or Text bounds retains the prior fingerprints.
func EntityFieldsSupportForFile(file *File, prior EntityFieldsSupport) EntityFieldsSupport {
	if file == nil || prior.State != EntityFieldsSupported || prior.Validate() != nil {
		return prior
	}
	if fileHasRecordReferences(file) {
		return EntityFieldsV7Support()
	}
	many := false
	for _, declaration := range file.Declarations {
		if entity, ok := declaration.(*EntityDecl); ok {
			for _, field := range entity.Fields {
				if field.TextBound.IsSet() {
					if prior.Profile.Version >= EntityFieldsV6ProfileVersion {
						return prior
					}
					return EntityFieldsV6Support()
				}
				if field.Cardinality == FieldCardinalityMany {
					many = true
				}
			}
		}
	}
	if many && prior.Profile.Version < EntityFieldsV5ProfileVersion {
		return EntityFieldsV5Support()
	}
	return prior
}
