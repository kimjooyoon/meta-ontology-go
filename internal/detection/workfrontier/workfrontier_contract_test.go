package workfrontier

import (
	"embed"
	"encoding/json"
	"fmt"
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
			if fixture.FairBaseline && len(oracle.SelectedIDs) == 0 {
				t.Fatal("fair baseline selected no compatible path")
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

func adaptLegacyInput(t *testing.T, raw []byte) []byte {
	t.Helper()
	var legacy oracleInput
	if err := json.Unmarshal(raw, &legacy); err != nil {
		t.Fatal(err)
	}
	input := map[string]any{
		"schema_version":             SchemaVersion,
		"snapshot_digest":            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"policy_digest":              "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"registry_digest":            "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		"minimum_selected_pressures": legacy.M,
		"capacity":                   map[string]any{"cpu_core_ns": legacy.CPUCapacity},
		"pressures":                  nil,
		"states":                     nil,
		"paths":                      nil,
	}
	if legacy.Pressures == nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	registered := make(map[string]struct{}, len(legacy.RegisteredPaths))
	pressures := make([]Pressure, 0, len(legacy.RegisteredPaths)+maxInt(legacy.M, 0))
	for _, path := range legacy.RegisteredPaths {
		registered[path] = struct{}{}
		pressures = append(pressures, Pressure{StableID: path})
	}
	requiredPressureIDs := make([]string, 0, maxInt(legacy.M, 0))
	for index := 0; index < legacy.M; index++ {
		id := fmt.Sprintf("pressure/registry/%d", index)
		requiredPressureIDs = append(requiredPressureIDs, id)
		pressures = append(pressures, Pressure{StableID: id})
	}
	states := make([]ObligationState, 0, len(*legacy.Pressures)*2)
	paths := make([]RepairPath, 0, len(*legacy.Pressures))
	for _, candidate := range *legacy.Pressures {
		obligationID := "obligation/" + candidate.ID
		prerequisiteID := "prerequisite/" + candidate.ID
		states = append(states,
			ObligationState{ObligationID: obligationID, Status: "PENDING"},
			ObligationState{ObligationID: prerequisiteID, Status: candidate.Prerequisite},
		)
		readSet, writeSet, claimsValid := legacyClaims(candidate.Claims, registered)
		required := append([]string(nil), requiredPressureIDs...)
		if !claimsValid || candidate.ID == "" || candidate.WorkID == "" || candidate.CPU <= 0 {
			required = append(required, "pressure/unresolved/"+candidate.ID)
		}
		paths = append(paths, RepairPath{
			StableID: candidate.ID, WorkID: candidate.WorkID, ObligationID: obligationID,
			PrerequisiteObligationIDs: []string{prerequisiteID}, ReadSet: readSet,
			WriteSet: writeSet, RequiredPressureIDs: required,
			PolicyPriority: uint32(maxInt(candidate.Priority, 0)), CPUCoreNSUpperBound: uint64(maxInt(candidate.CPU, 0)),
		})
	}
	input["pressures"] = pressures
	input["states"] = states
	input["paths"] = paths
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func legacyClaims(claims []oracleClaim, registered map[string]struct{}) ([]string, []string, bool) {
	readSet := make([]string, 0, len(claims))
	writeSet := make([]string, 0, len(claims))
	seen := make(map[string]struct{}, len(claims))
	valid := len(claims) != 0
	for _, claim := range claims {
		if claim.Path == "" || claim.Mode != "R" && claim.Mode != "W" {
			valid = false
			continue
		}
		if _, duplicate := seen[claim.Path]; duplicate {
			valid = false
		}
		seen[claim.Path] = struct{}{}
		if _, ok := registered[claim.Path]; !ok {
			valid = false
		}
		if claim.Mode == "R" {
			readSet = append(readSet, claim.Path)
		} else {
			writeSet = append(writeSet, claim.Path)
		}
	}
	return readSet, writeSet, valid
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
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
