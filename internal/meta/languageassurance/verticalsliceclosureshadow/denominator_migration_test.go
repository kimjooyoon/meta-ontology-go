package verticalsliceclosureshadow

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCurrentDenominatorPinsExistingCapabilities(t *testing.T) {
	raw := activeDenominator()
	if digestBytes(raw) != DenominatorMigrationV102Digest || activeDenominatorDigest() != DenominatorMigrationV102Digest {
		t.Fatal("active denominator is not the pinned v102 migration")
	}
	var observed struct {
		Version    int `json:"version"`
		Boundaries []struct {
			ID         string `json:"id"`
			Target     int    `json:"target"`
			LinkTarget int    `json:"link_target"`
		} `json:"boundaries"`
	}
	if err := json.Unmarshal(raw, &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Version != 102 || len(observed.Boundaries) != 6 ||
		observed.Boundaries[0].ID != "syntax" || observed.Boundaries[0].Target != 138 ||
		observed.Boundaries[1].ID != "semantics" || observed.Boundaries[1].Target != 46 ||
		observed.Boundaries[5].ID != "release" || observed.Boundaries[5].Target != 26 || observed.Boundaries[5].LinkTarget != 4 {
		t.Fatalf("migration changed the declared capability boundary: %#v", observed)
	}
	links := 0
	for _, boundary := range observed.Boundaries {
		links += boundary.LinkTarget
	}
	if links != 13 {
		t.Fatalf("migration changed the declared dependency count: %d", links)
	}
	if _, err := decodeDenominator(raw); err != nil {
		t.Fatalf("current pinned denominator was rejected: %v", err)
	}
}

func TestCurrentDenominatorRejectsLoweredTarget(t *testing.T) {
	raw := activeDenominator()
	lowered := bytes.Replace(raw, []byte("\"target\": 138"), []byte("\"target\": 137"), 1)
	if bytes.Equal(raw, lowered) {
		t.Fatal("counterexample did not change the syntax target")
	}
	if _, err := decodeDenominator(lowered); err == nil {
		t.Fatal("a lowered unapproved denominator was accepted")
	}
	if digestBytes(EmbeddedDenominator()) != DenominatorDigest {
		t.Fatal("the original denominator evidence was rewritten")
	}
}

func TestRecordMigrationPreservesPreviousBoundaryEvidence(t *testing.T) {
	if digestBytes(embeddedDenominatorV30) != DenominatorMigrationV30Digest ||
		digestBytes(embeddedDenominatorV31) != DenominatorMigrationV31Digest ||
		digestBytes(embeddedDenominatorV32) != DenominatorMigrationV32Digest ||
		digestBytes(embeddedDenominatorV33) != DenominatorMigrationV33Digest ||
		digestBytes(embeddedDenominatorV34) != DenominatorMigrationV34Digest ||
		digestBytes(embeddedDenominatorV35) != DenominatorMigrationV35Digest ||
		digestBytes(embeddedDenominatorV36) != DenominatorMigrationV36Digest ||
		digestBytes(embeddedDenominatorV37) != DenominatorMigrationV37Digest ||
		digestBytes(embeddedDenominatorV38) != DenominatorMigrationV38Digest ||
		digestBytes(embeddedDenominatorV39) != DenominatorMigrationV39Digest ||
		digestBytes(embeddedDenominatorV40) != DenominatorMigrationV40Digest ||
		digestBytes(embeddedDenominatorV41) != DenominatorMigrationV41Digest ||
		digestBytes(embeddedDenominatorV42) != DenominatorMigrationV42Digest ||
		digestBytes(embeddedDenominatorV43) != DenominatorMigrationV43Digest ||
		digestBytes(embeddedDenominatorV44) != DenominatorMigrationV44Digest ||
		digestBytes(embeddedDenominatorV45) != DenominatorMigrationV45Digest ||
		digestBytes(embeddedDenominatorV46) != DenominatorMigrationV46Digest ||
		digestBytes(embeddedDenominatorV47) != DenominatorMigrationV47Digest {
		t.Fatal("historical denominator evidence was rewritten")
	}
	var previous, prior, current denominator
	if err := json.Unmarshal(embeddedDenominatorV43, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV47, &current); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV44, &prior); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 43 || prior.Version != 44 || current.Version != 47 ||
		len(previous.Boundaries) != 6 || len(prior.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration changed the boundary inventory")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target += 2
		}
		if prior.Boundaries[index] != expected {
			t.Fatalf("v44 migration changed unrelated boundary %d: %#v", index, prior.Boundaries[index])
		}
	}
	for index, expected := range prior.Boundaries {
		if index == 0 || index == 1 {
			expected.Target += 2
		}
		if index == 5 {
			continue // v47 adds the native platform to the release boundary.
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("v45 migration changed unrelated boundary %d: %#v", index, current.Boundaries[index])
		}
	}
	if _, err := decodeDenominator(embeddedDenominatorV46); err != nil {
		t.Fatalf("historical v46 denominator was not preserved: %v", err)
	}
	if _, err := decodeDenominator(embeddedDenominatorV31); err == nil {
		t.Fatal("the historical denominator authorized the new capability")
	}
}

func TestGeneratedRegistrationMigrationV50PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV49) != DenominatorMigrationV49Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV49, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV50, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 49 || current.Version != 50 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV51PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV50) != DenominatorMigrationV50Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV50, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV51, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 50 || current.Version != 51 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV52PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV51) != DenominatorMigrationV51Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV51, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV52, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 51 || current.Version != 52 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV53PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV52) != DenominatorMigrationV52Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV52, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV53, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 52 || current.Version != 53 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV54PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV53) != DenominatorMigrationV53Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV53, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV54, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 53 || current.Version != 54 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV55PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV54) != DenominatorMigrationV54Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV54, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV55, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 54 || current.Version != 55 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV56PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV55) != DenominatorMigrationV55Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV55, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV56, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 55 || current.Version != 56 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV57PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV56) != DenominatorMigrationV56Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV56, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV57, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 56 || current.Version != 57 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV58PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV57) != DenominatorMigrationV57Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV57, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV58, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 57 || current.Version != 58 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV59PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV58) != DenominatorMigrationV58Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV58, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV59, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 58 || current.Version != 59 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV60PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV59) != DenominatorMigrationV59Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV59, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV60, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 59 || current.Version != 60 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV61PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV60) != DenominatorMigrationV60Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV60, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV61, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 60 || current.Version != 61 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV62PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV61) != DenominatorMigrationV61Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV61, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV62, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 61 || current.Version != 62 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV63PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV62) != DenominatorMigrationV62Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV62, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV63, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 62 || current.Version != 63 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV64PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV63) != DenominatorMigrationV63Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV63, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV64, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 63 || current.Version != 64 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV65PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV64) != DenominatorMigrationV64Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV64, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV65, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 64 || current.Version != 65 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV66PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV65) != DenominatorMigrationV65Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV65, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV66, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 65 || current.Version != 66 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV67PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV66) != DenominatorMigrationV66Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV66, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV67, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 66 || current.Version != 67 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV68PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV67) != DenominatorMigrationV67Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV67, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV68, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 67 || current.Version != 68 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV69PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV68) != DenominatorMigrationV68Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV68, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV69, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 68 || current.Version != 69 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV70PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV69) != DenominatorMigrationV69Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV69, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV70, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 69 || current.Version != 70 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV71PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV70) != DenominatorMigrationV70Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV70, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV71, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 70 || current.Version != 71 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV73PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV72) != DenominatorMigrationV72Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV72, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV73, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 72 || current.Version != 73 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV74PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV73) != DenominatorMigrationV73Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV73, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV74, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 73 || current.Version != 74 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV75PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV74) != DenominatorMigrationV74Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV74, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV75, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 74 || current.Version != 75 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV76PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV75) != DenominatorMigrationV75Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV75, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV76, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 75 || current.Version != 76 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV77PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV76) != DenominatorMigrationV76Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV76, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV77, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 76 || current.Version != 77 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV78PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV77) != DenominatorMigrationV77Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV77, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV78, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 77 || current.Version != 78 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV79PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV78) != DenominatorMigrationV78Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV78, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV79, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 78 || current.Version != 79 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV80PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV79) != DenominatorMigrationV79Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV79, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV80, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 79 || current.Version != 80 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV81PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV80) != DenominatorMigrationV80Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV80, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV81, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 80 || current.Version != 81 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV82PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV81) != DenominatorMigrationV81Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV81, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV82, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 81 || current.Version != 82 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV83PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV82) != DenominatorMigrationV82Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV82, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV83, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 82 || current.Version != 83 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV84PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV83) != DenominatorMigrationV83Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV83, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV84, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 83 || current.Version != 84 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV85PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV84) != DenominatorMigrationV84Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV84, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV85, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 84 || current.Version != 85 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV86PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV85) != DenominatorMigrationV85Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV85, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV86, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 85 || current.Version != 86 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV87PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV86) != DenominatorMigrationV86Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV86, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV87, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 86 || current.Version != 87 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV88PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV87) != DenominatorMigrationV87Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV87, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV88, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 87 || current.Version != 88 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV89PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV88) != DenominatorMigrationV88Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV88, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV89, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 88 || current.Version != 89 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV90PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV89) != DenominatorMigrationV89Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV89, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV90, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 89 || current.Version != 90 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV91PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV90) != DenominatorMigrationV90Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV90, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV91, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 90 || current.Version != 91 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV92PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV91) != DenominatorMigrationV91Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV91, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV92, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 91 || current.Version != 92 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV93PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV92) != DenominatorMigrationV92Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV92, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV93, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 92 || current.Version != 93 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV94PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV93) != DenominatorMigrationV93Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV93, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV94, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 93 || current.Version != 94 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV95PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV94) != DenominatorMigrationV94Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV94, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV95, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 94 || current.Version != 95 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV96PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV95) != DenominatorMigrationV95Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV95, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV96, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 95 || current.Version != 96 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV97PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV96) != DenominatorMigrationV96Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV96, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV97, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 96 || current.Version != 97 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV98PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV97) != DenominatorMigrationV97Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV97, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV98, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 97 || current.Version != 98 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV99PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV98) != DenominatorMigrationV98Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV98, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV99, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 98 || current.Version != 99 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV100PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV99) != DenominatorMigrationV99Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV99, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV100, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 99 || current.Version != 100 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV101PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV100) != DenominatorMigrationV100Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV100, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV101, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 100 || current.Version != 101 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}

func TestGeneratedRegistrationMigrationV102PreservesBaseline(t *testing.T) {
	if digestBytes(embeddedDenominatorV101) != DenominatorMigrationV101Digest {
		t.Fatal("baseline evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV101, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(embeddedDenominatorV102, &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 101 || current.Version != 102 || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration inventory changed")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
}
