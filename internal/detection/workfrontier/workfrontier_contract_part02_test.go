package workfrontier

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
)

func TestWorkfrontierPermutationReplayIsCanonical(t *testing.T) {
	for _, fixture := range loadFixtures(t) {
		if !fixture.PermutationTest {
			continue
		}
		t.Run(fixture.Name, func(t *testing.T) {
			baseInput, err := Decode(adaptLegacyInput(t, fixture.Input))
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
				input, err := Decode(adaptLegacyInput(t, raw))
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
