package syntax

import (
	"testing"
)

func TestStage0SyntaxContractHandlesMalformedUTF8(t *testing.T) {
	source := string([]byte{'e', 'n', 't', 'i', 't', 'y', ' ', 0xff})
	tokens, diagnostics := Lex(source)
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != TokenEOF {
		t.Fatalf("malformed input did not produce a terminating EOF: %#v", tokens)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != DiagUnexpectedCharacter {
		t.Fatalf("malformed UTF-8 diagnostics = %#v", diagnostics)
	}
}
