package cache

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type canonicalMapEntry struct {
	key   []byte
	value reflect.Value
}
type canonicalField struct {
	name  string
	index int
}

func canonicalFields(typ reflect.Type) ([]canonicalField, error) {
	fields := make([]canonicalField, 0, typ.NumField())
	seen := make(map[string]struct{}, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			return nil, fmt.Errorf("unsupported unexported field %s.%s", typ, field.Name)
		}
		name := field.Name
		if tag, ok := field.Tag.Lookup("json"); ok {
			name = strings.Split(tag, ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("duplicate canonical field name %q in %s", name, typ)
		}
		seen[name] = struct{}{}
		fields = append(fields, canonicalField{name: name, index: i})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].name < fields[j].name })
	return fields, nil
}
func typeDescriptor(typ reflect.Type) string {
	if typ.PkgPath() != "" && typ.Name() != "" {
		return typ.PkgPath() + "." + typ.Name()
	}
	return typ.String()
}
