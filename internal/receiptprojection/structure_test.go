package receiptprojection

import (
	"strings"
	"testing"
)

func TestStructuralIdentityIgnoresPresentationAndRetainsTypedMeaning(t *testing.T) {
	source := string(fixture(t))
	original, err := Compile("receipt.gooo", []byte(source), "CompletenessReceipt")
	if err != nil {
		t.Fatal(err)
	}
	want, err := original.StructureDigest()
	if err != nil {
		t.Fatal(err)
	}
	for name, changed := range map[string]string{
		"presentation": "// extra source context\n" + source,
		"counter-type": strings.Replace(source, "type Count required one", "type Text required one", 1),
		"stable-id":    strings.Replace(source, "field/dimension/numerator", "field/dimension/observed-count", 1),
		"cardinality":  strings.Replace(source, "type Text required many", "type Text required one", 1),
	} {
		t.Run(name, func(t *testing.T) {
			p, err := Compile("changed.gooo", []byte(changed), "CompletenessReceipt")
			if err != nil {
				t.Fatal(err)
			}
			got, err := p.StructureDigest()
			if err != nil {
				t.Fatal(err)
			}
			if (got == want) != (name == "presentation") || p.SourceSHA256 == original.SourceSHA256 {
				t.Fatal("structural identity conflated presentation and typed meaning")
			}
		})
	}
	original.Entities[0].Fields[0].TypeID = typePrefix + "count"
	if _, err := original.StructureDigest(); err == nil {
		t.Fatal("unbound in-memory structural mutation accepted")
	}
}
