package languageconcept

import (
	"encoding/base64"
	"go/format"
	"os"
	"testing"
)

func TestGo127FormatReceipt(t *testing.T) {
	for _, name := range []string{"evaluate.go", "evidence.go", "model.go"} {
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		formatted, err := format.Source(source)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("FORMAT_RECEIPT %s %s", name, base64.StdEncoding.EncodeToString(formatted))
	}
	t.Fail()
}
