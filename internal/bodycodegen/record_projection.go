package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"go/types"
	"strings"
)

func activityRecords(all []RecordType, activityInputs []InputParameter, output, body string) []RecordType {
	used := map[string]bool{output: true}
	for _, input := range activityInputs {
		used[input.Type] = true
	}
	var scan scanner.Scanner
	fset := token.NewFileSet()
	scan.Init(fset.AddFile("record-body", -1, len(body)), []byte(body), nil, 0)
	previous := ""
	for {
		_, kind, spelling := scan.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.LBRACE && previous != "" {
			used[previous] = true
		}
		previous = ""
		if kind == token.IDENT {
			previous = spelling
		}
	}
	closeRecordTypeUses(all, used, false)
	var result []RecordType
	for _, record := range all {
		if used[record.Name] {
			result = append(result, record)
		}
	}
	return result
}

// RecordDeclarations emits source-bound value structs with explicit JSON names.
// Stable IDs determine the native names; source field order determines layout.
func RecordDeclarations(records []RecordType, native bool) string {
	var out strings.Builder
	for _, record := range records {
		name := record.Name
		if native {
			name = record.GoName
		}
		fmt.Fprintf(&out, "//gooo:generated:start id=%q kind=\"entity\"\ntype %s struct {\n", record.ID, name)
		for _, field := range record.Fields {
			name := field.Name
			if native {
				name = field.GoName
			}
			fieldType := recordFieldGoType(field.TypeID, records, native)
			if field.Cardinality == "many" {
				fieldType = "[]" + fieldType
			}
			jsonOptions := ""
			if field.Presence == "optional" {
				fieldType = "*" + fieldType
				jsonOptions = ",omitempty"
			}
			tags := fmt.Sprintf("json:%q", field.Name+jsonOptions)
			if field.TextBound.IsSet() {
				maximum, _ := field.TextBound.Resolve(0)
				tags += fmt.Sprintf(" gooo_max_utf8_bytes:%q", fmt.Sprint(maximum))
			}
			fmt.Fprintf(&out, "%s %s `%s`\n", name, fieldType, tags)
		}
		fmt.Fprintf(&out, "}\n//gooo:generated:end id=%q kind=\"entity\"\n", record.ID)
	}
	return out.String()
}

func recordGoType(typeID string) string {
	if typeID == "urn:gooo:type:boolean" {
		return "bool"
	}
	if typeID == "urn:gooo:type:integer" {
		return "int64"
	}
	return "string"
}

func lowerRecordNames(file *ast.File, information *types.Info, records []RecordType) error {
	objects := make(map[types.Object]string)
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, spec := range general.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				return fmt.Errorf("record declaration requires one named value type")
			}
			for _, record := range records {
				if record.Name == typeSpec.Name.Name {
					object := information.Defs[typeSpec.Name]
					objects[object] = record.GoName
					structure := object.Type().Underlying().(*types.Struct)
					for i, field := range record.Fields {
						objects[structure.Field(i)] = field.GoName
					}
				}
			}
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if name, ok := node.(*ast.Ident); ok {
			object := information.Uses[name]
			if object == nil {
				object = information.Defs[name]
			}
			if replacement := objects[object]; replacement != "" {
				name.Name = replacement
			}
		}
		return true
	})
	return nil
}

func restoreRecordNames(file *ast.File, records []RecordType) {
	names := make(map[string]string)
	for _, record := range records {
		names[record.GoName] = record.Name
		for _, field := range record.Fields {
			names[field.GoName] = field.Name
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if name, ok := node.(*ast.Ident); ok && names[name.Name] != "" {
			name.Name = names[name.Name]
		}
		return true
	})
}

func typecheckAndLowerRecords(packageName string, file *ast.File, fset *token.FileSet, records []RecordType, fieldOrigins map[ast.Node]BodyFieldOrigin, returnOrigins *bodyReturnCapture, origins ...map[ast.Expr]BodyConditionOrigin) error {
	if len(records) == 0 {
		return nil
	}
	information := &types.Info{Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object)}
	if _, err := checkBodyTypes(packageName, file, fset, information); err != nil {
		return fmt.Errorf("typecheck record bodies: %w", err)
	}
	if err := ValidateRecordUseBudget(information, records, false); err != nil {
		return err
	}
	for _, conditions := range origins {
		captureConditionFields(file, information, records, conditions)
	}
	captureTypedFieldOrigins(file, information, records, fieldOrigins)
	returnOrigins.captureTyped(file, information, records)
	if err := lowerRecordNames(file, information, records); err != nil {
		return err
	}
	return typecheck(packageName, file, fset)
}

func recordFieldGoType(typeID string, records []RecordType, native bool) string {
	for _, record := range records {
		if record.ID == typeID {
			if native {
				return record.GoName
			}
			return record.Name
		}
	}
	return recordGoType(typeID)
}
