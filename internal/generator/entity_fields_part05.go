package generator

func prepareEntityFields(ir SemanticIR) SemanticIR {
	prepared := copyIR(ir)
	for entityIndex := range prepared.Entities {
		for fieldIndex := range prepared.Entities[entityIndex].Fields {
			field := &prepared.Entities[entityIndex].Fields[fieldIndex]
			field.GoName = field.Name
			field.GoType = "string"
			if field.TypeRefID == entityFieldsBooleanTypeID {
				field.GoType = "bool"
			}
		}
	}
	return prepared
}
