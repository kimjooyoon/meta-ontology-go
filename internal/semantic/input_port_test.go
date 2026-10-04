package semantic

import "testing"

func TestInputPortIndexesHaveOneCanonicalSpelling(t *testing.T) {
	for _, pair := range []struct {
		port         string
		arity, index int
	}{{"input", 1, 0}, {"input0", 2, 0}, {"input1", 2, 1}, {"input15", 16, 15}} {
		index, ok := InputPortIndex(pair.port, pair.arity)
		if !ok || index != pair.index {
			t.Fatal(pair)
		}
	}
	for _, pair := range []struct {
		port  string
		arity int
	}{{"input0", 1}, {"input", 2}, {"input01", 2}, {"input-1", 2}, {"input+1", 2}, {"input2", 2}, {"input", 0}, {"input999999999999999999999999999999999999999", 16}} {
		if _, ok := InputPortIndex(pair.port, pair.arity); ok {
			t.Fatal(pair)
		}
	}
}
