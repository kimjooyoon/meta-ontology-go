package main

import (
	"encoding/base64"
	"go/format"
	"os"
	"testing"
)

func TestGo127UseCaseFormatReceipt(t *testing.T) {
	files := map[string]string{
		"cmd/feedback-semantic-state-witness/fixture_test.go": "fixture_test.go",
		"internal/meta/feedbackstate/usecase_fixture_test.go":    "../../internal/meta/feedbackstate/usecase_fixture_test.go",
	}
	for logical, name := range files {
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		formatted, err := format.Source(source)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("FORMAT_RECEIPT %s %s", logical, base64.StdEncoding.EncodeToString(formatted))
	}
	t.Fail()
}
