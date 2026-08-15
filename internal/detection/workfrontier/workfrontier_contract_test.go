package workfrontier

import (
	"embed"
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
)

//go:embed testdata/cases.json
var contractFixtures embed.FS

func TestWorkfrontierFixtures(t *testing.T) {
	fixtures := loadFixtures(t)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			input, err := Decode(fixture.Input)
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
			if got.Status != oracle.Status || !reflect.DeepEqual(got.SelectedIDs, oracle.SelectedIDs) || !reflect.DeepEqual(got.WorkIDs, oracle.WorkIDs) {
				t.Fatalf("result = %#v, oracle = %#v", got, oracle)
			}
			if got.Status != fixture.Expected.Status || !reflect.DeepEqual(got.SelectedIDs, fixture.Expected.SelectedIDs) || !reflect.DeepEqual(got.WorkIDs, fixture.Expected.WorkIDs) {
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

func TestWorkfrontierPermutationReplayIsCanonical(t *testing.T) {
	for _, fixture := range loadFixtures(t) {
		if !fixture.PermutationTest {
			continue
		}
		t.Run(fixture.Name, func(t *testing.T) {
			baseInput, err := Decode(fixture.Input)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			baseResult := Select(baseInput)
			baseObserved := observeResult(t, baseResult)
			baseCanonical := canonicalResultBytes(t, baseResult)
			baseDigest, hasDigest := optionalResultDigest(t, baseResult)

			var object map[string]json.RawMessage
			if err := json.Unmarshal(fixture.Input, &object); err != nil {
				t.Fatal(err)
			}
			var pressures []json.RawMessage
			if err := json.Unmarshal(object["pressures"], &pressures); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 100; i++ {
				permutation := append([]json.RawMessage(nil), pressures...)
				rand.New(rand.NewSource(int64(i+1))).Shuffle(len(permutation), func(a, b int) {
					permutation[a], permutation[b] = permutation[b], permutation[a]
				})
				object["pressures"], _ = json.Marshal(permutation)
				raw, _ := json.Marshal(object)
				input, err := Decode(raw)
				if err != nil {
					t.Fatalf("permutation %d Decode() error = %v", i, err)
				}
				result := Select(input)
				if got := observeResult(t, result); !reflect.DeepEqual(got, baseObserved) {
					t.Fatalf("permutation %d result = %#v, base = %#v", i, got, baseObserved)
				}
				if got := canonicalResultBytes(t, result); !reflect.DeepEqual(got, baseCanonical) {
					t.Fatalf("permutation %d changed canonical result bytes", i)
				}
				if hasDigest {
					got, _ := optionalResultDigest(t, result)
					if got != baseDigest {
						t.Fatalf("permutation %d changed result digest: %q != %q", i, got, baseDigest)
					}
				}
			}
		})
	}
}

func loadFixtures(t *testing.T) []contractFixture {
	t.Helper()
	data, err := contractFixtures.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []contractFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}
