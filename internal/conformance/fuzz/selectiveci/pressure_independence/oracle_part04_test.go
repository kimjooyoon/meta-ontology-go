package pressureindependence

import (
	"testing"
)

func mustCorpusInput(t testing.TB, name string) Input {
	t.Helper()
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range corpus.Cases {
		if row.Name == name {
			return row.Input
		}
	}
	t.Fatalf("corpus case %q not found", name)
	return Input{}
}
func reverseRecords(values []PressureRecord) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
