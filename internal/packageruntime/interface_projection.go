package packageruntime

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func projectSourceInterface(source parsedPackageSource, ir semantic.IR) []InterfaceDeclaration {
	nodes := make(map[string]semantic.Node)
	for _, node := range ir.Graph.Nodes() {
		nodes[node.Name] = node
	}
	var result []InterfaceDeclaration
	for _, declaration := range source.file.Declarations {
		var name, kind, shape string
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			name, kind, shape = value.Name, "entity", "nominal"
			if value.FieldsPresent {
				shape = "record"
			}
		case *syntax.ActivityDecl:
			name, kind = value.Name, "activity"
		default:
			continue
		}
		node := nodes[name]
		item := InterfaceDeclaration{Kind: kind, Name: name, ID: node.ID.String(), Source: source.source.Filename, Shape: shape}
		for _, field := range node.Fields {
			item.Fields = append(item.Fields, InterfaceField{
				Name: field.Name, ID: field.ID.String(), Aliases: append([]string(nil), field.Aliases...),
				TypeID: field.TypeRef.ID.String(), Presence: string(field.Presence), Cardinality: string(field.Cardinality),
			})
		}
		result = append(result, item)
	}
	return result
}

func (compiled compiledPackage) publicInterface() (InterfacePackage, error) {
	image := compiled.image
	result := InterfacePackage{Path: image.Path, Name: image.Name, Namespace: image.Namespace,
		Imports: image.Imports, Sources: image.Sources, Declarations: compiled.declarations}
	exports := make(map[string]Export, len(image.Exports))
	for _, export := range image.Exports {
		exports[export.Name] = export
	}
	identities := make(map[string]bool)
	for index := range result.Declarations {
		item := &result.Declarations[index]
		if item.ID == "" || identities[item.ID] {
			return InterfacePackage{}, reject("PACKAGE_INTERFACE_ID_INVALID", "%s:%s has a missing or repeated identity", image.Path, item.Name)
		}
		identities[item.ID] = true
		if item.Kind == "activity" {
			export := exports[item.Name]
			item.Inputs, item.Output = append([]string(nil), export.InputTypes...), export.OutputType
		}
	}
	for _, item := range result.Declarations {
		for _, field := range item.Fields {
			if field.ID == "" || identities[field.ID] || field.TypeID == "" {
				return InterfacePackage{}, reject("PACKAGE_INTERFACE_FIELD_INVALID", "%s:%s has a missing or repeated field identity/type", image.Path, item.Name)
			}
			identities[field.ID] = true
		}
	}
	sort.Slice(result.Declarations, func(i, j int) bool { return result.Declarations[i].ID < result.Declarations[j].ID })
	return result, nil
}
