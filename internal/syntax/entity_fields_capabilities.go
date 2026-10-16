package syntax

// EntityFieldsCapabilities is derived only from an exact validated profile.
// A large version number or a familiar ID with a different digest grants none.
type EntityFieldsCapabilities struct {
	Boolean   bool
	Integer   bool
	Optional  bool
	Many      bool
	TextBound bool
}

func (p EntityFieldsProfile) Capabilities() (EntityFieldsCapabilities, error) {
	if err := p.Validate(); err != nil {
		return EntityFieldsCapabilities{}, err
	}
	return EntityFieldsCapabilities{
		Boolean: p.Version >= 2, Integer: p.Version >= 3, Optional: p.Version >= 4,
		Many: p.Version >= 5, TextBound: p.Version >= 6,
	}, nil
}

// EntityFieldsV6Support adds explicit decoded Text field limits to V5. The
// profile's canonical UTF-8 lines, each ending in LF, are:
// id=gooo.entityfields.go-projection.v6
// types=urn:gooo:type:string,urn:gooo:type:boolean,urn:gooo:type:integer
// presence=required,optional
// cardinality=one,many
// ordering=source
// optional-representation=pointer
// repeated-limit=256
// text-bound=max_utf8_bytes
// text-bound-unit=decoded-utf8-bytes
// text-bound-range=0..4294967295
// text-bound-absent=legacy
func EntityFieldsV6Support() EntityFieldsSupport {
	return EntityFieldsSupport{State: EntityFieldsSupported, Profile: EntityFieldsProfile{
		ID: EntityFieldsV6ProfileID, Version: EntityFieldsV6ProfileVersion, Digest: EntityFieldsV6ProfileDigest,
	}}
}
