package generator

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

func prepareEntityFields(ir SemanticIR, profile syntax.EntityFieldsProfile) (SemanticIR, error) {
	capabilities, err := profile.Capabilities()
	if err != nil {
		return SemanticIR{}, err
	}
	types, err := EntityFieldTypeNames(ir, profile)
	if err != nil {
		return SemanticIR{}, err
	}
	prepared := copyIR(ir)
	for entityIndex := range prepared.Entities {
		for fieldIndex := range prepared.Entities[entityIndex].Fields {
			field := &prepared.Entities[entityIndex].Fields[fieldIndex]
			field.GoName = field.Name
			base, allowed := types[field.TypeRefID]
			if !allowed {
				return SemanticIR{}, entityFieldsError(entityFieldsUnsupportedTypeDiagnostic, *field, "field type has no source-bound native name")
			}
			field.GoType = projectedEntityFieldType(*field, base, capabilities)
		}
	}
	return prepared, nil
}
