package bodycodegen

import (
	"go/ast"
	"go/types"
)

func (p bodyProjection) captureFieldOrigins(activityID, body string) map[ast.Node]BodyFieldOrigin {
	bodies := map[string]string{p.function.Name.Name: body}
	ids := map[string]string{p.function.Name.Name: activityID}
	for _, call := range p.calls {
		bodies[call.identity.Name], ids[call.identity.Name] = call.body, call.identity.ActivityID
	}
	result := map[ast.Node]BodyFieldOrigin{}
	for _, decl := range p.file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		capture := localOriginCapture{projection: p, text: bodies[fn.Name.Name], base: p.fset.Position(fn.Body.Lbrace).Offset + 2}
		WalkFieldSyntax(fn.Body, func(slot FieldSyntax) {
			value, valid := capture.span(slot.Value)
			target, targetValid := capture.span(slot.Target)
			if valid && targetValid {
				result[slot.Node] = BodyFieldOrigin{ActivityID: ids[fn.Name.Name], Kind: slot.Kind, BodySHA256: digest([]byte(capture.text)), Value: value, Target: target}
			}
		})
	}
	return result
}

func captureTypedFieldOrigins(file *ast.File, info *types.Info, records []RecordType, origins map[ast.Node]BodyFieldOrigin) {
	fields := map[types.Object][2]string{}
	for _, decl := range file.Decls {
		general, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range general.Specs {
			named, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			object := info.Defs[named.Name]
			if object == nil {
				continue
			}
			shape, ok := object.Type().Underlying().(*types.Struct)
			if !ok {
				continue
			}
			for _, record := range records {
				if record.Name != named.Name.Name {
					continue
				}
				for i, field := range record.Fields {
					fields[shape.Field(i)] = [2]string{record.ID, field.ID}
				}
			}
		}
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		WalkFieldSyntax(fn.Body, func(slot FieldSyntax) {
			origin, retained := origins[slot.Node]
			identity, typed := fields[info.Uses[slot.Field]]
			if retained && typed {
				origin.RecordID, origin.FieldID = identity[0], identity[1]
				origins[slot.Node] = origin
			}
		})
	}
}

func (p bodyProjection) emittedFieldOrigins() []BodyFieldOrigin {
	var result []BodyFieldOrigin
	for _, decl := range p.file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		counts := map[string]int{}
		first := len(result)
		WalkFieldSyntax(fn.Body, func(slot FieldSyntax) {
			ordinal := counts[slot.Kind]
			counts[slot.Kind]++
			if origin, ok := p.fields[slot.Node]; ok && origin.RecordID != "" && origin.FieldID != "" {
				origin.NativeOrdinal = ordinal
				result = append(result, origin)
			}
		})
		for i := first; i < len(result); i++ {
			result[i].NativePopulation = counts[result[i].Kind]
		}
	}
	return result
}
