package bodylibrary

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

func recordScopeFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../examples/body-codegen/selected-record-scope.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestLibraryRecordScopeUsesSelectedActivityClosure(t *testing.T) {
	source := recordScopeFixture(t)
	generated, err := Generate(context.Background(), "scope.gooo", source, "Evaluate", "scope")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Request", "Result", "Scratch", "Unused16"} {
		if !strings.Contains(generated, "type R_"+name+" struct") {
			t.Fatal("public declaration or selected type lost", name)
		}
	}
}

func TestLibraryRecordScopeStillRejectsSeventeenReachableRecords(t *testing.T) {
	var source strings.Builder
	source.WriteString("package active\nnamespace active\n")
	for i := range 17 {
		fmt.Fprintf(&source, "entity R%d id \"urn:active:%d\" fields {\nfield value id \"urn:active:%d:value\" type integer required one\n}\n", i, i, i)
	}
	source.WriteString("activity Evaluate(R0) -> R0 computes `\nlet total = input.value\n")
	for i := 1; i < 17; i++ {
		fmt.Fprintf(&source, "let r%d = R%d{value: input.value}\ntotal = total + r%d.value\n", i, i, i)
	}
	source.WriteString("return R0{value: total}\n`\n")
	generated, err := Generate(context.Background(), "active.gooo", []byte(source.String()), "Evaluate", "scope")
	if generated != "" || err == nil || !strings.Contains(err.Error(), "supports at most 16") {
		t.Fatal("active bound was relaxed", generated, err)
	}
}
