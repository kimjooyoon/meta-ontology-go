package bodycodegen

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ScalarEntityKind binds native value representation to a recognized type ID.
// Existing nominal Integer/Boolean/Text declarations keep their legacy profile.
func ScalarEntityKind(name, id string) string {
	if identity, err := semantic.ParseIdentity(id); err == nil {
		id = identity.String()
	}
	switch id {
	case string(semantic.BuiltinIntegerTypeID):
		return "Integer"
	case string(semantic.BuiltinBooleanTypeID):
		return "Boolean"
	case string(semantic.BuiltinStringTypeID):
		return "Text"
	}
	if _, ok := goTypeForEntity(name); ok {
		return name
	}
	return ""
}

func sourceScalarKind(file *syntax.File, name string) string {
	if file != nil {
		for _, declaration := range file.Declarations {
			entity, ok := declaration.(*syntax.EntityDecl)
			if !ok || entity.Name != name {
				continue
			}
			if entity.FieldsPresent || len(entity.Fields) != 0 {
				return ""
			}
			return ScalarEntityKind(name, entity.ID)
		}
	}
	return ScalarEntityKind(name, "")
}

// ScalarKindsFromSource preserves authored names while describing value kinds.
// An explicit record block supplies no scalar representation, including {}.
func ScalarKindsFromSource(file *syntax.File) map[string]string {
	result := make(map[string]string)
	if file == nil {
		return result
	}
	for _, declaration := range file.Declarations {
		if entity, ok := declaration.(*syntax.EntityDecl); ok {
			result[entity.Name] = sourceScalarKind(file, entity.Name)
		}
	}
	return result
}
