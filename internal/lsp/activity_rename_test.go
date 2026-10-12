package lsp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestActivityRenameRequestRetainsIDAndRewritesBodyCalls(t *testing.T) {
	const uri = "file:///rename.gooo"
	const source = `package p
namespace p
entity Integer id "urn:gooo:type:integer"
// 이름 앞의 유니코드도 위치 계산에 포함됩니다.
activity Fee(Integer) -> Integer computes "return input"
activity Total(Integer) -> Integer computes "return Fee(input)"
`
	server := NewServer()
	server.documents[uri] = &document{version: 1, text: source, result: SyntaxParser{}.Parse(uri, source)}
	params, _ := json.Marshal(RenameParams{TextDocument: TextDocumentIdentifier{URI: uri}, Position: Position{Line: 4, Character: 10}, NewName: "요금"})
	response, _, err := server.renameRequest(requestEnvelope{ID: json.RawMessage("1"), Params: params})
	if err != nil || response == nil || response.Error != nil {
		t.Fatal(response, err)
	}
	raw, _ := json.Marshal(response.Result)
	var workspace WorkspaceEdit
	if err := json.Unmarshal(raw, &workspace); err != nil {
		t.Fatal(err)
	}
	edits := workspace.Changes[uri]
	if len(edits) != 3 {
		t.Fatal("expected declaration, ID and body edits", string(raw))
	}
	updated := source
	for i := len(edits) - 1; i >= 0; i-- {
		start, err := PositionToOffset(source, edits[i].Range.Start)
		if err != nil {
			t.Fatal(err)
		}
		end, err := PositionToOffset(source, edits[i].Range.End)
		if err != nil {
			t.Fatal(err)
		}
		updated = updated[:start] + edits[i].NewText + updated[end:]
	}
	if !strings.Contains(updated, `activity 요금(Integer) -> Integer id "p://activity/fee"`) ||
		!strings.Contains(updated, `return 요금(input)`) {
		t.Fatal(updated)
	}
	if server.documents[uri].text != source {
		t.Fatal("server changed the client buffer")
	}
	params, _ = json.Marshal(RenameParams{TextDocument: TextDocumentIdentifier{URI: uri}, Position: Position{Line: 4, Character: 10}, NewName: "input"})
	response, _, err = server.renameRequest(requestEnvelope{ID: json.RawMessage("2"), Params: params})
	if err != nil || response.Error == nil || !strings.Contains(response.Error.Message, "captured") {
		t.Fatal(response, err)
	}
}
