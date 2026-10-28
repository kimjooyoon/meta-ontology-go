package bodycodegen

import (
	"encoding/json"
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const recordBodyRecordKind uint8 = 7
const recordBodyRecordSequenceKind uint8 = recordBodyRecordKind + 3

func recordBodyElementKind(kind uint8) bool {
	return kind >= 1 && kind <= 3 || kind == recordBodyRecordKind
}

func recordBodyCollectionKind(kind uint8) bool {
	return kind >= 4 && kind <= 6 || kind == recordBodyRecordSequenceKind
}

// Defaults form a bounded DAG of nominal layouts. Copies share immutable child
// defaults; assigning a field replaces a value instead of mutating that child.
func newRecordBodyZero(t types.Type) (recordBodyValue, error) {
	cache := make(map[*types.Named]*recordBodyValue)
	value, err := buildRecordBodyZero(t, cache, make(map[*types.Named]bool), 1)
	if err == nil {
		err = validateRecordBodyValueBudget(value)
	}
	return value, err
}

func buildRecordBodyZero(t types.Type, cache map[*types.Named]*recordBodyValue,
	visiting map[*types.Named]bool, depth int) (recordBodyValue, error) {
	named, ok := t.(*types.Named)
	if !ok {
		return recordBodyValue{}, fmt.Errorf("record requires a nominal value type")
	}
	if depth > syntax.RecordReferenceDepthLimit || visiting[named] {
		return recordBodyValue{}, fmt.Errorf("record reference graph exceeds its acyclic depth bound")
	}
	if prior := cache[named]; prior != nil {
		if depth+prior.height-1 > syntax.RecordReferenceDepthLimit {
			return recordBodyValue{}, fmt.Errorf("record reference depth exceeds %d", syntax.RecordReferenceDepthLimit)
		}
		return *prior, nil
	}
	if len(cache)+len(visiting) >= recordTypeLimit {
		return recordBodyValue{}, fmt.Errorf("record closure exceeds %d nominal types", recordTypeLimit)
	}
	structure, ok := named.Underlying().(*types.Struct)
	if !ok || structure.NumFields() < 1 || structure.NumFields() > RecordFieldLimit {
		return recordBodyValue{}, fmt.Errorf("record layout exceeds value bounds")
	}
	value := recordBodyValue{Type: named.Obj().Name(), Count: structure.NumFields(), layout: structure, height: 1}
	visiting[named] = true
	defer delete(visiting, named)
	for i := 0; i < value.Count; i++ {
		field, height, err := zeroRecordBodyField(structure.Field(i).Type(), cache, visiting, depth)
		if err != nil {
			return recordBodyValue{}, err
		}
		value.Values[i] = field
		value.height = max(value.height, height)
	}
	cache[named] = &value
	return value, nil
}

func zeroRecordBodyField(t types.Type, cache map[*types.Named]*recordBodyValue,
	visiting map[*types.Named]bool, depth int) (recordBodyScalar, int, error) {
	kind, optional, err := recordBodyScalarKind(t)
	field := recordBodyScalar{Kind: kind, Present: !optional}
	if err != nil {
		return field, 1, err
	}
	if pointer, ok := t.(*types.Pointer); ok {
		t = pointer.Elem()
	}
	if sequence, ok := t.(*types.Slice); ok {
		t = sequence.Elem()
	}
	if kind != recordBodyRecordKind && kind != recordBodyRecordSequenceKind {
		return field, 1, nil
	}
	child, err := buildRecordBodyZero(t, cache, visiting, depth+1)
	if err != nil {
		return field, 1, err
	}
	if kind == recordBodyRecordKind && !optional {
		field.Record = cache[t.(*types.Named)]
	}
	return field, child.height + 1, nil
}

func recordBodyHasReferences(value recordBodyValue) bool {
	for _, field := range value.Values[:value.Count] {
		if field.Kind == recordBodyRecordKind || field.Kind == recordBodyRecordSequenceKind {
			return true
		}
	}
	return false
}

// Count logical occurrences, including shared defaults at every field/index.
// Flat scalar profiles retain their existing independent collection limits.
func validateRecordBodyValueBudget(value recordBodyValue) error {
	if value.Count < 1 || value.Count > RecordFieldLimit {
		return fmt.Errorf("record layout exceeds value bounds")
	}
	if !recordBodyHasReferences(value) {
		return nil
	}
	nodes := 0
	return chargeRecordBodyValue(value, &nodes, 1)
}

func chargeRecordBodyValue(value recordBodyValue, nodes *int, depth int) error {
	if depth > syntax.RecordReferenceDepthLimit || value.Count < 1 || value.Count > RecordFieldLimit {
		return fmt.Errorf("record value exceeds its layout depth bound")
	}
	*nodes++
	if *nodes > syntax.RecordValueSlotLimit {
		return fmt.Errorf("record value exceeds %d slots", syntax.RecordValueSlotLimit)
	}
	for _, field := range value.Values[:value.Count] {
		if !field.Present {
			continue
		}
		*nodes++
		if *nodes > syntax.RecordValueSlotLimit {
			return fmt.Errorf("record value exceeds %d slots", syntax.RecordValueSlotLimit)
		}
		if field.Record != nil {
			if err := chargeRecordBodyValue(*field.Record, nodes, depth+1); err != nil {
				return err
			}
		}
		if field.Collection != nil {
			for _, item := range field.Collection.Values {
				if item.Record != nil {
					if err := chargeRecordBodyValue(*item.Record, nodes, depth+1); err != nil {
						return err
					}
				} else {
					*nodes++
				}
				if *nodes > syntax.RecordValueSlotLimit {
					return fmt.Errorf("record value exceeds %d slots", syntax.RecordValueSlotLimit)
				}
			}
		}
	}
	return nil
}

func (value recordBodyValue) MarshalJSON() ([]byte, error) {
	if err := validateRecordBodyValueBudget(value); err != nil {
		return nil, err
	}
	if value.layout == nil || value.layout.NumFields() != value.Count {
		return nil, fmt.Errorf("record JSON requires its declared field layout")
	}
	object := make(map[string]any, value.Count)
	for i, field := range value.Values[:value.Count] {
		if field.Present {
			object[recordBodyJSONName(value.layout, i)] = recordBodyScalarValue(field)
		}
	}
	return json.Marshal(object)
}

func recordBodyJSONName(layout *types.Struct, index int) string {
	name, _, _ := strings.Cut(reflect.StructTag(layout.Tag(index)).Get("json"), ",")
	if name != "" {
		return name
	}
	return layout.Field(index).Name()
}
