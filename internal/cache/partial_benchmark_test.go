package cache

import "testing"

func BenchmarkNewPartialKey(b *testing.B) {
	spec := partialSpec("body", 42, 7)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewPartialKey(spec); err != nil {
			b.Fatal(err)
		}
	}
}
