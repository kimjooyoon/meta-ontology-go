package main

import "testing"

func TestSelectionAlreadySplit12063(t *testing.T) {
	source := []byte(`package selection12063
func activeDenominatorChunk12063(total int) []byte {
 if total==73 {return []byte{7,3}}
 return nil
}
`)
	maximum, functions, parts := checkSelectionUnit11447(t, "already_split12063.go", source)
	if functions != 1 || parts != 1 || maximum > 6 {
		t.Fatal("already bounded unit was not accepted intact", maximum, functions, parts)
	}
	t.Log("new already-split source inspected intact;one selector/one unit;no split demanded;program executions0")
}
