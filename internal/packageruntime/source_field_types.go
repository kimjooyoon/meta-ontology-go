package packageruntime

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// FieldTypeReference retains the original file-local lookup and its source
// range while the temporary lowering view uses the existing scalar type ID.
type FieldTypeReference struct {
	OwnerID    string `json:"owner_id"`
	FieldID    string `json:"field_id"`
	FieldName  string `json:"field_name"`
	Spelling   string `json:"spelling"`
	ImportPath string `json:"import_path"`
	TypeID     string `json:"type_id"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
}

func scalarFieldTypeID(id string) bool {
	return id == string(semantic.BuiltinStringTypeID) || id == string(semantic.BuiltinBooleanTypeID) ||
		id == string(semantic.BuiltinIntegerTypeID)
}

func qualifiedFieldSpelling(spelling string) bool {
	return !strings.Contains(spelling, ":") && strings.Contains(spelling, ".")
}

func resolveSourceFieldTypes(spec PackageSpec, original *syntax.File, local map[string]string,
	dependencies map[string][]Export) (*syntax.File, []FieldTypeReference, error) {
	file := original.Clone()
	var references []FieldTypeReference
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			continue
		}
		for index := range entity.Fields {
			field := &entity.Fields[index]
			spelling := field.TypeRef.Spelling
			if !qualifiedFieldSpelling(spelling) {
				continue
			}
			if strings.Count(spelling, ".") != 1 {
				return nil, nil, reject("PACKAGE_FIELD_TYPE_REFERENCE_INVALID", "%s:%s field %q needs one import alias and member",
					spec.Path, field.TypeRef.Span.Filename, field.Name)
			}
			id, err := resolveSourceEntityType(spec, original, spelling, local, dependencies)
			if err != nil {
				return nil, nil, reject("PACKAGE_FIELD_TYPE_IMPORTED_INVALID", "%s:%s field %q type %q: %v",
					spec.Path, field.TypeRef.Span.Filename, field.Name, spelling, err)
			}
			if !scalarFieldTypeID(id) {
				return nil, nil, reject("PACKAGE_FIELD_TYPE_PROFILE_UNSUPPORTED", "%s:%s field %q type %q resolves to %q; V4 accepts single scalar fields",
					spec.Path, field.TypeRef.Span.Filename, field.Name, spelling, id)
			}
			imports, err := sourceImportAliases(spec, original)
			if err != nil {
				return nil, nil, err
			}
			alias, _, _ := strings.Cut(spelling, ".")
			references = append(references, FieldTypeReference{OwnerID: entity.ID, FieldID: field.ID,
				FieldName: field.Name, Spelling: spelling, ImportPath: imports[alias], TypeID: id,
				Start: field.TypeRef.Span.Start.Offset, End: field.TypeRef.Span.End.Offset})
			field.TypeRef.Spelling = id
		}
	}
	return file, references, nil
}

// ParseResolvedPackageSource uses the original source and the field references
// produced by Build. Original bytes and source spelling remain in SourceImage;
// this parsed view is temporary input to workspace lowering and construction.
func ParseResolvedPackageSource(image Image, packagePath string, source Source) (*syntax.File, error) {
	var matched *SourceImage
	for _, pkg := range image.Packages {
		if pkg.Path != packagePath {
			continue
		}
		for index := range pkg.Sources {
			if pkg.Sources[index].Filename == source.Filename {
				if matched != nil {
					return nil, fmt.Errorf("package source %q has duplicate type evidence", source.Filename)
				}
				matched = &pkg.Sources[index]
			}
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("package source %q has absent type evidence at image.packages[].sources[].source_digest", source.Filename)
	}
	if matched.SourceDigest != digestValue(source.Content) {
		return nil, fmt.Errorf("package source %q has stale type evidence at image.packages[].sources[].source_digest: expected sha256(json.Marshal(source.Content)); compare JSON string bytes, not raw file bytes", source.Filename)
	}
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(source.Filename, source.Content, syntax.EntityFieldsV4Support())
	if file == nil || diagnostics.HasErrors() {
		return nil, fmt.Errorf("workspace source %q: %s", source.Filename, diagnostics.Error())
	}
	if err := applySourceFieldTypes(file, matched.FieldTypes); err != nil {
		return nil, fmt.Errorf("package source %q: %w", source.Filename, err)
	}
	return file, nil
}

func applySourceFieldTypes(file *syntax.File, references []FieldTypeReference) error {
	byField := map[string]FieldTypeReference{}
	for _, ref := range references {
		key := ref.OwnerID + "\x00" + ref.FieldID
		if _, exists := byField[key]; exists || !scalarFieldTypeID(ref.TypeID) {
			return fmt.Errorf("duplicate or unsupported resolved field type %q", ref.FieldID)
		}
		byField[key] = ref
	}
	used := 0
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			continue
		}
		for index := range entity.Fields {
			field := &entity.Fields[index]
			if !qualifiedFieldSpelling(field.TypeRef.Spelling) {
				continue
			}
			ref, ok := byField[entity.ID+"\x00"+field.ID]
			if !ok || ref.Spelling != field.TypeRef.Spelling || ref.FieldName != field.Name ||
				ref.Start != field.TypeRef.Span.Start.Offset || ref.End != field.TypeRef.Span.End.Offset {
				return fmt.Errorf("qualified field %q has absent or stale reference evidence", field.ID)
			}
			field.TypeRef.Spelling = ref.TypeID
			used++
		}
	}
	if used != len(references) {
		return fmt.Errorf("resolved field type evidence contains unused references")
	}
	return nil
}
