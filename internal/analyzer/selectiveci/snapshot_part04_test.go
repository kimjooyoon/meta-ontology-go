package selectiveci

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"testing"
)

func TestBuildRejectsUncertainAuthorityWithoutPartialIDs(t *testing.T) {
	valid := testInput(t, "pkg/order.go", "Order", "urn:gooo:entity:order")

	cases := []struct {
		name string
		edit func(*SnapshotInput)
		code ErrorCode
	}{
		{name: "missing binding", edit: func(input *SnapshotInput) {
			input.Sources[0].Bindings = nil
		}, code: CodeMissingBinding},
		{name: "ambiguous source attachment", edit: func(input *SnapshotInput) {
			input.Sources[0].Path = "pkg/other.go"
		}, code: CodeAmbiguousBinding},
		{name: "duplicate binding", edit: func(input *SnapshotInput) {
			input.Sources[0].Bindings = append(append([]semanticbinding.Binding(nil), valid.Bindings...), valid.Bindings[0])
		}, code: CodeDuplicateBinding},
		{name: "unregistered ID", edit: func(input *SnapshotInput) {
			input.RegisteredIDs = []string{"urn:gooo:entity:other"}
		}, code: CodeUnregisteredID},
		{name: "malformed path", edit: func(input *SnapshotInput) {
			input.Sources[0].Path = "../outside.go"
		}, code: CodeMalformedPath},
		{name: "malformed digest", edit: func(input *SnapshotInput) {
			input.Sources[0].BlobDigest = "not-a-digest"
		}, code: CodeMalformedDigest},
		{name: "candidate-only identity", edit: func(input *SnapshotInput) {
			input.CandidateBindings = valid.Bindings
		}, code: CodeCandidateIdentity},
		{name: "derived-only identity", edit: func(input *SnapshotInput) {
			input.DerivedBindings = valid.Bindings
		}, code: CodeDerivedIdentity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := SnapshotInput{
				Sources:         []SourceInput{testInput(t, "pkg/order.go", "Order", "urn:gooo:entity:order")},
				SourceMapDigest: testDigest("source-map"),
				RegistryDigest:  testDigest("registry"),
				RegisteredIDs:   []string{"urn:gooo:entity:order"},
			}
			tc.edit(&input)
			got, err := Build(input)
			if err == nil {
				t.Fatal("Build accepted uncertain authority")
			}
			var typed *Error
			if !errors.As(err, &typed) || typed.Code != tc.code {
				t.Fatalf("error = %v, want code %q", err, tc.code)
			}
			if got.Status != StatusUnknown || !got.FullSuiteFallback || len(got.Sources) != 0 || got.Digest != "" {
				t.Fatalf("unknown snapshot retained partial authority: %#v", got)
			}
		})
	}
}
