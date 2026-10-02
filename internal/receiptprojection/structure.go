package receiptprojection

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// StructureDigest identifies the source-owned projection independently of
// formatting offsets. SourceSHA256 continues to identify the exact declaration.
func (p Projection) StructureDigest() (string, error) {
	if err := p.validateBinding(); err != nil {
		return "", err
	}
	entities := make([]Entity, len(p.Entities))
	for i, entity := range p.Entities {
		entities[i] = entity
		entities[i].Fields = append([]Field(nil), entity.Fields...)
		for j := range entities[i].Fields {
			entities[i].Fields[j].Span = Field{}.Span
		}
	}
	view := struct {
		Profile, Package, Root, Schema string
		Entities                       []Entity
	}{Profile, p.Package, p.Root, p.Schema, entities}
	data, err := json.Marshal(view)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data)), nil
}
