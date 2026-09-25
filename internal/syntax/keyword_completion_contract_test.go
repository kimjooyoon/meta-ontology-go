package syntax

import (
	"reflect"
	"testing"
)

func TestCanonicalKeywordNamesAreStableAndSorted(t *testing.T) {
	want := []string{"activity", "entity", "id", "namespace", "package"}
	got := CanonicalKeywordNames()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical keywords = %#v, want %#v", got, want)
	}
	got[0] = "mutated"
	if reflect.DeepEqual(CanonicalKeywordNames(), got) {
		t.Fatal("canonical keyword result aliases lexer state")
	}
}
