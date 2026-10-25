package extractor

import (
	"fmt"
	"strings"
	"testing"
)

func independentImportFixture(t *testing.T) (string, []byte) {
	t.Helper()
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module independent.origins\n\ngo 1.27.3\n")
	var source strings.Builder
	source.WriteString("package consumer\nimport (\n")
	for i := range 34 {
		fmt.Fprintf(&source, "d%d \"independent.origins/dep%d\"\n", i, i)
		writeImporterFixture(t, root, fmt.Sprintf("dep%d/value.go", i), fmt.Sprintf("package dep%d\nconst Value = %d\n", i, i+1))
	}
	source.WriteString(")\nfunc Total() int { return ")
	for i := range 34 {
		if i != 0 {
			source.WriteString(" + ")
		}
		fmt.Fprintf(&source, "d%d.Value", i)
	}
	source.WriteString(" }\n")
	writeImporterFixture(t, root, "consumer.go", source.String())
	return root, []byte(source.String())
}

func TestProjectedConformanceResolvesIndependentImportPopulation(t *testing.T) {
	root, source := independentImportFixture(t)
	if err := projectedFinalConformance(root, "consumer.go", map[string][]byte{"consumer.go": source}); err != nil {
		t.Fatal("valid 34-import projection did not close its dependency population", err)
	}
}
