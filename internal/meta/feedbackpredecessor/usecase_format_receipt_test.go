package feedbackpredecessor

import (
	"go/format"
	"os"
	"testing"
)

func TestUseCaseFormatReceipt(t *testing.T) {
	source, err := os.ReadFile("usecase_types_test.go")
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := format.Source(source)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatalf("GOFORMAT_CANONICAL usecase_types_test.go %q", canonical)
}
