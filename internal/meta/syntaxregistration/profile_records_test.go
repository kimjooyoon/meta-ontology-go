package syntaxregistration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestProfileObservationBindingRejectsChangedRecordContracts(t *testing.T) {
	if _, err := bindContract(); err != nil {
		t.Fatal(err)
	}
	row := `gooo://syntax-registration/profile-compatibility`
	passed := `field passed id "` + row + `/passed" type boolean required one`
	profiles := `field profiles id "gooo://syntax-registration/profile-observation/profiles" type ProfileCompatibility required many`
	for _, change := range []struct{ name, before, after string }{
		{"row identity", `entity ProfileCompatibility id "` + row + `"`, `entity ProfileCompatibility id "urn:other:row"`},
		{"field identity", passed, strings.Replace(passed, `/passed"`, `/different"`, 1)},
		{"field type", passed, strings.Replace(passed, "type boolean", "type string", 1)},
		{"field presence", passed, strings.Replace(passed, "required", "optional", 1)},
		{"field cardinality", passed, strings.Replace(passed, "one", "many", 1)},
		{"list type", profiles, strings.Replace(profiles, "type ProfileCompatibility", "type string", 1)},
		{"list cardinality", profiles, strings.Replace(profiles, "many", "one", 1)},
		{"extra row field", passed, passed + "\n    field extra id \"urn:extra:field\" type string required one"},
		{"unrelated node", "namespace syntaxregistration", "namespace syntaxregistration\nentity Extra id \"urn:extra:node\""},
	} {
		t.Run(change.name, func(t *testing.T) {
			if strings.Count(string(contractSource), change.before) != 1 {
				t.Fatal("counterexample did not identify exactly one source declaration")
			}
			source := strings.Replace(string(contractSource), change.before, change.after, 1)
			if _, err := bindContractSource([]byte(source)); err == nil {
				t.Fatal("changed record ABI was accepted")
			}
		})
	}
}

func TestProfileHintFindsActualNominalRecordSources(t *testing.T) {
	for _, path := range []string{"examples/record-collections/main.gooo", "examples/record-collections/assembly.gooo"} {
		r := InspectProfileHintForCapability(context.Background(), osProjectFS(), path, "entity-many")
		if r.Status != "HINT" || r.SuggestedProfile != syntax.EntityFieldsV7ProfileID ||
			r.MatchCount != 1 || len(r.CompatibleProfiles) != 1 || r.Exercise.Status != "EXERCISED" ||
			len(r.Profiles) != 11 || r.Authority != "ADVISORY_ONLY" {
			t.Fatalf("nominal source is missing from the kernel-backed catalog: %+v", r)
		}
	}
}

func TestProfilePolicyIteratesStructuredFactsWithoutProfileNameBranches(t *testing.T) {
	observation := profileObservation{Path: "synthetic.gooo", SourceDigest: "fixture-origin",
		Profiles: []profileCompatibility{
			{Profile: "unexecuted", Passed: false, Eligible: true},
			{Profile: "ineligible", Passed: true, Eligible: false},
			{Profile: "future-catalog-entry", Passed: true, Eligible: true},
		}}
	raw, err := json.Marshal([]profileObservation{observation})
	if err != nil {
		t.Fatal(err)
	}
	value, err := bodycodegen.EvaluateFixedValue(context.Background(), "contract.gooo", contractSource,
		"SuggestRegistrationProfile", raw)
	if err != nil {
		t.Fatal(err)
	}
	var hint profileHintValue
	if err := json.Unmarshal(value.Value, &hint); err != nil || hint.Profile != "future-catalog-entry" ||
		hint.MatchCount != 1 || hint.Status != "HINT" || hint.ExerciseStatus != "NOT_REQUESTED" {
		t.Fatal("source policy hard-coded a known profile name or lost ordered facts", hint, err)
	}
	// This synthetic source-policy check grants no kernel/profile admission.
	if _, known := syntax.EntityFieldsSupportForProfileID(hint.Profile); known {
		t.Fatal("a source hint registered an unknown kernel profile")
	}
}
