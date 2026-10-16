package syntax

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

const textBoundSource1905 = `package bounded
namespace bounded
entity Expanded id "bound1905://expanded" fields {
 field value id "bound1905://expanded/value" type string required one max_utf8_bytes "1030"
}
`

func TestTextBoundSyntaxRoundTripFresh1905(t *testing.T) {
	file, diagnostics := ParseFileWithEntityFieldsSupport("bounded.gooo", textBoundSource1905, EntityFieldsV6Support())
	if diagnostics.HasErrors() || file == nil {
		t.Fatal(diagnostics)
	}
	field := file.Declarations[0].(*EntityDecl).Fields[0]
	if field.TextBound != textbound.New(1030) || !strings.HasSuffix(textBoundSource1905[:field.Span.End.Offset], `"1030"`) {
		t.Fatal("bound or aggregate source span lost", field)
	}
	formatted, err := FormatWithEntityFieldsSupport(file, EntityFieldsV6Support())
	if err != nil || !strings.Contains(formatted, `max_utf8_bytes "1030"`) {
		t.Fatal("bound did not format", formatted, err)
	}
	other, diagnostics := ParseFileWithEntityFieldsSupport("bounded.gooo", formatted, EntityFieldsV6Support())
	if diagnostics.HasErrors() || other.Declarations[0].(*EntityDecl).Fields[0].TextBound != field.TextBound {
		t.Fatal("bound changed on round trip", diagnostics)
	}
	if _, err := FormatWithEntityFieldsSupport(file, EntityFieldsV5Support()); err == nil {
		t.Fatal("older profile formatted an unsupported bound")
	}
}

func TestTextBoundProfileSelectionFresh1905(t *testing.T) {
	for _, tc := range []struct {
		name   string
		suffix string
		want   int
	}{{"absent", "", 4}, {"zero", ` max_utf8_bytes "0"`, 6}, {"expanded", ` max_utf8_bytes "1030"`, 6}} {
		t.Run(tc.name, func(t *testing.T) {
			source := strings.Replace(textBoundSource1905, ` max_utf8_bytes "1030"`, tc.suffix, 1)
			file, diagnostics := ParseFileWithEntityFieldsSupport("bound.gooo", source, EntityFieldsV4Support())
			if diagnostics.HasErrors() {
				t.Fatal(diagnostics)
			}
			if got := EntityFieldsSupportForFile(file, EntityFieldsV4Support()); got.Profile.Version != tc.want {
				t.Fatal("source did not select its capability", got)
			}
		})
	}
}

func TestTextBoundV6ProfileDigestFresh1905(t *testing.T) {
	canonical := "id=gooo.entityfields.go-projection.v6\n" +
		"types=urn:gooo:type:string,urn:gooo:type:boolean,urn:gooo:type:integer\n" +
		"presence=required,optional\ncardinality=one,many\nordering=source\n" +
		"optional-representation=pointer\nrepeated-limit=256\n" +
		"text-bound=max_utf8_bytes\ntext-bound-unit=decoded-utf8-bytes\n" +
		"text-bound-range=0..4294967295\ntext-bound-absent=legacy\n"
	digest := sha256.Sum256([]byte(canonical))
	if hex.EncodeToString(digest[:]) != EntityFieldsV6ProfileDigest {
		t.Fatal("V6 digest does not bind its contract")
	}
	if err := EntityFieldsV6Support().Validate(); err != nil {
		t.Fatal(err)
	}
}
