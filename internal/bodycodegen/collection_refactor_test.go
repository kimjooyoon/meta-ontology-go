package bodycodegen

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

func TestCollectionCallRefactorRetainsSource7867(t *testing.T) {
	body := `let 설명 = "Original for x in values"; let total = 0; for 값 in input.items { total = total + Original(값) }; return total`
	renamed, count, err := renameBodyCalls(body, 1, "Original", "보정")
	want := strings.Replace(body, "Original(값)", "보정(값)", 1)
	if err != nil || count != 1 || renamed != want {
		t.Fatal("rename after expanded for-in changed unrelated original bytes", renamed, count, err)
	}
	source := []byte(`package refactor7867
namespace refactor7867
entity Integer id "urn:gooo:type:integer"
entity Batch id "urn:refactor7867:batch" fields {
 field items id "urn:refactor7867:items" type integer required many
}
activity Original(Integer) -> Integer id "urn:refactor7867:original" computes "return input + 6"
activity Sum(Batch) -> Integer id "urn:refactor7867:sum" computes ` + strconv.Quote(body) + "\n")
	offset := strings.LastIndex(string(source), "Original(")
	reference, found, err := ResolveActivityReference(context.Background(), "수정7867.gooo", source, offset+3)
	if err != nil || !found || reference.ActivityID != "urn:refactor7867:original" ||
		string(source[reference.Start:reference.End]) != "Original" {
		t.Fatal("source reference after collection lowering", reference, found, err)
	}
	textOffset := strings.Index(string(source), "Original for")
	if _, found, err := ResolveActivityReference(context.Background(), "수정7867.gooo", source, textOffset+3); err != nil || found {
		t.Fatal("a string became an activity reference", found, err)
	}
	t.Log("7867: for-in call rename changes one source name; Korean replacement and stable reference ID retained; quoted text unchanged; execution0 model0")
}
