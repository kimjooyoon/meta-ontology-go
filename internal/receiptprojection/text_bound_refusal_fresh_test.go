package receiptprojection

import (
	"strings"
	"testing"
)

func TestReceiptProfileCannotEraseTextBoundFresh1905(t *testing.T) {
	const source = `package receipt_bound
namespace receipt_bound
entity Text id "gooo://receipt-structure/type/text"
entity Envelope id "gooo://schema/receipt-bound1905/v1" fields {
 field schema id "bound1905://receipt/schema" type Text required one
}
`
	if _, err := Compile("receipt.gooo", []byte(source), "Envelope"); err != nil {
		t.Fatal("fresh ordinary receipt schema did not project", err)
	}
	for _, bound := range []string{"0", "1030"} {
		changed := strings.Replace(source, "required one", `required one max_utf8_bytes "`+bound+`"`, 1)
		if _, err := Compile("receipt.gooo", []byte(changed), "Envelope"); err == nil ||
			!strings.Contains(err.Error(), "max_utf8_bytes") || !strings.Contains(err.Error(), "bound1905://receipt/schema") {
			t.Fatal("independent profile silently erased the source domain", bound, err)
		}
	}
}
