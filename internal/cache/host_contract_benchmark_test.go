package cache

import "testing"

func BenchmarkNewKeyPerHostStage(b *testing.B) {
	spec := KeySpec{
		Version: "v1", Namespace: "billing", ToolVersion: "compiler-1",
		HostStage: GoHostedStage, Inputs: map[string]any{"source": "main.gooo"},
		Options: map[string]any{"mode": "fast"},
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewKey(spec); err != nil {
			b.Fatal(err)
		}
	}
}
