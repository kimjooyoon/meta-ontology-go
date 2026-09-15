package workfrontier

import (
	"embed"
	"testing"
)

//go:embed testdata/cases.json
var contractFixtures embed.FS

func TestWorkfrontierFixtures(t *testing.T) {
	fixtures := loadFixtures(t)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			input, err := Decode(adaptLegacyInput(t, fixture.Input))
			if fixture.DecodeError {
				if err == nil {
					t.Fatal("Decode accepted a pressure count below the minimum")
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			got := observeResult(t, Select(input))
			oracle := independentOracle(t, fixture.Input)
			if got.Status != oracle.Status || !sameStrings(got.SelectedIDs, oracle.SelectedIDs) || !sameStrings(got.WorkIDs, oracle.WorkIDs) {
				t.Fatalf("result = %#v, oracle = %#v", got, oracle)
			}
			if got.Status != fixture.Expected.Status || !sameStrings(got.SelectedIDs, fixture.Expected.SelectedIDs) || !sameStrings(got.WorkIDs, fixture.Expected.WorkIDs) {
				t.Fatalf("result = %#v, fixture = %#v", got, fixture.Expected)
			}
			if fixture.Expected.Quality != "" && got.Quality != fixture.Expected.Quality {
				t.Fatalf("quality = %q, want %q", got.Quality, fixture.Expected.Quality)
			}
			assertRequiredConflicts(t, fixture)
			if fixture.GreedyNonmaximum && oracle.MaximumSize <= len(oracle.SelectedIDs) {
				t.Fatalf("oracle did not prove a larger compatible set: maximum=%d selected=%d", oracle.MaximumSize, len(oracle.SelectedIDs))
			}
		})
	}
}
func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
