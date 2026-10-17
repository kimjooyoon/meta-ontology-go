package workspaceexecution

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

func TestWorkspaceCannotMergeDifferentTextDomainsFresh1905(t *testing.T) {
	base := syntax.EntityDecl{FieldsPresent: true, Fields: []syntax.FieldDecl{{
		ID: "bound1905://shared/value", Name: "value", TypeRef: syntax.TypeRefDecl{Spelling: "string"},
		Presence: syntax.FieldPresenceRequired, Cardinality: syntax.FieldCardinalityOne,
	}}}
	for _, bound := range []textbound.UTF8{0, textbound.New(0), textbound.New(1030)} {
		left, right := base, base
		left.Fields = append([]syntax.FieldDecl(nil), base.Fields...)
		right.Fields = append([]syntax.FieldDecl(nil), base.Fields...)
		left.Fields[0].TextBound, right.Fields[0].TextBound = bound, bound
		if !sameEntityShape(&left, &right) {
			t.Fatal("the same source domain stopped sharing its stable entity", bound)
		}
		right.Fields[0].TextBound = textbound.New(1031)
		if sameEntityShape(&left, &right) {
			t.Fatal("one package silently selected another package's field domain", bound)
		}
	}
}
