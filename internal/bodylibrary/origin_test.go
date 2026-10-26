package bodylibrary

import (
	"context"
	"strings"
	"testing"
)

func TestLibraryOriginIsOptionalAndSharesOutputBound(t *testing.T) {
	const source = `package origin
namespace origin
entity Number id "urn:origin:number" fields {
 field value id "urn:origin:value" type integer required one
}
activity Keep(Number) -> Number id "urn:origin:keep" computes "return input"
`
	plain, err := Generate(context.Background(), "origin.gooo", []byte(source), "Keep", "origin")
	if err != nil || strings.Contains(plain, "GoooGenerationOrigin") {
		t.Fatal("default output acquired optional metadata", err)
	}
	options := GenerationOptions{Origin: &GenerationContext{CompilerVersion: strings.Repeat("x", maxLibraryProjectionBytes)}}
	generated, err := GenerateWithOptions(context.Background(), "origin.gooo", []byte(source), "Keep", "origin", options)
	if generated != "" || err == nil || !strings.Contains(err.Error(), "exceeds 256 KiB") {
		t.Fatal("optional origin bypassed the output bound", err)
	}
}
