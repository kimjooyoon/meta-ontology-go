package pressureshadow

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestStrictWire(t *testing.T) {
	base, _ := json.Marshal(s1Input())
	nested := bytes.Replace(base, []byte(`"coverage":{`), []byte(`"coverage":{"unknown":1,`), 1)
	duplicate := s1Input()
	duplicate.PathCoverage = append(duplicate.PathCoverage, duplicate.PathCoverage[0])
	duplicateWire, _ := json.Marshal(duplicate)
	cases := map[string][]byte{
		"unknown top-level": addRootField(base, `"unknown":1`),
		"unknown nested":    nested,
		"duplicate key":     addRootField(base, `"schema":"other"`),
		"trailing value":    append(base, []byte(`{}`)...),
		"schema":            bytes.Replace(base, []byte(`"schema":"`+SchemaVersion+`"`), []byte(`"schema":"bad"`), 1),
		"invalid ID":        bytes.Replace(base, []byte(`path-a`), []byte(`path a`), 1),
		"duplicate row":     duplicateWire,
	}
	duplicates := map[string]func(*Input){
		"pressure": func(input *Input) {
			input.Selector.Pressures = append(input.Selector.Pressures, input.Selector.Pressures[0])
		},
		"state": func(input *Input) {
			input.Selector.States = append(input.Selector.States, input.Selector.States[0])
		},
	}
	for name, mutate := range duplicates {
		input := s1Input()
		mutate(&input)
		cases["duplicate "+name], _ = json.Marshal(input)
		if _, err := CanonicalInputBytes(input); err == nil {
			t.Fatalf("CanonicalInputBytes accepted duplicate %s ID", name)
		}
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeInput(data); err == nil {
				t.Fatal("DecodeInput accepted malformed wire")
			}
		})
	}
}
