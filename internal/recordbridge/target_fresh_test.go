package recordbridge

import "testing"

func TestNativeFieldIdentityBridgeFresh1863(t *testing.T) {
	for _, item := range []struct {
		name, fields string
		accept       bool
	}{
		{"native-name", "Name string `gooo:\"urn:bridge1863:value\"`", true},
		{"renamed-native", "Renamed string `gooo:\"urn:bridge1863:value\"`", true},
		{"missing-ID", "Name string `json:\"name\"`", false},
		{"duplicate-ID", "A string `gooo:\"same\"`; B string `gooo:\"same\"`", false},
		{"late-tag-corruption", "Name string `gooo:\"urn:bridge1863:value\" json:bad`", false},
	} {
		t.Run(item.name, func(t *testing.T) {
			fields, err := Fields("native1863.go", []byte("package sample; type Row struct {"+item.fields+"}"), "Row")
			if (err == nil) != item.accept {
				t.Fatalf("native identity boundary: %v", err)
			}
			if item.accept && (len(fields) != 1 || fields["urn:bridge1863:value"].Kind != "string") {
				t.Fatal("native field name became source identity")
			}
		})
	}
}
