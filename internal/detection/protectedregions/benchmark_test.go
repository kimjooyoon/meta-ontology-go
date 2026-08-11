package protectedregions

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkProtectedRegionValidateLocality64(b *testing.B) {
	before := benchmarkSource(64, false)
	after := benchmarkSource(64, true)
	b.SetBytes(int64(len(before) + len(after)))
	b.ReportAllocs()
	b.ResetTimer()
	for attempt := 0; attempt < b.N; attempt++ {
		if report := ValidateLocality(before, after); !report.Valid() {
			b.Fatal(report.Err())
		}
	}
}

func benchmarkSource(regionCount int, changed bool) []byte {
	var source strings.Builder
	source.WriteString("package benchmark\n\n")
	source.WriteString("var Keep = 7\n")
	for index := 0; index < regionCount; index++ {
		id := fmt.Sprintf("benchmark://activity/%d", index)
		fmt.Fprintf(&source, "\n//gooo:generated:start id=\"%s\" kind=\"activity\"\n", id)
		fmt.Fprintf(&source, "func Activity%d() int {\n", index)
		if index%8 == 0 {
			slotID := id + "/implementation"
			fmt.Fprintf(&source, "\t//gooo:slot:start id=\"%s\"\n\treturn 7\n\t//gooo:slot:end id=\"%s\"\n", slotID, slotID)
		}
		if changed {
			fmt.Fprintf(&source, "\t// regenerated body %d\n", index)
		}
		source.WriteString("}\n")
		fmt.Fprintf(&source, "//gooo:generated:end id=\"%s\"\n", id)
	}
	return []byte(source.String())
}
