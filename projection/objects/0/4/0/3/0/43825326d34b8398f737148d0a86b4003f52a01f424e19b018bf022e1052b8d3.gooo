package syntax

import (
	"testing"
)

func TestSyntheticFieldsAreNotSilentlyDroppedByFormatter(t *testing.T) {
	file := &File{
		Package:   &PackageDecl{Name: "billing"},
		Namespace: &NamespaceDecl{Name: "billing"},
		Decls: []Declaration{&EntityDecl{
			Name: "Order",
			ID:   "billing://entity/order",
			Fields: []FieldDecl{{
				ID:          "billing://field/name",
				Name:        "Name",
				TypeRef:     TypeRefDecl{Spelling: "string"},
				Presence:    FieldPresenceRequired,
				Cardinality: FieldCardinalityOne,
			}},
		}},
	}
	file.Declarations = file.Decls
	formatted, err := Format(file)
	if err != ErrLatentFieldsUnsupported || formatted != "" {
		t.Fatalf("latent field format result = %q, %v", formatted, err)
	}
}
func TestFieldEnumsExposeOnlyTechnologyIndependentValues(t *testing.T) {
	if !FieldPresenceRequired.Valid() || !FieldPresenceOptional.Valid() || FieldPresence("other").Valid() {
		t.Fatal("presence carrier values are not stable")
	}
	if !FieldCardinalityOne.Valid() || !FieldCardinalityMany.Valid() || FieldCardinality("many-items").Valid() {
		t.Fatal("cardinality carrier values are not stable")
	}
}
