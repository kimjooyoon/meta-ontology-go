package receiptprojection

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func collectReceiptEntities(file *syntax.File) (map[string]*syntax.EntityDecl, map[string]bool, error) {
	byName := map[string]*syntax.EntityDecl{}
	ids := map[string]bool{}
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			return nil, nil, fmt.Errorf("receipt schema profile accepts entity declarations only")
		}
		if _, duplicate := byName[entity.Name]; duplicate {
			return nil, nil, fmt.Errorf("duplicate entity name %q", entity.Name)
		}
		if err := claimID(ids, entity.ID); err != nil {
			return nil, nil, err
		}
		byName[entity.Name] = entity
	}
	return byName, ids, nil
}

func receiptSchema(byName map[string]*syntax.EntityDecl, root string) (string, error) {
	rootEntity, ok := byName[root]
	if !ok || !rootEntity.FieldsPresent {
		return "", fmt.Errorf("root receipt entity %q is absent or has no fields", root)
	}
	if !strings.HasPrefix(rootEntity.ID, "gooo://schema/") || rootEntity.ID == "gooo://schema/" {
		return "", fmt.Errorf("root receipt ID must use gooo://schema/")
	}
	boundSchema := false
	for _, field := range rootEntity.Fields {
		wireName := field.Name
		if wireName == strings.ToUpper(wireName) {
			wireName = strings.ToLower(wireName)
		}
		if wireName == "schema" {
			target, exists := byName[field.TypeRef.Spelling]
			boundSchema = exists && target.ID == typePrefix+"text" &&
				field.Presence == syntax.FieldPresenceRequired && field.Cardinality == syntax.FieldCardinalityOne
		}
	}
	if !boundSchema {
		return "", fmt.Errorf("root receipt requires a single required text schema field")
	}
	return "gooo/" + strings.TrimPrefix(rootEntity.ID, "gooo://schema/"), nil
}

func projectReceiptEntities(file *syntax.File, byName map[string]*syntax.EntityDecl,
	ids map[string]bool, root string) ([]Entity, error) {
	var result []Entity
	for _, declaration := range file.Declarations {
		entity := declaration.(*syntax.EntityDecl)
		if representation, scalar := scalarTypes[entity.ID]; scalar {
			if representation == "" || entity.FieldsPresent {
				return nil, fmt.Errorf("scalar entity %q has structural fields", entity.Name)
			}
			continue
		}
		if !entity.FieldsPresent || len(entity.Fields) == 0 || !exportedName(entity.Name) {
			return nil, fmt.Errorf("unsupported scalar or structural entity %q", entity.Name)
		}
		if entity.Name == root+"Schema" || entity.Name == "DeclarationSHA256" || entity.Name == "ProjectionProfile" {
			return nil, fmt.Errorf("entity name %q collides with projection metadata", entity.Name)
		}
		projected := Entity{ID: entity.ID, Name: entity.Name}
		names, goNames := map[string]bool{}, map[string]bool{}
		for _, field := range entity.Fields {
			projectedField, err := projectReceiptField(field, byName, ids, names, goNames)
			if err != nil {
				return nil, err
			}
			projected.Fields = append(projected.Fields, projectedField)
		}
		result = append(result, projected)
	}
	return result, nil
}
