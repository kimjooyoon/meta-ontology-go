package generator

import (
	"testing"
)

func TestEntityFieldsProductionEntryPointsRemainDeferred(t *testing.T) {
	ir := entityFieldsFixture()
	cases := []struct {
		name string
		call func() (Result, error)
	}{
		{name: "project", call: func() (Result, error) { return Project(ir, nil) }},
		{name: "generate", call: func() (Result, error) { return Generate(ir, nil) }},
		{name: "generator method", call: func() (Result, error) { return New(Options{}).Generate(ir, nil) }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := testCase.call()
			assertEntityFieldsDeferred(t, err)
			if result.Source != nil || result.SourceMap.Mappings != nil {
				t.Fatalf("deferred production entry point returned artifacts: %#v", result)
			}
		})
	}

	t.Run("generate from", func(t *testing.T) {
		source, sourceMap, err := GenerateFrom(ir, Options{})
		assertEntityFieldsDeferred(t, err)
		if source != nil || sourceMap.Mappings != nil {
			t.Fatalf("GenerateFrom returned artifacts: %q %#v", source, sourceMap)
		}
	})
	t.Run("projection metadata", func(t *testing.T) {
		result, err := GenerateProjectionV1(ir, nil)
		assertEntityFieldsDeferred(t, err)
		if result.Source != nil || result.SourceMap.Mappings != nil || result.Schema != "" {
			t.Fatalf("GenerateProjectionV1 returned artifacts: %#v", result)
		}
	})
	t.Run("adapter projection metadata", func(t *testing.T) {
		result, err := GenerateFromProjectionV1(ir, Options{})
		assertEntityFieldsDeferred(t, err)
		if result.Source != nil || result.SourceMap.Mappings != nil || result.Schema != "" {
			t.Fatalf("GenerateFromProjectionV1 returned artifacts: %#v", result)
		}
	})
	t.Run("metadata", func(t *testing.T) {
		result, err := GenerateWithMetadata(ir, nil)
		assertEntityFieldsDeferred(t, err)
		if result.Source != nil || result.SourceMap.Mappings != nil || result.Metadata.SourceDigest != "" {
			t.Fatalf("GenerateWithMetadata returned artifacts: %#v", result)
		}
	})
	t.Run("binding", func(t *testing.T) {
		result, err := GenerateWithBinding(ir, nil, ProjectionBinding{})
		assertEntityFieldsDeferred(t, err)
		if result.Source != nil || result.SourceMap.Mappings != nil || result.Schema != "" {
			t.Fatalf("GenerateWithBinding returned artifacts: %#v", result)
		}
	})
}
