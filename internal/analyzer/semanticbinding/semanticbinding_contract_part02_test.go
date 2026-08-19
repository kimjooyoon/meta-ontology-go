package semanticbinding

import (
	"testing"
)

func TestNonExactDirectiveMarkerDoesNotBind(t *testing.T) {
	source := []byte("package billing\n\n// gooo:bind id=\"billing://entity/order\" role=\"HANDWRITTEN_IMPL\"\ntype Order struct{}\n")
	result, err := Extract(Input{Sources: []SourceFile{{
		Filename: "non_exact_marker.go", PackagePath: "billing", Source: source,
	}}})
	if err != nil {
		t.Fatalf("Extract returned error for an ordinary comment: %v", err)
	}
	if result.Status != StatusBound || len(result.Bindings) != 0 || len(result.Obligations) != 0 {
		t.Fatalf("result = %#v, want no semantic records", result)
	}
}
