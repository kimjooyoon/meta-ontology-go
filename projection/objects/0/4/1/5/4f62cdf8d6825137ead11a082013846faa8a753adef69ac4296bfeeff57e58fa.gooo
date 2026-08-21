package freshness

import (
	"testing"
)

func assertState(t *testing.T, report Report, kind Kind, id string, expected State) {
	t.Helper()
	for _, item := range report.Items {
		if item.Kind == kind && item.ID == id {
			if item.State != expected {
				t.Fatalf("%s/%s state=%s, want %s: %s", kind, id, item.State, expected, item.Detail)
			}
			return
		}
	}
	t.Fatalf("missing result for %s/%s: %#v", kind, id, report.Items)
}
func reverseSources(values []Source) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseArtifacts(values []Artifact) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseEvidence(values []Evidence) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func TestDigestInputEncodingIsStable(t *testing.T) {
	input := HashBytes([]byte("a"))
	digest, err := DigestInputs([]string{"a"}, map[string]string{"a": input})
	if err != nil {
		t.Fatal(err)
	}
	want := HashBytes([]byte("a\x00" + input + "\n"))
	if digest != want {
		t.Fatalf("unexpected input encoding digest: got %s want %s", digest, want)
	}
}
