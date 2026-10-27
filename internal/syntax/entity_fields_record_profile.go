package syntax

// RecordReferenceDepthLimit bounds acyclic record layouts independently of
// selected record count, fields per layout, and elements per repeated field.
const RecordReferenceDepthLimit = 8
const RecordValueSlotLimit = 8192

// EntityFieldsV7Support explicitly enables source-declared record references.
// Its digest binds these canonical UTF-8 lines, each ending in LF:
// id=gooo.entityfields.go-projection.v7
// types=urn:gooo:type:string,urn:gooo:type:boolean,urn:gooo:type:integer,declared-record
// presence=required,optional
// cardinality=one,many
// ordering=source
// optional-representation=pointer
// repeated-limit=256
// text-bound=max_utf8_bytes
// text-bound-unit=decoded-utf8-bytes
// text-bound-range=0..4294967295
// text-bound-absent=legacy
// record-references=stable-id
// record-graph=acyclic
// record-depth-limit=8
// record-value-slot-limit=8192
func EntityFieldsV7Support() EntityFieldsSupport {
	return EntityFieldsSupport{State: EntityFieldsSupported, Profile: EntityFieldsProfile{
		ID: EntityFieldsV7ProfileID, Version: EntityFieldsV7ProfileVersion, Digest: EntityFieldsV7ProfileDigest,
	}}
}

func fileHasRecordReferences(file *File) bool {
	records := make(map[string]bool)
	for _, declaration := range file.Declarations {
		if entity, ok := declaration.(*EntityDecl); ok && len(entity.Fields) != 0 {
			records[entity.Name] = true
			if entity.ID != "" {
				records[entity.ID] = true
			}
			if file.Namespace != nil {
				records[file.Namespace.Name+"."+entity.Name] = true
			}
		}
	}
	for _, declaration := range file.Declarations {
		if entity, ok := declaration.(*EntityDecl); ok {
			for _, field := range entity.Fields {
				if records[field.TypeRef.Spelling] {
					return true
				}
			}
		}
	}
	return false
}
