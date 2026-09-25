package couplingmanifest

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"testing"
)

func TestIdentityClassificationPartitionsUnknownAndFailClosed(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	surfaces := []surfaceFixture{{Owner: source.Bindings[0].ID, Suffix: "order"}}
	base := testInput(t, []selectiveci.SourceInput{source}, []selectiveci.SourceInput{source}, surfaces)
	unknownSource := testSource(t, "pkg/unknown.go", "Unknown", "urn:gooo:entity:unknown")
	cases := []struct {
		name   string
		input  func() Input
		mutate func(*Input)
		status ConstructionStatus
		code   ConstructionCode
	}{
		{
			name: "missing source-map ID", input: func() Input { return cloneInput(base) },
			mutate: func(input *Input) { input.SourceMap.Head[0].SourceMapID = "" },
			status: ConstructionUnknown, code: CodeUnknownChangedSurface,
		},
		{
			name: "unregistered source-map ID", input: func() Input { return cloneInput(base) },
			mutate: func(input *Input) { input.SourceMap.Head[0].SourceMapID = fixtureID("unregistered-map") },
			status: ConstructionUnknown, code: CodeUnknownChangedSurface,
		},
		{
			name: "unregistered snapshot ID",
			input: func() Input {
				return testInput(t, []selectiveci.SourceInput{source}, []selectiveci.SourceInput{source, unknownSource}, surfaces)
			},
			status: ConstructionUnknown, code: CodeUnknownChangedSurface,
		},
		{
			name: "ambiguous duplicate registration", input: func() Input { return cloneInput(base) },
			mutate: func(input *Input) {
				input.Authority.Registry.Surfaces = append(input.Authority.Registry.Surfaces, input.Authority.Registry.Surfaces[0])
			},
			status: ConstructionFailClosed, code: CodeDuplicateBinding,
		},
		{
			name: "malformed ID syntax", input: func() Input { return cloneInput(base) },
			mutate: func(input *Input) { input.SourceMap.Head[0].SourceMapID = "not-an-id" },
			status: ConstructionFailClosed, code: CodeMalformedBinding,
		},
		{
			name: "contradictory digest", input: func() Input { return cloneInput(base) },
			mutate: func(input *Input) {
				input.SourceMap.Head[0].SourceMapBindingDigest = testDigest("contradictory-binding")
			},
			status: ConstructionFailClosed, code: CodeConflictingBinding,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.input()
			if tc.mutate != nil {
				tc.mutate(&input)
			}
			output, err := BuildDetailed(input)
			assertConstructionPartition(t, output, err, tc.status, tc.code)
		})
	}
}
