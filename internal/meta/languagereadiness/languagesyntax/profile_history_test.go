package languagesyntax

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

type retainedProfileTransition struct {
	Before CaseDefinition `json:"before_case"`
	After  CaseDefinition `json:"after_case"`
}

func recordedBridgeProfileTransition() (retainedProfileTransition, bool) {
	var transition retainedProfileTransition
	raw, err := os.ReadFile("../../../../docs/observations/record-iteration-origins/profile-v7.json")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) !=
		"5e687b365533355c083c8b72057d4768a332409a0086c7c09fe81c70c43f3b49" {
		return transition, false
	}
	err = json.Unmarshal(raw, &transition)
	return transition, err == nil
}

// An immutable, source-bound migration can change one parser profile. Every
// prior case still occupies its original position with its original identity.
func historicalSyntaxCasesRetained(prior, current []CaseDefinition) bool {
	if len(current) < len(prior) {
		return false
	}
	transition, valid := recordedBridgeProfileTransition()
	for i, old := range prior {
		if reflect.DeepEqual(old, current[i]) {
			continue
		}
		if !valid || !reflect.DeepEqual(old, transition.Before) ||
			!reflect.DeepEqual(current[i], transition.After) {
			return false
		}
	}
	return true
}

func TestHistoricalProfilesRequireExactRecordedTransition(t *testing.T) {
	migration, ok := recordedBridgeProfileTransition()
	if !ok || migration.Before.ID != "local-field-bridge-reader" ||
		migration.After.EntityFieldsProfile != "gooo.entityfields.go-projection.v7" {
		t.Fatal("source-bound profile migration missing")
	}
	prior := []CaseDefinition{migration.Before, {ID: "unchanged", Path: "unchanged.gooo"}}
	for _, name := range []string{"recorded", "id", "path", "scope", "proof", "profile", "reordered", "removed"} {
		t.Run(name, func(t *testing.T) {
			current := []CaseDefinition{migration.After, prior[1]}
			switch name {
			case "id":
				current[0].ID = "changed"
			case "path":
				current[0].Path = "different.gooo"
			case "scope":
				current[0].Scope = ScopeGovernanceObservation
			case "proof":
				current[0].ProofChoice = "FOUNDATION"
			case "profile":
				current[0].EntityFieldsProfile = "gooo.entityfields.go-projection.v6"
			case "reordered":
				current[0], current[1] = current[1], current[0]
			case "removed":
				current = current[:1]
			}
			if historicalSyntaxCasesRetained(prior, current) != (name == "recorded") {
				t.Fatal("historical case migration lost exact identity or order")
			}
		})
	}
}
