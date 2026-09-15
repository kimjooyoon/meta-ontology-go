package generator

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func prepareEntityFields(ir SemanticIR) SemanticIR {
	prepared := copyIR(ir)
	for entityIndex := range prepared.Entities {
		for fieldIndex := range prepared.Entities[entityIndex].Fields {
			field := &prepared.Entities[entityIndex].Fields[fieldIndex]
			field.GoName = field.Name
			field.GoType = "string"
		}
	}
	return prepared
}
func entityFieldsProfileMapping() syntax.EntityFieldsProfile {
	return syntax.CurrentEntityFieldsSupport().Profile
}
