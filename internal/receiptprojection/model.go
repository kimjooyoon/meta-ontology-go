// Package receiptprojection compiles the bounded receipt schema profile from
// source-owned Gooo declarations. It does not activate EntityFields V1's public
// semantic profile, execute activities, call a model, or grant write authority.
package receiptprojection

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/token"
	"net/url"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const Profile = "gooo/receipt-structure-projection/v1"
const typePrefix = "gooo://receipt-structure/type/"

type Field struct {
	ID, Name, GoName, TypeID, GoType string
	Optional, Many                   bool
	Span                             syntax.Span
}
type Entity struct {
	ID, Name string
	Fields   []Field
}
type Projection struct {
	Package, Root, Schema, SourceSHA256 string
	Entities                            []Entity
	source                              []byte
	filename, originalRoot              string
}

var scalarTypes = map[string]string{
	typePrefix + "text": "string", typePrefix + "state": "string",
	typePrefix + "count": "int", typePrefix + "number": "float64",
	typePrefix + "record": "map[string]any", typePrefix + "counts": "map[string]int",
}

func Compile(filename string, source []byte, root string) (Projection, error) {
	if len(source) == 0 || len(source) > 64<<10 {
		return Projection{}, fmt.Errorf("receipt schema source must contain 1..65536 bytes")
	}
	support := syntax.CurrentEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, string(source), support)
	if file == nil || len(diagnostics) != 0 {
		return Projection{}, fmt.Errorf("receipt schema parse: %v", diagnostics)
	}
	if !exportedName(root) || file.Package.Name == "_" || !token.IsIdentifier(file.Package.Name) || token.Lookup(file.Package.Name).IsKeyword() {
		return Projection{}, fmt.Errorf("invalid Go package or root name")
	}
	if len(file.Bindings) != 0 {
		return Projection{}, fmt.Errorf("receipt schema profile does not execute bindings")
	}
	p := Projection{Package: file.Package.Name, Root: root, SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(source)), source: bytes.Clone(source), filename: filename, originalRoot: root}
	byName := map[string]*syntax.EntityDecl{}
	ids := map[string]bool{}
	for _, decl := range file.Declarations {
		entity, ok := decl.(*syntax.EntityDecl)
		if !ok {
			return Projection{}, fmt.Errorf("receipt schema profile accepts entity declarations only")
		}
		if _, ok := byName[entity.Name]; ok {
			return Projection{}, fmt.Errorf("duplicate entity name %q", entity.Name)
		}
		if err := claimID(ids, entity.ID); err != nil {
			return Projection{}, err
		}
		byName[entity.Name] = entity
	}
	rootEntity, ok := byName[root]
	if !ok || !rootEntity.FieldsPresent {
		return Projection{}, fmt.Errorf("root receipt entity %q is absent or has no fields", root)
	}
	if !strings.HasPrefix(rootEntity.ID, "gooo://schema/") || rootEntity.ID == "gooo://schema/" {
		return Projection{}, fmt.Errorf("root receipt ID must use gooo://schema/")
	}
	p.Schema = "gooo/" + strings.TrimPrefix(rootEntity.ID, "gooo://schema/")
	boundSchema := false
	for _, field := range rootEntity.Fields {
		wireName := field.Name
		if wireName == strings.ToUpper(wireName) {
			wireName = strings.ToLower(wireName)
		}
		if wireName != "schema" {
			continue
		}
		target, exists := byName[field.TypeRef.Spelling]
		boundSchema = exists && target.ID == typePrefix+"text" && field.Presence == syntax.FieldPresenceRequired && field.Cardinality == syntax.FieldCardinalityOne
	}
	if !boundSchema {
		return Projection{}, fmt.Errorf("root receipt requires a single required text schema field")
	}
	for _, decl := range file.Declarations {
		entity := decl.(*syntax.EntityDecl)
		if representation, scalar := scalarTypes[entity.ID]; scalar {
			if representation == "" || entity.FieldsPresent {
				return Projection{}, fmt.Errorf("scalar entity %q has structural fields", entity.Name)
			}
			continue
		}
		if !entity.FieldsPresent || len(entity.Fields) == 0 || !exportedName(entity.Name) {
			return Projection{}, fmt.Errorf("unsupported scalar or structural entity %q", entity.Name)
		}
		if entity.Name == root+"Schema" || entity.Name == "DeclarationSHA256" || entity.Name == "ProjectionProfile" {
			return Projection{}, fmt.Errorf("entity name %q collides with projection metadata", entity.Name)
		}
		e := Entity{ID: entity.ID, Name: entity.Name}
		names := map[string]bool{}
		goNames := map[string]bool{}
		for _, field := range entity.Fields {
			if err := claimID(ids, field.ID); err != nil {
				return Projection{}, err
			}
			wireName := field.Name
			if wireName == strings.ToUpper(wireName) {
				wireName = strings.ToLower(wireName)
			}
			name := goFieldName(wireName)
			if !exportedName(name) || names[wireName] || goNames[name] {
				return Projection{}, fmt.Errorf("invalid or colliding field name %q", field.Name)
			}
			names[wireName] = true
			goNames[name] = true
			target, ok := byName[field.TypeRef.Spelling]
			if !ok {
				return Projection{}, fmt.Errorf("unresolved field type %q", field.TypeRef.Spelling)
			}
			goType, scalar := scalarTypes[target.ID]
			if !scalar {
				if !target.FieldsPresent || !exportedName(target.Name) {
					return Projection{}, fmt.Errorf("unsupported type %q", target.Name)
				}
				goType = target.Name
			}
			if !field.Presence.Valid() || !field.Cardinality.Valid() {
				return Projection{}, fmt.Errorf("unsupported presence or cardinality")
			}
			optional := field.Presence == syntax.FieldPresenceOptional
			many := field.Cardinality == syntax.FieldCardinalityMany
			if optional && many {
				return Projection{}, fmt.Errorf("optional-many is outside receipt profile")
			}
			if !scalar && !optional && !many {
				return Projection{}, fmt.Errorf("structural fields require optional-one or required-many; unbounded value recursion is rejected")
			}
			if many {
				goType = "[]" + goType
			} else if optional {
				goType = "*" + goType
			}
			e.Fields = append(e.Fields, Field{ID: field.ID, Name: wireName, GoName: name, TypeID: target.ID, GoType: goType, Optional: optional, Many: many, Span: field.Span})
		}
		p.Entities = append(p.Entities, e)
	}
	return p, nil
}

func exportedName(name string) bool {
	return token.IsIdentifier(name) && len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}
func goFieldName(name string) string {
	var b strings.Builder
	for part := range strings.SplitSeq(name, "_") {
		if part == "" {
			return ""
		}
		switch part {
		case "id", "sha", "rss", "cpu":
			b.WriteString(strings.ToUpper(part))
		default:
			b.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}
	return b.String()
}
func claimID(ids map[string]bool, id string) error {
	u, err := url.Parse(id)
	if err != nil || u.Scheme == "" || strings.ContainsAny(id, " \t\r\n") || ids[id] {
		return fmt.Errorf("invalid or colliding stable ID %q", id)
	}
	ids[id] = true
	return nil
}
