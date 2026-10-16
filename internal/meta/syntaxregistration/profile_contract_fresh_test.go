package syntaxregistration

import (
	"strings"
	"testing"
)

func TestProfileMigrationContractFresh1844(t *testing.T) {
	base, err := bindProfileContract(profileMigrationSource)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name   string
		source string
		accept bool
	}{
		{"source", string(profileMigrationSource), true},
		{"renamed-display", strings.ReplaceAll(strings.ReplaceAll(string(profileMigrationSource),
			"MigrateSyntaxProfile", "MoveProfile"), "ProfileCorpus", "CaseView"), true},
		{"wrong-program", strings.Replace(string(profileMigrationSource), "syntax.profile-migrate:v1", "other:v1", 1), false},
		{"wrong-input", strings.Replace(string(profileMigrationSource),
			"ReviseProfileCorpus(PinnedProfileMigrationInput)", "ReviseProfileCorpus(ProfileMigrationRequest)", 1), false},
		{"extra-node", string(profileMigrationSource) + "\nentity Extra id \"urn:profile1844:extra\"\n", false},
	} {
		t.Run(item.name, func(t *testing.T) {
			binding, err := bindProfileContract([]byte(item.source))
			if (err == nil) != item.accept {
				t.Fatalf("source contract acceptance: %+v %v", binding, err)
			}
			if item.accept && (binding.pin != base.pin || binding.operation != base.operation ||
				binding.activities != base.activities || binding.outputs != base.outputs || binding.source == "" || binding.semantic == "") {
				t.Fatalf("display change rewrote semantic roles: %+v", binding)
			}
		})
	}
}

func TestOwnedCaseLeavesFresh1845(t *testing.T) {
	inputs := map[string][]byte{
		"root1845.go": []byte(`package languagesyntax
func expectedRegistry() Registry { return Registry{Cases: allOwned()} }
func allOwned() []CaseDefinition { rows := firstOwned(); rows = append(rows, secondOwned()...); return rows }
`),
		"roles1845.go": []byte(`package languagesyntax
func firstOwned() []CaseDefinition { return []CaseDefinition{{ID: "first1845"}} }
func secondOwned() []CaseDefinition { return []CaseDefinition{{ID: "second1845"}} }
func unrelated() []CaseDefinition { return []CaseDefinition{{ID: "unowned1845"}} }
`),
	}
	source, err := parseSourceUnits(inputs, "./", "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	root, err := source.function("expectedRegistry")
	if err != nil {
		t.Fatal(err)
	}
	leaves, err := source.registryCaseLeaves(root)
	if err != nil || len(leaves) != 2 || !strings.Contains(source.text(leaves[0]), "first1845") ||
		!strings.Contains(source.text(leaves[1]), "second1845") || len(source.edits) != 0 {
		t.Fatalf("source-owned sequence changed: %v %v", leaves, err)
	}
}
