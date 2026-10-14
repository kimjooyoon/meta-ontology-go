package verticalsliceclosureshadow

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCurrentDenominatorPinsExistingCapabilities(t *testing.T) {
	raw := activeDenominator()
	if digestBytes(raw) != DenominatorMigrationV67Digest || activeDenominatorDigest() != DenominatorMigrationV67Digest {
		t.Fatal("active denominator is not the pinned v67 migration")
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
	if observed.Version != 67 || len(observed.Boundaries) != 6 ||
		observed.Boundaries[0].ID != "syntax" || observed.Boundaries[0].Target != 100 ||
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
	lowered := bytes.Replace(raw, []byte("\"target\": 100"), []byte("\"target\": 99"), 1)
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
