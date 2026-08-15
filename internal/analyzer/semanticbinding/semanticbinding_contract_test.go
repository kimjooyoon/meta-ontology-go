package semanticbinding

import "testing"

func TestSemanticBindingImplementationContract(t *testing.T) {
	t.Run("valid bind", func(t *testing.T) { assertFixture(t, "valid_bind") })
	t.Run("valid obligation", func(t *testing.T) { assertFixture(t, "valid_obligation") })
	t.Run("rename preserving ID", func(t *testing.T) {
		before := assertFixture(t, "rename_before")
		after := assertFixture(t, "rename_after")
		if len(before.records) != 1 || len(after.records) != 1 {
			t.Fatalf("rename records = %#v and %#v, want one record each", before.records, after.records)
		}
		if before.records[0].ID != after.records[0].ID ||
			before.records[0].Directive != after.records[0].Directive {
			t.Fatalf("rename changed semantic identity: before=%#v after=%#v", before.records, after.records)
		}
		if before.records[0].Name == after.records[0].Name {
			t.Fatalf("rename fixtures have the same display name: %q", before.records[0].Name)
		}
		if before.result.Bindings[0].DeclarationKey == after.result.Bindings[0].DeclarationKey {
			t.Fatalf("rename did not change the Go declaration key: %q", before.result.Bindings[0].DeclarationKey)
		}
		if before.result.Canonical() != after.result.Canonical() ||
			before.result.StableHash() != after.result.StableHash() {
			t.Fatalf("rename changed canonical semantic output: before=%s after=%s", before.result.StableHash(), after.result.StableHash())
		}
	})
	t.Run("same name without directive", func(t *testing.T) { assertFixture(t, "same_name_without_directive") })
	t.Run("detached comment", func(t *testing.T) { assertFixture(t, "detached_comment") })
	t.Run("unknown field", func(t *testing.T) { assertFixture(t, "unknown_field") })
	t.Run("duplicate field", func(t *testing.T) { assertFixture(t, "duplicate_field") })
	t.Run("duplicate ID", func(t *testing.T) { assertFixture(t, "duplicate_id") })
	t.Run("invalid URI", func(t *testing.T) { assertFixture(t, "invalid_uri") })
	t.Run("multi-name var declaration", func(t *testing.T) { assertFixture(t, "multi_name_var") })
	t.Run("two declarations with exact spans", func(t *testing.T) { assertFixture(t, "exact_spans") })
	t.Run("whitespace/comment canonical equality", func(t *testing.T) {
		a := assertFixture(t, "canonical_permutation_a")
		b := assertFixture(t, "canonical_permutation_b")
		if a.oracleCanonical != b.oracleCanonical || a.result.Canonical() != b.result.Canonical() ||
			a.result.StableHash() != b.result.StableHash() {
			t.Fatalf("presentation permutation changed canonical output: a=%q/%s b=%q/%s",
				a.oracleCanonical, a.result.StableHash(), b.oracleCanonical, b.result.StableHash())
		}
	})
	t.Run("filename identity versus location", func(t *testing.T) {
		a := assertFixture(t, "filename_identity_a")
		b := assertFixture(t, "filename_identity_b")
		if a.oracleCanonical != b.oracleCanonical || len(a.records) != 1 || len(b.records) != 1 {
			t.Fatalf("filename fixtures changed semantic records: a=%#v b=%#v", a.records, b.records)
		}
		if a.records[0].ID != b.records[0].ID || a.records[0].Name != b.records[0].Name {
			t.Fatalf("filename fixtures changed identity: a=%#v b=%#v", a.records, b.records)
		}
		if a.records[0].Location == nil || b.records[0].Location == nil ||
			a.records[0].Location.Filename == b.records[0].Location.Filename {
			t.Fatalf("filename locations were not retained independently: a=%#v b=%#v", a.records, b.records)
		}
		if a.records[0].Location.Start != b.records[0].Location.Start ||
			a.records[0].Location.End != b.records[0].Location.End {
			t.Fatalf("filename changed non-location span data: a=%#v b=%#v", a.records, b.records)
		}
		if a.result.Canonical() != b.result.Canonical() || a.result.StableHash() != b.result.StableHash() {
			t.Fatalf("filename relocation changed canonical semantic output: a=%s b=%s", a.result.StableHash(), b.result.StableHash())
		}
	})
}

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
