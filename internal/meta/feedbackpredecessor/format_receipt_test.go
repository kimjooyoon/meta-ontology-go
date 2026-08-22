package feedbackpredecessor

import (
	"go/format"
	"os"
	"testing"
)

func TestGoFormatReceipt(t *testing.T) {
	source, err := os.ReadFile("select_test.go")
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := format.Source(source)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatalf("GOFORMAT_CANONICAL select_test.go %q", canonical)
}
