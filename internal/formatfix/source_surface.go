package formatfix

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func needsSourceSurface(file *syntax.File) bool {
	if len(file.Imports) > 0 || len(file.Bindings) > 0 {
		return true
	}
	for _, declaration := range file.Decls {
		if entity, ok := declaration.(*syntax.EntityDecl); ok && (entity.FieldsPresent || len(entity.Fields) > 0) {
			return true
		}
		if activity, ok := declaration.(*syntax.ActivityDecl); ok &&
			(activity.ValueProgramPresent || activity.ValueProgram != "" || activity.Assembly != nil) {
			return true
		}
	}
	return false
}

func buildSourceSurface(plan Plan, file *syntax.File, source string) Plan {
	formatted, err := syntax.FormatWithEntityFieldsSupport(file, syntax.EntityFieldsSupportForFile(file, syntax.EntityFieldsV4Support()))
	if err != nil {
		plan.Diagnostics = append(plan.Diagnostics, err.Error())
		return lower(plan, "FORMAT_FIX_SOURCE_SURFACE_UNKNOWN")
	}
	after, diagnostics := syntax.ParseFileWithEntityFieldsSupport(plan.File, formatted, syntax.EntityFieldsV4Support())
	if diagnostics.HasErrors() {
		for _, diagnostic := range diagnostics {
			plan.Diagnostics = append(plan.Diagnostics, diagnostic.String())
		}
		return lower(plan, "FORMAT_FIX_SOURCE_SURFACE_ROUNDTRIP_UNKNOWN")
	}
	beforeHash, beforeErr := sourceSurfaceFingerprint(file)
	afterHash, afterErr := sourceSurfaceFingerprint(after)
	if beforeErr != nil || afterErr != nil {
		plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("source surface: before=%v after=%v", beforeErr, afterErr))
		return lower(plan, "FORMAT_FIX_SOURCE_SURFACE_FINGERPRINT_UNKNOWN")
	}
	plan.SemanticBefore, plan.SemanticAfter = beforeHash, afterHash
	plan.SemanticEqual = beforeHash != "" && beforeHash == afterHash
	if !plan.SemanticEqual {
		return lower(plan, "FORMAT_FIX_SOURCE_SURFACE_DRIFT")
	}
	plan.ResultDigest, plan.ResultBytes = digestBytes([]byte(formatted)), len(formatted)
	if formatted == source {
		plan.Decision, plan.Resolution = DecisionFixedPoint, ResolutionExact
		plan.ReasonCode = "FORMAT_FIX_ALREADY_CANONICAL"
		return seal(plan)
	}
	plan.Decision, plan.Resolution = DecisionChangePlanned, ResolutionExact
	plan.ReasonCode, plan.Changed = "FORMAT_FIX_CHANGE_PLANNED", true
	plan.Edits = []Edit{{Start: 0, End: len(source), Replacement: formatted}}
	return seal(plan)
}

// Compare the complete parsed carrier, retaining body bytes, binding ports,
// import paths, IDs, types and assembly intent. Only source positions and import
// ordering are presentation; assembly uses its existing source-owned canonical
// contract. This does not resolve imported types or establish runtime equivalence.
func sourceSurfaceFingerprint(file *syntax.File) (string, error) {
	clone := file.Clone()
	sort.Slice(clone.Imports, func(i, j int) bool { return clone.Imports[i].Path < clone.Imports[j].Path })
	value, err := sourceSurfaceValue(reflect.ValueOf(clone), 0)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(raw), nil
}

var sourceSpanType = reflect.TypeFor[syntax.Span]()
var sourcePositionType = reflect.TypeFor[syntax.Position]()
var assemblyIntentType = reflect.TypeFor[assemblyspec.Spec]()

func sourceSurfaceValue(value reflect.Value, depth int) (any, error) {
	if depth > 128 {
		return nil, fmt.Errorf("source carrier nesting exceeds 128")
	}
	if !value.IsValid() || value.Type() == sourceSpanType || value.Type() == sourcePositionType {
		return nil, nil
	}
	if value.Type() == assemblyIntentType {
		return value.Interface().(assemblyspec.Spec).Canonical()
	}
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		if value.IsNil() {
			return nil, nil
		}
		return sourceSurfaceValue(value.Elem(), depth+1)
	case reflect.Struct:
		result := map[string]any{"$type": value.Type().String()}
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if field.Type == sourceSpanType || field.Type == sourcePositionType {
				continue
			}
			if field.PkgPath != "" {
				return nil, fmt.Errorf("unexported source carrier field %s.%s", value.Type(), field.Name)
			}
			part, err := sourceSurfaceValue(value.Field(i), depth+1)
			if err != nil {
				return nil, err
			}
			result[field.Name] = part
		}
		return result, nil
	case reflect.Slice, reflect.Array:
		result := make([]any, value.Len())
		for i := range result {
			part, err := sourceSurfaceValue(value.Index(i), depth+1)
			if err != nil {
				return nil, err
			}
			result[i] = part
		}
		return result, nil
	case reflect.String:
		return value.String(), nil
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	default:
		return nil, fmt.Errorf("unsupported source carrier kind %s", value.Kind())
	}
}
