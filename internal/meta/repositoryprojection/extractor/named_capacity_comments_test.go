package extractor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNamedCapacityRetainsDeclarationComments(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "sample")
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	var baseline namedCapacityObservation
	for _, comments := range []int{0, 80} {
		source := "package sample\n" + strings.Repeat("// Retained declaration documentation.\n", comments) +
			"func Read() int { return 7 }\n"
		if err := os.WriteFile(filepath.Join(directory, "source.go"), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		observed, err := collectNamedCapacities(root, []string{"sample#Read"})
		if err != nil {
			t.Fatal(err)
		}
		got := observed["sample#Read"]
		if got.capacity.helperLines == nil {
			t.Fatal("declaration capacity was not measured")
		}
		if comments == 0 {
			baseline = got
			continue
		}
		if *got.capacity.helperLines != *baseline.capacity.helperLines+comments ||
			got.capacity.functionLines != baseline.capacity.functionLines ||
			got.capacity.helperStatus != renderedCapacityOverCap ||
			got.capacity.declarationStart == got.capacity.functionStart {
			t.Fatalf("declaration comments disappeared from the measured subject: %+v", got.capacity)
		}
	}
}
