package couplingmanifest

import (
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
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

func assertConstructionPartition(t *testing.T, output BuildOutput, err error, wantStatus ConstructionStatus, wantCode ConstructionCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected construction error, output=%#v", output)
	}
	var constructionErr *ConstructionError
	if !errors.As(err, &constructionErr) {
		t.Fatalf("error type = %T, want *ConstructionError: %v", err, err)
	}
	if constructionErr.Status != wantStatus || constructionErr.Code != wantCode {
		t.Fatalf("construction error = %#v, want status=%s code=%s", constructionErr, wantStatus, wantCode)
	}
	if output.Metadata.Status != wantStatus || output.Metadata.Reason != wantCode {
		t.Fatalf("metadata = %#v, want status=%s code=%s", output.Metadata, wantStatus, wantCode)
	}
	if output.Manifest.Complete || output.Manifest.Digest != "" || len(output.Manifest.Entries) != 0 {
		t.Fatalf("error output contains authoritative manifest data: %#v", output.Manifest)
	}
}
