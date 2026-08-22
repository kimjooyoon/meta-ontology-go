package main

import (
	"encoding/base64"
	"go/format"
	"os"
	"testing"
)

func TestGo127FormatReceipt(t *testing.T) {
	source, err := os.ReadFile("digest.go")
	if err != nil {
		t.Fatal(err)
	}
	formatted, err := format.Source(source)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("FORMAT_RECEIPT digest.go %s", base64.StdEncoding.EncodeToString(formatted))
	t.Fail()
}
